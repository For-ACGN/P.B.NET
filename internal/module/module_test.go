package module

import (
	"errors"
	"fmt"
)

type mockModule struct{}

func (mockModule) Name() string {
	return "mock module"
}

func (mockModule) Description() string {
	return "Mock module is used to test."
}

func (mockModule) Start() error {
	return nil
}

func (mockModule) Stop() {}

func (mockModule) Restart() error {
	return nil
}

func (mockModule) IsStarted() bool {
	return true
}

func (mockModule) Info() string {
	return "mock module information"
}

func (mockModule) Status() string {
	return "mock module status"
}

func (*mockModule) Methods() []*Method {
	scan := Method{
		Name: "Scan",
		Desc: "Scan is used to scan a host with port, it will return the port status",
		Args: []*Value{
			{Name: "host", Type: "string"},
			{Name: "port", Type: "uint16"},
		},
		Rets: []*Value{
			{Name: "open", Type: "bool"},
			{Name: "err", Type: "error"},
		},
	}
	return []*Method{&scan}
}

func (mod *mockModule) Call(method string, args ...interface{}) (interface{}, error) {
	switch method {
	case "Scan":
		if len(args) != 1 {
			return nil, errors.New("invalid argument number")
		}
		ip, ok := args[0].(string)
		if !ok {
			return nil, errors.New("argument 1 is not a string")
		}
		open, err := mod.Scan(ip)
		return []interface{}{open, err}, nil
	default:
		return nil, fmt.Errorf("unknown method: \"%s\"", method)
	}
}

func (*mockModule) Scan(ip string) (bool, error) {
	if ip == "" {
		return false, errors.New("empty ip address")
	}
	return true, nil
}
