package testnamer

import (
	"errors"

	"project/internal/namer"
)

// errors about test namer.
var (
	ErrLoad     = errors.New("failed to load namer resource")
	ErrGenerate = errors.New("failed to generate word")
)

// testNamer implemented namer.Namer.
type testNamer struct {
	prefixes map[string]struct{}
	stems    map[string]struct{}
	suffixes map[string]struct{}

	// return error flag
	loadErr bool
	genErr  bool
}

// Namer is used to create a namer for test.
func Namer() namer.Namer {
	prefixes := make(map[string]struct{})
	stems := make(map[string]struct{})
	suffixes := make(map[string]struct{})
	for _, text := range []string{
		"dis", "in", "im", "il",
	} {
		prefixes[text] = struct{}{}
	}
	for _, text := range []string{
		"agr", "ann", "astro", "audi",
	} {
		stems[text] = struct{}{}
	}
	for _, text := range []string{
		"st", "eer", "er", "or",
	} {
		suffixes[text] = struct{}{}
	}
	return &testNamer{
		prefixes: prefixes,
		stems:    stems,
		suffixes: suffixes,
	}
}

// Load is a padding function.
func (namer *testNamer) Load([]byte) error {
	if namer.loadErr {
		return ErrLoad
	}
	return nil
}

// Generate is used to generate a random word.
func (namer *testNamer) Generate(*namer.Options) (string, error) {
	if namer.genErr {
		return "", ErrGenerate
	}
	var (
		prefix string
		stem   string
		suffix string
	)
	for prefix = range namer.prefixes {
	}
	for stem = range namer.stems {
	}
	for suffix = range namer.suffixes {
	}
	return prefix + stem + suffix, nil
}

// Type is used to return the namer type.
func (namer *testNamer) Type() string {
	return "test"
}

// WithLoadFailed is used to create a namer that will failed to load resource.
func WithLoadFailed() namer.Namer {
	return &testNamer{loadErr: true}
}

// WithGenerateFailed is used to create a namer that will failed to generate word.
func WithGenerateFailed() namer.Namer {
	return &testNamer{genErr: true}
}
