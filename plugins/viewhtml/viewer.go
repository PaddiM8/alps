package alpsviewhtml

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"strings"

	"git.sr.ht/~migadu/alps"
	alpsbase "git.sr.ht/~migadu/alps/plugins/base"
	"github.com/emersion/go-message"
)

const tplSrc = `
<!-- allow-same-origin is required to resize the frame with its content -->
<!-- allow-popups is required for target="_blank" links -->
<iframe id="email-frame" srcdoc="{{.}}" sandbox="allow-same-origin allow-popups"></iframe>
<script src="/plugins/viewhtml/assets/script.js"></script>
<link rel="stylesheet" href="/plugins/viewhtml/assets/style.css">
`

var tpl = template.Must(template.New("view-html.html").Parse(tplSrc))

type viewer struct{}

func (viewer) ViewMessagePart(ctx *alps.Context, msg *alpsbase.IMAPMessage, part *message.Entity) (interface{}, error) {
	allowRemoteResources := ctx.QueryParam("allow-remote-resources") == "1"

	mimeType, _, err := part.Header.ContentType()
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(mimeType, "text/html") {
		return nil, alpsbase.ErrViewUnsupported
	}

	body, err := ioutil.ReadAll(part.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read part body: %v", err)
	}

	hasRemoteResources := false
	if ctx.QueryParam("sanitize") != "0" {
		san := sanitizer{
			msg:                  msg,
			allowRemoteResources: allowRemoteResources,
		}
		body, err = san.sanitizeHTML(body)
		if err != nil {
			return nil, fmt.Errorf("failed to sanitize HTML part: %v", err)
		}

		hasRemoteResources = san.hasRemoteResources
	}

	ctx.Set("viewhtml.hasRemoteResources", hasRemoteResources)

	var buf bytes.Buffer
	err = tpl.Execute(&buf, string(body))
	if err != nil {
		return nil, err
	}

	return template.HTML(buf.String()), nil
}

func init() {
	alpsbase.RegisterViewer(viewer{})
}
