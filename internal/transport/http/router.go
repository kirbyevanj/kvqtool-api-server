package http

import (
	"log/slog"

	"github.com/fasthttp/router"
	"github.com/kirbyevanj/kvqtool-api-server/internal/domain"
	"github.com/kirbyevanj/kvqtool-api-server/internal/storage"
	mw "github.com/kirbyevanj/kvqtool-api-server/internal/transport/http/middleware"
	"github.com/valyala/fasthttp"
)

type Server struct {
	router  *router.Router
	logger  *slog.Logger
	handler fasthttp.RequestHandler
}

func NewServer(
	logger *slog.Logger,
	projects domain.ProjectService,
	folders domain.FolderService,
	resources domain.ResourceService,
	workflows domain.WorkflowService,
	jobs domain.JobService,
	valkey *storage.ValkeyClient,
) *Server {
	r := router.New()
	s := &Server{router: r, logger: logger}

	ph := &projectHandler{svc: projects}
	fh := &folderHandler{svc: folders}
	rh := &resourceHandler{svc: resources}
	wh := &workflowHandler{svc: workflows}
	jh := &jobHandler{svc: jobs, valkey: valkey}

	r.GET("/healthz", s.healthHandler(valkey))

	v1 := r.Group("/v1")

	v1.GET("/projects", ph.list)
	v1.POST("/projects", ph.create)
	v1.GET("/projects/{project_id}", ph.get)
	v1.PUT("/projects/{project_id}", ph.update)
	v1.DELETE("/projects/{project_id}", ph.delete)

	v1.GET("/projects/{project_id}/folders", fh.tree)
	v1.POST("/projects/{project_id}/folders", fh.create)
	v1.PUT("/projects/{project_id}/folders/{folder_id}", fh.update)
	v1.DELETE("/projects/{project_id}/folders/{folder_id}", fh.delete)

	v1.GET("/projects/{project_id}/resources", rh.list)
	v1.POST("/projects/{project_id}/resources/upload-url", rh.uploadURL)
	v1.POST("/projects/{project_id}/resources/{resource_id}/confirm-upload", rh.confirmUpload)
	v1.GET("/projects/{project_id}/resources/{resource_id}/download-url", rh.downloadURL)
	v1.PUT("/projects/{project_id}/resources/{resource_id}", rh.update)
	v1.DELETE("/projects/{project_id}/resources/{resource_id}", rh.delete)

	v1.GET("/projects/{project_id}/workflows", wh.list)
	v1.POST("/projects/{project_id}/workflows", wh.create)
	v1.GET("/projects/{project_id}/workflows/{workflow_id}", wh.get)
	v1.PUT("/projects/{project_id}/workflows/{workflow_id}", wh.update)
	v1.DELETE("/projects/{project_id}/workflows/{workflow_id}", wh.delete)

	v1.POST("/projects/{project_id}/jobs", jh.create)
	v1.GET("/projects/{project_id}/jobs", jh.list)
	v1.GET("/projects/{project_id}/jobs/{job_id}", jh.get)
	v1.POST("/projects/{project_id}/jobs/{job_id}/cancel", jh.cancel)

	r.GET("/v1/jobs/{job_id}/ws", jh.wsProgress)

	hx := &htmxHandler{projects: projects, folders: folders, resources: resources}
	r.GET("/htmx/projects", hx.listProjects)
	r.POST("/htmx/projects", hx.createProject)
	r.GET("/htmx/projects/{project_id}/sidebar", hx.sidebar)

	s.handler = mw.Chain(r.Handler,
		mw.RequestID,
		mw.Logger(logger),
		mw.Recovery(logger),
		mw.CORS,
	)

	return s
}

func (s *Server) Handler() fasthttp.RequestHandler {
	return s.handler
}

func (s *Server) healthHandler(valkey *storage.ValkeyClient) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		valkeyStatus := "ok"
		if err := valkey.Ping(ctx); err != nil {
			valkeyStatus = "error"
		}
		respondJSON(ctx, fasthttp.StatusOK, map[string]string{
			"status": "ok",
			"valkey": valkeyStatus,
		})
	}
}
