package guard

import (
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
)

type ExportParam struct {
	Panel  table.Table
	Id     []string
	Prefix string
	IsAll  bool
}

func (g *Guard) Export(ctx *context.Context) {
	panel, prefix := g.table(ctx)
	if !panel.GetExportable() {
		alert(ctx, panel, errors.OperationNotAllow, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	var ids []string
	if id := ctx.FormValue("id"); id != "" {
		ids = strings.Split(id, ",")
	}

	ctx.SetUserValue(exportParamKey, &ExportParam{
		Panel:  panel,
		Id:     ids,
		Prefix: prefix,
		IsAll:  ctx.FormValue("is_all") == "true",
	})
	ctx.Next()
}

func GetExportParam(ctx *context.Context) *ExportParam {
	return ctx.UserValue[exportParamKey].(*ExportParam)
}
