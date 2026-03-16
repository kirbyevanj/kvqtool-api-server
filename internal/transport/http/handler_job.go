package http

import (
	"context"
	"fmt"

	"github.com/fasthttp/websocket"
	"github.com/kirbyevanj/kvqtool-api-server/internal/domain"
	"github.com/kirbyevanj/kvqtool-api-server/internal/storage"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/valyala/fasthttp"
)

type jobHandler struct {
	svc    domain.JobService
	valkey *storage.ValkeyClient
}

var wsUpgrader = websocket.FastHTTPUpgrader{}

func (h *jobHandler) list(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	status := string(ctx.QueryArgs().Peek("status"))
	list, err := h.svc.List(context.TODO(), projectID, status)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, list)
}

func (h *jobHandler) create(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	var req types.CreateJobRequest
	if err := decodeJSON(ctx, &req); err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, "invalid json")
		return
	}
	resp, err := h.svc.Create(context.TODO(), projectID, req)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusAccepted, resp)
}

func (h *jobHandler) get(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	jobID, err := parseUUIDParam(ctx, "job_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	job, err := h.svc.Get(context.TODO(), projectID, jobID)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, job)
}

func (h *jobHandler) cancel(ctx *fasthttp.RequestCtx) {
	projectID, err := parseUUIDParam(ctx, "project_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	jobID, err := parseUUIDParam(ctx, "job_id")
	if err != nil {
		respondError(ctx, fasthttp.StatusBadRequest, err.Error())
		return
	}
	job, err := h.svc.Cancel(context.TODO(), projectID, jobID)
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(ctx, fasthttp.StatusOK, job)
}

func (h *jobHandler) wsProgress(ctx *fasthttp.RequestCtx) {
	jobIDStr, ok := parseUUID(ctx, "job_id")
	if !ok {
		respondError(ctx, fasthttp.StatusBadRequest, "missing job_id")
		return
	}
	channel := fmt.Sprintf("job:%s:progress", jobIDStr)

	err := wsUpgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer conn.Close()
		subCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		msgCh, cancelSub := h.valkey.Subscribe(subCtx, channel)
		defer cancelSub()
		for msg := range msgCh {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return
			}
		}
	})
	if err != nil {
		respondError(ctx, fasthttp.StatusInternalServerError, "websocket upgrade failed")
	}
}
