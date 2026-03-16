package http

import (
	"context"

	"github.com/kirbyevanj/kvqtool-api-server/internal/domain"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/valyala/fasthttp"
)

type workflowHandler struct {
	svc domain.WorkflowService
}

func (h *workflowHandler) list(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	list, err := h.svc.List(context.TODO(), projectID)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, list)
}

func (h *workflowHandler) create(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var req types.CreateWorkflowRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" {
		respondError(ctx, 400, "name is required")
		return
	}
	wf, err := h.svc.Create(context.TODO(), projectID, req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusCreated, wf)
}

func (h *workflowHandler) get(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	workflowID, err := parseUUIDParam(ctx, "workflow_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	wf, err := h.svc.Get(context.TODO(), projectID, workflowID)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, wf)
}

func (h *workflowHandler) update(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	workflowID, err := parseUUIDParam(ctx, "workflow_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var req types.UpdateWorkflowRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	wf, err := h.svc.Update(context.TODO(), projectID, workflowID, req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, wf)
}

func (h *workflowHandler) delete(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	workflowID, err := parseUUIDParam(ctx, "workflow_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.Delete(context.TODO(), projectID, workflowID); err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusNoContent, nil)
}
