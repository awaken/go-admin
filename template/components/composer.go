package components

import (
	"html/template"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	template2 "github.com/GoAdminGroup/go-admin/template"
)

func ComposeHtml(temList map[string]string, separation bool, compo interface{}, templateName ...string) template.HTML {
	tmplName := ""
	if len(templateName) > 0 {
		tmplName = templateName[0] + " "
	}

	var err  error
	tmpl := template.New("comp").Funcs(template2.DefaultFuncMap)

	if separation {
		files := make([]string, len(templateName))
		root := config.GetAssetRootPath() + "pages/"
		for i, v := range templateName {
			files[i] = root + temList["components/" + v] + ".tmpl"
		}
		tmpl, err = tmpl.ParseFiles(files...)
	} else {
		var sb strings.Builder
		sb.Grow(1024)
		for _, v := range templateName {
			sb.WriteString(temList["components/" + v])
		}
		tmpl, err = tmpl.Parse(sb.String())
	}

	if err != nil {
		logger.Panic(tmplName + "ComposeHtml Error:" + err.Error())
		return ""
	}

	var sb strings.Builder
	defineName := utils.TableFormReplacer.Replace(templateName[0])

	err = tmpl.ExecuteTemplate(&sb, defineName, compo)
	if err != nil {
		logger.Error(tmplName+" ComposeHtml Error:", err)
	}
	return template.HTML(sb.String())
}
