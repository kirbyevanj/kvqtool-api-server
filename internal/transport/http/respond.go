package http

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

func respondJSON(ctx *fasthttp.RequestCtx, status int, v any) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(status)
	if v != nil {
		data, err := json.Marshal(v)
		if err != nil {
			ctx.Error(`{"error":"marshal failed"}`, fasthttp.StatusInternalServerError)
			return
		}
		ctx.SetBody(data)
	}
}

func respondError(ctx *fasthttp.RequestCtx, status int, msg string) {
	respondJSON(ctx, status, map[string]string{"error": msg})
}

func decodeJSON(ctx *fasthttp.RequestCtx, dst any) error {
	return json.Unmarshal(ctx.PostBody(), dst)
}

func parseUUID(ctx *fasthttp.RequestCtx, param string) (string, bool) {
	v, ok := ctx.UserValue(param).(string)
	return v, ok && v != ""
}

func parseUUIDParam(ctx *fasthttp.RequestCtx, param string) (uuid.UUID, error) {
	s, ok := parseUUID(ctx, param)
	if !ok {
		return uuid.Nil, fmt.Errorf("missing %s", param)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s", param)
	}
	return id, nil
}
