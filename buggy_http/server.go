package buggy_http

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// [buggyConfig] is the struct that holds the configuration for a BuggyServer.
type buggyConfig struct {
	// The base directory from which static files will be served.
	baseDir string

	// The maximum duration in seconds for reading the entire
	// request from the underling connection. If it is exceeded server respond with 408 code.
	readTimeout time.Duration

	// The maximum duration in seconds the server has to respond.
	// If it is exceeded server respond with 500 code.
	writeTimeout time.Duration

	// The maximum size of request the server will accept in MiB.
	maxRequestMiB int
}

// [buggyInstance] is the struct that implements the BuggyServer interface.
type buggyInstance struct {
	// The net.Listener that accepts tcp connections.
	listener net.Listener

	// A channel that can be closed to signal the server to stop.
	quit chan struct{}

	// The configuration settings for the server.
	config *buggyConfig
}

type BuggyServer interface {
	SetReadTimeout(seconds int) error
	SetWriteTimeout(seconds int) error
	SetmaxRequestMiB(size int) error
	SetBaseDir(path string) error
	StartBuggyServer(host string, port uint) error
	StopBuggyServer() error

	handleConnection(conn net.Conn)
	listenForConn()
}

// NewBuggyServer creates a BuggyServer with default values:
//
//	baseDir: "./"
//	readTimeout: 290 years -> NO timeout
//	writeTimeout: 290 years -> NO timeout
//	maxRequestMiB: -1 MiB -> NO maximum size
func NewBuggyServer() BuggyServer {

	// default values
	return &buggyInstance{
		config: &buggyConfig{
			baseDir:       "./",
			readTimeout:   (1<<63 - 1),
			writeTimeout:  (1<<63 - 1),
			maxRequestMiB: -1,
		},
		quit: make(chan struct{}),
	}

}

// StartBuggyServer starts a BuggyServer.
// It accepts TCP connections on the specified host and port.
// The server will serve static files from the configured base directory.
//
// Parameters:
//
//	host: The hostname or IP address on which the server should listen.
//	port: The port number on which the server should listen.
func (bs *buggyInstance) StartBuggyServer(host string, port uint) error {

	if bs.config.baseDir == "" ||
		bs.quit == nil ||
		bs.config.readTimeout == 0 ||
		bs.config.writeTimeout == 0 {
		return fmt.Errorf("StartBuggyServer(): Not all BuggyServer fields have a value, use NewBuggyServer()")
	}

	listenAddr := fmt.Sprintf("%s:%d", host, port)

	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("net.Listen(): %w", err)
	}

	log.Printf("Server started on: %s", listenAddr)

	bs.listener = l
	go bs.listenForConn()
	return nil

}

// SetReadTimeout set the maximum duration in seconds for reading the entire
// request from the underling connection. If it is exceeded server respond with 408 code.
// Zero or negative value means there will be no timeout.
func (bs *buggyInstance) SetReadTimeout(seconds int) error {
	if bs.listener != nil {
		return fmt.Errorf("SetReadTimeout(): BuggyServer has already been started, you can no longer change its configuration")
	}

	maxSeconds := (1<<63 - 1) / int(math.Pow(10, 9))

	if seconds <= 0 {
		bs.config.readTimeout = (1<<63 - 1)
		return nil
	} else if seconds > maxSeconds {
		return fmt.Errorf("SetReadTimeout(): number of seconds to large to fit in time.Duration ")
	}

	bs.config.readTimeout = time.Duration(seconds) * time.Second
	return nil
}

// SetWriteTimeout set the maximum duration in seconds the server has to respond.
// If it is exceeded server respond with 500 code.
// Zero or negative value means there will be no timeout.
func (bs *buggyInstance) SetWriteTimeout(seconds int) error {
	if bs.listener != nil {
		return fmt.Errorf("SetWriteTimeout(): BuggyServer has already been started, you can no longer change its configuration")
	}

	maxSeconds := (1<<63 - 1) / int(math.Pow(10, 9))

	if seconds <= 0 {
		bs.config.writeTimeout = (1<<63 - 1)
		return nil
	} else if seconds > maxSeconds {
		return fmt.Errorf("SetWriteTimeout(): number of seconds to large to fit in time.Duration")
	}

	bs.config.writeTimeout = time.Duration(seconds) * time.Second
	return nil
}

// SetmaxRequestMiB set the maximum size of request the server will accept in MiB.
// Zero or negative value means there will be no maximum request size.
func (bs *buggyInstance) SetmaxRequestMiB(size int) error {
	if bs.listener != nil {
		return fmt.Errorf("SetmaxRequestMiB(): BuggyServer has already been started, you can no longer change its configuration")
	}

	bs.config.maxRequestMiB = size
	return nil
}

// SetBaseDir set the base directory from which static files will be served.
// It accepts relative or absolute path.
func (bs *buggyInstance) SetBaseDir(path string) error {
	if bs.listener != nil {
		return fmt.Errorf("SetBaseDir(): BuggyServer has already been started, you can no longer change its configuration")
	}
	if path == "" {
		return fmt.Errorf("SetBaseDir(): cannot be and empty string, the value won't be updated")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("SetBaseDir(): the path is not valid, %w", err)
	}

	baseDir, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("SetBaseDir(): the path is not valid: %w", err)
	}
	if baseDir.IsDir() {
		bs.config.baseDir = path
		return nil
	}

	return fmt.Errorf("SetBaseDir(): %s is not a directory", path)

}

func (bs *buggyInstance) handleConnection(conn net.Conn) {

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error: handleConnection(): conn.Close(): %s: %s", conn.RemoteAddr(), err.Error())
		}
	}()

	bufReader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(bs.config.readTimeout))

		var response *response

		request, err := requestParser(bufReader, bs.config.maxRequestMiB)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNRESET) {
				log.Printf("error: handleConnection(): %s:%s, the underlying connection is closed", err.Error(), conn.RemoteAddr())
				break
			}
			if errors.Is(err, os.ErrDeadlineExceeded) {
				response = r408()
			} else {
				response = r400()
			}
			log.Printf("error: handleConnection(): %s. %d sent", err.Error(), response.code)

		} else {
			response, err = generateResponse(request, bs.config.writeTimeout, bs.config.baseDir)
			if err != nil {
				log.Printf("error: handleConnection(): %s", err.Error())
			}

			if headerFinder(request.headers, "connection", "close") {
				addCloseConnectionHeader(response)

			} else if _, ok := response.headers["connection"]; !ok {
				addKeepAliveHeaders(response, int(bs.config.readTimeout))
			}
		}

		err = sendResponse(conn, response)
		if err != nil {
			log.Printf("error: handleConnection(): %s", err.Error())
		} else {
			log.Printf("[ %s, %s, %s : %d ]", conn.RemoteAddr(), request.method, request.path, response.code)
		}

		if values, ok := response.headers["connection"]; ok && values[0] == "close" {
			break
		}

	}

}

func (bs *buggyInstance) listenForConn() {
	for {

		conn, err := bs.listener.Accept()
		if err != nil {
			select {
			case <-bs.quit:
				return
			default:
				log.Printf("error: listenForConn(): accepting connection: %s", err.Error())
				continue
			}

		}

		go bs.handleConnection(conn)

	}

}

func (bs *buggyInstance) StopBuggyServer() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("StopBuggyServer(): recovered panic: %s", r)
		}
	}()

	if bs.listener == nil {
		return fmt.Errorf("StopBuggyServer(): nil bs.listener, StopBuggyServer() called before StartBuggyServer()")
	}
	if bs.quit == nil {
		return fmt.Errorf("StopBuggyServer(): nil bs.quit, StopBuggyServer() called before NewBuggyServer()")
	}
	close(bs.quit)
	err = bs.listener.Close()
	if err != nil {
		return fmt.Errorf("StopBuggyServer(): during bs.listener.Close(), %w", err)
	}
	return nil
}

func sendResponse(conn net.Conn, response *response) error {
	if _, err := conn.Write([]byte(serializeResponse(response))); err != nil {
		return fmt.Errorf("sendResponse(): %s: %w", conn.RemoteAddr(), err)
	}
	return nil
}
