package middleware

import "github.com/valyala/fasthttp"

type Middleware func(fasthttp.RequestHandler) fasthttp.RequestHandler

func Chain(h fasthttp.RequestHandler, mw ...Middleware) fasthttp.RequestHandler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}
