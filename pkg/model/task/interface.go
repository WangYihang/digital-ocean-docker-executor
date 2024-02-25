package task

type Progress interface {
	String() string
	Done() bool
}

type Task interface {
	Start() (chan Progress, error)
}
