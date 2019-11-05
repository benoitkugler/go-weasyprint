package goweasyprint

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/ascii85"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"

	mt "github.com/benoitkugler/go-weasyprint/matrix"
	"github.com/benoitkugler/go-weasyprint/style/tree"

	"github.com/benoitkugler/go-weasyprint/utils"
)

// Post-process the PDF files created by cairo and add extra metadata (including
// attachments, embedded files, trim & bleed boxes).

// Rather than trying to parse any valid PDF, we make some assumptions
// that hold for cairo in order to simplify the code:

// * All newlines are "\n", not "\r" || "\r\n"
// * Except for number 0 (which is always free) there is no "free" object.
// * Most white space separators are made of a single 0x20 space.
// * Indirect dictionary objects do not contain ">>" at the start of a line
//   except to mark the end of the object, followed by "endobj".
//   (In other words, ">>" markers for sub-dictionaries are indented.)
// * The Page Tree is flat: all kids of the root page node are page objects,
//   not page tree nodes.

// However the code uses a lot of assert statements so that if an assumptions
// is not true anymore, the code should (hopefully) fail with an exception
// rather than silently behave incorrectly.

type bts = []byte

// Escape parentheses and backslashes in ``s``.
func pdfEscape(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "(", "\\(", -1)
	s = strings.Replace(s, ")", "\\)", -1)
	s = strings.Replace(s, "\r", "\\r", -1)
	return s
}

// add line breaks to Bytes() to mimic Python's next()
type myScan struct {
	s *bufio.Scanner
}

func newScanner(in io.Reader) myScan {
	return myScan{s: bufio.NewScanner(in)}
}

func (m myScan) Next() (line bts, err error) {
	if !m.s.Scan() {
		return nil, errors.New("unexpected end of file")
	}
	return m.Bytes(), nil
}

func (m myScan) Bytes() []byte {
	l := m.s.Bytes()
	return append(l, byte('\n'))
}

func pdfFormat(str string, args ...interface{}) bts {
	return bts(fmt.Sprintf(str, args...))
}

// return UTF-16 BE with a BOM, then backslash-escape parentheses.
func pdfString(s string) string {
	var buf bytes.Buffer
	encoder := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewEncoder().Writer(&buf)
	_, err := io.Copy(encoder, strings.NewReader("\ufeff"+s))
	if err != nil {
		log.Println("unable to encode %s in utf16-be", "\ufeff"+s)
		return ""
	}
	dec := charmap.ISO8859_1.NewDecoder().Reader(&buf)
	out, err := ioutil.ReadAll(dec)
	if err != nil {
		log.Println("unable to decode %s in latin1", string(buf.Bytes()))
		return ""
	}
	return "(" + pdfEscape(string(out)) + ")"
}

type PDFDictionary struct {
	objectNumber int
	byteString   bts
}

var reCache = map[[2]string]*regexp.Regexp{}

func (d PDFDictionary) getValue(key, valueRe string) (out bts, err error) {
	regex, in := reCache[[2]string{key, valueRe}]
	if !in {
		regex, err = regexp.Compile(string(pdfFormat("/%s %s", key, valueRe)))
		if err != nil {
			return nil, err
		}
		reCache[[2]string{key, valueRe}] = regex
	}
	match := regex.FindSubmatch(d.byteString)
	if len(match) < 1 {
		return nil, fmt.Errorf("no matching group for %s", regex.String())
	}
	return match[1], nil
}

// Get dictionary type.
// Returns the value for the /Type key.
func (d PDFDictionary) getType() (string, error) {
	// No end delimiter, + defaults to greedy
	v, err := d.getValue("Type", "/(\\w+)")
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(ascii85.NewDecoder(bytes.NewReader(v)))
	return string(b), err
}

// Read the value for `key` and follow the reference.
// We assume that it is an indirect dictionary object.
// :return: a new PDFDictionary instance.
func (d PDFDictionary) getIndirectDict(key string, pdfFile PDFFile) (PDFDictionary, error) {
	v, err := d.getValue(key, "(\\d+) 0 R")
	if err != nil {
		return PDFDictionary{}, err
	}
	objectNumber, err := strconv.Atoi(string(v))
	if err != nil {
		return PDFDictionary{}, err
	}
	bs, err := pdfFile.readObject(objectNumber)
	if err != nil {
		return PDFDictionary{}, err
	}
	return PDFDictionary{objectNumber: objectNumber, byteString: bs}, nil
}

// Read the value for `key` and follow the references.
// We assume that it is an array of indirect dictionary objects.
// :return: a list of new PDFDictionary instance.
func (d PDFDictionary) getIndirectDictArray(key string, pdfFile PDFFile) ([]PDFDictionary, error) {
	v, err := d.getValue(key, "\\[(.+?)\\]")
	if err != nil {
		return nil, err
	}
	parts := bytes.Split(v, bts(" 0 R"))
	// The array looks like this: " <a> 0 R <b> 0 R <c> 0 R "
	// so `parts` ends up like this [" <a>", " <b>", " <c>", " "]
	// With the trailing white space in the list.
	trail := parts[len(parts)-1]
	parts = parts[:len(parts)-1]
	if s := string(bytes.TrimSpace(trail)); s != "" {
		log.Fatalf("expected empty trail, got %s", s)
	}
	out := make([]PDFDictionary, len(parts))
	for i, ns := range parts {
		n, err := strconv.Atoi(string(ns))
		if err != nil {
			return nil, err
		}
		obj, err := pdfFile.readObject(n)
		if err != nil {
			return nil, err
		}
		out[i] = PDFDictionary{objectNumber: n, byteString: obj}
	}
	return out, nil
}

type PDFFile struct {
	finished bool
	fileobj  io.ReadWriteSeeker

	// Maps object number -> bytes from the start of the file
	objectsOffsets            []int64
	newObjectsOffsets         []int64
	overwrittenObjectsOffsets map[int]int64

	startxref               int
	info, catalog, pageTree PDFDictionary
	pages                   []PDFDictionary
}

var trailerRe = regexp.MustCompile("(?s)\ntrailer\n(.+)\nstartxref\n(\\d+)\n%%EOF\n$")

func NewPDFFile(fileobj io.ReadWriteSeeker) (p PDFFile, err error) {
	// cairo’s trailer only has Size, Root and Info.
	// The trailer + startxref + EOF is typically under 100 bytes
	_, err = fileobj.Seek(-200, io.SeekEnd)
	if err != nil {
		return p, err
	}
	content, err := ioutil.ReadAll(fileobj)
	if err != nil {
		return p, err
	}
	tmp := trailerRe.FindSubmatch(content)
	trailer_, startxref_ := tmp[1], tmp[2]
	trailer := PDFDictionary{objectNumber: -1, byteString: trailer_}
	startxref, err := strconv.Atoi(string(startxref_))
	if err != nil {
		return p, err
	}

	_, err = fileobj.Seek(int64(startxref), io.SeekStart)
	if err != nil {
		return p, err
	}
	scanner := newScanner(fileobj)
	line, err := scanner.Next()
	if err != nil {
		return p, err
	}
	if s := string(line); s != "xref\n" {
		return p, fmt.Errorf("unexpected line %s", s)
	}
	line, err = scanner.Next()
	if err != nil {
		return p, err
	}
	tmp = bytes.Split(line, bts(" "))
	firstObject, totalObjects_ := tmp[0], tmp[1]
	if s := string(firstObject); s != "0" {
		return p, fmt.Errorf("expected 0, got %s", s)
	}
	totalObjects, err := strconv.Atoi(string(totalObjects_))
	if err != nil {
		return p, err
	}
	line, err = scanner.Next()
	if err != nil {
		return p, err
	}
	if s := string(line); s != "0000000000 65535 f \n" {
		return p, fmt.Errorf("unexpected line %s", s)
	}

	objectsOffsets := []int64{-1}
	for objectNumber := 1; objectNumber < totalObjects; objectNumber += 1 {
		line, err = scanner.Next()
		if err != nil {
			return p, err
		}

		if s := string(line[10:]); s != " 00000 n \n" {
			return p, fmt.Errorf("unexpected line %s", s)
		}
		offset, err := strconv.Atoi(string(line[:10]))
		if err != nil {
			return p, err
		}
		objectsOffsets = append(objectsOffsets, int64(offset))
	}

	p.fileobj = fileobj
	p.objectsOffsets = objectsOffsets

	info, err := trailer.getIndirectDict("Info", p)
	if err != nil {
		return p, err
	}
	catalog, err := trailer.getIndirectDict("Root", p)
	if err != nil {
		return p, err
	}
	pageTree, err := catalog.getIndirectDict("Pages", p)
	if err != nil {
		return p, err
	}
	pages, err := pageTree.getIndirectDictArray("Kids", p)
	if err != nil {
		return p, err
	}
	// Check that the tree is flat
	for _, pa := range pages {
		if t, err := pa.getType(); err != nil || t != "Page" {
			return p, fmt.Errorf("tree is not flat (page type : %s, err %s)", t, err)
		}
	}

	p.startxref = startxref
	p.info = info
	p.catalog = catalog
	p.pageTree = pageTree
	p.pages = pages

	p.overwrittenObjectsOffsets = map[int]int64{}
	return p, nil
}

// objectNumber is an integer N so that 1 <= N < len(self.objectsOffsets)
func (p PDFFile) readObject(objectNumber int) (bts, error) {
	p.fileobj.Seek(p.objectsOffsets[objectNumber], io.SeekStart)
	scanner := newScanner(p.fileobj)
	line, err := scanner.Next()
	if err != nil {
		return nil, err
	}
	if !bytes.HasSuffix(line, bts(" 0 obj\n")) {
		return nil, fmt.Errorf("bad suffix for line %s", string(line))
	}
	if i, err := strconv.Atoi(string(line[:len(line)-7])); err != nil || i != objectNumber { // len(b" 0 obj\n") == 7
		return nil, fmt.Errorf("bad objectNumber for line %s", string(line))
	}
	var objectLines []bts
	for scanner.s.Scan() {
		line := scanner.Bytes()
		if string(line) == ">>\n" {
			line, err = scanner.Next()
			if err != nil || string(line) != "endobj\n" {
				return nil, fmt.Errorf("unexpected line %s", string(line))
			}
			// No newline, we’ll add it when writing.
			objectLines = append(objectLines, bts(">>"))
			return bytes.Join(objectLines, bts("")), nil
		}
		objectLines = append(objectLines, line)
	}
	return nil, nil
}

// Write the new content for an existing object at the end of the file.
// objectNumber verifies 1 <= N < len(p.objectsOffsets)
// byteString is the new object content as a byte string.
func (p PDFFile) overwriteObject(objectNumber int, byteString bts) error {
	offset, err := p.writeObject(objectNumber, byteString)
	if err != nil {
		return err
	}
	p.overwrittenObjectsOffsets[objectNumber] = offset
	return nil
}

// Overwrite a dictionary object.
// Content is added inside the << >> delimiters.
func (p PDFFile) extendDict(dictionary PDFDictionary, newContent bts) error {
	if !bytes.HasSuffix(dictionary.byteString, bts(">>")) {
		log.Fatalf("expected >> at then end, got %s", string(dictionary.byteString))
	}
	return p.overwriteObject(dictionary.objectNumber, append(append(dictionary.byteString[:len(dictionary.byteString)-2], newContent...), bts("\n>>")...))
}

// Return object number that would be used by writeNewObject().
func (p PDFFile) nextObjectNumber() int {
	return len(p.objectsOffsets) + len(p.newObjectsOffsets)
}

// Write a new object at the end of the file.
// Returns the new object number.
func (p *PDFFile) writeNewObject(byteString bts) (int, error) {
	objectNumber := p.nextObjectNumber()
	offset, err := p.writeObject(objectNumber, byteString)
	if err != nil {
		return 0, err
	}
	p.newObjectsOffsets = append(p.newObjectsOffsets, offset)
	return objectNumber, nil
}

// Write cross-ref table and trailer for new and overwritten objects.
// This makes `fileobj` a valid (updated) PDF file.
func (p *PDFFile) finish() error {
	newStartxref, err := p.startWriting()
	if err != nil {
		return err
	}
	p.finished = true
	_, err = p.fileobj.Write(bts("xref\n"))
	if err != nil {
		return err
	}

	// Don’t bother sorting or finding contiguous numbers,
	// just write a new sub-section for each overwritten object.
	for objectNumber, offset := range p.overwrittenObjectsOffsets {
		_, err = p.fileobj.Write(pdfFormat("%d 1\n%010d 00000 n \n", objectNumber, offset))
	}

	if len(p.newObjectsOffsets) != 0 {
		firstNewObject := len(p.objectsOffsets)
		_, err = p.fileobj.Write(pdfFormat("%d %d\n", firstNewObject, len(p.newObjectsOffsets)))
		if err != nil {
			return err
		}
		for _, offset := range p.newObjectsOffsets {
			_, err = p.fileobj.Write(pdfFormat("%010d 00000 n \n", offset))
			if err != nil {
				return err
			}
		}
	}

	_, err = p.fileobj.Write(pdfFormat(
		"trailer\n<< "+
			"/Size %d /Root %d 0 R /Info %d 0 R /Prev %d"+
			" >>\nstartxref\n%d\n%%EOF\n",
		p.nextObjectNumber(), p.catalog.objectNumber, p.info.objectNumber, p.startxref, newStartxref))
	return err
}

func (p PDFFile) writeObject(objectNumber int, byteString bts) (int64, error) {
	offset, err := p.startWriting()
	if err != nil {
		return 0, err
	}
	_, err = p.fileobj.Write(bts(pdfFormat("%d 0 obj\n", objectNumber)))
	if err != nil {
		return 0, err
	}
	_, err = p.fileobj.Write(byteString)
	if err != nil {
		return 0, err
	}
	_, err = p.fileobj.Write(bts("\nendobj\n"))
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func (p PDFFile) startWriting() (int64, error) {
	if p.finished {
		return 0, errors.New("start writing on finished PDFFile !")
	}
	offset, err := p.fileobj.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func hexdigest(data []byte) string {
	tmp := md5.Sum(data)
	sl := make([]byte, len(tmp))
	for i, v := range tmp {
		sl[i] = v
	}
	return hex.EncodeToString(sl)
}

// Write a compressed file like object as ``/EmbeddedFile``.
// Compressing is done with deflate. In fact, this method writes multiple PDF
// objects to include length, compressed length and MD5 checksum.
// Returns the object number of the compressed file stream object
func writeCompressedFileObject(pdf *PDFFile, file io.Reader) (int, error) {
	objectNumber := pdf.nextObjectNumber()
	// Make sure we stay in sync with our object numbers
	expectedNextObjectNumber := objectNumber + 4

	lengthNumber := objectNumber + 1
	md5Number := objectNumber + 2
	uncompressedLengthNumber := objectNumber + 3

	offset, err := pdf.startWriting()
	if err != nil {
		return 0, err
	}
	_, err = pdf.fileobj.Write(pdfFormat("%d 0 obj\n", objectNumber))
	if err != nil {
		return 0, nil
	}
	_, err = pdf.fileobj.Write(pdfFormat("<< /Type /EmbeddedFile /Length %d 0 R /Filter "+
		"/FlateDecode /Params << /CheckSum %d 0 R /Size %d 0 R >> >>\n",
		lengthNumber, md5Number, uncompressedLengthNumber))
	if err != nil {
		return 0, nil
	}
	_, err = pdf.fileobj.Write(bts("stream\n"))
	if err != nil {
		return 0, nil
	}

	uncompressedLength := 0
	compressedLength := 0

	var (
		md5            []byte
		data           = make([]byte, 4096)
		zLibCompressed bytes.Buffer
		compress       = zlib.NewWriter(&zLibCompressed)
	)
	for {
		nRead, err1 := file.Read(data)
		uncompressedLength += nRead
		md5 = append(md5, data[:nRead]...)

		nCompressed, err2 := compress.Write(data[:nRead])
		compressedLength += nCompressed
		if err2 != nil {
			return 0, err
		}
		_, err = pdf.fileobj.Write(zLibCompressed.Bytes())
		if err != nil {
			return 0, nil
		}
		zLibCompressed.Reset()
		if err1 == io.EOF {
			break
		}
		if err1 != nil {
			return 0, err1
		}
	}
	if err = compress.Flush(); err != nil {
		return 0, err
	}
	compressed := zLibCompressed.Bytes()
	compressedLength += len(compressed)
	_, err = pdf.fileobj.Write(compressed)
	if err != nil {
		return 0, nil
	}

	_, err = pdf.fileobj.Write(bts("\nendstream\n"))
	if err != nil {
		return 0, nil
	}
	_, err = pdf.fileobj.Write(bts("endobj\n"))
	if err != nil {
		return 0, nil
	}

	pdf.newObjectsOffsets = append(pdf.newObjectsOffsets, offset)

	_, err = pdf.writeNewObject(pdfFormat("%d", compressedLength))
	if err != nil {
		return 0, err
	}
	_, err = pdf.writeNewObject(pdfFormat("<%s>", hexdigest(md5)))
	if err != nil {
		return 0, err
	}
	_, err = pdf.writeNewObject(pdfFormat("%d", uncompressedLength))
	if err != nil {
		return 0, err
	}

	if pdf.nextObjectNumber() != expectedNextObjectNumber {
		return 0, fmt.Errorf("expected %d as nextObjectNumber, got %d", expectedNextObjectNumber, pdf.nextObjectNumber())
	}

	return objectNumber, nil
}

// Derive a filename from a fetched resource.
// This is either the filename returned by the URL fetcher, the last URL path
// component or a synthetic name if the URL has no path.
func getFilenameFromResult(rawurl string) string {
	var filename string

	// The URL path likely contains a filename, which is a good second guess
	if rawurl != "" {
		u, err := url.Parse(rawurl)
		if err == nil {
			if u.Scheme != "data" {
				filename = path.Base(u.Path)
			}
		}
	}

	if filename == "" {
		// The URL lacks a path altogether. Use a synthetic name.

		// Using guessExtension is a great idea, but sadly the extension is
		// probably random, depending on the alignment of the stars, which car
		// you're driving and which software has been installed on your machine.

		// Unfortuneatly this isn't even imdepodent on one machine, because the
		// extension can depend on PYTHONHASHSEED if mimetypes has multiple
		// extensions to offer
		extension := ".bin"
		filename = "attachment" + extension
	} else {
		filename = utils.Unquote(filename)
	}

	return filename
}

// Write attachments as embedded files (document attachents).
// Returns the object number of the name dictionary or nil
func writePdfEmbeddedFiles(pdf *PDFFile, attachments []utils.Attachment, urlFetcher utils.UrlFetcher) (*int, error) {
	var fileSpecIds []int
	for _, attachment := range attachments {
		fileSpecId := writePdfAttachment(pdf, attachment, urlFetcher)
		if fileSpecId != nil {
			fileSpecIds = append(fileSpecIds, *fileSpecId)
		}
	}

	// We might have failed to write any attachment at all
	if len(fileSpecIds) == 0 {
		return nil, nil
	}

	content := []bts{bts("<< /Names [")}
	for _, fs := range fileSpecIds {
		content = append(content, pdfFormat("\n(attachment%d) %d 0 R ", fs, fs))
	}
	content = append(content, bts("\n] >>"))
	out, err := pdf.writeNewObject(bytes.Join(content, bts("")))
	return &out, err
}

// Write an attachment to the PDF stream.
// Returns the object number of the ``/Filespec`` object or nil if the
// attachment couldn't be read.
func writePdfAttachment(pdf *PDFFile, attachment utils.Attachment, urlFetcher utils.UrlFetcher) *int {
	// try {
	// Attachments from document links like <link> or <a> can only be URLs.

	tmp, err := tree.SelectSource(tree.InputUrl(attachment.Url), "", urlFetcher, false)
	if err != nil {
		log.Printf("Failed to load attachment: %s\n", err)
		return nil
	}
	source, url := tmp.Content, tmp.BaseUrl
	// attachment = Attachment(url=url, urlFetcher=urlFetcher, description=description)
	// } else if ! isinstance(attachment, Attachment) {
	//     attachment = Attachment(guess=attachment, urlFetcher=urlFetcher)
	// }

	fileStreamId, err := writeCompressedFileObject(pdf, bytes.NewReader(source))
	if err != nil {
		log.Printf("Failed to compress and include attachment : %s\n", err)
		return nil
	}

	// TODO: Use the result object from a URL fetch operation to provide more
	// details on the possible filename
	filename := getFilenameFromResult(url)

	num, err := pdf.writeNewObject(pdfFormat(
		"<< /Type /Filespec /F () /UF %s /EF << /F {1} 0 R >> "+
			"/Desc %s\n>>",
		pdfString(filename),
		fileStreamId,
		pdfString(attachment.Title)))
	if err != nil {
		log.Printf("Failed to write attachment filename and title: %s\n", err)
		return nil
	}
	return &num
}

// Add PDF metadata that are not handled by cairo.
//
// Includes:
// - attachments
// - embedded files
// - trim box
// - bleed box
func writePdfMetadata(fileobj io.ReadWriteSeeker, scale float64, urlFetcher utils.UrlFetcher, attachments []utils.Attachment,
	attachmentLinks [][]linkData, pages []Page) {

	pdf, err := NewPDFFile(fileobj)
	if err != nil {
		log.Printf("failed to read pdf : %s", err)
		return
	}
	// Add embedded files

	embeddedFilesId, err := writePdfEmbeddedFiles(&pdf, attachments, urlFetcher)
	if err != nil {
		log.Printf("failed to embed files : %s\n", err)
	}
	if embeddedFilesId != nil {
		params := pdfFormat(" /Names << /EmbeddedFiles {0} 0 R >>", embeddedFilesId)
		pdf.extendDict(pdf.catalog, params)
	}

	// Add attachments

	// A single link can be split in multiple regions. We don't want to embed
	// a file multiple times of course, so keep a reference to every embedded
	// URL and reuse the object number.
	// TODO: If we add support for descriptions this won't always be correct,
	// because two links might have the same href, but different titles.
	annotFiles := map[string]*int{}
	for _, pageLinks := range attachmentLinks {
		for _, v := range pageLinks {
			// linkType, target, rectangle = v
			if _, targetInAnnotFiles := annotFiles[v.target]; v.type_ == "attachment" && !targetInAnnotFiles {
				// TODO: use the title attribute as description
				annotFiles[v.target] = writePdfAttachment(&pdf, utils.Attachment{v.target, ""}, urlFetcher)
			}
		}
	}

	for i, pdfPage := range pdf.pages {
		documentPage, pageLinks := pages[i], attachmentLinks[i]

		// Add bleed box

		mediaBox_, err := pdfPage.getValue("MediaBox", "\\[(.+?)\\]")
		if err != nil {
			log.Printf("can't read MediaBox : %s", err)
			continue
		}
		mediaBoxParts := strings.Split(string(bytes.TrimSpace(mediaBox_)), " ")
		left, err := strconv.ParseFloat(mediaBoxParts[0], 64)
		if err != nil {
			log.Printf("invalid left MediaBox value : %s", mediaBoxParts[0])
		}
		top, err := strconv.ParseFloat(mediaBoxParts[1], 64)
		if err != nil {
			log.Printf("invalid top MediaBox value : %s", mediaBoxParts[1])
		}
		right, err := strconv.ParseFloat(mediaBoxParts[2], 64)
		if err != nil {
			log.Printf("invalid right MediaBox value : %s", mediaBoxParts[2])
		}
		bottom, err := strconv.ParseFloat(mediaBoxParts[3], 64)
		if err != nil {
			log.Printf("invalid bottom MediaBox value : %s", mediaBoxParts[3])
		}
		// Convert pixels into points
		bleed := bleedData{
			Left:   documentPage.bleed.Left * 0.75,
			Top:    documentPage.bleed.Top * 0.75,
			Right:  documentPage.bleed.Right * 0.75,
			Bottom: documentPage.bleed.Bottom * 0.75,
		}

		trimLeft := left + bleed.Left
		trimTop := top + bleed.Top
		trimRight := right - bleed.Right
		trimBottom := bottom - bleed.Bottom

		// Arbitrarly set PDF BleedBox between CSS bleed box (PDF MediaBox) and
		// CSS page box (PDF TrimBox), at most 10 points from the TrimBox.
		bleedLeft := trimLeft - math.Min(10, bleed.Left)
		bleedTop := trimTop - math.Min(10, bleed.Top)
		bleedRight := trimRight + math.Min(10, bleed.Right)
		bleedBottom := trimBottom + math.Min(10, bleed.Bottom)

		pdf.extendDict(pdfPage, pdfFormat(
			"/TrimBox [ %f %f %f %f ] /BleedBox [ %f %f %f %f ]",
			trimLeft, trimTop, trimRight, trimBottom,
			bleedLeft, bleedTop, bleedRight, bleedBottom))

		// Add links to attachments

		// TODO: splitting a link into multiple independent rectangular
		// annotations works well for pure links, but rather mediocre for other
		// annotations and fails completely for transformed (CSS) or complex
		// link shapes (area). It would be better to use /AP for all links and
		// coalesce link shapes that originate from the same HTML link. This
		// would give a feeling similiar to what browsers do with links that
		// span multiple lines.
		var annotations []int
		for _, v := range pageLinks {
			linkType, target, r := v.type_, v.target, v.rectangle
			if anTarget := annotFiles[target]; linkType == "attachment" && anTarget != nil {
				// xx=scale, yy=-scale, y0=documentPage.height * scale
				matrix := mt.New(scale, 0, 0, -scale, 0, documentPage.height*scale)
				rectX, rectY, width, height := r[0], r[1], r[2], r[3]
				rectX, rectY = matrix.TransformPoint(rectX, rectY)
				width, height = matrix.TransformDistance(width, height)
				// x, y, w, h => x0, y0, x1, y1
				r = [4]float64{rectX, rectY, rectX + width, rectY + height}
				content := []bts{pdfFormat("<< /Type /Annot /Rect [%.2f %.2f %.2f %.2f] /Border [0 0 0]\n",
					r[0], r[1], r[2], r[3])}
				linkAp, err := pdf.writeNewObject(pdfFormat("<< /Type /XObject /Subtype /Form "+
					"/BBox [%.2f %.2f %.2f %.2f] /Length 0 >>\n"+
					"stream\n"+
					"endstream", r[0], r[1], r[2], r[3]))
				if err != nil {
					log.Printf("failed to add link to attachment %v : %s", v.target, err)
					continue
				}
				content = append(content, bts("/Subtype /FileAttachment "))
				// evince needs /T or fails on an internal assertion. PDF
				// doesn't require it.
				content = append(content, pdfFormat("/T () /FS %d 0 R /AP << /N %d 0 R >>",
					*anTarget, linkAp))
				content = append(content, bts(">>"))
				annot, err := pdf.writeNewObject(bytes.Join(content, bts("")))
				if err != nil {
					log.Printf("failed to add link to attachment %v : %s", v.target, err)
					continue
				}
				annotations = append(annotations, annot)
			}
		}

		if len(annotations) != 0 {
			chunks := make([]bts, len(annotations))
			for i, n := range annotations {
				chunks[i] = pdfFormat("%d 0 R", n)
			}
			pdf.extendDict(pdfPage, pdfFormat("/Annots [%s]", bytes.Join(chunks, bts(" "))))
		}
	}

	if err = pdf.finish(); err != nil {
		log.Printf("error closing file : %s", err)
	}
}
