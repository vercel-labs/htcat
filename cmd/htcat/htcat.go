package main

import (
	"crypto/tls"
	"flag"
	"github.com/htcat/htcat"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
)

const version = "1.0.2"

var onlyPrintVersion = flag.Bool("version", false, "print the htcat version")
var outputFile = flag.String("o", "", "write output to file instead of stdout")

const (
	_        = iota
	KB int64 = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

func printUsage() {
	log.Printf("usage: %v [-o output_file] URL", os.Args[0])
}

func main() {
	flag.Parse()
	args := flag.Args()

	if *onlyPrintVersion {
		os.Stdout.Write([]byte(version + "\n"))
		os.Exit(0)
	}

	if len(args) != 1 {
		printUsage()
		log.Fatalf("aborting: incorrect usage")
	}

	u, err := url.Parse(args[0])
	if err != nil {
		log.Fatalf("aborting: could not parse given URL: %v", err)
	}

	client := *http.DefaultClient

	// Only support HTTP and HTTPS schemes
	switch u.Scheme {
	case "https":
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{},
		}
	case "http":
	default:
		// This error path can be hit with common alphanumeric
		// lexemes like "help", which also parse as URLs,
		// which makes this error message somewhat
		// incomprehensible.  Try to help out a user by
		// printing the usage here.
		printUsage()
		log.Fatalf("aborting: unsupported URL scheme %v", u.Scheme)
	}

	// On fast links (~= saturating gigabit), parallel execution
	// gives a large speedup.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Begin the GET.
	htc := htcat.New(&client, u, 5)

	// Determine output destination
	var output io.Writer = os.Stdout
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			log.Fatalf("aborting: could not create output file %v: %v",
				*outputFile, err)
		}
		defer file.Close()
		output = file
	}

	if _, err := htc.WriteTo(output); err != nil {
		log.Fatalf("aborting: could not write to output stream: %v",
			err)
	}
}
