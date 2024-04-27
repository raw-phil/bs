package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raw-phil/bs/buggy_http"
)

var (
	port          = flag.Uint("p", 8080, "Sets the port")
	host          = flag.String("h", "0.0.0.0", "Sets the host")
	directory     = flag.String("d", "./", "Directory from which files are served")
	noBanner      = flag.Bool("no-banner", false, "Suppress the initial banner")
	readTimeout   = flag.Int("read-timeout", -1, "Maximum duration in seconds server has for reading the entire request from the underling connection.\nZero or negative value means there will be no timeout.")
	writeTimeout  = flag.Int("write-timeout", -1, "Maximum duration in seconds the server has to respond.\nZero or negative value means there will be no timeout.")
	maxRequestMiB = flag.Int("max-request-size", -1, "Maximum size of request the server will accept in MiB.\nZero or negative value means there will be no maximum size.")
)

func main() {

	flag.Parse()

	if !*noBanner {
		fmt.Print(`
	 ______                           _____
	| ___ \                         /  ___|
	| |_/ /_   _  __ _  __ _ _   _  \ ` + "`" + `--.  ___ _ ____   _____ _ __
	| ___ \ | | |/ _` + "`" + ` |/ _` + "`" + ` | | | |  ` + "`" + `--. \/ _ \ '__\ \ / / _ \ '__|
	| |_/ / |_| | (_| | (_| | |_| | /\__/ /  __/ |   \ V /  __/ |
	\____/ \__,_|\__, |\__, |\__, | \____/ \___|_|    \_/ \___|_|
	              __/ | __/ | __/ |
	             |___/ |___/ |___/` + "\n\n")
	}

	bs := buggy_http.NewBuggyServer()

	if err := bs.SetBaseDir(*directory); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	if err := bs.SetReadTimeout(time.Duration(*readTimeout)); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	if err := bs.SetWriteTimeout(time.Duration(*writeTimeout)); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	if err := bs.SetmaxRequestMiB(*maxRequestMiB); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	if err := bs.StartBuggyServer(*host, *port); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Signal: %s received\n", <-c)
	if err := bs.StopBuggyServer(); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}

}
