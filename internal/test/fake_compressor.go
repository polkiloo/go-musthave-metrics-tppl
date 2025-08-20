package test

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

type FakeCompressor struct {
	Encoding string

	ErrNewWriter error
	ErrNewReader error

	CaptureWrites bool

	mu       sync.Mutex
	written  bytes.Buffer
	closeWEr error
	closeREr error
}

func NewFakeCompressor(enc string) *FakeCompressor {
	if enc == "" {
		enc = "gzip"
	}
	return &FakeCompressor{Encoding: enc}
}

func (f *FakeCompressor) ContentEncoding() string { return f.Encoding }

func (f *FakeCompressor) NewWriter(w io.Writer) (io.WriteCloser, error) {
	if f.ErrNewWriter != nil {
		return nil, f.ErrNewWriter
	}
	return &fakeWriteCloser{
		dst:           w,
		parent:        f,
		captureWrites: f.CaptureWrites,
	}, nil
}

func (f *FakeCompressor) NewReader(r io.Reader) (io.ReadCloser, error) {
	if f.ErrNewReader != nil {
		return nil, f.ErrNewReader
	}
	return &fakeReadCloser{src: r, closeErr: &f.closeREr}, nil
}

func (f *FakeCompressor) Written() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]byte(nil), f.written.Bytes()...)
}

func (f *FakeCompressor) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.written.Reset()
}

type fakeWriteCloser struct {
	dst           io.Writer
	parent        *FakeCompressor
	captureWrites bool
	closed        bool
}

func (w *fakeWriteCloser) Write(p []byte) (int, error) {
	if w.captureWrites {
		w.parent.mu.Lock()
		w.parent.written.Write(p)
		w.parent.mu.Unlock()
	}
	return w.dst.Write(p)
}

func (w *fakeWriteCloser) Close() error {
	if w.closed {
		return errors.New("fakeWriteCloser: double close")
	}
	w.closed = true
	if w.parent.closeWEr != nil {
		return w.parent.closeWEr
	}
	return nil
}

type fakeReadCloser struct {
	src      io.Reader
	closeErr *error
}

func (r *fakeReadCloser) Read(p []byte) (int, error) { return r.src.Read(p) }
func (r *fakeReadCloser) Close() error {
	if r.closeErr != nil && *r.closeErr != nil {
		return *r.closeErr
	}
	return nil
}
