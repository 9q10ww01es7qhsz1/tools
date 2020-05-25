package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base32"
	"flag"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"

	"golang.org/x/crypto/sha3"
)

var (
	addressVersion = []byte{0x03}
)

func worker(re *regexp.Regexp, closeCh <-chan interface{}, addressCh chan<- string) {
	var (
		checksumBuf bytes.Buffer
		addressBuf  bytes.Buffer
	)

	for {
		select {
		case <-closeCh:
			return
		default:
		}

		pub, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			log.Println(fmt.Errorf("failed to generate key: %w", err))
			return
		}

		checksumBuf.Reset()
		checksumBuf.Write([]byte(".onion checksum"))
		checksumBuf.Write(pub)
		checksumBuf.Write(addressVersion)

		checksum := sha3.Sum256(checksumBuf.Bytes())

		addressBuf.Reset()
		addressBuf.Write(pub)
		addressBuf.Write(checksum[:2])
		addressBuf.Write(addressVersion)

		address := base32.StdEncoding.EncodeToString(addressBuf.Bytes())
		address = strings.ToLower(address) + ".onion"

		if re.MatchString(address) {
			addressCh <- address
			break
		}
	}
}

func main() {
	var (
		addrRegexp string
		workers    int
	)

	flag.StringVar(&addrRegexp, "regexp", "", "address regexp")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "workers number")

	flag.Parse()

	if workers < 1 {
		log.Fatalln(fmt.Errorf("unexpected workers value: %d", workers))
	}

	re, err := regexp.Compile(addrRegexp)
	if err != nil {
		log.Fatalln(fmt.Errorf("failed to compile address regexp"))
	}

	closeCh := make(chan interface{})
	addressCh := make(chan string)

	for i := 0; i < workers; i++ {
		go worker(re, closeCh, addressCh)
	}

	log.Println(<-addressCh)
	close(closeCh)
	close(addressCh)

	log.Println("done")
}
