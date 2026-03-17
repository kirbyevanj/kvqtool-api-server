package storage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"go.temporal.io/sdk/client"
)

type TemporalClient struct {
	client client.Client
	logger *slog.Logger
}

func NewTemporalClient(hostPort string, logger *slog.Logger) (*TemporalClient, error) {
	c, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: types.TemporalNamespace,
	})
	if err != nil {
		return nil, fmt.Errorf("temporal dial: %w", err)
	}
	logger.Info("connected to temporal", "host", hostPort)
	return &TemporalClient{client: c, logger: logger}, nil
}

func (t *TemporalClient) StartWorkflow(ctx context.Context, workflowID string, dag types.WorkflowDAG) (string, error) {
	opts := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: types.TemporalTaskQueue,
	}
	run, err := t.client.ExecuteWorkflow(ctx, opts, "InterpreterWorkflow", dag)
	if err != nil {
		return "", fmt.Errorf("start workflow: %w", err)
	}
	return run.GetRunID(), nil
}

func (t *TemporalClient) GetWorkflowStatus(ctx context.Context, workflowID, runID string) (*types.WorkflowStatusResponse, error) {
	desc, err := t.client.DescribeWorkflowExecution(ctx, workflowID, runID)
	if err != nil {
		return nil, fmt.Errorf("describe workflow: %w", err)
	}
	status := desc.WorkflowExecutionInfo.Status.String()
	return &types.WorkflowStatusResponse{
		RunID:  runID,
		Status: status,
	}, nil
}

func (t *TemporalClient) Ping(ctx context.Context) error {
	_, err := t.client.CheckHealth(ctx, &client.CheckHealthRequest{})
	return err
}

func (t *TemporalClient) Close() {
	t.client.Close()
}
