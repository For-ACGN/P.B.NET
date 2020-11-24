package plugin

import (
	"context"

	"github.com/pkg/errors"

	"project/internal/module"
)

// supported modes.
const (
	ModeAnko  = "anko"  // support go 1.10
	ModeYaegi = "yaegi" // support go 1.13
)

// New is used to create a new plugin from script.
// external include role functions like Sender.Send(),
// the script can use external to call Role self function.
// [warning]: script string will covered after call.
func New(ctx context.Context, external interface{}, mode, script string) (module.Module, error) {
	switch mode {
	case ModeAnko:
		return NewAnko(ctx, external, script)
	case ModeYaegi:
		return nil, nil
	default:
		return nil, errors.New("unsupported mode: " + mode)
	}
}
