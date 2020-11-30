// +build go1.10, !go1.11

package os

// From go1.11
const ModeIrregular FileMode = 1 << (32 - 1 - 12)
