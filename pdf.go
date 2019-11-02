package goweasyprint

import (
	"regexp"
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"

	"github.com/benoitkugler/go-weasyprint/style/tree"

	"github.com/benoitkugler/go-weasyprint/utils"

	bo "github.com/benoitkugler/go-weasyprint/boxes"
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

// class PDFFormatter(string.Formatter) {
//     """Like str.format except:

//     * Results are byte strings
//     * The new !P conversion flags encodes a PDF string.
//       (UTF-16 BE with a BOM, then backslash-escape parentheses.)

//     Except for fields marked !P, everything should be ASCII-only.

//     """
//     def convertField(self, value, conversion) {
//         if conversion == "P" {
//             // Make a round-trip back through Unicode for the .translate()
//             // method. (bytes.translate only maps to single bytes.)
//             // Use latin1 to map all byte values.
//             return "({0})".format(pdfEscape(
//                 ("\ufeff" + value).encode("utf-16-be").decode("latin1")))
//         } else {
//             return super(PDFFormatter, self).convertField(value, conversion)
//         }
//     }
// }
//     def vformat(self, formatString, args, kwargs) {
//         result = super(PDFFormatter, self).vformat(formatString, args, kwargs)
//         return result.encode("latin1")
//     }

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
	objectNumber int64 
	byteString bts
}

 var reCache = map[string]*regexp.Regexp{}

    def getValue(self, key, valueRe) {
        regex = self.ReCache.get((key, valueRe))
        if ! regex {
            regex = re.compile(pdfFormat("/{0} {1}", key, valueRe))
            self.ReCache[key, valueRe] = regex
        } return regex.search(self.byteString).group(1)
    }

    def getType(self) {
        """Get dictionary type.

        :returns: the value for the /Type key.

        """
        // No end delimiter, + defaults to greedy
        return self.getValue("Type", "/(\\w+)").decode("ascii")
    }

    def getIndirectDict(self, key, pdfFile) {
        """Read the value for `key` && follow the reference.

        We assume that it is an indirect dictionary object.

        :return: a new PDFDictionary instance.

        """
        objectNumber = int(self.getValue(key, "(\\d+) 0 R"))
        return type(self)(objectNumber, pdfFile.readObject(objectNumber))
    }

    def getIndirectDictArray(self, key, pdfFile) {
        """Read the value for `key` && follow the references.

        We assume that it is an array of indirect dictionary objects.

        :return: a list of new PDFDictionary instance.

        """
        parts = self.getValue(key, "\\[(.+?)\\]").split(b" 0 R")
        // The array looks like this: " <a> 0 R <b> 0 R <c> 0 R "
        // so `parts` ends up like this [" <a>", " <b>", " <c>", " "]
        // With the trailing white space := range the list.
        trail = parts.pop()
        assert ! trail.strip()
        class_ = type(self)
        read = pdfFile.readObject
        return [class(n, read(n)) for n := range map(int, parts)]
    }

type PDFFile struct {
	finished          bool
	fileobj           io.WriteSeeker
	objectsOffsets    []*int64
	newObjectsOffsets []*int64
}

//     trailerRe = re.compile(
//         b"\ntrailer\n(.+)\nstartxref\n(\\d+)\n%%EOF\n$", re.DOTALL)
// }
//     def _Init_(self, fileobj) {
//         // cairo’s trailer only has Size, Root && Info.
//         // The trailer + startxref + EOF is typically under 100 bytes
//         fileobj.seek(-200, os.SEEKEND)
//         trailer, startxref = self.trailerRe.search(fileobj.read()).groups()
//         trailer = PDFDictionary(None, trailer)
//         startxref = int(startxref)
//

//         fileobj.seek(startxref)
//         line = next(fileobj)
//         assert line == b"xref\n"

//         line = next(fileobj)
//         firstObject, totalObjects = line.split()
//         assert firstObject == b"0"
//         totalObjects = int(totalObjects)

//         line = next(fileobj)
//         assert line == b"0000000000 65535 f \n"

//         objectsOffsets = [None]
//         for objectNumber := range range(1, totalObjects) {
//             line = next(fileobj)
//             assert line[10:] == b" 00000 n \n"
//             objectsOffsets.append(int(line[:10]))
//         }

//         self.fileobj = fileobj
//         #: Maps object number -> bytes from the start of the file
//         self.objectsOffsets = objectsOffsets

//         info = trailer.getIndirectDict("Info", self)
//         catalog = trailer.getIndirectDict("Root", self)
//         pageTree = catalog.getIndirectDict("Pages", self)
//         pages = pageTree.getIndirectDictArray("Kids", self)
//         // Check that the tree is flat
//         assert all(p.getType() == "Page" for p := range pages)

//         self.startxref = startxref
//         self.info = info
//         self.catalog = catalog
//         self.pageTree = pageTree
//         self.pages = pages

//         self.finished = false
//         self.overwrittenObjectsOffsets = {}
//         self.newObjectsOffsets = []

//     def readObject(self, objectNumber) {
//         """
//         :param objectNumber:
//             An integer N so that 1 <= N < len(self.objectsOffsets)
//         :returns:
//             The object content as a byte string.

//         """
//         fileobj = self.fileobj
//         fileobj.seek(self.objectsOffsets[objectNumber])
//         line = next(fileobj)
//         assert line.endswith(b" 0 obj\n")
//         assert int(line[:-7]) == objectNumber  // len(b" 0 obj\n") == 7
//         objectLines = []
//         for line := range fileobj {
//             if line == b">>\n" {
//                 assert next(fileobj) == b"endobj\n"
//                 // No newline, we’ll add it when writing.
//                 objectLines.append(b">>")
//                 return b"".join(objectLines)
//             } objectLines.append(line)
//         }
//     }

//     def overwriteObject(self, objectNumber, byteString) {
//         """Write the new content for an existing object at the end of the file.

//         :param objectNumber:
//             An integer N so that 1 <= N < len(self.objectsOffsets)
//         :param byteString:
//             The new object content as a byte string.

//         """
//         self.overwrittenObjectsOffsets[objectNumber] = (
//             self.writeObject(objectNumber, byteString))
//     }

// Overwrite a dictionary object.
// 
// Content is added inside the << >> delimiters.
    func (p PDFFile) extendDict(dictionary, newContent) {
        if !bytes.HasSuffix(dictionary.byteString,bts(">>")) {
			log.Fatalf("expected >> at then end, got %s", string(dictionary.byteString))
		}
        self.overwriteObject(dictionary.objectNumber,
            dictionary.byteString[:-2] + newContent + b"\n>>")
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
	p.newObjectsOffsets = append(p.newObjectsOffsets, &offset)
	return objectNumber, nil
}

//     def finish(self) {
//         """Write cross-ref table && trailer for new && overwritten objects.

//         This makes `fileobj` a valid (updated) PDF file.

//         """
//         newStartxref, write = self.startWriting()
//         self.finished = true
//         write(b"xref\n")
//     }

//         // Don’t bother sorting || finding contiguous numbers,
//         // just write a new sub-section for each overwritten object.
//         for objectNumber, offset := range self.overwrittenObjectsOffsets.items() {
//             write(pdfFormat(
//                 "{0} 1\n{1:010} 00000 n \n", objectNumber, offset))
//         }

//         if self.newObjectsOffsets {
//             firstNewObject = len(self.objectsOffsets)
//             write(pdfFormat(
//                 "{0} {1}\n", firstNewObject, len(self.newObjectsOffsets)))
//             for objectNumber, offset := range enumerate(
//                     self.newObjectsOffsets, start=firstNewObject) {
//                     }
//                 write(pdfFormat("{0:010} 00000 n \n", offset))
//         }

//         write(pdfFormat(
//             "trailer\n<< "
//             "/Size {size} /Root {root} 0 R /Info {info} 0 R /Prev {prev}"
//             " >>\nstartxref\n{startxref}\n%%EOF\n",
//             size=self.nextObjectNumber(),
//             root=self.catalog.objectNumber,
//             info=self.info.objectNumber,
//             prev=self.startxref,
//             startxref=newStartxref))

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

	pdf.newObjectsOffsets = append(pdf.newObjectsOffsets, &offset)

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
func writePdfMetadata(fileobj io.Reader, scale float64, urlFetcher utils.UrlFetcher, attachments []utils.Attachment,
	attachmentLinks [][]linkData, pages []Page) {

	pdf := PDFFile(fileobj)

	// Add embedded files

	embeddedFilesId, err := writePdfEmbeddedFiles(pdf, attachments, urlFetcher)
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
			if _, targetInAnnotFiles := annotFiles[v.target]; linkType == "attachment" && !targetInAnnotFiles {
				// TODO: use the title attribute as description
				annotFiles[target] = writePdfAttachment(pdf, utils.Attachment{v.target, ""}, urlFetcher)
			}
		}
	}

	for i, pdfPage := range pdf.pages {
		documentPage, pageLinks := pages[i], attachmentLinks[i]

		// FIXME: // Add bleed box

		// mediaBox = pdfPage.getValue(
		//     "MediaBox", "\\[(.+?)\\]").decode("ascii").strip()
		// left, top, right, bottom = (
		//     float(value) for value := range mediaBox.split(" "))
		// // Convert pixels into points
		// bleed = {
		//     key: value * 0.75 for key, value := range documentPage.bleed.items()}

		// trimLeft = left + bleed["left"]
		// trimTop = top + bleed["top"]
		// trimRight = right - bleed["right"]
		// trimBottom = bottom - bleed["bottom"]

		// // Arbitrarly set PDF BleedBox between CSS bleed box (PDF MediaBox) and
		// // CSS page box (PDF TrimBox), at most 10 points from the TrimBox.
		// bleedLeft = trimLeft - min(10, bleed["left"])
		// bleedTop = trimTop - min(10, bleed["top"])
		// bleedRight = trimRight + min(10, bleed["right"])
		// bleedBottom = trimBottom + min(10, bleed["bottom"])

		// pdf.extendDict(pdfPage, pdfFormat(
		//     "/TrimBox [ {} {} {} {} ] /BleedBox [ {} {} {} {} ]".format(
		//         trimLeft, trimTop, trimRight, trimBottom,
		//         bleedLeft, bleedTop, bleedRight, bleedBottom)))

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
				var matrix bo.Matrix = cairo.Matrix(scale, -scale, documentPage.height*scale)
				rectX, rectY, width, height := r[0], r[1], r[2], r[3]
				rectX, rectY = matrix.TransformPoint(rectX, rectY)
				width, height = matrix.TransformDistance(width, height)
				// x, y, w, h => x0, y0, x1, y1
				r = [4]float64{rectX, rectY, rectX + width, rectY + height}
				content := []bts{bts{pdfFormat("<< /Type /Annot /Rect [%.2f %.2f %.2f %.2f] /Border [0 0 0]\n",
					r[0], r[1], r[2], r[3])}}
				linkAp := pdf.writeNewObject("<< /Type /XObject /Subtype /Form "+
					"/BBox [%.2f %.2f %.2f %.2f] /Length 0 >>\n"+
					"stream\n"+
					"endstream", r[0], r[1], r[2], r[3])
				content = append(content, bts{"/Subtype /FileAttachment "})
				// evince needs /T or fails on an internal assertion. PDF
				// doesn't require it.
				content = append(content, bts{pdfFormat("/T () /FS %d 0 R /AP << /N %d 0 R >>",
					*anTarget, linkAp)})
				content = append(content, bts{">>"})
				annotations = append(annotations, pdf.writeNewObject(bytes.Join(content, "")))
			}
		}

		if len(annotations) != 0 {
			chunks := make([]string, len(annotations))
			for i, n := range annotations {
				chunks[i] = pdfFormat("%d 0 R", n)
			}
			pdf.extendDict(pdfPage, bts{pdfFormat("/Annots [%s]", strings.Join(chunks, " "))})
		}
	}

	pdf.finish()
}
