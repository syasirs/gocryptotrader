package log

import (
	"errors"
	"io"
)

var (
	errWriterAlreadyLoaded = errors.New("io.Writer already loaded")
	errWriterNotFound      = errors.New("io.Writer not found")
)

// Add appends a new writer to the multiwriter slice
func (mw *multiWriter) Add(writer io.Writer) error {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	for i := range mw.writers {
		if mw.writers[i] == writer {
			return errWriterAlreadyLoaded
		}
	}
	mw.writers = append(mw.writers, writer)
	return nil
}

// Remove removes existing writer from multiwriter slice
func (mw *multiWriter) Remove(writer io.Writer) error {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	for i := range mw.writers {
		if mw.writers[i] != writer {
			continue
		}
		mw.writers[i] = mw.writers[len(mw.writers)-1]
		mw.writers[len(mw.writers)-1] = nil
		mw.writers = mw.writers[:len(mw.writers)-1]
		return nil
	}
	return errWriterNotFound
}

// Write concurrent safe Write for each writer
func (mw *multiWriter) Write(p []byte) (n int, err error) {
	type data struct {
		n   int
		err error
	}
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	results := make(chan data, len(mw.writers))
	for _, wr := range mw.writers {
		go func(w io.Writer, p []byte, ch chan data) {
			n, err = w.Write(p)
			if err != nil {
				ch <- data{n, err}
				return
			}
			if n != len(p) {
				ch <- data{n, io.ErrShortWrite}
				return
			}
			ch <- data{n, nil}
		}(wr, p, results)
	}

	for range mw.writers {
		// NOTE: These results do not necessarily reflect the current io.writer
		d := <-results
		if d.err != nil {
			return d.n, d.err
		}
	}
	return len(p), nil
}

// MultiWriter make and return a new copy of multiWriter
func MultiWriter(writers ...io.Writer) (*multiWriter, error) {
	mw := &multiWriter{}
	for x := range writers {
		err := mw.Add(writers[x])
		if err != nil {
			return nil, err
		}
	}
	return mw, nil
}
