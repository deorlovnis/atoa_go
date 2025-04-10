package task

// Task represents a task in the A2A system
type Task struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"sessionId"`
	Message   TaskMessage            `json:"message"`
	Status    TaskStatus             `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TaskMessage represents the content of a task
type TaskMessage struct {
	Text string `json:"text"`
}

// TaskStatus represents the current state of a task
type TaskStatus struct {
	State   string `json:"state"`
	Message string `json:"message,omitempty"`
}

// TaskArtifactUpdateEvent represents an update about a task artifact
type TaskArtifactUpdateEvent struct {
	Type       string `json:"type"`
	ArtifactID string `json:"artifact.id"`
}

// TaskStatusUpdateEvent represents an update about a task's status
type TaskStatusUpdateEvent struct {
	Type   string     `json:"type"`
	Status TaskStatus `json:"status"`
}

// SendTaskRequest represents a request to send a task
type SendTaskRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  Task        `json:"params"`
}

// SendTaskResponse represents the response to a task send request
type SendTaskResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  Task        `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// SendTaskStreamingRequest represents a request to subscribe to task updates
type SendTaskStreamingRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  struct {
		TaskID string `json:"taskId"`
	} `json:"params"`
}

// SendTaskStreamingResponse represents a streaming response for task updates
type SendTaskStreamingResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// SetTaskPushNotificationRequest represents a request to set push notification preferences
type SetTaskPushNotificationRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  struct {
		TaskID   string `json:"taskId"`
		Endpoint string `json:"endpoint"`
	} `json:"params"`
}

// SetTaskPushNotificationResponse represents the response to setting push notification preferences
type SetTaskPushNotificationResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  struct {
		TaskID   string `json:"taskId"`
		Endpoint string `json:"endpoint"`
	} `json:"result,omitempty"`
	Error *RPCError `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
