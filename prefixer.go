// Copyright 2015 The Vanadium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"

	"v.io/x/lib/textutil"
)

// prefixer is a io.WriteCloser that adds given prefix to each line of the output.
type prefixer struct {
	textutil.WriteFlusher
}

// newPrefixer returns a new prefixer that uses the given prefix.
func newPrefixer(writer io.Writer, prefix string) *prefixer {
	return &prefixer{
		WriteFlusher: textutil.PrefixLineWriter(writer, fmt.Sprintf("[%v]\t", prefix)),
	}
}

// Close flushes the remaining buffer content with prefix, and closes the underlying writer if applicable.
// This internally calls Flush() on the underlying textutil.WriteFlusher.
func (p *prefixer) Close() error {
	return p.WriteFlusher.Flush()
}
