package simupro

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// SimulationType represents the type of simulation
type SimulationType string

const (
	ProcessSimulation        SimulationType = "process-simulation"
	VirtualCommissioning     SimulationType = "virtual-commissioning"
	ManufacturingSimulation  SimulationType = "manufacturing-simulation"
	RoboticsSimulation       SimulationType = "robotics-simulation"
	PLCSimulation            SimulationType = "plc-simulation"
	SupplyChain              SimulationType = "supply-chain"
	QualityControl           SimulationType = "quality-control"
	EnergyOptimization       SimulationType = "energy-optimization"
	PredictiveMaintenance    SimulationType = "predictive-maintenance"
	DigitalTwin              SimulationType = "digital-twin"
)

// JobStatus represents the status of a simulation job
type JobStatus string

const (
	JobStatusWaiting   JobStatus = "waiting"
	JobStatusActive    JobStatus = "active"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// Client is the main SimuPro client
type Client struct {
	apiKey        string
	baseURL       string
	wsURL         string
	organizationID string
	httpClient    *resty.Client
	wsConn        *websocket.Conn
	logger        *zap.Logger
	mu            sync.RWMutex
	subscriptions map[string]chan JobUpdate
}

// Config holds client configuration
type Config struct {
	APIKey         string
	BaseURL        string
	OrganizationID string
	Timeout        time.Duration
	EnableWebSocket bool
	Logger         *zap.Logger
}

// SimulationJob represents a job to be submitted
type SimulationJob struct {
	Type        SimulationType         `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// Job represents a job status response
type Job struct {
	ID          string                 `json:"id"`
	Status      JobStatus              `json:"status"`
	Progress    int                    `json:"progress"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	StartedAt   *time.Time             `json:"startedAt,omitempty"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
}

// JobUpdate represents a real-time job update
type JobUpdate struct {
	JobID    string    `json:"jobId"`
	Status   JobStatus `json:"status"`
	Progress int       `json:"progress"`
	Message  string    `json:"message,omitempty"`
}

// SimulationResult represents the result of a completed simulation
type SimulationResult struct {
	JobID         string                 `json:"jobId"`
	Type          SimulationType         `json:"type"`
	Status        string                 `json:"status"`
	ExecutionTime int64                  `json:"executionTime"`
	Results       map[string]interface{} `json:"results"`
	Metadata      map[string]interface{} `json:"metadata"`
	Files         []string               `json:"files,omitempty"`
}

// NewClient creates a new SimuPro client
func NewClient(config Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.simupro.io"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.Logger == nil {
		config.Logger, _ = zap.NewProduction()
	}

	httpClient := resty.New().
		SetTimeout(config.Timeout).
		SetHeader("Authorization", "Bearer "+config.APIKey).
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "SimuPro-Go-SDK/1.0.0")

	if config.OrganizationID != "" {
		httpClient.SetHeader("X-Organization-ID", config.OrganizationID)
	}

	client := &Client{
		apiKey:         config.APIKey,
		baseURL:        config.BaseURL,
		organizationID: config.OrganizationID,
		httpClient:     httpClient,
		logger:         config.Logger,
		subscriptions:  make(map[string]chan JobUpdate),
	}

	if config.EnableWebSocket {
		wsURL := config.BaseURL
		wsURL = "wss" + wsURL[5:] // Replace https with wss
		client.wsURL = wsURL
		if err := client.connectWebSocket(); err != nil {
			return nil, fmt.Errorf("failed to connect websocket: %w", err)
		}
	}

	return client, nil
}

// connectWebSocket establishes WebSocket connection
func (c *Client) connectWebSocket() error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(c.wsURL+"/ws", nil)
	if err != nil {
		return err
	}

	c.wsConn = conn

	// Send authentication
	authMsg := map[string]string{
		"type":  "auth",
		"token": c.apiKey,
	}
	if err := conn.WriteJSON(authMsg); err != nil {
		return err
	}

	// Start listening for messages
	go c.handleWebSocketMessages()

	return nil
}

// handleWebSocketMessages processes incoming WebSocket messages
func (c *Client) handleWebSocketMessages() {
	for {
		var update JobUpdate
		err := c.wsConn.ReadJSON(&update)
		if err != nil {
			c.logger.Error("WebSocket read error", zap.Error(err))
			// Attempt reconnection
			time.Sleep(5 * time.Second)
			if err := c.connectWebSocket(); err != nil {
				c.logger.Error("WebSocket reconnection failed", zap.Error(err))
			}
			continue
		}

		c.mu.RLock()
		ch, exists := c.subscriptions[update.JobID]
		c.mu.RUnlock()

		if exists {
			select {
			case ch <- update:
			default:
				// Channel full, skip update
			}
		}
	}
}

// SubmitJob submits a new simulation job
func (c *Client) SubmitJob(ctx context.Context, job SimulationJob) (string, error) {
	var response struct {
		JobID string `json:"jobId"`
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(job).
		SetResult(&response).
		Post(c.baseURL + "/jobs/submit")

	if err != nil {
		return "", fmt.Errorf("failed to submit job: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	// Subscribe to updates if WebSocket is enabled
	if c.wsConn != nil {
		c.SubscribeToJob(response.JobID)
	}

	return response.JobID, nil
}

// GetJobStatus retrieves the current status of a job
func (c *Client) GetJobStatus(ctx context.Context, jobID string) (*Job, error) {
	var job Job

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetResult(&job).
		Get(c.baseURL + "/jobs/" + jobID + "/status")

	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return &job, nil
}

// GetJobResult retrieves the result of a completed job
func (c *Client) GetJobResult(ctx context.Context, jobID string) (*SimulationResult, error) {
	var result SimulationResult

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetResult(&result).
		Get(c.baseURL + "/jobs/" + jobID + "/result")

	if err != nil {
		return nil, fmt.Errorf("failed to get job result: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return &result, nil
}

// CancelJob cancels a running job
func (c *Client) CancelJob(ctx context.Context, jobID string) error {
	resp, err := c.httpClient.R().
		SetContext(ctx).
		Delete(c.baseURL + "/jobs/" + jobID)

	if err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	// Unsubscribe from updates
	if c.wsConn != nil {
		c.UnsubscribeFromJob(jobID)
	}

	return nil
}

// WaitForCompletion waits for a job to complete with exponential backoff
func (c *Client) WaitForCompletion(ctx context.Context, jobID string) (*SimulationResult, error) {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Minute

	var result *SimulationResult
	operation := func() error {
		job, err := c.GetJobStatus(ctx, jobID)
		if err != nil {
			return err
		}

		switch job.Status {
		case JobStatusCompleted:
			result, err = c.GetJobResult(ctx, jobID)
			return err
		case JobStatusFailed:
			return backoff.Permanent(fmt.Errorf("job failed: %s", job.Error))
		default:
			return fmt.Errorf("job still running")
		}
	}

	if err := backoff.Retry(operation, b); err != nil {
		return nil, err
	}

	return result, nil
}

// SubscribeToJob subscribes to real-time updates for a job
func (c *Client) SubscribeToJob(jobID string) <-chan JobUpdate {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan JobUpdate, 10)
	c.subscriptions[jobID] = ch

	if c.wsConn != nil {
		msg := map[string]string{
			"type":  "subscribe",
			"jobId": jobID,
		}
		c.wsConn.WriteJSON(msg)
	}

	return ch
}

// UnsubscribeFromJob unsubscribes from job updates
func (c *Client) UnsubscribeFromJob(jobID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ch, exists := c.subscriptions[jobID]; exists {
		close(ch)
		delete(c.subscriptions, jobID)
	}

	if c.wsConn != nil {
		msg := map[string]string{
			"type":  "unsubscribe",
			"jobId": jobID,
		}
		c.wsConn.WriteJSON(msg)
	}
}

// BatchSubmit submits multiple jobs in batch
func (c *Client) BatchSubmit(ctx context.Context, jobs []SimulationJob) ([]string, error) {
	var response struct {
		JobIDs []string `json:"jobIds"`
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetBody(map[string]interface{}{
			"jobs": jobs,
		}).
		SetResult(&response).
		Post(c.baseURL + "/jobs/batch")

	if err != nil {
		return nil, fmt.Errorf("failed to submit batch: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return response.JobIDs, nil
}

// Close closes the client and cleans up resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close all subscription channels
	for _, ch := range c.subscriptions {
		close(ch)
	}
	c.subscriptions = make(map[string]chan JobUpdate)

	// Close WebSocket connection
	if c.wsConn != nil {
		c.wsConn.Close()
	}

	return nil
}