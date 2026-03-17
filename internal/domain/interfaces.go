package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-kvq-models/models"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
)

type ProjectService interface {
	List(ctx context.Context) ([]types.ProjectSummary, error)
	Get(ctx context.Context, id uuid.UUID) (*models.Project, error)
	Create(ctx context.Context, req types.CreateProjectRequest) (*models.Project, error)
	Update(ctx context.Context, id uuid.UUID, req types.UpdateProjectRequest) (*models.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type FolderService interface {
	Tree(ctx context.Context, projectID uuid.UUID) ([]types.FolderNode, error)
	Create(ctx context.Context, projectID uuid.UUID, req types.CreateFolderRequest) (*models.VirtualFolder, error)
	Update(ctx context.Context, projectID, folderID uuid.UUID, req types.UpdateFolderRequest) (*models.VirtualFolder, error)
	Delete(ctx context.Context, projectID, folderID uuid.UUID) error
}

type ResourceService interface {
	List(ctx context.Context, projectID uuid.UUID, folderID *uuid.UUID, resourceType string) ([]models.Resource, error)
	GenerateUploadURL(ctx context.Context, projectID uuid.UUID, req types.UploadURLRequest) (*types.UploadURLResponse, error)
	ConfirmUpload(ctx context.Context, projectID, resourceID uuid.UUID) (*models.Resource, error)
	GenerateDownloadURL(ctx context.Context, projectID, resourceID uuid.UUID) (*types.DownloadURLResponse, error)
	Update(ctx context.Context, projectID, resourceID uuid.UUID, req types.UpdateResourceRequest) (*models.Resource, error)
	Delete(ctx context.Context, projectID, resourceID uuid.UUID) error
	Register(ctx context.Context, projectID uuid.UUID, filename, contentType, s3Key string) (*models.Resource, error)
}

type WorkflowService interface {
	List(ctx context.Context, projectID uuid.UUID) ([]models.WorkflowDefinition, error)
	Get(ctx context.Context, projectID, workflowID uuid.UUID) (*models.WorkflowDefinition, error)
	Create(ctx context.Context, projectID uuid.UUID, req types.CreateWorkflowRequest) (*models.WorkflowDefinition, error)
	Update(ctx context.Context, projectID, workflowID uuid.UUID, req types.UpdateWorkflowRequest) (*models.WorkflowDefinition, error)
	Delete(ctx context.Context, projectID, workflowID uuid.UUID) error
}

type JobService interface {
	List(ctx context.Context, projectID uuid.UUID, status string) ([]models.Job, error)
	Get(ctx context.Context, projectID, jobID uuid.UUID) (*models.Job, error)
	Create(ctx context.Context, projectID uuid.UUID, req types.CreateJobRequest) (*types.CreateJobResponse, error)
	Cancel(ctx context.Context, projectID, jobID uuid.UUID) (*models.Job, error)
}
