package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-api-server/internal/storage"
	"github.com/kirbyevanj/kvqtool-kvq-models/models"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/uptrace/bun"
)

var defaultUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

type ProjectService struct {
	db  *bun.DB
	s3  *storage.S3Client
	log *slog.Logger
}

func NewProjectService(db *bun.DB, s3 *storage.S3Client, log *slog.Logger) *ProjectService {
	return &ProjectService{db: db, s3: s3, log: log}
}

type projectListRow struct {
	ID            uuid.UUID
	Name          string
	Description   string
	CreatedAt     time.Time
	ResourceCount int
	JobCount      int
}

func (s *ProjectService) List(ctx context.Context) ([]types.ProjectSummary, error) {
	var rows []projectListRow
	err := s.db.NewSelect().
		TableExpr("projects AS p").
		ColumnExpr("p.id, p.name, p.description, p.created_at").
		ColumnExpr("(SELECT COUNT(*) FROM resources WHERE project_id = p.id) AS resource_count").
		ColumnExpr("(SELECT COUNT(*) FROM jobs WHERE project_id = p.id) AS job_count").
		Where("p.user_id = ?", defaultUserID).
		Order("p.created_at DESC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	out := make([]types.ProjectSummary, len(rows))
	for i, r := range rows {
		out[i] = types.ProjectSummary{
			ID:            r.ID,
			Name:          r.Name,
			Description:   r.Description,
			CreatedAt:     r.CreatedAt.Format(time.RFC3339),
			ResourceCount: r.ResourceCount,
			JobCount:      r.JobCount,
		}
	}
	return out, nil
}

func (s *ProjectService) Get(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	proj := &models.Project{}
	err := s.db.NewSelect().Model(proj).Where("id = ? AND user_id = ?", id, defaultUserID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *ProjectService) Create(ctx context.Context, req types.CreateProjectRequest) (*models.Project, error) {
	proj := &models.Project{
		UserID:      defaultUserID,
		Name:        req.Name,
		Description: req.Description,
	}
	_, err := s.db.NewInsert().Model(proj).Returning("*").Exec(ctx)
	if err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, req types.UpdateProjectRequest) (*models.Project, error) {
	proj := &models.Project{}
	q := s.db.NewUpdate().Model(proj).Where("id = ? AND user_id = ?", id, defaultUserID).Set("updated_at = NOW()")
	if req.Name != "" {
		q = q.Set("name = ?", req.Name)
	}
	if req.Description != "" {
		q = q.Set("description = ?", req.Description)
	}
	_, err := q.Returning("*").Exec(ctx, proj)
	if err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.NewDelete().Model((*models.Job)(nil)).Where("project_id = ?", id).Exec(ctx)
	tx.NewDelete().Model((*models.Resource)(nil)).Where("project_id = ?", id).Exec(ctx)
	tx.NewDelete().Model((*models.WorkflowDefinition)(nil)).Where("project_id = ?", id).Exec(ctx)
	tx.NewDelete().Model((*models.VirtualFolder)(nil)).Where("project_id = ?", id).Exec(ctx)
	tx.NewDelete().Model((*models.Project)(nil)).Where("id = ? AND user_id = ?", id, defaultUserID).Exec(ctx)

	if err := tx.Commit(); err != nil {
		return err
	}

	prefix := "projects/" + id.String() + "/"
	return s.s3.DeletePrefix(ctx, prefix)
}
