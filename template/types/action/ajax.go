package action

import (
	"html/template"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/template/types"
)

type AjaxAction struct {
	BaseAction
	Url         string
	Method      string
	Data        AjaxData
	Alert       bool
	AlertData   AlertData
	SuccessJS   template.JS
	ErrorJS     template.JS
	ParameterJS template.JS
	Event       Event
	Handlers    []context.Handler
}

type AlertData struct {
	Title              string `json:"title"`
	Type               string `json:"type"`
	ShowCancelButton   bool   `json:"showCancelButton"`
	ConfirmButtonColor string `json:"confirmButtonColor"`
	ConfirmButtonText  string `json:"confirmButtonText"`
	CloseOnConfirm     bool   `json:"closeOnConfirm"`
	CancelButtonText   string `json:"cancelButtonText"`
}

func Ajax(id string, handler types.Handler) *AjaxAction {
	if id == "" {
		panic("wrong ajax action parameter, empty id")
	}
	return &AjaxAction{
		Url:    URL(id),
		Method: "post",
		Data:   NewAjaxData(),
		SuccessJS: `swal(data.msg, '', data.code===0 ? 'success' : 'error');`,
		ErrorJS: `if(data.responseText !== "") {
				swal(data.responseJSON.msg, '', 'error');
			} else {
				swal('error', '', 'error');
			}`,
		Handlers: context.Handlers{handler.Wrap()},
		Event:    EventClick,
	}
}

func (ajax *AjaxAction) WithAlert(data ...AlertData) *AjaxAction {
	ajax.Alert = true
	if len(data) > 0 {
		ajax.AlertData = data[0]
	} else {
		ajax.AlertData = AlertData{
			Title:              "",
			Type:               "warning",
			ShowCancelButton:   true,
			ConfirmButtonColor: "#DD6B55",
			ConfirmButtonText:  language.Get("yes"),
			CloseOnConfirm:     false,
			CancelButtonText:   language.Get("cancel"),
		}
	}
	return ajax
}

func (ajax *AjaxAction) AddData(data map[string]interface{}) *AjaxAction {
	ajax.Data = ajax.Data.Add(data)
	return ajax
}

func (ajax *AjaxAction) SetData(data map[string]interface{}) *AjaxAction {
	ajax.Data = data
	return ajax
}

func (ajax *AjaxAction) SetUrl(url string) *AjaxAction {
	ajax.Url = url
	return ajax
}

func (ajax *AjaxAction) SetSuccessJS(successJS template.JS) *AjaxAction {
	ajax.SuccessJS = successJS
	return ajax
}

func (ajax *AjaxAction) ChangeHTMLWhenSuccess(identify string, text ...string) *AjaxAction {
	data := "data.data"
	if len(text) > 0 {
		if t := text[0]; len(t) > 5 && t[:5] == "data." {
			data = t
		} else {
			data = utils.StrConcat(`"`, t, `"`)
		}
	}
	if !strings.ContainsRune(identify, '$') {
		identify = utils.StrConcat(`$("`, identify, `")`)
	}
	ajax.SuccessJS = template.JS(utils.StrConcat(`
	if(data.code === 0) {
		if(`, identify, `.is("input")) {
			`, identify, `.val(`, data, `);
		} else if(`, identify, `.is("select")) {
			`, identify, `.val(`, data, `);
		} else {
			`, identify, `.html(`, data, `);
		}
	} else {
		swal(data.msg, '', 'error');
	}`))
	return ajax
}

func (ajax *AjaxAction) SetEvent(event Event) *AjaxAction {
	ajax.Event = event
	return ajax
}

func (ajax *AjaxAction) SetErrorJS(errorJS template.JS) *AjaxAction {
	ajax.ErrorJS = errorJS
	return ajax
}

func (ajax *AjaxAction) SetParameterJS(parameterJS template.JS) *AjaxAction {
	ajax.ParameterJS += parameterJS
	return ajax
}

func (ajax *AjaxAction) SetMethod(method string) *AjaxAction {
	ajax.Method = method
	return ajax
}

func (ajax *AjaxAction) GetCallbacks() context.Node {
	return context.Node{
		Path:     ajax.Url,
		Method:   ajax.Method,
		Handlers: ajax.Handlers,
		Value:    map[string]interface{}{ constant.ContextNodeNeedAuth: 1 },
	}
}

func (ajax *AjaxAction) Js() template.JS {
	ajaxStatement := utils.StrConcat(`$.ajax({
		method: '`, ajax.Method, `',
		url: "`, ajax.Url, `",
		data: data,
		success: function (data){`, string(ajax.SuccessJS), `},
		error: function(data){`, string(ajax.ErrorJS), `},
	});`)

	if ajax.Alert {
		b, _ := utils.JsonMarshal(ajax.AlertData)
		ajaxStatement = utils.StrConcat("swal(", string(b), `, function(){`, ajaxStatement, `});`)
	}

	return template.JS(utils.StrConcat(`$('`, ajax.BtnId, `').on('`, string(ajax.Event), `',function(event){
		let data = `, ajax.Data.JSON(), `;
		`, string(ajax.ParameterJS), `
		let id = $(this).attr("data-id");
		if(id) { data["id"] = id }
		`, ajaxStatement, `
	});`))
}

func (ajax *AjaxAction) BtnAttribute() template.HTML { return `href="javascript:;"` }
