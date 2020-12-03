// +build go1.10,!go1.16

package net

import (
	"internal/poll"
)

// errClosed exists just so that the docs for ErrClosed don't mention
// the internal package poll.
var errClosed = poll.ErrNetClosing

// ErrClosed is the error returned by an I/O call on a network
// connection that has already been closed, or that is closed by
// another goroutine before the I/O is completed. This may be wrapped
// in another error, and should normally be tested using
// errors.Is(err, net.ErrClosed).
var ErrClosed = errClosed
