package utils

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/benoitkugler/go-weasyprint/version"
	"github.com/benoitkugler/textlayout/fonts"

	"github.com/vincent-petithory/dataurl"
)

// warn if baseUrl is required but missing.
func UrlJoin(baseUrl, urlS string, allowRelative bool, context ...interface{}) string {
	out, err := SafeUrljoin(baseUrl, urlS, allowRelative)
	if err != nil {
		log.Println(err, context)
	}
	return out
}

func BasicUrlJoin(baseUrl, urls string) (string, error) {
	parsedBase, err := url.Parse(baseUrl)
	if err != nil {
		return "", fmt.Errorf("Invalid base url : %s", baseUrl)
	}
	if path.Ext(parsedBase.Path) != "" {
		parsedBase.Path = path.Join(path.Dir(parsedBase.Path), urls)
	} else {
		parsedBase.Path = path.Join(parsedBase.Path, urls)
	}
	return parsedBase.String(), nil
}

// defaut: allowRelative = false
func SafeUrljoin(baseUrl, urls string, allowRelative bool) (string, error) {
	parsed, err := url.Parse(urls)
	if err != nil {
		return "", fmt.Errorf("Invalid url : %s", urls)
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	} else if baseUrl != "" {
		return BasicUrlJoin(baseUrl, urls)
	} else if allowRelative {
		return parsed.String(), nil
	} else {
		return "", errors.New("Relative URI reference without a base URI: " + urls)
	}
}

// Get the URI corresponding to the ``attrName`` attribute.
// Return "" if:
// * the attribute is empty or missing or,
// * the value is a relative URI but the document has no base URI and
//   ``allowRelative`` is ``False``.
// Otherwise return an URI, absolute if possible.
func (element HTMLNode) GetUrlAttribute(attrName, baseUrl string, allowRelative bool) string {
	value := strings.TrimSpace(element.Get(attrName))
	if value != "" {
		return UrlJoin(baseUrl, value, allowRelative,
			fmt.Sprintf("<%s %s='%s'>", element.Data, attrName, value))
	}
	return ""
}

func Unquote(s string) string {
	unescaped, err := url.PathUnescape(s)
	if err != nil {
		log.Println(err)
		return ""
	}
	return unescaped
}

// Url represent an url which can be either internal or external
type Url struct {
	Url      string
	Internal bool
}

func (u Url) IsNone() bool {
	return u == Url{}
}

// Return ('external', absolute_uri) or
// ('internal', unquoted_fragment_id) or nil.s
func GetLinkAttribute(element HTMLNode, attrName string, baseUrl string) []string {
	attrValue := strings.TrimSpace(element.Get(attrName))
	if strings.HasPrefix(attrValue, "#") && len(attrValue) > 1 {
		// Do not require a baseUrl when the value is just a fragment.
		unescaped := Unquote(attrValue[1:])
		return []string{"internal", unescaped}
	}
	uri := element.GetUrlAttribute(attrName, baseUrl, true)
	if uri != "" {
		if baseUrl != "" {
			parsed, err := url.Parse(uri)
			if err != nil {
				log.Println(err)
				return nil
			}
			baseParsed, err := url.Parse(baseUrl)
			if err != nil {
				log.Println(err)
				return nil
			}
			if parsed.Scheme == baseParsed.Scheme && parsed.Host == baseParsed.Host && parsed.Path == baseParsed.Path && parsed.RawQuery == baseParsed.RawQuery {
				// Compare with fragments removed
				unescaped := Unquote(parsed.Fragment)
				return []string{"internal", unescaped}
			}
		}
		return []string{"external", uri}
	}
	return nil
}

// Return file URL of `path`.
func Path2url(urlPath string) (out string, err error) {
	urlPath, err = filepath.Abs(urlPath)
	if err != nil {
		return "", err
	}
	fileinfo, err := os.Lstat(urlPath)
	if err != nil {
		return "", err
	}
	if fileinfo.IsDir() {
		// Make sure directory names have a trailing slash.
		// Otherwise relative URIs are resolved from the parent directory.
		urlPath += string(filepath.Separator)
	}
	urlPath = filepath.ToSlash(urlPath)
	return "file://" + urlPath, nil
}

// Get a ``scheme://path`` URL from ``string``.
//
// If ``string`` looks like an URL, return it unchanged. Otherwise assume a
// filename and convert it to a ``file://`` URL.
func EnsureUrl(urlS string) (string, error) {
	parsed, err := url.Parse(urlS)
	if err != nil {
		return "", fmt.Errorf("Invalid url : %s", urlS)
	}
	if parsed.IsAbs() {
		return urlS, nil
	}
	return Path2url(urlS)
}

type content interface {
	io.Closer
	fonts.Resource
}

type RemoteRessource struct {
	Content content

	// Optionnals values

	// MIME type extracted e.g. from a *Content-Type* header. If not provided, the type is guessed from the
	// 	file extension in the URL.
	MimeType string

	// actual URL of the resource
	// 	if there were e.g. HTTP redirects.
	RedirectedUrl string

	// filename of the resource. Usually
	// 	derived from the *filename* parameter in a *Content-Disposition*
	// 	header
	Filename string

	ProtocolEncoding string
}

type UrlFetcher = func(url string) (RemoteRessource, error)

type BytesCloser bytes.Reader

func (g *BytesCloser) Read(p []byte) (n int, err error) {
	return (*bytes.Reader)(g).Read(p)
}

func (g *BytesCloser) ReadAt(p []byte, off int64) (n int, err error) {
	return (*bytes.Reader)(g).ReadAt(p, off)
}

func (g *BytesCloser) Seek(off int64, whence int) (n int64, err error) {
	return (*bytes.Reader)(g).Seek(off, whence)
}

func (BytesCloser) Close() error { return nil }

func NewBytesCloser(s string) *BytesCloser {
	return (*BytesCloser)(bytes.NewReader([]byte(s)))
}

// Fetch an external resource such as an image or stylesheet.
func DefaultUrlFetcher(urlTarget string) (RemoteRessource, error) {
	if strings.HasPrefix(strings.ToLower(urlTarget), "data:") {
		// data url can't contains spaces and the strings comming from css
		// may contain tabs when separated on several lines with \
		urlTarget = htmlSpacesRe.ReplaceAllString(urlTarget, "")
		data, err := dataurl.DecodeString(urlTarget)
		if err != nil {
			return RemoteRessource{}, err
		}
		return RemoteRessource{
			Content:          (*BytesCloser)(bytes.NewReader(data.Data)),
			MimeType:         data.ContentType(),
			RedirectedUrl:    urlTarget,
			ProtocolEncoding: data.Params["charset"],
		}, nil
	}

	data, err := url.Parse(urlTarget)
	if err != nil {
		return RemoteRessource{}, err
	}
	if !data.IsAbs() {
		return RemoteRessource{}, fmt.Errorf("Not an absolute URI: %s", urlTarget)
	}
	urlTarget = data.String()

	if data.Scheme == "file" {
		var f content
		f, err = os.Open(data.Path)
		if err != nil {
			return RemoteRessource{}, fmt.Errorf("local file not found : %s", err)
		}
		return RemoteRessource{
			Content:       f,
			Filename:      filepath.Base(data.Path),
			MimeType:      mime.TypeByExtension(filepath.Ext(data.Path)),
			RedirectedUrl: urlTarget,
		}, nil
	}

	req, err := http.NewRequest(http.MethodGet, urlTarget, nil)
	if err != nil {
		return RemoteRessource{}, err
	}
	req.Header.Set("User-Agent", version.VersionString)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return RemoteRessource{}, err
	}
	defer response.Body.Close()

	result := RemoteRessource{}
	redirect, err := response.Location()
	if err == nil {
		result.RedirectedUrl = redirect.String()
	}
	mediaType, params, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err == nil {
		result.MimeType = mediaType
		result.ProtocolEncoding = params["charset"]
	}
	_, params, err = mime.ParseMediaType(response.Header.Get("Content-Disposition"))
	if err == nil {
		result.Filename = params["filename"]
	}

	contentEncoding := response.Header.Get("Content-Encoding")
	var r io.Reader
	if contentEncoding == "gzip" {
		r, err = gzip.NewReader(response.Body)
		if err != nil {
			return RemoteRessource{}, err
		}
	} else if contentEncoding == "deflate" {
		r, err = zlib.NewReader(response.Body)
		if err != nil {
			return RemoteRessource{}, err
		}
	} else {
		r = response.Body
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return RemoteRessource{}, err
	}
	result.Content = (*BytesCloser)(bytes.NewReader(buf.Bytes()))

	return result, nil
}
