package controller

import (
	"html/template"

	"github.com/GoAdminGroup/go-admin/template/types"
)

var sysInfo SystemInfoData

type SystemInfoData struct {
	AppName    string
	AppVersion string
	AppMode    string
	AppRoles   string
	AppHost    string
	AppIp      string
	AppBuildAt string
	AppCommit  string
	AppEnv     string
}

func SetSystemInfoData(sid SystemInfoData) {
	sysInfo = sid
}

func sysInfoItemsForApplication() []map[string]types.InfoItem {
	return []map[string]types.InfoItem{
		{
			"key":   types.InfoItem{Content: lg("app_name")},
			"value": types.InfoItem{Content: template.HTML(sysInfo.AppName)},
		}, {
			"key":   types.InfoItem{Content: lg("app_version")},
			"value": types.InfoItem{Content: template.HTML(sysInfo.AppVersion)},
		}, {
			"key":   types.InfoItem{Content: lg("app_mode")},
			"value": types.InfoItem{Content: lg(template.HTML("mode_" + sysInfo.AppMode))},
		}, {
			"key":   types.InfoItem{Content: lg("app_build_at")},
			"value": types.InfoItem{Content: template.HTML(sysInfo.AppBuildAt)},
		}, {
			"key":   types.InfoItem{Content: lg("app_commit")},
			"value": types.InfoItem{Content: template.HTML(sysInfo.AppCommit)},
		}, {
			"key":   types.InfoItem{Content: lg("app_env")},
			"value": types.InfoItem{Content: lg(template.HTML("env_" + sysInfo.AppEnv))},
		}, {
			"key":   types.InfoItem{Content: lg("app_roles")},
			"value": types.InfoItem{Content: template.HTML(sysInfo.AppRoles)},
		}, {
			"key":   types.InfoItem{Content: lg("app_host_ip")},
			"value": types.InfoItem{Content: template.HTML(sysInfo.AppHost + " / " + sysInfo.AppIp)},
		},
	}
}
