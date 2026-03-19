package http

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/google/uuid"
	"github.com/kirbyevanj/kvqtool-api-server/internal/domain"
	"github.com/kirbyevanj/kvqtool-kvq-models/types"
	"github.com/valyala/fasthttp"
)

type htmxHandler struct {
	projects  domain.ProjectService
	folders   domain.FolderService
	resources domain.ResourceService
	workflows domain.WorkflowService
}

func (h *htmxHandler) listProjects(ctx *fasthttp.RequestCtx) {
	projects, err := h.projects.List(context.TODO())
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString("<p>Error loading projects</p>")
		return
	}

	var b strings.Builder
	for _, p := range projects {
		b.WriteString(fmt.Sprintf(
			`<div class="project-row" onclick="window.location='/workspace.html?project=%s'">
				<div>
					<span class="name">%s</span>
					<span class="meta">%s — %d resources, %d jobs</span>
				</div>
				<div class="actions">
					<button class="btn btn-sm" onclick="event.stopPropagation()" hx-delete="/v1/projects/%s" hx-target="closest .project-row" hx-swap="outerHTML" hx-confirm="Delete project '%s'?">Delete</button>
				</div>
			</div>`,
			p.ID, html.EscapeString(p.Name), p.CreatedAt,
			p.ResourceCount, p.JobCount,
			p.ID, html.EscapeString(p.Name),
		))
	}

	if len(projects) == 0 {
		b.WriteString(`<p style="color:var(--text-dim)">No projects yet. Create one to get started.</p>`)
	}

	ctx.SetContentType("text/html")
	ctx.SetStatusCode(200)
	ctx.SetBodyString(b.String())
}

func (h *htmxHandler) createProject(ctx *fasthttp.RequestCtx) {
	name := string(ctx.FormValue("name"))
	desc := string(ctx.FormValue("description"))

	if name == "" {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("<p>Name is required</p>")
		return
	}

	_, err := h.projects.Create(context.TODO(), types.CreateProjectRequest{
		Name:        name,
		Description: desc,
	})
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBodyString(fmt.Sprintf("<p>Error: %s</p>", html.EscapeString(err.Error())))
		return
	}

	h.listProjects(ctx)
}

func (h *htmxHandler) sidebar(ctx *fasthttp.RequestCtx) {
	pidStr, ok := parseUUID(ctx, "project_id")
	if !ok {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("<p>Invalid project ID</p>")
		return
	}
	projectID, err := uuid.Parse(pidStr)
	if err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBodyString("<p>Invalid project ID</p>")
		return
	}

	folders, fErr := h.folders.Tree(context.TODO(), projectID)
	resources, rErr := h.resources.List(context.TODO(), projectID, nil, "")

	var b strings.Builder

	if fErr == nil {
		for _, f := range folders {
			renderFolderNode(&b, f, projectID.String())
		}
	}

	if rErr == nil {
		for _, r := range resources {
			if r.FolderID == nil {
				b.WriteString(renderResourceItem(r.ID.String(), r.ResourceType, r.Name))
			}
		}
	}

	if b.Len() == 0 {
		b.WriteString(`<p style="padding:12px;color:var(--text-dim)">No resources yet.</p>`)
	}

	ctx.SetContentType("text/html")
	ctx.SetStatusCode(200)
	ctx.SetBodyString(b.String())
}

func renderFolderNode(b *strings.Builder, f types.FolderNode, projectID string) {
	b.WriteString(fmt.Sprintf(
		`<div class="folder-item" hx-get="/htmx/projects/%s/folders/%s/contents" hx-target="next .folder-children" hx-swap="innerHTML">📁 %s</div><div class="folder-children">`,
		projectID, f.ID, html.EscapeString(f.Name),
	))
	for _, child := range f.Children {
		renderFolderNode(b, child, projectID)
	}
	b.WriteString(`</div>`)
}

func renderResourceItem(id, resType, name string) string {
	icon := resourceIcon(resType)
	escapedName := html.EscapeString(name)

	var typeActions string
	switch resType {
	case "media":
		typeActions = fmt.Sprintf(`<button onclick="onResourceClick('%s','media')">Play</button><button onclick="addToCompare('%s','%s')">Add to Compare</button>`, id, id, escapedName)
	case "report":
		typeActions = fmt.Sprintf(`<button onclick="viewReport('%s')">View Charts</button>`, id)
	case "file":
		if strings.HasSuffix(name, ".json") {
			typeActions = fmt.Sprintf(`<button onclick="viewReport('%s')">View Charts</button>`, id)
		}
	case "workflow":
		typeActions = fmt.Sprintf(`<button onclick="openWorkflow('%s')">Edit</button>`, id)
	}

	return fmt.Sprintf(
		`<div class="resource-item" data-id="%s" data-type="%s" draggable="true" ondragstart="onResDragStart(event,'%s','%s')" ondblclick="onResourceClick('%s','%s')" oncontextmenu="event.preventDefault();toggleResMenu('%s')">`+
			`<span class="res-label">%s %s</span>`+
			`<button class="res-menu-btn" onclick="event.stopPropagation();toggleResMenu('%s')">&#8230;</button>`+
			`<div class="res-menu" id="res-menu-%s" style="display:none">`+
			`%s`+
			`<button onclick="downloadResource('%s')">Download</button>`+
			`<button onclick="copyResource('%s','%s')">Duplicate</button>`+
			`<button onclick="renameResource('%s')">Rename</button>`+
			`<button onclick="deleteResource('%s')">Delete</button>`+
			`</div></div>`,
		id, resType,
		id, escapedName,
		id, resType,
		id,
		icon, escapedName,
		id,
		id,
		typeActions,
		id,
		id, escapedName,
		id,
		id,
	)
}

func resourceIcon(t string) string {
	switch t {
	case "media":
		return "🎬"
	case "workflow":
		return "⚙️"
	case "report":
		return "📊"
	default:
		return "📄"
	}
}

func (h *htmxHandler) workflowList(ctx *fasthttp.RequestCtx) {
	pidStr, ok := parseUUID(ctx, "project_id")
	if !ok {
		ctx.SetStatusCode(400)
		return
	}
	projectID, err := uuid.Parse(pidStr)
	if err != nil {
		ctx.SetStatusCode(400)
		return
	}

	workflows, err := h.workflows.List(context.TODO(), projectID)
	if err != nil {
		ctx.SetStatusCode(500)
		return
	}

	var b strings.Builder
	for _, w := range workflows {
		eName := html.EscapeString(w.Name)
		b.WriteString(fmt.Sprintf(
			`<div class="resource-item" data-id="%s" data-type="workflow" ondblclick="editWorkflow('%s')" oncontextmenu="event.preventDefault();toggleWfMenu('%s')">`+
				`<span class="res-label">⚙️ %s</span>`+
				`<button class="res-menu-btn" onclick="event.stopPropagation();toggleWfMenu('%s')">&#8230;</button>`+
				`<div class="res-menu" id="wf-menu-%s" style="display:none">`+
				`<button onclick="editWorkflow('%s')">Edit</button>`+
				`<button onclick="downloadWorkflow('%s')">Download</button>`+
				`<button onclick="copyWorkflow('%s')">Duplicate</button>`+
				`<button onclick="deleteWorkflow('%s','%s')">Delete</button>`+
				`</div></div>`,
			w.ID, w.ID, w.ID,
			eName,
			w.ID,
			w.ID,
			w.ID,
			w.ID,
			w.ID,
			w.ID, eName,
		))
	}

	if len(workflows) == 0 {
		b.WriteString(`<p style="padding:8px 12px;color:var(--text-muted);font-size:12px">No workflows yet.</p>`)
	}

	ctx.SetContentType("text/html")
	ctx.SetStatusCode(200)
	ctx.SetBodyString(b.String())
}
