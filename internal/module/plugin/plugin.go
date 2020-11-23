package plugin

import (
	"github.com/pkg/errors"

	"project/internal/module"
)

// supported modes.
const (
	ModeAnko  = "anko"  // support go 1.10
	ModeYaegi = "yaegi" // support go 1.13
)

// New is used to create a new plugin.
// [warning]: script string will covered after call.
func New(mode string, script string) (module.Module, error) {
	switch mode {
	case ModeAnko:
		return nil, nil
	case ModeYaegi:
		return nil, nil
	default:
		return nil, errors.New("unsupported mode: " + mode)
	}
}
