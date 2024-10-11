# BuggyServer
        ______                           _____
        | ___ \                         /  ___|
        | |_/ /_   _  __ _  __ _ _   _  \ `--.  ___ _ ____   _____ _ __
        | ___ \ | | |/ _` |/ _` | | | |  `--. \/ _ \ '__\ \ / / _ \ '__|
        | |_/ / |_| | (_| | (_| | |_| | /\__/ /  __/ |   \ V /  __/ |
        \____/ \__,_|\__, |\__, |\__, | \____/ \___|_|    \_/ \___|_|
                      __/ | __/ | __/ |
                     |___/ |___/ |___/       
                     
[![Go Reference](https://pkg.go.dev/badge/golang.org/x/example.svg)](https://pkg.go.dev/golang.org/x/example)
[![Go Report Card](https://goreportcard.com/badge/github.com/raw-phil/bs)](https://goreportcard.com/report/github.com/raw-phil/bs)


Free time project to learn [GO](https://go.dev/) language and something more about HTTP/1.1 protocol. 

BuggyServer is a minimal HTTP/1.1 server built from scratch that serves static files, it does not use any HTTP packages like [net/http](https://pkg.go.dev/net/http).

I have followed the [HTTP/1.1 spec](https://www.rfc-editor.org/rfc/rfc9112), and [HTTP/1.1 Semantics](https://www.rfc-editor.org/rfc/rfc9110).

> [!WARNING]
> BuggyServer is not suitable for production use. It lacks critical features necessary for a robust, secure, and reliable production server.
> Use it solely for educational and exploration purposes.

- [About](#buggyserver)
- [Usage](#usage)
  - [CLI](#cli)
    - [Install](#install)
    - [Help](#help)
    - [Run](#run)
  - [Import in other GO packages](#import-in-other-go-packages)
- [Functionalities: Currently Implemented](#white_check_mark-functionalities-currently-implemented)
  - [GET](#get)
  - [HEAD](#head)
  - [OPTIONS](#options)
  - [Request and Response Timeout](#request-and-response-timeout)
  - [Request size limit](#reqest-size-limit)
  - [Connection reuse and pipelining](#connection-reuse-and-pipelining)
- [Functionalities: Not Yet Implemented](#hourglass-functionalities-not-yet-implemented)
  - [Level based logging](#level-based-logging)


## Usage
### CLI

> [!IMPORTANT]  
> Tested only on Linux and MacOS

#### Install:
```bash
go install github.com/raw-phil/bs@latest
```
#### Help:
```bash
bs --help
```
```
  -d string
        Directory from which files are served (default "./")
  -h string
        Sets the host (default "0.0.0.0")
  -p uint
        Sets the port (default 8080)
  -no-banner
        Suppress the initial banner
  -read-timeout int
        Maximum duration in seconds server has for reading the entire request from the underlying connection.
        Zero or negative value means there will be no timeout. (default -1)
  -write-timeout int
        Maximum duration in seconds the server has to respond.
        Zero or negative value means there will be no timeout. (default -1)
  -max-request-size int
        Maximum size of request the server will accept in MiB.
        Zero or negative value means there will be no maximum size. (default -1)
```

#### Run:
This command launch server on port 3333, and make it serves files from `./foo` directory.
```bash
bs -p 3333 -d ./foo
```


### Import in other GO packages

```go
package mypackage

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/raw-phil/bs/buggy_http"
)

func StartServer(directory, host string, port uint) {
	log.Println("Starting Buggy Server...")

	bs := buggy_http.NewBuggyServer()

	if err := bs.SetBaseDir(directory); err != nil {
		log.Fatalf("Failed to set base directory: %s", err.Error())
	}

	if err := bs.StartBuggyServer(host, port); err != nil {
		log.Fatalf("Failed to start server: %s", err.Error())
	}

	log.Println("Buggy Server started successfully.")
}
```


## :white_check_mark: Functionalities: Currently Implemented

BuggyServer implements **GET**, **HEAD** and **OPTIONS** [HTTP Methods](https://www.rfc-editor.org/rfc/rfc9110#section-9),     
It has Read and Write timeout basic mechanisms and configurable maximum requests size.

### GET
It serves static files from the selected base directory of the host filesystem.  
Requests to `/` are the same as `index.html`.

```bash
$ curl -i 127.0.0.1:8080/

HTTP/1.1 200 OK
date: Tue, 09 Apr 2024 10:35:37 GMT
server: BuggyServer
content-type: text/html; charset=utf-8
content-length: 697

<!DOCTYPE html>
<html lang="en">
        . . . 
</html>
```

### HEAD
Same as GET, but does not send the response content, only headers are sent.
```bash
$ curl -I 127.0.0.1:8080/

HTTP/1.1 200 OK
date: Tue, 09 Apr 2024 10:35:37 GMT
server: BuggyServer
content-type: text/html; charset=utf-8
content-length: 697
```

### OPTIONS
Return allowed [HTTP Methods](https://www.rfc-editor.org/rfc/rfc9110#section-9), for a given endpoint.  
Requests to `*` ( OPTIONS * HTTP/1.1 ) refer to the entire server.

```bash
$ curl -i --request-target "*" -X OPTIONS 127.0.0.1:8080

HTTP/1.1 204 No Content
allow: GET, HEAD, OPTIONS
cache-control: max-age=604800
date: Mon, 15 Apr 2024 11:50:58 GMT
server: BuggyServer
```

### Request and Response Timeout

BuggyServer uses two fields to implement timeouts:

- `ReadTimeout` set the maximum duration in seconds for reading the entire   
request from the underlying connection. If it is exceeded server responds with code 408.    
Zero or negative value means that there will be no timeout.

- `WriteTimeout` set the maximum duration in seconds that the server has to respond within.   
If it is exceeded server responds with code 500.   
Zero or negative value means that there will be no timeout.

### Request size limit


The `maxRequestMiB` field sets the maximum MiB size the server will accept. 
It indicates how many MiB could be read from the underlying connection for each request.
Zero or negative value means there will be no maximum request size.

### Connection reuse and pipelining

BuggyServer supports connection reuse, which allows multiple HTTP requests and responses to be sent over a single TCP connection.   
The server sends the `connection: keep-alive` header and the `keep-alive` header with a timeout parameter ( equal to `ReadTimeout` )   
that indicates the maximum time in seconds the server will keep an idle connection open before closing it"

Additionally, BuggyServer supports pipelined requests, which enable sending multiple HTTP requests in a single TCP connection without waiting for each response.  

```bash
$ echo -ne "GET /foo HTTP/1.1\r\n\r\nGET / HTTP/1.1\r\n\r\n" |  nc 127.0.0.1 8080

HTTP/1.1 404 Not Found
date: Fri, 20 Sep 2024 16:52:50 GMT
server: BuggyServer
connection: keep-alive

HTTP/1.1 200 OK
content-length: 684
connection: keep-alive
date: Fri, 20 Sep 2024 16:52:50 GMT
server: BuggyServer
content-type: text/html; charset=utf-8

<!DOCTYPE html>
<html lang="en">
        . . . 
</html>
```

> [!WARNING]
> Although pipelined requests can improve performance by reducing latency, they are not widely used ([is not activated by default in modern browsers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Connection_management_in_HTTP_1.x#http_pipelining))  
> or recommended due to potential issues with head-of-line blocking and compatibility with some intermediaries and clients.   



## :hourglass: Functionalities: Not Yet Implemented

### Level based logging

Currently there are two type of log: 
- error log, that is displayed every time an exception occurs while replying to a request.
- log in the format [ `client IP`, `method`, `path`, `status code sent`] that indicates the responses the server sends.    

All logs are only printed to STDOUT.
