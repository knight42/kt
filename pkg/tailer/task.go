package tailer

import "context"

type Task struct {
	Job       func()
	Cancel    context.CancelFunc
	Completed bool
}
