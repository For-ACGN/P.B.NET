// +build go1.10, !go1.13

package log

import "io"

// Writer returns the output destination for the logger.
//
// From go1.13
func (l *Logger) Writer() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}
