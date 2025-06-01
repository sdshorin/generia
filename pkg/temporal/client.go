package temporal

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

// Client wraps Temporal client with convenience methods
type Client struct {
	client.Client
}

// NewClient creates a new Temporal client
func NewClient(hostPort string) (*Client, error) {
	c, err := client.Dial(client.Options{
		HostPort: hostPort,
		DataConverter: converter.NewCompositeDataConverter(
			converter.NewNilPayloadConverter(),
			converter.NewByteSlicePayloadConverter(),
			converter.NewProtoJSONPayloadConverter(),
			converter.NewProtoPayloadConverter(),
			converter.NewJSONPayloadConverter(),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %w", err)
	}

	return &Client{Client: c}, nil
}

// WorkflowInput represents input for InitWorldCreationWorkflow
type InitWorldCreationInput struct {
	WorldID         string `json:"world_id"`
	WorldName       string `json:"world_name"`
	WorldPrompt     string `json:"world_prompt"`
	CharactersCount int    `json:"characters_count"`
	PostsCount      int    `json:"posts_count"`
}

// ExecuteInitWorldCreationWorkflow starts the world creation workflow
func (c *Client) ExecuteInitWorldCreationWorkflow(ctx context.Context, input InitWorldCreationInput) (client.WorkflowRun, error) {
	options := client.StartWorkflowOptions{
		ID:           fmt.Sprintf("init-world-%s", input.WorldID),
		TaskQueue:    "ai-worker-main",
		WorkflowExecutionTimeout: 30 * time.Minute,
		WorkflowRunTimeout:       30 * time.Minute,
		WorkflowTaskTimeout:      5 * time.Minute,
	}

	return c.ExecuteWorkflow(ctx, options, "InitWorldCreationWorkflow", input)
}

// GetWorkflowResult gets the result of a workflow execution
func (c *Client) GetWorkflowResult(ctx context.Context, workflowID string, runID string, valuePtr interface{}) error {
	workflowRun := c.GetWorkflow(ctx, workflowID, runID)
	return workflowRun.Get(ctx, valuePtr)
}

// Close closes the Temporal client
func (c *Client) Close() {
	c.Client.Close()
}