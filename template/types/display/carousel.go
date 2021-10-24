package display

import (
	"html/template"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/template/types"
)

type Carousel struct {
	types.BaseDisplayFnGenerator
}

func init() {
	types.RegisterDisplayFnGenerator("carousel", new(Carousel))
}

func (c *Carousel) Get(args ...interface{}) types.FieldFilterFn {
	size := args[1].([]int)
	fn   := args[0].(types.FieldGetImgArrFn)

	width  := "300"
	height := "200"
	switch len(size) {
	case 0:
	case 1:
		width = strconv.Itoa(size[0])
	default:
		_ = size[1]
		width  = strconv.Itoa(size[0])
		height = strconv.Itoa(size[1])
	}

	return func(value types.FieldModel) interface{} {
		var indicators, items strings.Builder
		indicators.Grow(512)
		items.Grow(512)

		for i, img := range fn(value.Value) {
			indicators.WriteString(`<li data-target="#carousel-value-`)
			indicators.WriteString(value.ID)
			indicators.WriteString(`" data-slide-to="`)
			indicators.WriteString(strconv.Itoa(i))
			indicators.WriteString(`" class=""></li>`)
			items.WriteString(`<div class="item`)
			if i == 0 { items.WriteString(" active") }
			items.WriteString(`"><img src="`)
			items.WriteString(img)
			items.WriteString(`" alt="" style="max-width:`)
			items.WriteString(width)
			items.WriteString(`px;max-height:`)
			items.WriteString(height)
			items.WriteString(`px;display: block;margin-left: auto;margin-right: auto;" /><div class="carousel-caption"></div></div>`)
		}

		return template.HTML(utils.StrConcat(``+
			`<div id="carousel-value-`, value.ID, `" class="carousel slide" data-ride="carousel" width="`, width, `" height="`, height,
				`" style="padding: 5px;border: 1px solid #f4f4f4;background-color:white;width:`, width, `px;">`+
				`<ol class="carousel-indicators">`, indicators.String(), `</ol>`+
				`<div class="carousel-inner">`, items.String(), `</div>`+
				`<a class="left carousel-control" href="#carousel-value-`, value.ID, `" data-slide="prev"><span class="fa fa-angle-left"></span></a>`+
				`<a class="right carousel-control" href="#carousel-value-`, value.ID, `" data-slide="next"><span class="fa fa-angle-right"></span></a>`+
			`</div>`))
	}
}
