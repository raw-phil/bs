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


Free time project for learning [GO](https://go.dev/) language and something more about HTTP/1.1 protocol. 

BuggyServer is a minimal HTTP/1.1 server build from scratch that serves static files, it does not use any HTTP packages like [net/http](https://pkg.go.dev/net/http).

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
- [Functionalities: Not Yet Implemented](#hourglass-functionalities-not-yet-implemented)
  - [Connection reuse](#connection-reuse)
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
        Maximum duration in seconds server has for reading the entire request from the underling connection.
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
connection: close

<!DOCTYPE html>
<html lang="en">
        . . . 
</html>
```

### HEAD
Same as GET, but does not send the response content, only the headers are sent in response.
```bash
$ curl -I 127.0.0.1:8080/

HTTP/1.1 200 OK
date: Tue, 09 Apr 2024 10:35:37 GMT
server: BuggyServer
content-type: text/html; charset=utf-8
content-length: 697
connection: close
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
connection: close
```

### Request and Response Timeout

BuggyServer uses two fields to implement timeouts:

- `ReadTimeout` set the maximum duration in seconds for reading the entire   
request from the underling connection. If it is exceeded server responds with 408 code.    
Zero or negative value means that there will be no timeout.

- `WriteTimeout` set the maximum duration in seconds that the server has to respond.   
If it is exceeded server responds with 500 code.   
Zero or negative value means that there will be no timeout.

### Reqest size limit

BuggyServer use the `maxRequestMiB` field to set the maximum size of request the server will accept in MiB.    
It indicates the amounts of MiB that could be read from the underling connection for each request.    
Zero or negative value means there will be no maximum request size.
 

## :hourglass: Functionalities: Not Yet Implemented

### Connection reuse
Actually BuggyServer does not support reuse of connection, after each response the server closes the connection and sends
`connection: close` header in response.


### Level based logging
Actually there are two type of log: 
- error log, that are displayed every time that an exception occur in the process of replying to a request.
- log in the format [ `client IP`, `method`, `path`, `status code sent`] that indicates the responses that the server sends.    

Actually all logs are only printed to STDOUT.




