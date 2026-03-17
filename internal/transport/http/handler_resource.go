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

func (h *resourceHandler) register(ctx *fasthttp.RequestCtx) {
	pidStr, ok := parseUUID(ctx, "project_id")
	if !ok {
		respondError(ctx, 400, "invalid project_id")
		return
	}
	pid, err := uuid.Parse(pidStr)
	if err != nil {
		respondError(ctx, 400, "invalid project_id")
		return
	}

	var req struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		S3Key       string `json:"s3_key"`
	}
	if err := decodeJSON(ctx, &req); err != nil || req.S3Key == "" {
		respondError(ctx, 400, "filename, content_type, and s3_key required")
		return
	}

	resource, err := h.svc.Register(context.TODO(), pid, req.Filename, req.ContentType, req.S3Key)
	if err != nil {
		respondError(ctx, 500, err.Error())
		return
	}
	respondJSON(ctx, 201, resource)
}
