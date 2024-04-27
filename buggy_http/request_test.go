package buggy_http

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeaderLineParser(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		header := ""
		name, value, err := headerLineParser(header)
		assert.NotNil(t, err)
		assert.Empty(t, name)
		assert.Nil(t, value)
	})

	t.Run("Missing header name", func(t *testing.T) {
		header := ": NoHeaderName"
		name, value, err := headerLineParser(header)
		assert.NotNil(t, err)
		assert.Empty(t, name)
		assert.Nil(t, value)
	})

	t.Run("Header line with extra colons", func(t *testing.T) {
		line := "Accept: text/plain: text/html"
		expectedName := "accept"
		expectedValue := []string{"text/plain: text/html"}

		name, value, err := headerLineParser(line)

		assert.NoError(t, err)
		assert.Equal(t, expectedName, name)
		assert.Equal(t, expectedValue, value)
	})

	t.Run("Header line with only whitespace", func(t *testing.T) {
		line := "     "
		expectedName := ""
		var expectedValue []string

		name, value, err := headerLineParser(line)

		assert.Error(t, err)
		assert.Equal(t, expectedName, name)
		assert.Equal(t, expectedValue, value)
	})
	t.Run("header line with multiple spaces", func(t *testing.T) {
		line := "Accept:    text/plain,   text/html"
		expectedName := "accept"
		expectedValue := []string{"text/plain", "text/html"}

		name, value, err := headerLineParser(line)

		assert.NoError(t, err)
		assert.Equal(t, expectedName, name)
		assert.Equal(t, expectedValue, value)
	})

	t.Run("header line with tab characters", func(t *testing.T) {
		line := "Accept:\ttext/plain,\ttext/html"
		expectedName := "accept"
		expectedValue := []string{"text/plain", "text/html"}

		name, value, err := headerLineParser(line)

		assert.NoError(t, err)
		assert.Equal(t, expectedName, name)
		assert.Equal(t, expectedValue, value)
	})

	t.Run("header line with mixed spaces and tabs", func(t *testing.T) {
		line := "Accept: \t text/plain, \t text/html"
		expectedName := "accept"
		expectedValue := []string{"text/plain", "text/html"}

		name, value, err := headerLineParser(line)

		assert.NoError(t, err)
		assert.Equal(t, expectedName, name)
		assert.Equal(t, expectedValue, value)
	})

	t.Run("header line with no value", func(t *testing.T) {
		line := "Accept:"
		expectedName := "accept"
		expectedValue := []string{""}

		name, value, err := headerLineParser(line)

		assert.NoError(t, err)
		assert.Equal(t, expectedName, name)
		assert.Equal(t, expectedValue, value)
	})
}

func TestRequestLineParser(t *testing.T) {
	testCases := []struct {
		name          string
		line          string
		expectedError error
		expectedReq   *Request
	}{
		{
			name:          "Valid GET request",
			line:          "GET / HTTP/1.1",
			expectedError: nil,
			expectedReq:   &Request{method: "GET", path: "/", proto: "HTTP/1.1", headers: make(map[string][]string)},
		},
		{
			name:          "Valid POST request",
			line:          "POST /login HTTP/1.1",
			expectedError: nil,
			expectedReq:   &Request{method: "POST", path: "/login", proto: "HTTP/1.1", headers: make(map[string][]string)},
		},
		{
			name:          "Invalid request line with 2 parts",
			line:          "GET /",
			expectedError: fmt.Errorf("requestLineParser(): invalid request line: %q", "GET /"),
			expectedReq:   &Request{},
		},
		{
			name:          "Invalid request line with 4 parts",
			line:          "GET / HTTP/1.1 extra",
			expectedError: fmt.Errorf("requestLineParser(): invalid request line: %q", "GET / HTTP/1.1 extra"),
			expectedReq:   &Request{},
		},
		{
			name:          "Invalid empty request line",
			line:          "",
			expectedError: fmt.Errorf("requestLineParser(): invalid request line: \"\""),
			expectedReq:   &Request{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := requestLineParser(tc.line)
			assert.Equal(t, tc.expectedError, err)
			assert.Equal(t, tc.expectedReq, req)
		})
	}
}
