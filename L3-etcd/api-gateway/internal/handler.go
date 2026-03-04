package internal

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type HttpHandler struct {
	client Clients
}

func NewHttpHandler(client Clients) *HttpHandler {
	return &HttpHandler{
		client: client,
	}
}

func (h *HttpHandler) Register(c context.Context, ctx *app.RequestContext) {

}

func (h *HttpHandler) Login(c context.Context, ctx *app.RequestContext) {

}

func (h *HttpHandler) CreatePost(c context.Context, ctx *app.RequestContext) {}

func (h *HttpHandler) DeletePost(c context.Context, ctx *app.RequestContext) {}
