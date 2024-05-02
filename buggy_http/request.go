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

func requestParser(reader *bufio.Reader, maxRequestMiB int) (*request, error) {

	var maxRequestBytes int = 0
	if maxRequestMiB > 0 {
		maxRequestBytes = maxRequestMiB * (1 << 20)
	}

	var byteCount int = 0

	startLine, err := readLine(reader, &byteCount, maxRequestBytes)
	if err != nil {
		return &request{}, fmt.Errorf("requestParser(): %w", err)
	}

	parsedRequest, err := requestLineParser(strings.TrimSpace(string(startLine)))
	if err != nil {
		return parsedRequest, fmt.Errorf("requestParser(): %w", err)
	}

	for {
		byteLine, err := readLine(reader, &byteCount, maxRequestBytes)
		if err != nil {
			return parsedRequest, fmt.Errorf("requestParser(): %w", err)
		}

		line := strings.TrimSpace(string(byteLine))
		if line == "" {
			break
		}

		name, value, err := headerLineParser(line)
		if err != nil {
			return parsedRequest, fmt.Errorf("requestParser(): %w", err)
		}

		if _, ok := parsedRequest.headers[name]; ok {
			parsedRequest.headers[name] = append(parsedRequest.headers[name], value...)
		} else {
			parsedRequest.headers[name] = value
		}

	}

	if headerFinder(parsedRequest.headers, "transfer-encoding", "chunked") {
		return parsedRequest, fmt.Errorf("requestParser(): BuggyServer does not support transfer-encoding: chunked")
	}

	if err = readBody(reader, parsedRequest); err != nil {
		return parsedRequest, fmt.Errorf("requestParser(): %w", err)
	}

	//fmt.Printf("%+v", parsedRequest)
	return parsedRequest, nil
}

func readLine(reader *bufio.Reader, byteCount *int, maxRequestBytes int) ([]byte, error) {
	line, isPrefix, err := reader.ReadLine()
	if err != nil {
		return nil, err
	}
	if isPrefix {
		return nil, fmt.Errorf("readLine(): line exceeded max size")
	}
	if maxRequestBytes > 0 {
		*byteCount += len(line)
		if *byteCount > int(maxRequestBytes) {
			return nil, fmt.Errorf("readLine(): request exceeded max size")
		}
	}
	return line, nil
}

// This function serves to find if a heder exist in the headersMap
// and if it has a given value.
func headerFinder(headersMap map[string][]string, header, value string) bool {
	if values, ok := headersMap[strings.ToLower(header)]; ok {
		for _, v := range values {
			if strings.EqualFold(v, value) {
				return true
			}
		}
	}
	return false
}

func readBody(reader *bufio.Reader, req *request) error {
	if value, ok := req.headers["content-length"]; ok {
		contentLength, err := strconv.Atoi(value[0])
		if err != nil {
			return err
		}

		req.body = make([]byte, contentLength)
		_, err = io.ReadFull(reader, req.body)
		if err != nil {
			return fmt.Errorf("parseBody(): io.ReadFull(): %w", err)
		}
	} else {
		req.body = make([]byte, 0)
	}
	return nil
}
