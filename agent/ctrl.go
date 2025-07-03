package agent

type ControllerType int

const (
	CtrlTypeCmdLine ControllerType = iota
	CtrlTypeHTTP
)

type Controller interface {
	Run() error
}
