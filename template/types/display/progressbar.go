package display

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"html/template"
	"strconv"

	"github.com/GoAdminGroup/go-admin/template/types"
)

type ProgressBar struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("progressbar", new(ProgressBar))
}

func (p *ProgressBar) Get(args ...interface{}) types.FieldFilterFn {
	param := args[0].([]types.FieldProgressBarData)
	style := "primary"
	size  := "sm"
	max   := 100
	if len(param) > 0 {
		par := param[0]
		if par.Style != "" { style = par.Style }
		if par.Size  != "" { size  = par.Size  }
		if par.Max   != 0  { max   = par.Max   }
	}
	fMax := 100 / float64(max)
	sMax := strconv.Itoa(max)

	return func(value types.FieldModel) interface{} {
		base, _ := strconv.Atoi(value.Value)
		perc := fmt.Sprintf("%.0f", float64(base) * fMax)

		return template.HTML(utils.StrConcat(
		`<div class="row" style="min-width: 100px;">`+
				`<span class="col-sm-3" style="color:#777;width: 60px">`, perc, `%</span>`+
				`<div class="progress progress-`, size, ` col-sm-9" style="padding-left: 0;width: 100px;margin-left: -13px;">`+
					`<div class="progress-bar progress-bar-`, style, `" role="progressbar" aria-valuenow="1" aria-valuemin="0" aria-valuemax="`, sMax, `" style="width: `, perc, `%">`+
					`</div>`+
				`</div>`+
			`</div>`))
	}
}
