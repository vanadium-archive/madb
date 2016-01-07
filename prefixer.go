// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"io"
	"os"
)

// prefixer is a io.WriteCloser that adds given prefix to each line of the output.
type prefixer struct {
	prefix []byte
	writer io.Writer
	buffer bytes.Buffer
}

// newPrefixer returns a new prefixer that uses the given prefix.
func newPrefixer(prefix string, writer io.Writer) *prefixer {
	return &prefixer{
		prefix: []byte("[" + prefix + "]\t"),
		writer: writer,
		buffer: bytes.Buffer{},
	}
}

func (p *prefixer) Write(b []byte) (int, error) {
	// write the bytes to the buffer.
	p.buffer.Write(b)

	// For each line in the unread buffer, add the prefix and write it to the underlying writer.
	for {
		idx := bytes.IndexByte(p.buffer.Bytes(), '\n')
		if idx == -1 {
			break
		}

		prefixed := append(p.prefix, p.buffer.Next(idx+1)...)
		if _, err := p.writer.Write(prefixed); err != nil {
			return len(b), err
		}
	}

	return len(b), nil
}

// Close flushes the remaining buffer content with prefix, and closes the underlying writer if applicable.
func (p *prefixer) Close() error {
	// Flush the remaining buffer content if any.
	// Add the prefix at the beginning, and a newline character at the end.
	if p.buffer.Len() > 0 {
		prefixed := append(p.prefix, p.buffer.Bytes()...)
		prefixed = append(prefixed, '\n')
		if _, err := p.writer.Write(prefixed); err != nil {
			return err
		}
	}

	// Close the underlying writer unless the writer is stdout or stderr.
	if p.writer != os.Stdout && p.writer != os.Stderr {
		if closer, ok := p.writer.(io.Closer); ok {
			return closer.Close()
		}
	}

	return nil
}
