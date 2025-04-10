package task

import (
	"encoding/json"
	"testing"
)

func TestSendTask(t *testing.T) {
	service := NewService()

	req := SendTaskRequest{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "tasks/send",
		Params: Task{
			ID:        "task-123",
			SessionID: "session-abc",
			Message: TaskMessage{
				Text: "Please process this file",
			},
		},
	}

	resp, err := service.SendTask(req)
	if err != nil {
		t.Fatalf("SendTask failed: %v", err)
	}

	if resp.Result.ID != "task-123" {
		t.Errorf("Expected task ID 'task-123', got '%s'", resp.Result.ID)
	}

	if resp.Result.Status.State != "submitted" {
		t.Errorf("Expected status 'submitted', got '%s'", resp.Result.Status.State)
	}
}

func TestSubscribeToTaskUpdates(t *testing.T) {
	service := NewService()

	// First create a task
	req := SendTaskRequest{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "tasks/send",
		Params: Task{
			ID:        "task-123",
			SessionID: "session-abc",
			Message: TaskMessage{
				Text: "Please process this file",
			},
		},
	}
	_, err := service.SendTask(req)
	if err != nil {
		t.Fatalf("SendTask failed: %v", err)
	}

	// Then subscribe to updates
	subReq := SendTaskStreamingRequest{
		JSONRPC: "2.0",
		ID:      "2",
		Method:  "tasks/sendSubscribe",
		Params: struct {
			TaskID string `json:"taskId"`
		}{
			TaskID: "task-123",
		},
	}

	resp, err := service.SubscribeToTaskUpdates(subReq)
	if err != nil {
		t.Fatalf("SubscribeToTaskUpdates failed: %v", err)
	}

	if resp.Result != "subscribed" {
		t.Errorf("Expected result 'subscribed', got '%v'", resp.Result)
	}
}

func TestSetPushNotification(t *testing.T) {
	service := NewService()

	req := SetTaskPushNotificationRequest{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "tasks/pushNotification/set",
		Params: struct {
			TaskID   string `json:"taskId"`
			Endpoint string `json:"endpoint"`
		}{
			TaskID:   "task-123",
			Endpoint: "https://example.com/webhook",
		},
	}

	resp, err := service.SetPushNotification(req)
	if err != nil {
		t.Fatalf("SetPushNotification failed: %v", err)
	}

	if resp.Result.TaskID != "task-123" {
		t.Errorf("Expected task ID 'task-123', got '%s'", resp.Result.TaskID)
	}

	if resp.Result.Endpoint != "https://example.com/webhook" {
		t.Errorf("Expected endpoint 'https://example.com/webhook', got '%s'", resp.Result.Endpoint)
	}
}

func TestProcessJSONRPCRequest(t *testing.T) {
	service := NewService()

	// Test sending a task
	sendReq := SendTaskRequest{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "tasks/send",
		Params: Task{
			ID:        "task-123",
			SessionID: "session-abc",
			Message: TaskMessage{
				Text: "Please process this file",
			},
		},
	}

	reqJSON, err := json.Marshal(sendReq)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	respJSON, err := service.ProcessJSONRPCRequest(reqJSON)
	if err != nil {
		t.Fatalf("ProcessJSONRPCRequest failed: %v", err)
	}

	var resp SendTaskResponse
	if err := json.Unmarshal(respJSON, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Result.ID != "task-123" {
		t.Errorf("Expected task ID 'task-123', got '%s'", resp.Result.ID)
	}
}

func TestSendTaskWithMissingMessage(t *testing.T) {
	service := NewService()

	req := SendTaskRequest{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "tasks/send",
		Params: Task{
			ID:        "task-123",
			SessionID: "session-abc",
			// Message field is intentionally omitted
		},
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	respJSON, err := service.ProcessJSONRPCRequest(reqJSON)
	if err != nil {
		t.Fatalf("ProcessJSONRPCRequest failed: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respJSON, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	errorObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error object in response")
	}

	code, ok := errorObj["code"].(float64)
	if !ok || int(code) != -32602 {
		t.Errorf("Expected error code -32602, got %v", code)
	}

	message, ok := errorObj["message"].(string)
	if !ok || message != "Invalid params" {
		t.Errorf("Expected error message 'Invalid params', got '%v'", message)
	}
}

func TestTaskStateLifecycle(t *testing.T) {
	service := NewService()

	// Create a task
	req := SendTaskRequest{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "tasks/send",
		Params: Task{
			ID:        "task-123",
			SessionID: "session-abc",
			Message: TaskMessage{
				Text: "Please process this file",
			},
		},
	}

	resp, err := service.SendTask(req)
	if err != nil {
		t.Fatalf("SendTask failed: %v", err)
	}

	// Verify initial state
	if resp.Result.Status.State != "submitted" {
		t.Errorf("Expected initial state 'submitted', got '%s'", resp.Result.Status.State)
	}

	// Update to working state
	err = service.UpdateTaskStatus("task-123", TaskStatus{State: "working"})
	if err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}

	task, err := service.GetTask("task-123")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if task.Status.State != "working" {
		t.Errorf("Expected state 'working', got '%s'", task.Status.State)
	}

	// Update to completed state
	err = service.UpdateTaskStatus("task-123", TaskStatus{State: "completed"})
	if err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}

	task, err = service.GetTask("task-123")
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if task.Status.State != "completed" {
		t.Errorf("Expected state 'completed', got '%s'", task.Status.State)
	}
}

func TestTaskNotFoundError(t *testing.T) {
	service := NewService()

	// Try to get a non-existent task
	_, err := service.GetTask("unknown-task")
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}

	taskErr, ok := err.(*TaskNotFoundError)
	if !ok {
		t.Fatalf("Expected TaskNotFoundError, got %T", err)
	}

	if taskErr.TaskID != "unknown-task" {
		t.Errorf("Expected TaskID 'unknown-task', got '%s'", taskErr.TaskID)
	}

	// Verify RPC error conversion
	rpcErr := ErrorToRPCError(taskErr)
	if rpcErr.Code != -32001 {
		t.Errorf("Expected error code -32001, got %d", rpcErr.Code)
	}
	if rpcErr.Message != "Task not found" {
		t.Errorf("Expected error message 'Task not found', got '%s'", rpcErr.Message)
	}
}

func TestUnsupportedOperation(t *testing.T) {
	service := NewService()

	// Try to call an unsupported method
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "tasks/selfDestruct",
		"params":  map[string]interface{}{},
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	respJSON, err := service.ProcessJSONRPCRequest(reqJSON)
	if err != nil {
		t.Fatalf("ProcessJSONRPCRequest failed: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(respJSON, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	errorObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error object in response")
	}

	code, ok := errorObj["code"].(float64)
	if !ok || int(code) != -32004 {
		t.Errorf("Expected error code -32004, got %v", code)
	}

	message, ok := errorObj["message"].(string)
	if !ok || message != "This operation is not supported" {
		t.Errorf("Expected error message 'This operation is not supported', got '%v'", message)
	}
}

func TestTaskWithMetadata(t *testing.T) {
	service := NewService()

	metadata := map[string]interface{}{
		"priority": "high",
		"tags":     []string{"urgent", "processing"},
	}

	req := SendTaskRequest{
		JSONRPC: "2.0",
		ID:      "1",
		Method:  "tasks/send",
		Params: Task{
			ID:        "task-123",
			SessionID: "session-abc",
			Message: TaskMessage{
				Text: "Please process this file",
			},
			Metadata: metadata,
		},
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	respJSON, err := service.ProcessJSONRPCRequest(reqJSON)
	if err != nil {
		t.Fatalf("ProcessJSONRPCRequest failed: %v", err)
	}

	var resp SendTaskResponse
	if err := json.Unmarshal(respJSON, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify metadata is preserved
	if len(resp.Result.Metadata) != len(metadata) {
		t.Errorf("Expected %d metadata items, got %d", len(metadata), len(resp.Result.Metadata))
	}

	priority, ok := resp.Result.Metadata["priority"].(string)
	if !ok || priority != "high" {
		t.Errorf("Expected priority 'high', got '%v'", priority)
	}

	tags, ok := resp.Result.Metadata["tags"].([]interface{})
	if !ok || len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %v", tags)
	}
}
