package alpsviewmultipart

import (
	"fmt"
	"html/template"
	"strings"

	"git.sr.ht/~migadu/alps"
	alpsbase "git.sr.ht/~migadu/alps/plugins/base"
	"github.com/emersion/go-message"
)

type viewer struct{}

const tplMixedStr = "{{- range $i, $template := .Templates}}{{- if $i}}{{$.Divider}}{{end -}}{{$template}} {{- end}}"

var tplMixed *template.Template

func init() {
	tplMixed = template.Must(template.New("view-mixed.html").Parse(tplMixedStr))
}

type mixedRenderData struct {
	Templates []interface{}
	Divider   template.HTML
}

func viewFirst(ctx *alps.Context, msg *alpsbase.IMAPMessage, part *message.Entity) (interface{}, error) {
	var multipartReader = part.MultipartReader()
	defer multipartReader.Close()

	next, err := multipartReader.NextPart()
	if err != nil {
		return nil, fmt.Errorf("failed to read part body: %v", err)
	}

	return alpsbase.ViewMessagePart(ctx, msg, next)
}

func viewAlternative(ctx *alps.Context, msg *alpsbase.IMAPMessage, part *message.Entity) (interface{}, error) {
	var multipartReader = part.MultipartReader()
	defer multipartReader.Close()

	preferredContentType := ctx.QueryParam("preferredContentType")
	if preferredContentType == "" {
		return viewFirst(ctx, msg, part)
	}

	// TODO: For better reliability, traverse the tree rather than relying on convention.
	first, err := multipartReader.NextPart()
    rendered, _ := alpsbase.ViewMessagePart(ctx, msg, first)

	second, err := multipartReader.NextPart()
	if preferredContentType == "text/html" && err == nil {
		rendered, err = alpsbase.ViewMessagePart(ctx, msg, second)
	}

	return rendered, err
}

func viewMixed(ctx *alps.Context, msg *alpsbase.IMAPMessage, part *message.Entity) (interface{}, error) {
	var multipartReader = part.MultipartReader()
	defer multipartReader.Close()

	first, err := multipartReader.NextPart()
	if err != nil {
		return nil, err
	}

	if err != nil {
		return "", err
	}

	var data mixedRenderData
	if ctx.QueryParam("preferredContentType") == "text/html" {
		data.Divider = template.HTML("<hr>")
	} else {
		data.Divider = template.HTML("\n---\n")
	}

	for next := first; next != nil; next, _ = multipartReader.NextPart() {
		result, err := alpsbase.ViewMessagePart(ctx, msg, next)
		if err != nil {
			continue
		}

		data.Templates = append(data.Templates, result)
	}

	var sb strings.Builder
	tplMixed.ExecuteTemplate(&sb, "view-mixed.html", data)

	return template.HTML(sb.String()), nil
}

func (viewer) ViewMessagePart(ctx *alps.Context, msg *alpsbase.IMAPMessage, part *message.Entity) (interface{}, error) {
	mimeType, _, err := part.Header.ContentType()
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(mimeType, "multipart/alternative") {
		return viewAlternative(ctx, msg, part)
	}

	if strings.EqualFold(mimeType, "multipart/related") {
		return viewFirst(ctx, msg, part)
	}

	if strings.EqualFold(mimeType, "multipart/mixed") ||
		strings.EqualFold(mimeType, "multipart/report") {
		return viewMixed(ctx, msg, part)
	}

	return nil, alpsbase.ErrViewUnsupported
}

func init() {
	alpsbase.RegisterViewer(viewer{})
}
