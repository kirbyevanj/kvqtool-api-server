package middleware

import (
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const RequestIDHeader = "X-Request-ID"

func RequestID(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		id := string(ctx.Request.Header.Peek(RequestIDHeader))
		if id == "" {
			id = uuid.New().String()
		}
		ctx.Response.Header.Set(RequestIDHeader, id)
		ctx.SetUserValue("request_id", id)
		next(ctx)
	}
}
