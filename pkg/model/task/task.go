package task

import "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/executor/secureshell"

type TaskStatus int

const (
	// The task is pending
	PENDING TaskStatus = iota
	// The task is running
	RUNNING
	// The task is finished, which means all the nano tasks are finished (either with success or error)
	FINISHED
)

type Stringer interface {
	// Get the string representation of the object
	String() string
}

type StatusInterface interface {
	Stringer
	// GetStatus gets the status of the entire task
	GetStatus() TaskStatus
	// Get the number of total nano tasks
	NumTotal() int
	// Get the number of suceeded nano tasks
	NumDoneWithSuccess() int
	// Get the number of failed nano tasks
	NumDoneWithError() int
}

type TaskInterface interface {
	Stringer
	// Assign the executor to the task, the caller can use the executor to perform the task
	Assign(e *secureshell.SSHExecutor) error
	// Prepare the required conditions (e.g. upload input files, create directories, etc.)
	Prepare() error
	// Start the task
	Start() error
	// Stop the task
	Stop() error
	// Get the status of the task
	Status() (StatusInterface, error)
	// Download the output files
	Download() error
}
