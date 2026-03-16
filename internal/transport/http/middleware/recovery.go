package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/valyala/fasthttp"
)

func Recovery(log *slog.Logger) Middleware {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			defer func() {
				if r := recover(); r != nil {
					log.Error("panic recovered",
						"err", fmt.Sprint(r),
						"stack", string(debug.Stack()),
						"path", string(ctx.Path()),
					)
					ctx.Error("internal server error", fasthttp.StatusInternalServerError)
				}
			}()
			next(ctx)
		}
	}
}
