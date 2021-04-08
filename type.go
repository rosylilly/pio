package pio

import (
	"io/fs"
)

type fileLike interface {
	Stat() (fs.FileInfo, error)
}

type bufferLike interface {
	Len() int
}
