package buggy_http

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePath(t *testing.T) {

	t.Run("file does not exist", func(t *testing.T) {
		_, err := validatePath("./testBaseDir", "nonexistentfile")
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("path traversal", func(t *testing.T) {
		expectedError := fmt.Errorf("validatePath(): invalid path: path is outside the base directory")
		_, err := validatePath("./testBaseDir", "../../../etc/passwd")
		assert.Equal(t, expectedError, err)
	})

}

func TestSerializeResponse(t *testing.T) {
	response := &response{
		proto:        "HTTP/1.1",
		code:         200,
		reasonPhrase: "OK",
		headers:      map[string][]string{"content-type": {"text/html"}, "content-length": {"13"}},
		body:         []byte("Hello, world!"),
	}

	result := serializeResponse(response)

	expected1 := "HTTP/1.1 200 OK\r\ncontent-type: text/html\r\ncontent-length: 13\r\n\r\nHello, world!"
	expected2 := "HTTP/1.1 200 OK\r\ncontent-length: 13\r\ncontent-type: text/html\r\n\r\nHello, world!"

	assert.True(t, result == expected1 || result == expected2, "The result does not match any of the expected strings.")
}
