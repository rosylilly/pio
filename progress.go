package pio

import (
	"log"
	"math"
)

type ProgressInfo struct {
	Name     string
	Read     int64
	Size     int64
	Progress float64
}

type ProgressFunc func(info ProgressInfo)

var voidProgressFunc = func(info ProgressInfo) {}

type Logger interface {
	Printf(format string, a ...interface{})
}

func WithLogger(logger Logger) ReaderOption {
	if logger == nil {
		logger = log.Default()
	}
	return func(r *Reader) error {
		r.OnProgress = func(info ProgressInfo) {
			width := 1
			if info.Size > 0 {
				width = int(1 + math.Log10(float64(info.Size)))
			}
			logger.Printf("%s: %06.2f(%*d / %*d)\n", info.Name, info.Progress, width, info.Read, width, info.Size)
		}
		return nil
	}
}

func WithFunc(f ProgressFunc) ReaderOption {
	return func(r *Reader) error {
		r.OnProgress = f
		return nil
	}
}
