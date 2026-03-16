package http

import (
	"context"

	"github.com/kirbyevanj/kvqtool-api-server/internal/domain"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/valyala/fasthttp"
)

type folderHandler struct {
	svc domain.FolderService
}

func (h *folderHandler) tree(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	tree, err := h.svc.Tree(context.TODO(), projectID)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, tree)
}

func (h *folderHandler) create(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var req types.CreateFolderRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		respondError(ctx, 400, "name is required")
		return
	}
	folder, err := h.svc.Create(context.TODO(), projectID, req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusCreated, folder)
}

func (h *folderHandler) update(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	folderID, err := parseUUIDParam(ctx, "folder_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var req types.UpdateFolderRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	folder, err := h.svc.Update(context.TODO(), projectID, folderID, req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, folder)
}

func (h *folderHandler) delete(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	folderID, err := parseUUIDParam(ctx, "folder_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.Delete(context.TODO(), projectID, folderID); err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusNoContent, nil)
}
