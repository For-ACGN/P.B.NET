package namer

import (
	"archive/zip"
	"bufio"
	"bytes"
	"sync"

	"github.com/pkg/errors"

	"project/internal/security"
)

// Namer is used to generate a random name from dictionary.
type Namer interface {
	// Load is used to load resource about namer.
	Load(res []byte) error

	// Generate is used to generate a random word.
	Generate(opts *Options) (string, error)

	// Type is used to get the namer type.
	Type() string
}

// Options contains options about all namer.
type Options struct {
	DisablePrefix bool `toml:"disable_prefix"`
	DisableStem   bool `toml:"disable_stem"`
	DisableSuffix bool `toml:"disable_suffix"`
}

// registered new functions
var newNamers = map[string]func() Namer{
	"english": func() Namer { return NewEnglish() },
}
var newNamersRWM sync.RWMutex

// Register is used to register a new namer function.
func Register(typ string, fn func() Namer) error {
	newNamersRWM.Lock()
	defer newNamersRWM.Unlock()
	if _, ok := newNamers[typ]; ok {
		return errors.Errorf("namer \"%s\" is already registered", typ)
	}
	newNamers[typ] = fn
	return nil
}

// Unregister is used to unregister a new namer function.
func Unregister(typ string) {
	newNamersRWM.Lock()
	defer newNamersRWM.Unlock()
	delete(newNamers, typ)
}

// Load is used to load resource and create a namer.
func Load(typ string, res []byte) (Namer, error) {
	namer, err := newNamer(typ)
	if err != nil {
		return nil, err
	}
	err = namer.Load(res)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to load namer \"%s\"", typ)
	}
	return namer, nil
}

func newNamer(typ string) (Namer, error) {
	newNamersRWM.RLock()
	defer newNamersRWM.RUnlock()
	nn, ok := newNamers[typ]
	if !ok {
		return nil, errors.Errorf("namer \"%s\" is not registered", typ)
	}
	return nn(), nil
}

func loadWordsFromZipFile(file *zip.File) (*security.Bytes, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer func() { _ = rc.Close() }()
	data, err := security.ReadAll(rc, 16*1024)
	if err != nil {
		return nil, errors.Errorf("%s, maybe zip file is larger than 16KB", err)
	}
	defer security.CoverBytes(data)
	return security.NewBytes(data), nil
}

func loadWordsFromSecurityBytes(sb *security.Bytes) map[string]struct{} {
	data := sb.Get()
	defer sb.Put(data)
	words := make(map[string]struct{}, 256)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		word := scanner.Text()
		if word != "" {
			words[word] = struct{}{}
		}
	}
	return words
}
