package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-api-server/internal/domain"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/valyala/fasthttp"
)

type resourceHandler struct {
	svc domain.ResourceService
}

func (h *resourceHandler) list(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var folderID *uuid.UUID
	if s := string(ctx.QueryArgs().Peek("folder_id")); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			respondError(ctx, fasthttp.StatusBadRequest, "invalid folder_id")
			return
		}
		folderID = &id
	}
	resourceType := string(ctx.QueryArgs().Peek("type"))
	list, err := h.svc.List(context.TODO(), projectID, folderID, resourceType)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, list)
}

func (h *resourceHandler) uploadURL(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var req types.UploadURLRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	if req.Filename == "" {
		respondError(ctx, 400, "filename is required")
		return
	}
	resp, err := h.svc.GenerateUploadURL(context.TODO(), projectID, req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, resp)
}

func (h *resourceHandler) confirmUpload(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	resourceID, err := parseUUIDParam(ctx, "resource_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	resource, err := h.svc.ConfirmUpload(context.TODO(), projectID, resourceID)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, resource)
}

func (h *resourceHandler) downloadURL(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	resourceID, err := parseUUIDParam(ctx, "resource_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.svc.GenerateDownloadURL(context.TODO(), projectID, resourceID)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, resp)
}

func (h *resourceHandler) update(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	resourceID, err := parseUUIDParam(ctx, "resource_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var req types.UpdateResourceRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	resource, err := h.svc.Update(context.TODO(), projectID, resourceID, req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, resource)
}

func (h *resourceHandler) delete(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	resourceID, err := parseUUIDParam(ctx, "resource_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.Delete(context.TODO(), projectID, resourceID); err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusNoContent, nil)
}
