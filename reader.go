package pio

import (
	"fmt"
	"io"
	"sync/atomic"
)

var ErrNotImplemented = fmt.Errorf("Not implemented")

type ReaderOption func(r *Reader) error

type Reader struct {
	src     io.Reader
	read    *int64
	size    *int64
	asyncCh chan ProgressInfo

	Name          string
	OnProgress    ProgressFunc
	AsyncProgress bool
}

func NewReader(src io.Reader, options ...ReaderOption) (*Reader, error) {
	read := int64(0)
	size := int64(0)

	pr := &Reader{
		src:           src,
		read:          &read,
		size:          &size,
		Name:          fmt.Sprintf("%p", src),
		OnProgress:    voidProgressFunc,
		AsyncProgress: false,
	}

	switch src := src.(type) {
	case fileLike:
		fi, err := src.Stat()
		if err != nil {
			return nil, err
		}
		size := fi.Size()
		pr.Name = fi.Name()
		pr.size = &size
	case bufferLike:
		size := int64(src.Len())
		pr.size = &size
	}

	for _, option := range options {
		if err := option(pr); err != nil {
			return pr, err
		}
	}

	if pr.AsyncProgress {
		pr.asyncCh = make(chan ProgressInfo, 10)
		go func(pr *Reader) {
			for {
				info, ok := <-pr.asyncCh
				if !ok {
					return
				}
				pr.OnProgress(info)
			}
		}(pr)
	}

	return pr, nil
}

func WithName(name string) ReaderOption {
	return func(r *Reader) error {
		r.Name = name
		return nil
	}
}

func WithAsync() ReaderOption {
	return func(r *Reader) error {
		r.AsyncProgress = true
		return nil
	}
}

// Read implements the io.Reader interface.
func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.src.Read(p)
	read := atomic.AddInt64(r.read, int64(n))
	r.progress(read)

	return n, err
}

// ReadAt implements the io.ReaderAt interface.
func (r *Reader) ReadAt(b []byte, offset int64) (n int, err error) {
	if ra, ok := r.src.(io.ReaderAt); ok {
		n, err = ra.ReadAt(b, offset)
		read := offset + int64(n)
		atomic.StoreInt64(r.read, read)
		r.progress(read)
	} else {
		err = fmt.Errorf("ReadAt: %w", ErrNotImplemented)
	}
	return
}

// Seek implements the io.Seeker interface.
func (r *Reader) Seek(offset int64, whence int) (ret int64, err error) {
	if seeker, ok := r.src.(io.Seeker); ok {
		ret, err = seeker.Seek(offset, whence)
		atomic.StoreInt64(r.read, ret)
	} else {
		err = fmt.Errorf("Seek: %w", ErrNotImplemented)
	}
	return
}

// ReadByte implements the io.ByteReader interface.
func (r *Reader) ReadByte() (byte, error) {
	if br, ok := r.src.(io.ByteReader); ok {
		b, err := br.ReadByte()
		read := atomic.AddInt64(r.read, 1)
		r.progress(read)
		return b, err
	} else {
		buf := make([]byte, 1)
		_, err := r.Read(buf)
		return buf[0], err
	}
}

// Close implements the io.Closer interface.
func (r *Reader) Close() error {
	if r.asyncCh != nil {
		close(r.asyncCh)
	}

	if closer, ok := r.src.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (r *Reader) progress(read int64) {
	size := atomic.LoadInt64(r.size)
	progress := float64(0)
	if size > 0 {
		progress = float64(read) / float64(size) * 100
	}
	info := ProgressInfo{
		Name:     r.Name,
		Read:     read,
		Size:     size,
		Progress: progress,
	}
	if r.asyncCh != nil {
		r.asyncCh <- info
	} else {
		r.OnProgress(info)
	}
}
