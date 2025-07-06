package agent

import "fmt"

type ControllerType int

const (
	CtrlTypeCmdLine ControllerType = iota
	CtrlTypeHTTP
)

func CtrlTypeFromString(s string) (ControllerType, error) {
	switch s {
	case "cmdline", "cmd":
		return CtrlTypeCmdLine, nil
	case "http", "rest":
		return CtrlTypeHTTP, nil
	default:
		return -1, fmt.Errorf("unknown controller type: %s, available types are: cmd, http", s)
	}
}

type Controller interface {
	Run() error
}
