// Copyright 2017 Josef Fröhle. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package context

import (
	"io"
	"sync"

	"github.com/google/brotli/tree/master/go/cbrotli"
)

// brotliCompressionPool is a wrapper of sync.Pool, to initialize a new compression writer pool
type brotliCompressionPool struct {
	sync.Pool
	Quality int
}

//  +------------------------------------------------------------+
//  |Brotli raw io.writer, our brotli response writer will use that. |
//  +------------------------------------------------------------+

// default writer pool with Compressor's level setted to 1
var brotliPool = &brotliCompressionPool{Quality: 5}

// acquireGzipWriter prepares a brotli writer and returns it.
//
// see releaseGzipWriter too.
func acquireBrotliWriter(w io.Writer) *cbrotli.Writer {
	v := brotliPool.Get()
	if v == nil {
		brotliWriter, err := cbrotli.NewWriterLevel(w, brotliPool.Quality)
		if err != nil {
			return nil
		}
		return brotliWriter
	}
	brotliWriter := v.(*cbrotli.Writer)
	brotliWriter.Reset(w)
	return brotliWriter
}

// releaseGzipWriter called when flush/close and put the gzip writer back to the pool.
//
// see acquireGzipWriter too.
func releaseBrotliWriter(brotliWriter *cbrotli.Writer) {
	brotliWriter.Close()
	brotliPool.Put(brotliWriter)
}

// writeGzip writes a compressed form of p to the underlying io.Writer. The
// compressed bytes are not necessarily flushed until the Writer is closed.
func writeBrotli(w io.Writer, b []byte) (int, error) {
	brotliWriter := acquireBrotliWriter(w)
	n, err := brotliWriter.Write(b)
	releaseBrotliWriter(brotliWriter)
	return n, err
}

var brpool = sync.Pool{New: func() interface{} { return &BrotliResponseWriter{} }}

// AcquireGzipResponseWriter returns a new *BrotliResponseWriter from the pool.
// Releasing is done automatically when request and response is done.
func AcquireBrotliResponseWriter() *BrotliResponseWriter {
	w := brpool.Get().(*BrotliResponseWriter)
	return w
}

func releaseBrotliResponseWriter(w *BrotliResponseWriter) {
	releaseBrotliWriter(w.gzipWriter)
	brpool.Put(w)
}

// BrotliResponseWriter is an upgraded response writer which writes compressed data to the underline ResponseWriter.
//
// It's a separate response writer because Siris gives you the ability to "fallback" and "roll-back" the gzip encoding if something
// went wrong with the response, and write http errors in plain form instead.
type BrotliResponseWriter struct {
	ResponseWriter
	gzipWriter *gzip.Writer
	chunks     []byte
	disabled   bool
}

var _ ResponseWriter = &BrotliResponseWriter{}

// BeginGzipResponse accepts a ResponseWriter
// and prepares the new gzip response writer.
// It's being called per-handler, when caller decide
// to change the response writer type.
func (w *BrotliResponseWriter) BeginGzipResponse(underline ResponseWriter) {
	w.ResponseWriter = underline
	w.gzipWriter = acquireGzipWriter(w.ResponseWriter)
	w.chunks = w.chunks[0:0]
	w.disabled = false
}

// EndResponse called right before the contents of this
// response writer are flushed to the client.
func (w *BrotliResponseWriter) EndResponse() {
	releaseGzipResponseWriter(w)
	w.ResponseWriter.EndResponse()
}

// Write compresses and writes that data to the underline response writer
func (w *BrotliResponseWriter) Write(contents []byte) (int, error) {
	// save the contents to serve them (only gzip data here)
	w.chunks = append(w.chunks, contents...)
	return len(w.chunks), nil
}

// FlushResponse validates the response headers in order to be compatible with the gzip written data
// and writes the data to the underline ResponseWriter.
func (w *BrotliResponseWriter) FlushResponse() {
	if w.disabled {
		w.ResponseWriter.Write(w.chunks)
		// remove gzip headers: no need, we just add two of them if gzip was enabled, below
		// headers := w.ResponseWriter.Header()
		// headers[contentType] = nil
		// headers["X-Content-Type-Options"] = nil
		// headers[varyHeader] = nil
		// headers[contentEncodingHeader] = nil
		// headers[contentLength] = nil
	} else {
		// if it's not disable write all chunks gzip compressed with the correct response headers.
		w.ResponseWriter.Header().Add(varyHeaderKey, "Accept-Encoding")
		w.ResponseWriter.Header().Set(contentEncodingHeaderKey, "br")
		w.gzipWriter.Write(w.chunks) // it writes to the underline ResponseWriter.
	}
	w.ResponseWriter.FlushResponse()
}

// ResetBody resets the response body.
func (w *BrotliResponseWriter) ResetBody() {
	w.chunks = w.chunks[0:0]
}

// Disable turns off the gzip compression for the next .Write's data,
// if called then the contents are being written in plain form.
func (w *BrotliResponseWriter) Disable() {
	w.disabled = true
}
