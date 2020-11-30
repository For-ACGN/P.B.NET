// +build go1.10 !go1.11

package log

import "io"

// Writer returns the output destination for the logger.
func (l *Logger) Writer() io.Writer {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out
}
