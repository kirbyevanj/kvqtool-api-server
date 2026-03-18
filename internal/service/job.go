package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-api-server/internal/storage"
	"github.com/kirbyevanj/kvqtool-kvq-models/models"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/uptrace/bun"
)

type JobService struct {
	db       *bun.DB
	temporal *storage.TemporalClient
	log      *slog.Logger
}

func NewJobService(db *bun.DB, temporal *storage.TemporalClient, log *slog.Logger) *JobService {
	return &JobService{db: db, temporal: temporal, log: log}
}

func (s *JobService) List(ctx context.Context, projectID uuid.UUID, status string) ([]models.Job, error) {
	var jobs []models.Job
	q := s.db.NewSelect().Model(&jobs).Where("project_id = ?", projectID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Order("created_at DESC").Scan(ctx)
	return jobs, err
}

func (s *JobService) Get(ctx context.Context, projectID, jobID uuid.UUID) (*models.Job, error) {
	job := &models.Job{}
	err := s.db.NewSelect().Model(job).Where("id = ? AND project_id = ?", jobID, projectID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *JobService) Create(ctx context.Context, projectID uuid.UUID, req types.CreateJobRequest) (*types.CreateJobResponse, error) {
	var wf models.WorkflowDefinition
	err := s.db.NewSelect().Model(&wf).Where("id = ? AND project_id = ?", req.WorkflowID, projectID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	var dag types.WorkflowDAG
	if err := json.Unmarshal(wf.DAGJson, &dag); err != nil {
		return nil, fmt.Errorf("parse dag: %w", err)
	}

	if err := s.resolveResourceParams(ctx, projectID, &dag); err != nil {
		return nil, fmt.Errorf("resolve resources: %w", err)
	}

	// Merge runtime InputParams into GlobalInputs defaults.
	if len(req.InputParams) > 0 && dag.GlobalInputs != nil {
		var inputValues map[string]string
		if json.Unmarshal(req.InputParams, &inputValues) == nil {
			for key, val := range inputValues {
				if gi, ok := dag.GlobalInputs[key]; ok {
					gi.Default = val
				}
			}
		}
	}

	job := &models.Job{
		ProjectID:   projectID,
		WorkflowID:  req.WorkflowID,
		Status:      models.JobStatusQueued,
		InputParams: req.InputParams,
	}
	_, err = s.db.NewInsert().Model(job).Returning("*").Exec(ctx)
	if err != nil {
		return nil, err
	}

	runID, err := s.temporal.StartWorkflow(ctx, job.ID.String(), dag)
	if err != nil {
		return nil, fmt.Errorf("start workflow: %w", err)
	}

	return &types.CreateJobResponse{JobID: job.ID, Status: job.Status, RunID: runID}, nil
}

func (s *JobService) Cancel(ctx context.Context, projectID, jobID uuid.UUID) (*models.Job, error) {
	job := &models.Job{}
	_, err := s.db.NewUpdate().Model(job).Set("status = ?", models.JobStatusCancelled).Where("id = ? AND project_id = ?", jobID, projectID).Returning("*").Exec(ctx, job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// nodeTypesWithResourceID lists all node types that can reference a resource_id.
var nodeTypesWithResourceID = map[string]bool{
	types.ActivityResDownload:        true,
	types.ActivityResUpload:          true,
	types.ActivityX264RemoteTranscode: true,
	types.ActivityRemoteFileMetric:   true,
	types.ActivityRemoteSceneCut:     true,
	types.ActivityRemoteSegmentMedia: true,
}

// resourceParamMappings defines extra resource_id-style params beyond the standard "resource_id".
// Maps node type → list of (paramKey, s3KeyTarget) pairs.
var resourceParamMappings = map[string][][2]string{
	types.ActivityRemoteFileMetric: {
		{"reference_resource_id", "reference_s3_key"},
		{"distorted_resource_id", "distorted_s3_key"},
	},
}

func (s *JobService) resolveResourceParams(ctx context.Context, projectID uuid.UUID, dag *types.WorkflowDAG) error {
	for _, node := range dag.Nodes {
		if node.Params == nil {
			node.Params = make(map[string]string)
		}
		if node.Params["project_id"] == "" {
			node.Params["project_id"] = projectID.String()
		}

		if !nodeTypesWithResourceID[node.Type] {
			continue
		}

		// Standard single resource_id → s3_key mapping.
		if resID := node.Params["resource_id"]; resID != "" {
			if res, err := s.lookupResource(ctx, projectID, resID); err == nil {
				node.Params["s3_key"] = res.S3Key
				node.Params["resource_name"] = res.Name
			}
		}

		// Additional per-node-type resource mappings (e.g. reference/distorted for metrics).
		for _, mapping := range resourceParamMappings[node.Type] {
			srcKey, dstKey := mapping[0], mapping[1]
			if resID := node.Params[srcKey]; resID != "" {
				if res, err := s.lookupResource(ctx, projectID, resID); err == nil {
					node.Params[dstKey] = res.S3Key
				} else {
					return fmt.Errorf("resource %s (%s) not found: %w", resID, srcKey, err)
				}
			}
		}
	}
	return nil
}

func (s *JobService) lookupResource(ctx context.Context, projectID uuid.UUID, resIDStr string) (*models.Resource, error) {
	rid, err := uuid.Parse(resIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid resource id %q: %w", resIDStr, err)
	}
	res := &models.Resource{}
	if err := s.db.NewSelect().Model(res).Where("id = ? AND project_id = ?", rid, projectID).Scan(ctx); err != nil {
		return nil, err
	}
	return res, nil
}
