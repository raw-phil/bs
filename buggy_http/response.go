package buggy_http

import (
	"fmt"
	"math"
	net_http "net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type response struct {
	proto        string
	code         int
	reasonPhrase string
	headers      map[string][]string
	body         []byte
}

// generateResponse generates a response for a give request.
// If error is not nil, the returned response have the HTTP code associated with that error.
func generateResponse(request *request, t time.Duration, baseDir string) (*response, error) {

	ch := make(chan *struct {
		r   *response
		err error
	})

	go func() {
		r, err := reply(request, baseDir)
		ch <- &struct {
			r   *response
			err error
		}{r: r, err: err}
	}()

	select {
	case result := <-ch:
		return result.r, result.err

	case <-time.After(t):
		return r500(), fmt.Errorf("generateResponse() -> %s, %s: the server has exceeded the time limit to generate a response. 500 sent", request.method, request.path)
	}
}

func reply(request *request, baseDir string) (*response, error) {

	if request.proto != "HTTP/1.1" {
		return addCloseConnectionHeader(r505()), fmt.Errorf("reply() -> %s, %s: HTTP version not supported. 505 sent", request.method, request.path)
	}

	switch request.method {
	case "OPTIONS":
		return replyToOPTIONS(request, baseDir)

	case "GET":
		return replyToGET(request, baseDir)

	case "HEAD":
		return replyToHEAD(request, baseDir)

	default:
		return r405(), fmt.Errorf("reply() -> %s, %s: HTTP method not allowed. 405 sent", request.method, request.path)
	}

}

func replyToOPTIONS(request *request, baseDir string) (*response, error) {

	// asterisk (*) refer to the entire server.
	if request.path != "*" {

		_, err := validatePath(baseDir, request.path)
		if err != nil {
			return r404(), fmt.Errorf("replyToOPTIONS() -> %s, %s : %w. 404 sent", request.method, request.path, err)
		}
	}

	t := time.Now().UTC()

	headers := map[string][]string{
		"allow":         {"GET", "HEAD", "OPTIONS"},
		"cache-control": {"max-age=604800"},
		"date":          {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":        {"BuggyServer"},
	}

	return &response{
		proto:        "HTTP/1.1",
		code:         204,
		reasonPhrase: "No Content",
		headers:      headers,
		body:         make([]byte, 0),
	}, nil

}

func replyToGET(request *request, baseDir string) (*response, error) {

	path, err := url.QueryUnescape(request.path)
	if err != nil {
		return r404(), fmt.Errorf("replyToGET() -> %s, %s : %w. 404 sent", request.method, request.path, err)
	}

	if request.path == "/" {
		path = "/index.html"
	}

	path, err = validatePath(baseDir, path)
	if err != nil {
		return r404(), fmt.Errorf("replyToGET() -> %s, %s : %w. 404 sent", request.method, request.path, err)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return r500(), fmt.Errorf("replyToGET() -> %s, %s : %w. 500 sent", request.method, request.path, err)
	}

	// Implements the algorithm described at https://mimesniff.spec.whatwg.org/
	mimeType := net_http.DetectContentType(file)

	t := time.Now().UTC()

	headers := map[string][]string{
		"date":           {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":         {"BuggyServer"},
		"content-type":   {mimeType},
		"content-length": {fmt.Sprintf("%v", len(file))},
	}

	return &response{
		proto:        "HTTP/1.1",
		code:         200,
		reasonPhrase: "OK",
		headers:      headers,
		body:         file,
	}, nil

}

func replyToHEAD(request *request, baseDir string) (*response, error) {
	path, _ := url.QueryUnescape(request.path)

	path, err := validatePath(baseDir, path)
	if err != nil {
		return r404(), fmt.Errorf("replyToHEAD() -> %s, %s : %w. 404 sent", request.method, request.path, err)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return r500(), fmt.Errorf("replyToHEAD() -> %s, %s : %w. 500 sent", request.method, request.path, err)
	}

	// Implements the algorithm described at https://mimesniff.spec.whatwg.org/
	mimeType := net_http.DetectContentType(file)

	t := time.Now().UTC()

	headers := map[string][]string{
		"date":           {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":         {"BuggyServer"},
		"content-type":   {mimeType},
		"content-length": {fmt.Sprintf("%v", len(file))},
	}

	return &response{
		proto:        "HTTP/1.1",
		code:         200,
		reasonPhrase: "OK",
		headers:      headers,
		body:         make([]byte, 0),
	}, nil

}

func serializeResponse(response *response) string {

	r := fmt.Sprintf("%s %d %s\r\n", response.proto, response.code, response.reasonPhrase)

	for key, values := range response.headers {
		r += key + ": "
		for i, value := range values {
			r += value
			if i != len(values)-1 {
				r += ", "
			}
		}
		r += "\r\n"
	}

	r += "\r\n" + string(response.body)

	return r
}

func validatePath(baseDir string, p string) (string, error) {

	path := filepath.Join(baseDir, p)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}

	// Check for path traversal
	if !strings.HasPrefix(absPath, absBaseDir) {
		return "", fmt.Errorf("validatePath(): invalid path: path is outside the base directory")
	}

	// Check if the file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}
	if fileInfo.IsDir() {
		return "", fmt.Errorf("validatePath(): invalid path path is to a directory")
	}

	return absPath, nil
}

func r400() *response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &response{
		proto:        "HTTP/1.1",
		code:         400,
		reasonPhrase: "Bad Request",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r404() *response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":   {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server": {"BuggyServer"},
	}
	return &response{
		proto:        "HTTP/1.1",
		code:         404,
		reasonPhrase: "Not Found",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r405() *response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"allow":  {"GET", "HEAD", "OPTIONS"},
		"date":   {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server": {"BuggyServer"},
	}
	return &response{
		proto:        "HTTP/1.1",
		code:         405,
		reasonPhrase: "Method Not Allowed",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r408() *response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &response{
		proto:        "HTTP/1.1",
		code:         408,
		reasonPhrase: "Request Timeout",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r500() *response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &response{
		proto:        "HTTP/1.1",
		code:         500,
		reasonPhrase: "Internal Server Error",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r505() *response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &response{
		proto:        "HTTP/1.1",
		code:         505,
		reasonPhrase: "HTTP Version Not Supported",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

// Add 'connection: close' header to a response
func addCloseConnectionHeader(r *response) *response {
	r.headers["connection"] = []string{"close"}
	return r
}

// Add 'connection: keep-alive' and 'keep-alive: timeout=X, max=10' headers to a response
func addKeepAliveHeaders(r *response, timeout int) *response {
	r.headers["connection"] = []string{"keep-alive"}
	if timeout != (1<<63 - 1) {
		r.headers["keep-alive"] = []string{fmt.Sprintf("timeout=%d", timeout/int(math.Pow(10, 9)))}
	}
	return r
}
