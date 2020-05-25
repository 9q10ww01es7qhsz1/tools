package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base32"
	"encoding/hex"
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

type result struct {
	Address string
	Priv    ed25519.PrivateKey
}

func worker(ctx context.Context, re *regexp.Regexp, wg *sync.WaitGroup, resultCh chan<- *result) {
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

		pub, priv, err := ed25519.GenerateKey(nil)
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
			resultCh <- &result{Address: address, Priv: priv}
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
	resultCh := make(chan *result, workers)

	wg := sync.WaitGroup{}
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go worker(ctx, re, &wg, resultCh)
	}

	result := <-resultCh

	cancel()
	wg.Wait()
	close(resultCh)

	log.Println(result.Address, hex.EncodeToString(result.Priv))
}
