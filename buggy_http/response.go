package buggy_http

import (
	"fmt"
	net_http "net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Response struct {
	proto        string
	code         int
	reasonPhrase string
	headers      map[string][]string
	body         []byte
}

func reply(request *Request, baseDir string) (*Response, error) {

	if request.proto != "HTTP/1.1" {
		return r505(), fmt.Errorf("reply() -> %s, %s: HTTP version not supported. 505 sent", request.method, request.path)
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

func generateResponse(request *Request, t time.Duration, baseDir string) (*Response, error) {

	ch := make(chan *struct {
		response *Response
		err      error
	})

	go func() {
		response, err := reply(request, baseDir)
		ch <- &struct {
			response *Response
			err      error
		}{response: response, err: err}
	}()

	select {
	case result := <-ch:
		return result.response, result.err

	case <-time.After(t):
		return r500(), fmt.Errorf("generateResponse() -> %s, %s: the server has exceeded the time limit to generate a response. 500 sent", request.method, request.path)
	}
}

func replyToOPTIONS(request *Request, baseDir string) (*Response, error) {

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
		"connection":    {"close"},
	}

	return &Response{
		proto:        "HTTP/1.1",
		code:         204,
		reasonPhrase: "No Content",
		headers:      headers,
		body:         make([]byte, 0),
	}, nil

}

func replyToGET(request *Request, baseDir string) (*Response, error) {

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
		"connection":     {"close"},
	}

	return &Response{
		proto:        "HTTP/1.1",
		code:         200,
		reasonPhrase: "OK",
		headers:      headers,
		body:         file,
	}, nil

}

func replyToHEAD(request *Request, baseDir string) (*Response, error) {
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
		"connection":     {"close"},
	}

	return &Response{
		proto:        "HTTP/1.1",
		code:         200,
		reasonPhrase: "OK",
		headers:      headers,
		body:         make([]byte, 0),
	}, nil

}

func serializeResponse(response *Response) string {

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

func r400() *Response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &Response{
		proto:        "HTTP/1.1",
		code:         400,
		reasonPhrase: "Bad Request",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r404() *Response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &Response{
		proto:        "HTTP/1.1",
		code:         404,
		reasonPhrase: "Not Found",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r405() *Response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"allow":      {"GET", "HEAD", "OPTIONS"},
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &Response{
		proto:        "HTTP/1.1",
		code:         405,
		reasonPhrase: "Method Not Allowed",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r408() *Response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &Response{
		proto:        "HTTP/1.1",
		code:         408,
		reasonPhrase: "Request Timeout",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r500() *Response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &Response{
		proto:        "HTTP/1.1",
		code:         500,
		reasonPhrase: "Internal Server Error",
		headers:      headers,
		body:         make([]byte, 0),
	}
}

func r505() *Response {
	t := time.Now().UTC()

	headers := map[string][]string{
		"date":       {t.Format("Mon, 02 Jan 2006 15:04:05 GMT")},
		"server":     {"BuggyServer"},
		"connection": {"close"},
	}
	return &Response{
		proto:        "HTTP/1.1",
		code:         505,
		reasonPhrase: "HTTP Version Not Supported",
		headers:      headers,
		body:         make([]byte, 0),
	}
}
