package buggy_http

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type request struct {
	method  string
	path    string
	proto   string
	headers map[string][]string
	body    []byte
}

func requestLineParser(line string) (*request, error) {

	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		return &request{}, fmt.Errorf("requestLineParser(): invalid request line: %q", line)
	}

	return &request{
		method:  parts[0],
		path:    parts[1],
		proto:   parts[2],
		headers: make(map[string][]string),
	}, nil
}

func headerLineParser(line string) (string, []string, error) {

	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 2 {
		return "", nil, fmt.Errorf("headerLineParser(): invalid header line: %q", line)
	}
	if parts[0] == "" {
		return "", nil, fmt.Errorf("headerLineParser(): missing header name: %q", line)
	}

	name := strings.ToLower(strings.TrimSpace(parts[0]))
	value := strings.Split(parts[1], ",")

	for i := range value {
		value[i] = strings.TrimSpace(value[i])
	}

	return name, value, nil
}

func requestParser(r io.Reader) (*request, error) {
	reader := bufio.NewReader(r)

	startLine, _, err := reader.ReadLine()
	if err != nil {
		return &request{}, fmt.Errorf("requestParser(): %w", err)
	}

	request, err := requestLineParser(strings.TrimSpace(string(startLine)))
	if err != nil {
		return request, fmt.Errorf("requestParser(): %w", err)
	}

	for {
		byteLine, _, err := reader.ReadLine()
		if err != nil {
			return request, fmt.Errorf("requestParser(): %w", err)
		}

		line := strings.TrimSpace(string(byteLine))
		if line == "" {
			break
		}

		name, value, err := headerLineParser(line)
		if err != nil {
			return request, fmt.Errorf("requestParser(): %w", err)
		}

		if _, ok := request.headers[name]; ok {
			request.headers[name] = append(request.headers[name], value...)
		} else {
			request.headers[name] = value
		}

	}

	// Check for chunked transfer-encoding
	if values, ok := request.headers["transfer-encoding"]; ok {
		for _, value := range values {
			if strings.ToLower(value) == "chunked" {
				return request, fmt.Errorf("requestParser(): BuggyServer does not support 'Transfer-Encoding: chunked'")
			}
		}
	}

	// Read body
	if value, ok := request.headers["content-length"]; ok {

		contentLength, err := strconv.Atoi(value[0])
		if err != nil {
			return request, fmt.Errorf("requestParser(): %w", err)
		}

		request.body = make([]byte, contentLength)
		_, err = io.ReadFull(reader, request.body)
		if err != nil {
			return request, fmt.Errorf("requestParser(): io.ReadFull(): %w", err)
		}
	} else {
		request.body = make([]byte, 0)
	}

	//fmt.Printf("%+v", request)
	return request, nil
}
