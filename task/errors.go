package task

import "fmt"

// TaskNotFoundError represents an error when a task is not found
type TaskNotFoundError struct {
	TaskID string
}

func (e *TaskNotFoundError) Error() string {
	return "Task not found"
}

// TaskNotCancelableError represents an error when a task cannot be canceled
type TaskNotCancelableError struct {
	TaskID string
	State  string
}

func (e *TaskNotCancelableError) Error() string {
	return fmt.Sprintf("task %s cannot be canceled in state %s", e.TaskID, e.State)
}

// UnsupportedOperationError represents an error when an operation is not supported
type UnsupportedOperationError struct {
	Operation string
}

func (e *UnsupportedOperationError) Error() string {
	return fmt.Sprintf("unsupported operation: %s", e.Operation)
}

// ErrorToRPCError converts a task error to an RPCError
func ErrorToRPCError(err error) *RPCError {
	switch e := err.(type) {
	case *TaskNotFoundError:
		return &RPCError{
			Code:    -32001,
			Message: e.Error(),
		}
	case *TaskNotCancelableError:
		return &RPCError{
			Code:    -32002,
			Message: e.Error(),
		}
	case *UnsupportedOperationError:
		return &RPCError{
			Code:    -32003,
			Message: e.Error(),
		}
	default:
		return &RPCError{
			Code:    -32603,
			Message: "Internal error",
		}
	}
}
