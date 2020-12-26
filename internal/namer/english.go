package namer

import (
	"archive/zip"
	"bytes"
	"sync"

	"github.com/pkg/errors"

	"project/internal/random"
	"project/internal/security"
)

// English is a english word generator, it implement Namer interface.
type English struct {
	rand *random.Rand

	prefix *security.Bytes
	stem   *security.Bytes
	suffix *security.Bytes
	rwm    sync.RWMutex
}

// NewEnglish is used to create a english word generator.
func NewEnglish() *English {
	return &English{rand: random.NewRand()}
}

// Load is used to load prefix.txt, stem.txt and suffix.txt from a zip file.
func (eng *English) Load(res []byte) error {
	reader, err := zip.NewReader(bytes.NewReader(res), int64(len(res)))
	if err != nil {
		return errors.WithStack(err)
	}
	eng.rwm.Lock()
	defer eng.rwm.Unlock()
	items := map[string]**security.Bytes{
		"prefix.txt": &eng.prefix,
		"stem.txt":   &eng.stem,
		"suffix.txt": &eng.suffix,
	}
	for _, file := range reader.File {
		item := items[file.Name]
		if item == nil {
			continue
		}
		*item, err = loadWordsFromZipFile(file)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return eng.checkWordNumber()
}

// Generate is used to generate a word.
func (eng *English) Generate(opts *Options) (string, error) {
	if opts == nil {
		opts = new(Options)
	}

	eng.rwm.RLock()
	defer eng.rwm.RUnlock()

	err := eng.checkWordNumber()
	if err != nil {
		return "", err
	}

	var (
		prefix string
		stem   string
		suffix string
	)
	// select prefix
	if !opts.DisablePrefix {
		// map is include random, but we use random.Rand.
		times := 10 + eng.rand.Int(10)
		counter := 0
		prefixes := loadWordsFromSecurityBytes(eng.prefix)

		for prefix = range prefixes {
			if counter > times {
				break
			}
			counter++
		}
	}
	// select stem
	if !opts.DisableStem {
		// map is include random, but we use random.Rand.
		times := 10 + eng.rand.Int(10)
		counter := 0
		stems := loadWordsFromSecurityBytes(eng.stem)

		for stem = range stems {
			if counter > times {
				break
			}
			counter++
		}
	}
	// select suffix
	if !opts.DisableSuffix {
		// map is include random, but we use random.Rand.
		times := 10 + eng.rand.Int(10)
		counter := 0
		suffixes := loadWordsFromSecurityBytes(eng.suffix)

		for suffix = range suffixes {
			if counter > times {
				break
			}
			counter++
		}
	}
	// merge
	word := prefix + stem + suffix
	if word == "" {
		return "", errors.New("generated a empty word")
	}
	return word, nil
}

func (eng *English) checkWordNumber() error {
	for _, item := range [...]*struct {
		sb   *security.Bytes
		name string
	}{
		{eng.prefix, "prefix"},
		{eng.stem, "stem"},
		{eng.suffix, "suffix"},
	} {
		err := func() error {
			if item.sb == nil {
				return errors.New("not load " + item.name)
			}
			data := item.sb.Get()
			defer item.sb.Put(data)
			if len(data) == 0 {
				return errors.New("empty " + item.name)
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

// Type is used to return the namer type.
func (eng *English) Type() string {
	return "english"
}
