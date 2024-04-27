package buggy_http

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	method  string
	path    string
	proto   string
	headers map[string][]string
	body    []byte
}

func requestLineParser(line string) (*Request, error) {

	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		return &Request{}, fmt.Errorf("requestLineParser(): invalid request line: %q", line)
	}

	return &Request{
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

func requestParser(r io.Reader) (*Request, error) {
	reader := bufio.NewReader(r)

	startLine, _, err := reader.ReadLine()
	if err != nil {
		return &Request{}, fmt.Errorf("requestParser(): %w", err)
	}

	request, err := requestLineParser(strings.TrimSpace(string(startLine)))
	if err != nil {
		return request, fmt.Errorf("requestParser(): %w", err)
	}

	// After the headers there could be a body BUT we will ignore it since we only accept GET, OPTIONS and HEAD,
	// and because keep-alive is not implemented and after response the conn is closed.

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

		request.headers[name] = value

	}

	return request, nil
}
