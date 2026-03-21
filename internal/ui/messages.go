package ui

// taskCompleteMsg indicates a task action finished.
type taskCompleteMsg struct {
	err error
}

// allTasksCompleteMsg indicates all queued tasks have finished.
type allTasksCompleteMsg struct{}
