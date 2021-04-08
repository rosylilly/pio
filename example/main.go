package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/rosylilly/pio"
)

func main() {
	file, err := os.Open("README.md")
	if err != nil {
		log.Fatalln(err)
	}

	logger := log.New(os.Stdout, "", log.Ltime)
	reader, err := pio.NewReader(file, pio.WithLogger(logger))
	if err != nil {
		log.Fatalln(err)
	}

	buffer := bytes.NewBuffer([]byte{})
	for {
		_, err := io.CopyN(buffer, reader, 100)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				log.Fatalln(err)
			}
		}
	}

	fmt.Println(buffer.String())
}
