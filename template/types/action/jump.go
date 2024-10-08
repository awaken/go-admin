package action

import (
	"html/template"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/utils"
)

type JumpAction struct {
	BaseAction
	Url         string
	Target      string
	Ext         template.HTML
	NewTabTitle string
}

func Jump(url string, ext ...template.HTML) *JumpAction {
	url = utils.JumpTmplReplacer.Replace(url)
	if len(ext) > 0 {
		return &JumpAction{ Url: url, Ext: ext[0] }
	}
	return &JumpAction{ Url: url }
}

func JumpInNewTab(url, title string, ext ...template.HTML) *JumpAction {
	url = utils.JumpTmplReplacer.Replace(url)
	if len(ext) > 0 {
		return &JumpAction{ Url: url, NewTabTitle: title, Ext: ext[0] }
	}
	return &JumpAction{ Url: url, NewTabTitle: title }
}

func JumpWithTarget(url, target string, ext ...template.HTML) *JumpAction {
	url = utils.JumpTmplReplacer.Replace(url)
	if len(ext) > 0 {
		return &JumpAction{ Url: url, Target: target, Ext: ext[0] }
	}
	return &JumpAction{ Url: url, Target: target }
}

func (jump *JumpAction) GetCallbacks() context.Node {
	return context.Node{ Path: jump.Url, Method: "GET" }
}

func (jump *JumpAction) BtnAttribute() template.HTML {
	var sb strings.Builder
	sb.Grow(256)
	sb.WriteString(`href="`)
	sb.WriteString(jump.Url)
	sb.WriteByte('"')
	if jump.NewTabTitle != "" {
		sb.WriteString(` data-title="`)
		sb.WriteString(jump.NewTabTitle)
		sb.WriteByte('"')
	}
	if jump.Target != "" {
		sb.WriteString(` target="`)
		sb.WriteString(jump.Target)
		sb.WriteByte('"')
	}
	return template.HTML(sb.String())
	/*html := template.HTML(`href="` + jump.Url + `"`)
	if jump.NewTabTitle != "" {
		html += template.HTML(` data-title="` + jump.NewTabTitle + `"`)
	}
	if jump.Target != "" {
		html += template.HTML(` target="` + jump.Target + `"`)
	}
	return html*/
}

func (jump *JumpAction) BtnClass() template.HTML {
	if jump.NewTabTitle == "" { return "" }
	return "new-tab-link"
}

func (jump *JumpAction) ExtContent() template.HTML {
	return jump.Ext
}
