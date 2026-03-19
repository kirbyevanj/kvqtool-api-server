package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-api-server/internal/storage"
	"github.com/kirbyevanj/kvqtool-kvq-models/models"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/uptrace/bun"
)

const presignExpiry = time.Hour

type ResourceService struct {
	db  *bun.DB
	s3  *storage.S3Client
	log *slog.Logger
}

func NewResourceService(db *bun.DB, s3 *storage.S3Client, log *slog.Logger) *ResourceService {
	return &ResourceService{db: db, s3: s3, log: log}
}

func (s *ResourceService) List(ctx context.Context, projectID uuid.UUID, folderID *uuid.UUID, resourceType string) ([]models.Resource, error) {
	var resources []models.Resource
	q := s.db.NewSelect().Model(&resources).Where("project_id = ?", projectID)
	if folderID != nil {
		q = q.Where("folder_id = ?", *folderID)
	}
	if resourceType != "" {
		q = q.Where("resource_type = ?", resourceType)
	}
	err := q.Order("created_at DESC").Scan(ctx)
	return resources, err
}

func (s *ResourceService) GenerateUploadURL(ctx context.Context, projectID uuid.UUID, req types.UploadURLRequest) (*types.UploadURLResponse, error) {
	resourceID := uuid.New()
	s3Key := fmt.Sprintf("projects/%s/media/%s-%s", projectID, resourceID, req.Filename)
	resource := &models.Resource{
		ID:           resourceID,
		ProjectID:    projectID,
		FolderID:     req.FolderID,
		ResourceType: inferResourceType(req.ContentType),
		Name:         req.Filename,
		S3Key:        s3Key,
	}
	_, err := s.db.NewInsert().Model(resource).Exec(ctx)
	if err != nil {
		return nil, err
	}
	url, err := s.s3.PresignPut(ctx, s3Key, req.ContentType, presignExpiry)
	if err != nil {
		return nil, err
	}
	return &types.UploadURLResponse{
		ResourceID:       resourceID,
		UploadURL:        url,
		S3Key:            s3Key,
		ExpiresInSeconds: int(presignExpiry.Seconds()),
	}, nil
}

func (s *ResourceService) Copy(ctx context.Context, projectID, resourceID uuid.UUID) (*models.Resource, error) {
	src := &models.Resource{}
	err := s.db.NewSelect().Model(src).Where("id = ? AND project_id = ?", resourceID, projectID).Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	newID := uuid.New()
	newKey := fmt.Sprintf("projects/%s/media/%s-%s", projectID, newID, src.Name)

	if err := s.s3.Copy(ctx, src.S3Key, newKey); err != nil {
		return nil, fmt.Errorf("s3 copy: %w", err)
	}

	copy := &models.Resource{
		ProjectID:    projectID,
		FolderID:     src.FolderID,
		ResourceType: src.ResourceType,
		Name:         src.Name + " (copy)",
		S3Key:        newKey,
		SizeBytes:    src.SizeBytes,
		Metadata:     src.Metadata,
	}
	_, err = s.db.NewInsert().Model(copy).Returning("*").Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("insert copy: %w", err)
	}
	return copy, nil
}

func (s *ResourceService) Register(ctx context.Context, projectID uuid.UUID, filename, contentType, s3Key string) (*models.Resource, error) {
	res := &models.Resource{
		ProjectID:    projectID,
		ResourceType: inferResourceType(contentType),
		Name:         filename,
		S3Key:        s3Key,
	}

	head, err := s.s3.Head(ctx, s3Key)
	if err == nil && head.ContentLength != nil {
		res.SizeBytes = *head.ContentLength
	}

	_, err = s.db.NewInsert().Model(res).Returning("*").Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("insert resource: %w", err)
	}
	s.log.Info("resource registered", "id", res.ID, "name", filename, "s3_key", s3Key)
	return res, nil
}

func inferResourceType(contentType string) string {
	if contentType == "" {
		return "file"
	}
	switch {
	case len(contentType) >= 6 && contentType[:6] == "video/":
		return "media"
	case len(contentType) >= 6 && contentType[:6] == "image/":
		return "media"
	case len(contentType) >= 5 && contentType[:5] == "audio/":
		return "media"
	case contentType == "application/x-metric-report":
		return "report"
	default:
		return "file"
	}
}

func (s *ResourceService) ConfirmUpload(ctx context.Context, projectID, resourceID uuid.UUID) (*models.Resource, error) {
	var res models.Resource
	err := s.db.NewSelect().Model(&res).Where("id = ? AND project_id = ?", resourceID, projectID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	head, err := s.s3.Head(ctx, res.S3Key)
	if err != nil {
		return nil, fmt.Errorf("s3 head: %w", err)
	}
	size := int64(0)
	if head.ContentLength != nil {
		size = *head.ContentLength
	}
	_, err = s.db.NewUpdate().Model(&res).Set("size_bytes = ?", size).Where("id = ? AND project_id = ?", resourceID, projectID).Returning("*").Exec(ctx, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *ResourceService) GenerateDownloadURL(ctx context.Context, projectID, resourceID uuid.UUID) (*types.DownloadURLResponse, error) {
	var res models.Resource
	err := s.db.NewSelect().Model(&res).Where("id = ? AND project_id = ?", resourceID, projectID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	url, err := s.s3.PresignGet(ctx, res.S3Key, presignExpiry)
	if err != nil {
		return nil, err
	}
	return &types.DownloadURLResponse{
		DownloadURL:      url,
		ExpiresInSeconds: int(presignExpiry.Seconds()),
	}, nil
}

func (s *ResourceService) Update(ctx context.Context, projectID, resourceID uuid.UUID, req types.UpdateResourceRequest) (*models.Resource, error) {
	res := &models.Resource{}
	q := s.db.NewUpdate().Model(res).Where("id = ? AND project_id = ?", resourceID, projectID)
	if req.Name != "" {
		q = q.Set("name = ?", req.Name)
	}
	if req.FolderID != nil {
		q = q.Set("folder_id = ?", req.FolderID)
	}
	_, err := q.Set("updated_at = NOW()").Returning("*").Exec(ctx, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ResourceService) Delete(ctx context.Context, projectID, resourceID uuid.UUID) error {
	var res models.Resource
	err := s.db.NewSelect().Model(&res).Where("id = ? AND project_id = ?", resourceID, projectID).Scan(ctx)
	if err != nil {
		return err
	}
	_, err = s.db.NewDelete().Model(&res).Where("id = ? AND project_id = ?", resourceID, projectID).Exec(ctx)
	if err != nil {
		return err
	}
	return s.s3.Delete(ctx, res.S3Key)
}
