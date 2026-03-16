package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-kvq-models/models"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/uptrace/bun"
)

type WorkflowService struct {
	db  *bun.DB
	log *slog.Logger
}

func NewWorkflowService(db *bun.DB, log *slog.Logger) *WorkflowService {
	return &WorkflowService{db: db, log: log}
}

func (s *WorkflowService) List(ctx context.Context, projectID uuid.UUID) ([]models.WorkflowDefinition, error) {
	var workflows []models.WorkflowDefinition
	err := s.db.NewSelect().Model(&workflows).Where("project_id = ?", projectID).Order("created_at DESC").Scan(ctx)
	return workflows, err
}

func (s *WorkflowService) Get(ctx context.Context, projectID, workflowID uuid.UUID) (*models.WorkflowDefinition, error) {
	wf := &models.WorkflowDefinition{}
	err := s.db.NewSelect().Model(wf).Where("id = ? AND project_id = ?", workflowID, projectID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return wf, nil
}

func (s *WorkflowService) Create(ctx context.Context, projectID uuid.UUID, req types.CreateWorkflowRequest) (*models.WorkflowDefinition, error) {
	wf := &models.WorkflowDefinition{
		ProjectID:   projectID,
		Name:        req.Name,
		DAGJson:     req.DAGJson,
		InputSchema: req.InputSchema,
	}
	_, err := s.db.NewInsert().Model(wf).Returning("*").Exec(ctx)
	if err != nil {
		return nil, err
	}
	return wf, nil
}

func (s *WorkflowService) Update(ctx context.Context, projectID, workflowID uuid.UUID, req types.UpdateWorkflowRequest) (*models.WorkflowDefinition, error) {
	wf := &models.WorkflowDefinition{}
	err := s.db.NewSelect().Model(wf).Where("id = ? AND project_id = ?", workflowID, projectID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		wf.Name = req.Name
	}
	if len(req.DAGJson) > 0 {
		wf.DAGJson = req.DAGJson
	}
	if req.InputSchema != nil {
		wf.InputSchema = req.InputSchema
	}
	_, err = s.db.NewUpdate().Model(wf).Column("name", "dag_json", "input_schema", "updated_at").Where("id = ? AND project_id = ?", workflowID, projectID).Returning("*").Exec(ctx)
	if err != nil {
		return nil, err
	}
	return wf, nil
}

func (s *WorkflowService) Delete(ctx context.Context, projectID, workflowID uuid.UUID) error {
	_, err := s.db.NewDelete().Model((*models.WorkflowDefinition)(nil)).Where("id = ? AND project_id = ?", workflowID, projectID).Exec(ctx)
	return err
}
