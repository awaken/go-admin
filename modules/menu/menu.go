// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package menu

import (
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"html/template"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
)

// Item is an menu item.
type Item struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	Url          string `json:"url"`
	IsLinkUrl    bool   `json:"isLinkUrl"`
	Icon         string `json:"icon"`
	Header       string `json:"header"`
	Active       string `json:"active"`
	ChildrenList []Item `json:"childrenList"`
}

// Menu contains list of menu items and other info.
type Menu struct {
	List        []Item              `json:"list"`
	Options     []map[string]string `json:"options"`
	MaxOrder    int64               `json:"maxOrder"`
	PluginName  string              `json:"pluginName"`
	ForceUpdate bool                `json:"forceUpdate"`
}

func (menu *Menu) GetUpdateJS(updateFlag bool) template.JS {
	if !updateFlag {
		return ""
	}
	forceUpdate := "false"
	if menu.ForceUpdate {
		forceUpdate = "true"
	}
	return template.JS(utils.StrConcat(`$(function () {
	let curMenuPlug = $(".main-sidebar section.sidebar ul.sidebar-menu").attr("data-plug");
	if(curMenuPlug !== '`, menu.PluginName, `' || `, forceUpdate, `) {
		$(".main-sidebar section.sidebar").html($("#sidebar-menu-tmpl").html())
	}
});`))
}

// SetMaxOrder set the max order of menu.
func (menu *Menu) SetMaxOrder(order int64) {
	menu.MaxOrder = order
}

// AddMaxOrder add the max order of menu.
func (menu *Menu) AddMaxOrder() {
	menu.MaxOrder++
}

// SetActiveClass set the active class of menu.
func (menu *Menu) SetActiveClass(path string) *Menu {
	path = utils.RexMenuActiveClass.ReplaceAllString(path, "")

	for i, item := range menu.List {
		item.Active = ""
		for j, child := range item.ChildrenList {
			child.Active = ""
			item.ChildrenList[j] = child
		}
		menu.List[i] = item
	}

	for i, item := range menu.List {
		if item.Url == path && len(item.ChildrenList) == 0 {
			item.Active  = "active"
			menu.List[i] = item
			break
		}

		for j, child := range item.ChildrenList {
			if child.Url == path {
				item.Active = "active"
				child.Active = "active"
				item.ChildrenList[j] = child
				menu.List[i] = item
				return menu
			}
		}
	}

	return menu
}

// FormatPath get template.HTML for front-end.
func (menu Menu) FormatPath() template.HTML {
	var sb strings.Builder
	sb.Grow(1024)
	//res := template.HTML(``)
	for _, l := range menu.List {
		if l.Active != "" {
			if l.Url != "#" && l.Url != "" && len(l.ChildrenList) > 0 {
				sb.WriteString(`<li><a href="`)
				sb.WriteString(l.Url)
				sb.WriteString(`">`)
				sb.WriteString(l.Name)
				sb.WriteString(`</a></li>`)
				//res += template.HTML(`<li><a href="` + l.Url + `">` + l.Name + `</a></li>`)
			} else {
				sb.WriteString(`<li>`)
				sb.WriteString(l.Name)
				sb.WriteString(`</li>`)
				//res += template.HTML(`<li>` + l.Name + `</li>`)
				if len(l.ChildrenList) == 0 {
					break
				}
			}
			for _, c := range l.ChildrenList {
				if c.Active != "" {
					sb.WriteString(`<li>`)
					sb.WriteString(c.Name)
					sb.WriteString(`</li>`)
					return template.HTML(sb.String())
					//return res + template.HTML(`<li>`+c.Name+`</li>`)
				}
			}
		}
	}
	return template.HTML(sb.String())
	//return res
}

// GetEditMenuList return menu items list.
func (menu *Menu) GetEditMenuList() []Item {
	return menu.List
}

type NewMenuData struct {
	ParentId   int64  `json:"parent_id"`
	Type       int64  `json:"type"`
	Order      int64  `json:"order"`
	Title      string `json:"title"`
	Icon       string `json:"icon"`
	PluginName string `json:"plugin_name"`
	Uri        string `json:"uri"`
	Header     string `json:"header"`
	Uuid       string `json:"uuid"`
}

func NewMenu(conn db.Connection, data NewMenuData) (int64, error) {
	maxOrder := data.Order
	checkOrder, _ := db.WithDriver(conn).Table("goadmin_menu").
		Where("plugin_name", "=", data.PluginName).
		OrderBy("order", "desc").
		First()

	if checkOrder != nil {
		maxOrder = checkOrder["order"].(int64)
	}

	id, err := db.WithDriver(conn).Table("goadmin_menu").
		Insert(dialect.H{
			"parent_id":   data.ParentId,
			"type":        data.Type,
			"order":       maxOrder,
			"title":       data.Title,
			"uuid":        data.Uuid,
			"icon":        data.Icon,
			"plugin_name": data.PluginName,
			"uri":         data.Uri,
			"header":      data.Header,
		})
	if !db.CheckError(err, db.INSERT) {
		return id, nil
	}
	return id, err
}

// GetGlobalMenu return Menu of given user model.
func GetGlobalMenu(user models.UserModel, conn db.Connection, lang string, pluginNames ...string) *Menu {
	var (
		menus    []map[string]interface{}
		plugName string
	)
	if len(pluginNames) > 0 {
		plugName = pluginNames[0]
	}

	user.WithRoles().WithMenus()

	if user.IsSuperAdmin() {
		menus, _ = db.WithDriver(conn).Table("goadmin_menu").
			Where("id", ">", 0).
			Where("plugin_name", "=", plugName).
			OrderBy("order", "asc").
			All()
	} else {
		ids := make([]interface{}, len(user.MenuIds))
		for i, id := range user.MenuIds {
			ids[i] = id
		}
		menus, _ = db.WithDriver(conn).Table("goadmin_menu").
			WhereIn("id", ids).
			Where("plugin_name", "=", plugName).
			OrderBy("order", "asc").
			All()
	}

	menuOptions := make([]map[string]string, len(menus))
	for i, menu := range menus {
		menuOptions[i] = map[string]string{
			"id"   : strconv.Itoa(int(menu["id"].(int64))),
			"title": language.GetWithLang(menu["title"].(string), lang),
		}
	}

	if lang != "" {
		lang = "__ga_lang=" + lang
	}
	menuList := constructMenuTree(menus, 0, lang)
	maxOrder := int64(0)
	if len(menus) > 0 {
		maxOrder = menus[len(menus)-1]["parent_id"].(int64)
	}

	return &Menu{
		List:       menuList,
		Options:    menuOptions,
		MaxOrder:   maxOrder,
		PluginName: plugName,
	}
}

func constructMenuTree(menus []map[string]interface{}, parentID int64, langParam string) []Item {
	branch := make([]Item, 0, len(menus))

	for _, menu := range menus {
		if parentID == menu["parent_id"].(int64) {
			var title string
			if menu["type"].(int64) == 1 {
				title = language.Get(menu["title"].(string))
			} else {
				title = menu["title"].(string)
			}

			menuId    := menu["id"].(int64)
			header, _ := menu["header"].(string)
			uri       := menu["uri"].(string)

			if langParam != "" {
				var sb strings.Builder
				sb.Grow(len(uri) + 1 + len(langParam))
				sb.WriteString(uri)
				if strings.ContainsRune(uri, '?') {
					sb.WriteByte('&')
				} else {
					sb.WriteByte('?')
				}
				sb.WriteString(langParam)
				uri = sb.String()
			}

			branch = append(branch, Item{
				Name:         title,
				ID:           strconv.Itoa(int(menuId)),
				Url:          uri,
				Icon:         menu["icon"].(string),
				Header:       header,
				ChildrenList: constructMenuTree(menus, menuId, langParam),
			})
		}
	}

	return branch
}
