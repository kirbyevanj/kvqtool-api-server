package middleware

import (
	"log/slog"
	"time"

	"github.com/valyala/fasthttp"
)

func Logger(log *slog.Logger) Middleware {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			start := time.Now()
			next(ctx)

			reqID, _ := ctx.UserValue("request_id").(string)
			log.Info("request",
				"method", string(ctx.Method()),
				"path", string(ctx.Path()),
				"status", ctx.Response.StatusCode(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", reqID,
			)
		}
	}
}
