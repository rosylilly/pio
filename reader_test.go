package pio_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/rosylilly/pio"
	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	srcString := "Hello, World"
	src := bytes.NewBufferString(srcString)
	reader, err := pio.NewReader(src)
	assert.NoError(t, err)
	defer reader.Close()

	count := 0
	for {
		buf := make([]byte, 1) // read each bytes
		_, err := reader.Read(buf)
		if err != nil {
			assert.ErrorIs(t, err, io.EOF)
			break
		}
		assert.Equal(t, srcString[count], buf[0])
		count++
	}
	assert.Equal(t, len(srcString), count)
}

func TestReaderOption(t *testing.T) {
	errOnOption := fmt.Errorf("raise error test")
	reader, err := pio.NewReader(nil, func(r *pio.Reader) error {
		return errOnOption
	})
	assert.ErrorIs(t, err, errOnOption)
	defer reader.Close()
}

func TestFileReader(t *testing.T) {
	srcString := "Hello, World"

	filesystem := fstest.MapFS{
		"test": &fstest.MapFile{
			Data: []byte(srcString),
			Mode: 0666,
		},
	}

	fp, err := filesystem.Open("test")
	assert.NoError(t, err)

	reader, err := pio.NewReader(fp, pio.WithLogger(log.Default()))
	assert.NoError(t, err)
	defer reader.Close()

	count := 0
	for {
		buf := make([]byte, 1) // read each bytes
		_, err := reader.Read(buf)
		if err != nil {
			assert.ErrorIs(t, err, io.EOF)
			break
		}
		assert.Equal(t, srcString[count], buf[0])
		count++
	}
	assert.Equal(t, len(srcString), count)
}

func TestAsyncProgressReader(t *testing.T) {
	wg := sync.WaitGroup{}
	srcString := "Hello, World"
	src := bytes.NewBufferString(srcString)
	reader, err := pio.NewReader(src, pio.WithName("foo"), pio.WithAsync(), pio.WithFunc(func(info pio.ProgressInfo) {
		wg.Done()
	}))
	assert.NoError(t, err)
	defer reader.Close()

	count := 0
	wg.Add(1)
	for {
		buf := make([]byte, 1) // read each bytes
		wg.Add(1)
		_, err := reader.Read(buf)
		if err != nil {
			assert.ErrorIs(t, err, io.EOF)
			break
		}
		assert.Equal(t, srcString[count], buf[0])
		count++
	}
	assert.Equal(t, len(srcString), count)

	wg.Done()
	wg.Wait()
}

func TestReadAt(t *testing.T) {
	skipString := "Hello, World."
	readString := "If you catch me."
	src := bytes.NewReader(append([]byte(skipString), []byte(readString)...))
	reader, err := pio.NewReader(src, pio.WithLogger(nil))
	assert.NoError(t, err)
	defer reader.Close()

	ret := make([]byte, len(readString))
	n, err := reader.ReadAt(ret, int64(len(skipString)))
	assert.Equal(t, len(readString), n)
	assert.Equal(t, readString, string(ret))
}

func TestReadAtNotImplemented(t *testing.T) {
	skipString := "Hello, World."
	readString := "If you catch me."
	src := bytes.NewBufferString(skipString + readString)
	reader, err := pio.NewReader(src)
	assert.NoError(t, err)
	defer reader.Close()

	ret := make([]byte, len(readString))
	_, err = reader.ReadAt(ret, int64(len(skipString)))
	assert.ErrorIs(t, err, pio.ErrNotImplemented)
}

func TestSeek(t *testing.T) {
	skipString := "Catch You, "
	readString := "Catch Me"
	src := bytes.NewReader(append([]byte(skipString), []byte(readString)...))
	reader, err := pio.NewReader(src, pio.WithLogger(nil))
	assert.NoError(t, err)
	defer reader.Close()

	ret := make([]byte, len(readString))
	_, err = reader.Seek(int64(len(skipString)), io.SeekStart)
	assert.NoError(t, err)
	n, err := reader.Read(ret)
	assert.Equal(t, len(readString), n)
	assert.Equal(t, readString, string(ret))
}

func TestSeekNotImplemented(t *testing.T) {
	skipString := "Hello, World."
	readString := "If you catch me."
	src := bytes.NewBufferString(skipString + readString)
	reader, err := pio.NewReader(src)
	assert.NoError(t, err)
	defer reader.Close()

	_, err = reader.Seek(int64(len(skipString)), io.SeekStart)
	assert.ErrorIs(t, err, pio.ErrNotImplemented)
}

func TestReadByte(t *testing.T) {
	srcString := "Hello, World."
	src := bytes.NewBufferString(srcString)
	reader, err := pio.NewReader(src, pio.WithName("ReadByte"), pio.WithLogger(nil))
	assert.NoError(t, err)
	defer reader.Close()

	for i := 0; i < len(srcString); i++ {
		chr, err := reader.ReadByte()
		assert.NoError(t, err)
		assert.Equal(t, srcString[i], chr)
	}
}

type StringReader struct {
	source string
	i      int
}

func (s *StringReader) Read(b []byte) (int, error) {
	n := copy(b, []byte(s.source)[s.i:])
	s.i += n
	return n, nil
}

func TestReadByteNotImplemented(t *testing.T) {
	srcString := "Hello, World."
	src := &StringReader{source: srcString}
	reader, err := pio.NewReader(src)
	assert.NoError(t, err)
	defer reader.Close()

	for i := 0; i < len(srcString); i++ {
		chr, err := reader.ReadByte()
		assert.NoError(t, err)
		assert.Equal(t, srcString[i], chr)
	}
}
