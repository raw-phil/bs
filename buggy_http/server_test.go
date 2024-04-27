package buggy_http

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartBuggyServer(t *testing.T) {
	t.Run("missing baseDir", func(t *testing.T) {
		bs := &BuggyInstance{
			config: &BuggyConfig{
				baseDir:       "",
				readTimeout:   10,
				writeTimeout:  10,
				maxRequestMiB: 10,
			},
			quit: make(chan struct{}),
		}

		err := bs.StartBuggyServer("localhost", 8080)
		assert.Error(t, err, "StartBuggyServer(): Not all BuggyServer fields have a value, use NewBuggyServer()")
	})

	t.Run("missing readTimeout", func(t *testing.T) {
		bs := &BuggyInstance{
			config: &BuggyConfig{
				baseDir:       "./foo",
				readTimeout:   0,
				writeTimeout:  10,
				maxRequestMiB: 10,
			},
			quit: make(chan struct{}),
		}

		err := bs.StartBuggyServer("localhost", 8080)
		assert.Error(t, err, "StartBuggyServer(): Not all BuggyServer fields have a value, use NewBuggyServer()")
	})
}

func TestSetReadTimeout(t *testing.T) {
	bs := &BuggyInstance{config: &BuggyConfig{}}

	t.Run("Error when listener is not nil", func(t *testing.T) {
		bs.listener = &net.TCPListener{}
		err := bs.SetReadTimeout(10 * time.Second)
		assert.Error(t, err)
	})

	t.Run("Success when timeout is zero", func(t *testing.T) {
		bs.listener = nil
		err := bs.SetReadTimeout(0)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(1<<63-1), bs.config.readTimeout)
	})

	t.Run("Success when listener is nil and timeout is positive", func(t *testing.T) {
		bs.listener = nil
		err := bs.SetReadTimeout(10 * time.Second)
		assert.NoError(t, err)
		assert.Equal(t, 10*time.Second, bs.config.readTimeout)
	})
}

func TestSetWriteTimeout(t *testing.T) {
	bs := &BuggyInstance{config: &BuggyConfig{}}

	t.Run("Error when listener is not nil", func(t *testing.T) {
		bs.listener = &net.TCPListener{}
		err := bs.SetWriteTimeout(10 * time.Second)
		assert.Error(t, err)
	})

	t.Run("Success when timeout is zero", func(t *testing.T) {
		bs.listener = nil
		err := bs.SetWriteTimeout(0)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(1<<63-1), bs.config.writeTimeout)
	})

	t.Run("Success when listener is nil and timeout is positive", func(t *testing.T) {
		bs.listener = nil
		err := bs.SetWriteTimeout(10 * time.Second)
		assert.NoError(t, err)
		assert.Equal(t, 10*time.Second, bs.config.writeTimeout)
	})
}

func TestSetmaxRequestMiB(t *testing.T) {
	bs := &BuggyInstance{config: &BuggyConfig{}}

	t.Run("Error when listener is not nil", func(t *testing.T) {
		bs.listener = &net.TCPListener{}
		err := bs.SetmaxRequestMiB(10)
		assert.Error(t, err)
	})

	t.Run("Error when size is zero", func(t *testing.T) {
		bs.listener = nil
		err := bs.SetmaxRequestMiB(0)
		assert.NoError(t, err)
		assert.Equal(t, 0, bs.config.maxRequestMiB)
	})

	t.Run("Success when listener is nil", func(t *testing.T) {
		bs.listener = nil
		err := bs.SetmaxRequestMiB(10)
		assert.NoError(t, err)
		assert.Equal(t, 10, bs.config.maxRequestMiB)
	})
}

func TestSetBaseDir(t *testing.T) {
	bs := &BuggyInstance{config: &BuggyConfig{}}

	t.Run("Error when listener is not nil", func(t *testing.T) {
		bs.listener = &net.TCPListener{}
		err := bs.SetBaseDir("./foo")
		assert.Error(t, err)
	})

	t.Run("Error when baseDir is empty", func(t *testing.T) {
		bs.listener = nil
		err := bs.SetBaseDir("")
		assert.Error(t, err)
	})

	t.Run("Error when listener is nil and baseDir does not exist", func(t *testing.T) {
		bs.listener = nil
		err := bs.SetBaseDir("./non-existing-directory")
		assert.Error(t, err)
	})
}

func TestStopBuggyServer(t *testing.T) {
	t.Run("Test when listener is nil", func(t *testing.T) {
		bs := &BuggyInstance{
			listener: nil,
			quit:     make(chan struct{}),
			config:   &BuggyConfig{},
		}
		err := bs.StopBuggyServer()
		assert.Error(t, err)
		assert.Equal(t, "StopBuggyServer(): nil bs.listener, StopBuggyServer() called before StartBuggyServer()", err.Error())
	})

	t.Run("Test when quit is nil", func(t *testing.T) {
		bs := &BuggyInstance{
			listener: &net.TCPListener{},
			quit:     nil,
			config:   &BuggyConfig{},
		}
		err := bs.StopBuggyServer()
		assert.Error(t, err)
		assert.Equal(t, "StopBuggyServer(): nil bs.quit, StopBuggyServer() called before NewBuggyServer()", err.Error())
	})

	t.Run("Test successful stop", func(t *testing.T) {
		ln, _ := net.Listen("tcp", "127.0.0.1:0")
		bs := &BuggyInstance{
			listener: ln,
			quit:     make(chan struct{}),
			config:   &BuggyConfig{},
		}
		err := bs.StopBuggyServer()
		assert.NoError(t, err)
	})

	t.Run("Test StopBuggyServer() called twice", func(t *testing.T) {
		ln, _ := net.Listen("tcp", "127.0.0.1:0")
		bs := &BuggyInstance{
			listener: ln,
			quit:     make(chan struct{}),
			config:   &BuggyConfig{},
		}
		_ = bs.StopBuggyServer()
		err := bs.StopBuggyServer()
		assert.Error(t, err)
		assert.Equal(t, "StopBuggyServer(): recovered panic: close of closed channel", err.Error())
	})

	t.Run("Test stop on already closed connection", func(t *testing.T) {
		ln, _ := net.Listen("tcp", "127.0.0.1:8080")
		bs := &BuggyInstance{
			listener: ln,
			quit:     make(chan struct{}),
			config:   &BuggyConfig{},
		}
		bs.listener.Close()
		err := bs.StopBuggyServer()
		assert.Error(t, err)
		assert.Equal(t, "StopBuggyServer(): during bs.listener.Close(), close tcp 127.0.0.1:8080: use of closed network connection", err.Error())
	})

}
