package plugin

import (
	"context"

	"project/external/anko/ast"
	"project/internal/module"

	"project/internal/anko"
)

type Anko struct {
	stmt ast.Stmt // store parsed anko script

	ctx    context.Context
	cancel context.CancelFunc
}

func NewAnko(script string) (module.Module, error) {
	anko.NewEnv()

	return nil, nil
}

func (m *Anko) Start() error {
	return nil
}

func (m *Anko) Stop() {

}

func (m *Anko) Restart() error {
	return nil
}

func (m *Anko) Name() string {
	return ""
}

func (m *Anko) Info() string {
	return ""
}

func (m *Anko) Status() string {
	return ""
}
