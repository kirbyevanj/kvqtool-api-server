package http

import (
	"context"

	"github.com/kirbyevanj/kvqtool-api-server/internal/domain"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/valyala/fasthttp"
)

type projectHandler struct {
	svc domain.ProjectService
}

func (h *projectHandler) list(ctx *fasthttp.RequestCtx) {
	list, err := h.svc.List(context.TODO())
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, list)
}

func (h *projectHandler) create(ctx *fasthttp.RequestCtx) {
	var req types.CreateProjectRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	proj, err := h.svc.Create(context.TODO(), req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusCreated, proj)
}

func (h *projectHandler) get(ctx *fasthttp.RequestCtx) {
	id, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	proj, err := h.svc.Get(context.TODO(), id)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, proj)
}

func (h *projectHandler) update(ctx *fasthttp.RequestCtx) {
	id, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var req types.UpdateProjectRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	proj, err := h.svc.Update(context.TODO(), id, req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, proj)
}

func (h *projectHandler) delete(ctx *fasthttp.RequestCtx) {
	id, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	if err := h.svc.Delete(context.TODO(), id); err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusNoContent, nil)
}
