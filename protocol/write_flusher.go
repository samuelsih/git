package protocol

import (
	"io"
	"net/http"
)

type WriteFlusher interface {
	io.Writer
	http.Flusher
}

type writeFlusher struct {
	wf WriteFlusher
}

func newWriteFlusher(w http.ResponseWriter) io.Writer {
	return writeFlusher{w.(WriteFlusher)}
}

func (w writeFlusher) Write(p []byte) (int, error) {
	defer w.wf.Flush()
	return w.wf.Write(p)
}
