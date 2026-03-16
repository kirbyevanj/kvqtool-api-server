package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-kvq-models/models"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/uptrace/bun"
)

type FolderService struct {
	db  *bun.DB
	log *slog.Logger
}

func NewFolderService(db *bun.DB, log *slog.Logger) *FolderService {
	return &FolderService{db: db, log: log}
}

func (s *FolderService) Tree(ctx context.Context, projectID uuid.UUID) ([]types.FolderNode, error) {
	var folders []models.VirtualFolder
	err := s.db.NewSelect().Model(&folders).Where("project_id = ?", projectID).Order("path ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}
	return buildFolderTree(folders, nil), nil
}

func buildFolderTree(folders []models.VirtualFolder, parentID *uuid.UUID) []types.FolderNode {
	var nodes []types.FolderNode
	for _, f := range folders {
		if (parentID == nil && f.ParentID == nil) || (parentID != nil && f.ParentID != nil && *f.ParentID == *parentID) {
			node := types.FolderNode{
				ID:       f.ID,
				Name:     f.Name,
				Path:     f.Path,
				ParentID: f.ParentID,
				Children: buildFolderTree(folders, &f.ID),
			}
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (s *FolderService) Create(ctx context.Context, projectID uuid.UUID, req types.CreateFolderRequest) (*models.VirtualFolder, error) {
	path := req.Name
	if req.ParentID != nil {
		var parent models.VirtualFolder
		err := s.db.NewSelect().Model(&parent).Where("id = ? AND project_id = ?", *req.ParentID, projectID).Scan(ctx)
		if err != nil {
			return nil, err
		}
		path = strings.TrimSuffix(parent.Path, "/") + "/" + req.Name
	}
	folder := &models.VirtualFolder{
		ProjectID: projectID,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      path,
	}
	_, err := s.db.NewInsert().Model(folder).Returning("*").Exec(ctx)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *FolderService) Update(ctx context.Context, projectID, folderID uuid.UUID, req types.UpdateFolderRequest) (*models.VirtualFolder, error) {
	folder := &models.VirtualFolder{}
	err := s.db.NewSelect().Model(folder).Where("id = ? AND project_id = ?", folderID, projectID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	name := folder.Name
	parentID := folder.ParentID
	if req.Name != "" {
		name = req.Name
	}
	if req.ParentID != nil {
		parentID = req.ParentID
	}
	path := name
	if parentID != nil && *parentID != uuid.Nil {
		var parent models.VirtualFolder
		err := s.db.NewSelect().Model(&parent).Where("id = ? AND project_id = ?", *parentID, projectID).Scan(ctx)
		if err != nil {
			return nil, err
		}
		path = strings.TrimSuffix(parent.Path, "/") + "/" + name
	}
	folder.Name = name
	folder.ParentID = parentID
	folder.Path = path
	_, err = s.db.NewUpdate().Model(folder).Column("name", "parent_id", "path").Where("id = ? AND project_id = ?", folderID, projectID).Returning("*").Exec(ctx)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

func (s *FolderService) Delete(ctx context.Context, projectID, folderID uuid.UUID) error {
	_, err := s.db.NewUpdate().Model((*models.Resource)(nil)).Set("folder_id = NULL").Where("folder_id = ? AND project_id = ?", folderID, projectID).Exec(ctx)
	if err != nil {
		return err
	}
	res, err := s.db.NewDelete().Model((*models.VirtualFolder)(nil)).Where("id = ? AND project_id = ?", folderID, projectID).Exec(ctx)
	if err != nil {
		return err
	}
	rows, raErr := res.RowsAffected()
	if raErr != nil {
		return raErr
	}
	if rows == 0 {
		return fmt.Errorf("folder not found")
	}
	return nil
}
