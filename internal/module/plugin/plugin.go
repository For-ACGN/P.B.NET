package plugin

import (
	"io"
	"time"

	"github.com/pkg/errors"

	"project/internal/module"
)

// supported modes.
const (
	ModeAnko  = "anko"  // support go 1.10
	ModeYaegi = "yaegi" // support go 1.12
)

const operationTimeout = 30 * time.Second

// New is used to create a new plugin from script.
// external include role functions like Sender.Send(),
// the script can use external to call Role self function.
// [warning]: script string will covered after call.
func New(external interface{}, output io.Writer, mode, script string) (module.Module, error) {
	switch mode {
	case ModeAnko:
		return NewAnko(external, output, script)
	case ModeYaegi:
		return nil, errors.New("not implemented")
	default:
		return nil, errors.New("unsupported mode: " + mode)
	}
}
