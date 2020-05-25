package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base32"
	"flag"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/crypto/sha3"
)

var (
	addressVersion = []byte{0x03}
)

func worker(ctx context.Context, re *regexp.Regexp, wg *sync.WaitGroup, addressCh chan<- string) {
	defer wg.Done()

	var (
		checksumBuf bytes.Buffer
		addressBuf  bytes.Buffer
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		pub, _, err := ed25519.GenerateKey(nil)
		if err != nil {
			log.Println(fmt.Errorf("failed to generate key: %w", err))
			break
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

	ctx, cancel := context.WithCancel(context.Background())
	addressCh := make(chan string, workers)

	wg := sync.WaitGroup{}
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go worker(ctx, re, &wg, addressCh)
	}

	log.Println(<-addressCh)
	cancel()
	wg.Wait()
	close(addressCh)

	log.Println("done")
}
