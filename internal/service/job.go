package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-api-server/internal/storage"
	"github.com/kirbyevanj/kvqtool-kvq-models/messages"
	"github.com/kirbyevanj/kvqtool-kvq-models/models"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/uptrace/bun"
)

const jobCancelChannel = "kvq:jobs:cancel"

type cancelJobMessage struct {
	JobID uuid.UUID `json:"job_id"`
}

type JobService struct {
	db     *bun.DB
	valkey *storage.ValkeyClient
	log    *slog.Logger
}

func NewJobService(db *bun.DB, valkey *storage.ValkeyClient, log *slog.Logger) *JobService {
	return &JobService{db: db, valkey: valkey, log: log}
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
	msg := messages.JobMessage{
		JobID:      job.ID,
		ProjectID:  projectID,
		WorkflowID: req.WorkflowID,
		DAGJson:    wf.DAGJson,
		Params:     job.InputParams,
	}
	err = s.valkey.PushJob(ctx, messages.JobQueueKey, msg)
	if err != nil {
		return nil, err
	}
	return &types.CreateJobResponse{JobID: job.ID, Status: job.Status}, nil
}

func (s *JobService) Cancel(ctx context.Context, projectID, jobID uuid.UUID) (*models.Job, error) {
	job := &models.Job{}
	_, err := s.db.NewUpdate().Model(job).Set("status = ?", models.JobStatusCancelled).Where("id = ? AND project_id = ?", jobID, projectID).Returning("*").Exec(ctx, job)
	if err != nil {
		return nil, err
	}
	_ = s.valkey.Publish(ctx, jobCancelChannel, cancelJobMessage{JobID: jobID})
	return job, nil
}
