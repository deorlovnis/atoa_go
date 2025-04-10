package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

// Service represents the task service
type Service struct {
	mu            sync.RWMutex
	tasks         map[string]Task
	subscribers   map[string][]chan TaskStatusUpdateEvent
	pushEndpoints map[string]string
}

// NewService creates a new task service
func NewService() *Service {
	return &Service{
		tasks:         make(map[string]Task),
		subscribers:   make(map[string][]chan TaskStatusUpdateEvent),
		pushEndpoints: make(map[string]string),
	}
}

// SendTask sends a new task
func (s *Service) SendTask(req SendTaskRequest) (SendTaskResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate required fields
	if req.Params.Message.Text == "" {
		return SendTaskResponse{}, fmt.Errorf("message text is required")
	}

	// Create task with metadata
	task := Task{
		ID:        req.Params.ID,
		SessionID: req.Params.SessionID,
		Message:   req.Params.Message,
		Status: TaskStatus{
			State: "submitted",
		},
		Metadata: req.Params.Metadata,
	}

	// Store task
	s.tasks[task.ID] = task

	// Create response
	resp := SendTaskResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  task,
	}

	// Notify subscribers
	s.notifySubscribers(task.ID, TaskStatusUpdateEvent{
		Type:   "status_update",
		Status: task.Status,
	})

	return resp, nil
}

// SubscribeToTaskUpdates subscribes to task updates
func (s *Service) SubscribeToTaskUpdates(req SendTaskStreamingRequest) (*SendTaskStreamingResponse, error) {
	if req.JSONRPC != "2.0" {
		return nil, errors.New("invalid JSON-RPC version")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a channel for updates
	ch := make(chan TaskStatusUpdateEvent, 10)
	s.subscribers[req.Params.TaskID] = append(s.subscribers[req.Params.TaskID], ch)

	// Send initial status if task exists
	if task, exists := s.tasks[req.Params.TaskID]; exists {
		ch <- TaskStatusUpdateEvent{
			Type:   "status_update",
			Status: task.Status,
		}
	}

	return &SendTaskStreamingResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  "subscribed",
	}, nil
}

// SetPushNotification sets push notification preferences for a task
func (s *Service) SetPushNotification(req SetTaskPushNotificationRequest) (*SetTaskPushNotificationResponse, error) {
	if req.JSONRPC != "2.0" {
		return nil, errors.New("invalid JSON-RPC version")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Store the push notification endpoint
	s.pushEndpoints[req.Params.TaskID] = req.Params.Endpoint

	return &SetTaskPushNotificationResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: struct {
			TaskID   string `json:"taskId"`
			Endpoint string `json:"endpoint"`
		}{
			TaskID:   req.Params.TaskID,
			Endpoint: req.Params.Endpoint,
		},
	}, nil
}

// GetTask retrieves a task by ID
func (s *Service) GetTask(taskID string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, &TaskNotFoundError{TaskID: taskID}
	}
	return &task, nil
}

// UpdateTaskStatus updates the status of a task
func (s *Service) UpdateTaskStatus(taskID string, status TaskStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return &TaskNotFoundError{TaskID: taskID}
	}

	// Update the task status
	task.Status = status
	s.tasks[taskID] = task

	// Notify subscribers
	s.notifySubscribers(taskID, TaskStatusUpdateEvent{
		Type:   "status_update",
		Status: status,
	})

	return nil
}

// notifySubscribers sends updates to all subscribers
func (s *Service) notifySubscribers(taskID string, event TaskStatusUpdateEvent) {
	if subscribers, exists := s.subscribers[taskID]; exists {
		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// Channel is full, skip this subscriber
			}
		}
	}
}

// ProcessJSONRPCRequest processes a JSON-RPC request
func (s *Service) ProcessJSONRPCRequest(request []byte) ([]byte, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	// Validate JSON-RPC version
	version, ok := req["jsonrpc"].(string)
	if !ok || version != "2.0" {
		return json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
			},
		})
	}

	// Get method
	method, ok := req["method"].(string)
	if !ok {
		return json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid Request",
			},
		})
	}

	// Handle different methods
	switch method {
	case "tasks/send":
		var sendReq SendTaskRequest
		if err := json.Unmarshal(request, &sendReq); err != nil {
			return json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params",
				},
			})
		}

		// Validate required fields
		if sendReq.Params.Message.Text == "" {
			return json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params",
				},
			})
		}

		resp, err := s.SendTask(sendReq)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)

	case "tasks/sendSubscribe":
		var subReq SendTaskStreamingRequest
		if err := json.Unmarshal(request, &subReq); err != nil {
			return json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params",
				},
			})
		}
		resp, err := s.SubscribeToTaskUpdates(subReq)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)

	case "tasks/pushNotification/set":
		var notifReq SetTaskPushNotificationRequest
		if err := json.Unmarshal(request, &notifReq); err != nil {
			return json.Marshal(map[string]interface{}{
				"jsonrpc": "2.0",
				"error": map[string]interface{}{
					"code":    -32602,
					"message": "Invalid params",
				},
			})
		}
		resp, err := s.SetPushNotification(notifReq)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)

	default:
		return json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0",
			"error": map[string]interface{}{
				"code":    -32004,
				"message": "This operation is not supported",
			},
		})
	}
}
