package main

import (
	"io"
)

// A Reader implements a sliding window over an io.Reader.
type Reader struct {
	Data   []byte
	Offset int
	R      io.Reader
	Err    error
}

// Release discards n bytes from the front of the window.
func (b *Reader) Release(n int) {
	b.Offset += n
}

// Window returns the current Window.
// The Window is invalidated by calls to release or extend.
func (b *Reader) Window() []byte {
	return b.Data[b.Offset:]
}

// tuning constants for byteReader.extend.
const (
	newBufferSize = 4096
	minReadSize   = newBufferSize >> 2
)

// Extend extends the window with data from the underlying reader.
func (b *Reader) Extend() int {
	if b.Err != nil {
		return 0
	}

	remaining := len(b.Data) - b.Offset
	if remaining == 0 {
		b.Data = b.Data[:0]
		b.Offset = 0
	}
	if cap(b.Data)-len(b.Data) >= minReadSize {
		// nothing to do, enough space exists between len and cap.
	} else if cap(b.Data)-remaining >= minReadSize {
		// buffer has enough space if we move the data to the front.
		b.compact()
	} else {
		// otherwise, we must allocate/extend a new buffer
		b.grow()
	}
	remaining += b.Offset
	n, err := b.R.Read(b.Data[remaining:cap(b.Data)])
	// reduce length to the existing plus the data we read.
	b.Data = b.Data[:remaining+n]
	b.Err = err
	return n
}

// grow grows the buffer, moving the active data to the front.
func (b *Reader) grow() {
	buf := make([]byte, max(cap(b.Data)*2, newBufferSize))
	copy(buf, b.Data[b.Offset:])
	b.Data = buf
	b.Offset = 0
}

// compact moves the active data to the front of the buffer.
func (b *Reader) compact() {
	copy(b.Data, b.Data[b.Offset:])
	b.Offset = 0
}
