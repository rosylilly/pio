# pio - io with progress

[![Go Reference](https://pkg.go.dev/badge/github.com/rosylilly/pio.svg)](https://pkg.go.dev/github.com/rosylilly/pio)
[![test](https://github.com/rosylilly/pio/actions/workflows/test.yml/badge.svg)](https://github.com/rosylilly/pio/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/rosylilly/pio/branch/main/graph/badge.svg?token=kMUjFmnxRY)](https://codecov.io/gh/rosylilly/pio)

## Example

```golang
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
```
