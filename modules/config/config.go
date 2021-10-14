// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
)

// Database is a type of database connection config.
//
// Because a little difference of different database driver.
// The Config has multiple options but may not be used.
// Such as the sqlite driver only use the File option which
// can be ignored when the driver is mysql.
//
// If the Dsn is configured, when driver is mysql/postgresql/
// mssql, the other configurations will be ignored, except for
// MaxIdleCon and MaxOpenCon.
type Database struct {
	Host       string            `json:"host,omitempty" yaml:"host,omitempty" ini:"host,omitempty"`
	Port       string            `json:"port,omitempty" yaml:"port,omitempty" ini:"port,omitempty"`
	User       string            `json:"user,omitempty" yaml:"user,omitempty" ini:"user,omitempty"`
	Pwd        string            `json:"pwd,omitempty" yaml:"pwd,omitempty" ini:"pwd,omitempty"`
	Name       string            `json:"name,omitempty" yaml:"name,omitempty" ini:"name,omitempty"`
	MaxIdleCon int               `json:"max_idle_con,omitempty" yaml:"max_idle_con,omitempty" ini:"max_idle_con,omitempty"`
	MaxOpenCon int               `json:"max_open_con,omitempty" yaml:"max_open_con,omitempty" ini:"max_open_con,omitempty"`
	Driver     string            `json:"driver,omitempty" yaml:"driver,omitempty" ini:"driver,omitempty"`
	DriverMode string            `json:"driver_mode,omitempty" yaml:"driver_mode,omitempty" ini:"driver_mode,omitempty"`
	File       string            `json:"file,omitempty" yaml:"file,omitempty" ini:"file,omitempty"`
	Dsn        string            `json:"dsn,omitempty" yaml:"dsn,omitempty" ini:"dsn,omitempty"`
	Params     map[string]string `json:"params,omitempty" yaml:"params,omitempty" ini:"params,omitempty"`
}

func (d Database) GetDSN() string {
	if d.Dsn != "" { return d.Dsn }
	switch d.Driver {
	case DriverMysql:
		return d.User + ":" + d.Pwd + "@tcp(" + d.Host + ":" + d.Port + ")/" +
			d.Name + d.ParamStr()
	case DriverPostgresql:
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s" + d.ParamStr(),
			d.Host, d.Port, d.User, d.Pwd, d.Name)
	case DriverMssql:
		return fmt.Sprintf("user id=%s;password=%s;server=%s;port=%s;database=%s;" + d.ParamStr(),
			d.User, d.Pwd, d.Host, d.Port, d.Name)
	case DriverSqlite:
		return d.File + d.ParamStr()
	}
	return ""
}

func (d Database) ParamStr() string {
	p := ""
	if d.Params == nil { d.Params = make(map[string]string) }
	switch d.Driver {
	case DriverMysql:
		if _, ok := d.Params["charset"]; !ok {
			d.Params["charset"] = "utf8mb4"
		}
		fallthrough
	case DriverSqlite:
		if len(d.Params) > 0 {
			p = "?"
			for k, v := range d.Params {
				p += k + "=" + v + "&"
			}
			p = p[:len(p)-1]
		}
	case DriverMssql:
		if _, ok := d.Params["encrypt"]; !ok {
			d.Params["encrypt"] = "disable"
		}
		for k, v := range d.Params {
			p += k + "=" + v + ";"
		}
		p = p[:len(p)-1]
	case DriverPostgresql:
		if _, ok := d.Params["sslmode"]; !ok {
			d.Params["sslmode"] = "disable"
		}
		p = " "
		for k, v := range d.Params {
			p += k + "=" + v + " "
		}
		p = p[:len(p)-1]
	}
	return p
}

// DatabaseList is a map of Database.
type DatabaseList map[string]Database

// GetDefault get the default Database.
func (d DatabaseList) GetDefault() Database {
	return d["default"]
}

// Add add a Database to the DatabaseList.
func (d DatabaseList) Add(key string, db Database) {
	d[key] = db
}

// GroupByDriver group the Databases with the drivers.
func (d DatabaseList) GroupByDriver() map[string]DatabaseList {
	drivers := make(map[string]DatabaseList)
	for key, item := range d {
		if driverList, ok := drivers[item.Driver]; ok {
			driverList.Add(key, item)
		} else {
			drivers[item.Driver] = make(DatabaseList)
			drivers[item.Driver].Add(key, item)
		}
	}
	return drivers
}

func (d DatabaseList) JSON() string {
	return utils.JSON(d)
}

func (d DatabaseList) Copy() DatabaseList {
	res := make(DatabaseList)
	for k, v := range d { res[k] = v }
	return res
}

func (d DatabaseList) Connections() []string {
	conns := make([]string, len(d))
	i     := 0
	for key := range d {
		conns[i] = key
		i++
	}
	return conns
}

func GetDatabaseListFromJSON(m string) DatabaseList {
	var d = make(DatabaseList)
	if m == "" { panic("wrong config") }
	_ = utils.JsonUnmarshal([]byte(m), &d)
	return d
}

const (
	// EnvDev is a const value of development environment.
	EnvDev = "dev"
	// EnvTest is a const value of test environment.
	EnvTest = "test"
	// EnvLocal is a const value of local environment.
	EnvLocal = "local"
	// EnvProd is a const value of production environment.
	EnvProd = "prod"

	// DriverMysql is a const value of mysql driver.
	DriverMysql = "mysql"
	// DriverSqlite is a const value of sqlite driver.
	DriverSqlite = "sqlite"
	// DriverPostgresql is a const value of postgresql driver.
	DriverPostgresql = "postgresql"
	// DriverMssql is a const value of mssql driver.
	DriverMssql = "mssql"
)

// Store is the file store config. Path is the local store path.
// and prefix is the url prefix used to visit it.
type Store struct {
	Path   string `json:"path,omitempty" yaml:"path,omitempty" ini:"path,omitempty"`
	Prefix string `json:"prefix,omitempty" yaml:"prefix,omitempty" ini:"prefix,omitempty"`
}

func (s Store) URL(suffix string) string {
	if len(suffix) > 4 && suffix[:4] == "http" {
		return suffix
	}
	if s.Prefix == "" {
		if suffix[0] == '/' {
			return suffix
		}
		return "/" + suffix
	}
	if s.Prefix[0] == '/' {
		if suffix[0] == '/' {
			return s.Prefix + suffix
		}
		return s.Prefix + "/" + suffix
	}
	if suffix[0] == '/' {
		if len(s.Prefix) > 4 && s.Prefix[:4] == "http" {
			return s.Prefix + suffix
		}
		return "/" + s.Prefix + suffix
	}
	if len(s.Prefix) > 4 && s.Prefix[:4] == "http" {
		return s.Prefix + "/" + suffix
	}
	return "/" + s.Prefix + "/" + suffix
}

func (s Store) JSON() string {
	if s.Path == "" && s.Prefix == "" { return "" }
	return utils.JSON(s)
}

func GetStoreFromJSON(m string) Store {
	var s Store
	if m == "" { return s }
	_ = utils.JsonUnmarshal([]byte(m), &s)
	return s
}

// Config type is the global config of goAdmin. It will be
// initialized in the engine.
type Config struct {
	// An map supports multi database connection. The first
	// element of Databases is the default connection. See the
	// file connection.go.
	Databases DatabaseList `json:"database,omitempty" yaml:"database,omitempty" ini:"database,omitempty"`

	// The cookie domain used in the auth modules. see
	// the session.go.
	Domain string `json:"domain,omitempty" yaml:"domain,omitempty" ini:"domain,omitempty"`

	// Used to set as the localize language which show in the
	// interface.
	Language string `json:"language,omitempty" yaml:"language,omitempty" ini:"language,omitempty"`

	// The global url prefix.
	UrlPrefix string `json:"prefix,omitempty" yaml:"prefix,omitempty" ini:"prefix,omitempty"`

	// The theme name of template.
	Theme string `json:"theme,omitempty" yaml:"theme,omitempty" ini:"theme,omitempty"`

	// The path where files will be stored into.
	Store Store `json:"store,omitempty" yaml:"store,omitempty" ini:"store,omitempty"`

	// The title of web page.
	Title string `json:"title,omitempty" yaml:"title,omitempty" ini:"title,omitempty"`

	// Logo is the top text in the sidebar.
	Logo template.HTML `json:"logo,omitempty" yaml:"logo,omitempty" ini:"logo,omitempty"`

	// Mini-logo is the top text in the sidebar when folding.
	MiniLogo template.HTML `json:"mini_logo,omitempty" yaml:"mini_logo,omitempty" ini:"mini_logo,omitempty"`

	// The url redirect to after login.
	IndexUrl string `json:"index,omitempty" yaml:"index,omitempty" ini:"index,omitempty"`

	// Login page URL
	LoginUrl string `json:"login_url,omitempty" yaml:"login_url,omitempty" ini:"login_url,omitempty"`

	// Debug mode
	Debug bool `json:"debug,omitempty" yaml:"debug,omitempty" ini:"debug,omitempty"`

	// Env is the environment,which maybe local,test,prod.
	Env string `json:"env,omitempty" yaml:"env,omitempty" ini:"env,omitempty"`

	// Access assets log off
	AccessAssetsLogOff bool `json:"access_assets_log_off,omitempty" yaml:"access_assets_log_off,omitempty" ini:"access_assets_log_off,omitempty"`

	// Sql operator record log switch.
	SqlLog bool `json:"sql_log,omitempty" yaml:"sql_log,omitempty" ini:"sql_log,omitempty"`

	AccessLogOff bool `json:"access_log_off,omitempty" yaml:"access_log_off,omitempty" ini:"access_log_off,omitempty"`
	InfoLogOff   bool `json:"info_log_off,omitempty" yaml:"info_log_off,omitempty" ini:"info_log_off,omitempty"`
	ErrorLogOff  bool `json:"error_log_off,omitempty" yaml:"error_log_off,omitempty" ini:"error_log_off,omitempty"`

	Logger Logger `json:"logger,omitempty" yaml:"logger,omitempty" ini:"logger,omitempty"`

	// Color scheme.
	ColorScheme string `json:"color_scheme,omitempty" yaml:"color_scheme,omitempty" ini:"color_scheme,omitempty"`

	// Session valid time duration,units are seconds. Default 7200.
	SessionLifeTime int `json:"session_life_time,omitempty" yaml:"session_life_time,omitempty" ini:"session_life_time,omitempty"`

	// Assets visit link.
	AssetUrl string `json:"asset_url,omitempty" yaml:"asset_url,omitempty" ini:"asset_url,omitempty"`

	// File upload engine,default "local"
	FileUploadEngine FileUploadEngine `json:"file_upload_engine,omitempty" yaml:"file_upload_engine,omitempty" ini:"file_upload_engine,omitempty"`

	// Custom html in the tag head.
	CustomHeadHtml template.HTML `json:"custom_head_html,omitempty" yaml:"custom_head_html,omitempty" ini:"custom_head_html,omitempty"`

	// Custom html after body.
	CustomFootHtml template.HTML `json:"custom_foot_html,omitempty" yaml:"custom_foot_html,omitempty" ini:"custom_foot_html,omitempty"`

	// Footer Info html
	FooterInfo template.HTML `json:"footer_info,omitempty" yaml:"footer_info,omitempty" ini:"footer_info,omitempty"`

	// Login page title
	LoginTitle string `json:"login_title,omitempty" yaml:"login_title,omitempty" ini:"login_title,omitempty"`

	// Login page logo
	LoginLogo template.HTML `json:"login_logo,omitempty" yaml:"login_logo,omitempty" ini:"login_logo,omitempty"`

	// Auth user table
	AuthUserTable string `json:"auth_user_table,omitempty" yaml:"auth_user_table,omitempty" ini:"auth_user_table,omitempty"`

	// Extra config info
	Extra ExtraInfo `json:"extra,omitempty" yaml:"extra,omitempty" ini:"extra,omitempty"`

	// Page animation
	Animation PageAnimation `json:"animation,omitempty" yaml:"animation,omitempty" ini:"animation,omitempty"`

	// Limit login with different IPs
	NoLimitLoginIP bool `json:"no_limit_login_ip,omitempty" yaml:"no_limit_login_ip,omitempty" ini:"no_limit_login_ip,omitempty"`

	// When site off is true, website will be closed
	SiteOff bool `json:"site_off,omitempty" yaml:"site_off,omitempty" ini:"site_off,omitempty"`

	// Hide config center entrance flag
	HideConfigCenterEntrance bool `json:"hide_config_center_entrance,omitempty" yaml:"hide_config_center_entrance,omitempty" ini:"hide_config_center_entrance,omitempty"`

	// Prohibit config modification
	ProhibitConfigModification bool `json:"prohibit_config_modification,omitempty" yaml:"prohibit_config_modification,omitempty" ini:"prohibit_config_modification,omitempty"`

	// Hide tool entrance flag
	HideToolEntrance bool `json:"hide_tool_entrance,omitempty" yaml:"hide_tool_entrance,omitempty" ini:"hide_tool_entrance,omitempty"`

	HidePluginEntrance bool `json:"hide_plugin_entrance,omitempty" yaml:"hide_plugin_entrance,omitempty" ini:"hide_plugin_entrance,omitempty"`

	Custom404HTML template.HTML `json:"custom_404_html,omitempty" yaml:"custom_404_html,omitempty" ini:"custom_404_html,omitempty"`

	Custom403HTML template.HTML `json:"custom_403_html,omitempty" yaml:"custom_403_html,omitempty" ini:"custom_403_html,omitempty"`

	Custom500HTML template.HTML `json:"custom_500_html,omitempty" yaml:"custom_500_html,omitempty" ini:"custom_500_html,omitempty"`

	// Update Process Function
	UpdateProcessFn UpdateConfigProcessFn `json:"-" yaml:"-" ini:"-"`

	// Favicon string `json:"favicon,omitempty" yaml:"favicon,omitempty" ini:"favicon,omitempty"`

	// Is open admin plugin json api
	OpenAdminApi bool `json:"open_admin_api,omitempty" yaml:"open_admin_api,omitempty" ini:"open_admin_api,omitempty"`

	HideVisitorUserCenterEntrance bool `json:"hide_visitor_user_center_entrance,omitempty" yaml:"hide_visitor_user_center_entrance,omitempty" ini:"hide_visitor_user_center_entrance,omitempty"`

	ExcludeThemeComponents []string `json:"exclude_theme_components,omitempty" yaml:"exclude_theme_components,omitempty" ini:"exclude_theme_components,omitempty"`

	OperationLogOff bool `json:"operation_log_off,omitempty" yaml:"operation_log_off,omitempty" ini:"operation_log_off,omitempty"`

	AssetRootPath string `json:"asset_root_path,omitempty" yaml:"asset_root_path,omitempty" ini:"asset_root_path,omitempty"`

	URLFormat URLFormat `json:"url_format,omitempty" yaml:"url_format,omitempty" ini:"url_format,omitempty"`

	prefix string
	//lock   sync.RWMutex
}

type Logger struct {
	Encoder EncoderCfg `json:"encoder,omitempty" yaml:"encoder,omitempty" ini:"encoder,omitempty"`
	Level   int8       `json:"level,omitempty" yaml:"level,omitempty" ini:"level,omitempty"`
}

type EncoderCfg struct {
	TimeKey       string `json:"time_key,omitempty" yaml:"time_key,omitempty" ini:"time_key,omitempty"`
	LevelKey      string `json:"level_key,omitempty" yaml:"level_key,omitempty" ini:"level_key,omitempty"`
	NameKey       string `json:"name_key,omitempty" yaml:"name_key,omitempty" ini:"name_key,omitempty"`
	CallerKey     string `json:"caller_key,omitempty" yaml:"caller_key,omitempty" ini:"caller_key,omitempty"`
	MessageKey    string `json:"message_key,omitempty" yaml:"message_key,omitempty" ini:"message_key,omitempty"`
	StacktraceKey string `json:"stacktrace_key,omitempty" yaml:"stacktrace_key,omitempty" ini:"stacktrace_key,omitempty"`
	Level         string `json:"level,omitempty" yaml:"level,omitempty" ini:"level,omitempty"`
	Time          string `json:"time,omitempty" yaml:"time,omitempty" ini:"time,omitempty"`
	Duration      string `json:"duration,omitempty" yaml:"duration,omitempty" ini:"duration,omitempty"`
	Caller        string `json:"caller,omitempty" yaml:"caller,omitempty" ini:"caller,omitempty"`
	Encoding      string `json:"encoding,omitempty" yaml:"encoding,omitempty" ini:"encoding,omitempty"`
}

type URLFormat struct {
	Info       string `json:"info,omitempty" yaml:"info,omitempty" ini:"info,omitempty"`
	Detail     string `json:"detail,omitempty" yaml:"detail,omitempty" ini:"detail,omitempty"`
	Create     string `json:"create,omitempty" yaml:"create,omitempty" ini:"create,omitempty"`
	Delete     string `json:"delete,omitempty" yaml:"delete,omitempty" ini:"delete,omitempty"`
	Export     string `json:"export,omitempty" yaml:"export,omitempty" ini:"export,omitempty"`
	Edit       string `json:"edit,omitempty" yaml:"edit,omitempty" ini:"edit,omitempty"`
	ShowEdit   string `json:"show_edit,omitempty" yaml:"show_edit,omitempty" ini:"show_edit,omitempty"`
	ShowCreate string `json:"show_create,omitempty" yaml:"show_create,omitempty" ini:"show_create,omitempty"`
	Update     string `json:"update,omitempty" yaml:"update,omitempty" ini:"update,omitempty"`
}

func (f URLFormat) SetDefault() URLFormat {
	f.Detail     = utils.SetDefault(f.Detail    , "", "/info/:__prefix/detail")
	f.ShowEdit   = utils.SetDefault(f.ShowEdit  , "", "/info/:__prefix/edit")
	f.ShowCreate = utils.SetDefault(f.ShowCreate, "", "/info/:__prefix/new")
	f.Edit       = utils.SetDefault(f.Edit      , "", "/edit/:__prefix")
	f.Create     = utils.SetDefault(f.Create    , "", "/new/:__prefix")
	f.Delete     = utils.SetDefault(f.Delete    , "", "/delete/:__prefix")
	f.Export     = utils.SetDefault(f.Export    , "", "/export/:__prefix")
	f.Info       = utils.SetDefault(f.Info      , "", "/info/:__prefix")
	f.Update     = utils.SetDefault(f.Update    , "", "/update/:__prefix")
	return f
}

type ExtraInfo map[string]interface{}

type UpdateConfigProcessFn func(values form.Values) (form.Values, error)

// see more: https://daneden.github.io/animate.css/
type PageAnimation struct {
	Type     string  `json:"type,omitempty" yaml:"type,omitempty" ini:"type,omitempty"`
	Duration float32 `json:"duration,omitempty" yaml:"duration,omitempty" ini:"duration,omitempty"`
	Delay    float32 `json:"delay,omitempty" yaml:"delay,omitempty" ini:"delay,omitempty"`
}

func (p PageAnimation) JSON() string {
	if p.Type == "" { return "" }
	return utils.JSON(p)
}

// FileUploadEngine is a file upload engine.
type FileUploadEngine struct {
	Name   string                 `json:"name,omitempty" yaml:"name,omitempty" ini:"name,omitempty"`
	Config map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty" ini:"config,omitempty"`
}

func (f FileUploadEngine) JSON() string {
	if f.Name == "" { return "" }
	if len(f.Config) == 0 { f.Config = nil }
	return utils.JSON(f)
}

func GetFileUploadEngineFromJSON(m string) FileUploadEngine {
	var f FileUploadEngine
	if m == "" { return f }
	_ = utils.JsonUnmarshal([]byte(m), &f)
	return f
}

// GetIndexURL get the index url with prefix.
func (c *Config) GetIndexURL() string {
	index := c.Index()
	if index == "/" { return c.Prefix() }
	return c.Prefix() + index
}

// Url get url with the given suffix.
func (c *Config) Url(suffix string) string {
	if c.prefix == "/" { return   suffix }
	if   suffix == "/" { return c.prefix }
	return c.prefix + suffix
}

func (c *Config) SetupPrefix() *Config {
	if c.UrlPrefix == "" {
		c.prefix = "/"
	} else if c.UrlPrefix[0] != '/' {
		c.prefix = "/" + c.UrlPrefix
	} else {
		c.prefix = c.UrlPrefix
	}
	return c
}

// IsDevEnvironment check the environment if it is development.
func (c *Config) IsDevEnvironment() bool {
	return c.Env == EnvDev
}

// IsTestEnvironment check the environment if it is test.
func (c *Config) IsTestEnvironment() bool {
	return c.Env == EnvTest
}

// IsLocalEnvironment check the environment if it is local.
func (c *Config) IsLocalEnvironment() bool {
	return c.Env == EnvLocal
}

// IsProductionEnvironment check the environment if it is production.
func (c *Config) IsProductionEnvironment() bool {
	return c.Env == EnvProd
}

// IsNotProductionEnvironment check the environment if it is not production.
func (c *Config) IsNotProductionEnvironment() bool {
	return c.Env != EnvProd
}

func (c *Config) IsAllowConfigModification() bool {
	return !c.ProhibitConfigModification
}

// URLRemovePrefix remove prefix from the given url.
func (c *Config) URLRemovePrefix(url string) string {
	if url == c.prefix { return "/" }
	if c.prefix == "/" { return url }
	return strings.Replace(url, c.prefix, "", 1)
}

// Index return the index url without prefix.
func (c *Config) Index() string {
	if c.IndexUrl    == ""  { return "/" }
	if c.IndexUrl[0] != '/' { return "/" + c.IndexUrl }
	return c.IndexUrl
}

// Prefix return the prefix.
func (c *Config) Prefix() string {
	return c.prefix
}

// AssertPrefix return the prefix of assert.
func (c *Config) AssertPrefix() string {
	if c.prefix == "/" { return "" }
	return c.prefix
}

func (c *Config) AddUpdateProcessFn(fn UpdateConfigProcessFn) *Config {
	c.UpdateProcessFn = fn
	return c
}

// PrefixFixSlash return the prefix fix the slash error.
func (c *Config) PrefixFixSlash() string {
	if c.UrlPrefix == "/" {
		return ""
	}
	if c.UrlPrefix != "" && c.UrlPrefix[0] != '/' {
		return "/" + c.UrlPrefix
	}
	return c.UrlPrefix
}

func (c *Config) Copy() *Config {
	//c.lock.RLock()
	//defer c.lock.RUnlock()
	var (
		newCfg   = new(Config)
		srcType  = reflect.TypeOf(c).Elem()
		srcVal   = reflect.ValueOf(c).Elem()
		distType = reflect.TypeOf(newCfg).Elem()
		distVal  = reflect.ValueOf(newCfg).Elem()
	)

	for i := 0; i < distType.NumField(); i++ {
		v := distVal.Field(i)
		if distType.Field(i).Type.String() == "config.DatabaseList" {
			newCfg.Databases = c.Databases.Copy()
		} else if v.CanInterface() {
			for j := 0; j < srcType.NumField(); j++ {
				if distType.Field(i).Name == srcType.Field(j).Name {
					v.Set(reflect.ValueOf(srcVal.Field(i).Interface()))
					break
				}
			}
		}
	}

	newCfg.prefix = c.prefix
	return newCfg
}

func (c *Config) ToMap() map[string]string {
	//c.lock.RLock()
	//defer c.lock.RUnlock()
	var (
		m     = make(map[string]string)
		rType = reflect.TypeOf(c).Elem()
		rVal  = reflect.ValueOf(c).Elem()
	)

	for i := 0; i < rType.NumField(); i++ {
		v := rVal.Field(i)
		if !v.CanInterface() {
			continue
		}
		t := rType.Field(i)
		keyName := t.Tag.Get("json")
		if keyName == "-" {
			continue
		}
		keyName = keyName[:len(keyName)-10]
		switch t.Type.Kind() {
		case reflect.Bool:
			m[keyName] = strconv.FormatBool(v.Bool())
		case reflect.String:
			if keyName == "prefix" {
				keyName = "url_prefix"
			} else if keyName == "index" {
				keyName = "index_url"
			}
			m[keyName] = v.String()
		case reflect.Int:
			m[keyName] = fmt.Sprintf("%d", v.Int())
		case reflect.Struct:
			switch t.Type.String() {
			case "config.PageAnimation":
				m["animation_type"] = c.Animation.Type
				m["animation_duration"] = fmt.Sprintf("%.2f", c.Animation.Duration)
				m["animation_delay"] = fmt.Sprintf("%.2f", c.Animation.Delay)
			case "config.Logger":
				m["logger_encoder_time_key"] = c.Logger.Encoder.TimeKey
				m["logger_encoder_level_key"] = c.Logger.Encoder.LevelKey
				m["logger_encoder_name_key"] = c.Logger.Encoder.NameKey
				m["logger_encoder_caller_key"] = c.Logger.Encoder.CallerKey
				m["logger_encoder_message_key"] = c.Logger.Encoder.MessageKey
				m["logger_encoder_stacktrace_key"] = c.Logger.Encoder.StacktraceKey
				m["logger_encoder_level"] = c.Logger.Encoder.Level
				m["logger_encoder_time"] = c.Logger.Encoder.Time
				m["logger_encoder_duration"] = c.Logger.Encoder.Duration
				m["logger_encoder_caller"] = c.Logger.Encoder.Caller
				m["logger_encoder_encoding"] = c.Logger.Encoder.Encoding
				m["logger_level"] = strconv.Itoa(int(c.Logger.Level))
			case "config.DatabaseList":
				m["databases"] = utils.JSON(v.Interface())
			case "config.FileUploadEngine":
				m["file_upload_engine"] = c.FileUploadEngine.JSON()
			}
		case reflect.Map:
			if t.Type.String() == "config.ExtraInfo" {
				if len(c.Extra) == 0 {
					m["extra"] = ""
				} else {
					m["extra"] = utils.JSON(c.Extra)
				}
			}
		default:
			m[keyName] = utils.JSON(v.Interface())
		}
	}

	return m
}

func (c *Config) Update(m map[string]string) error {
	//c.lock.Lock()
	//defer c.lock.Unlock()
	rType := reflect.TypeOf(c).Elem()
	rVal := reflect.ValueOf(c).Elem()
	for i := 0; i < rType.NumField(); i++ {
		v := rVal.Field(i)
		if !v.CanInterface() {
			continue
		}
		t := rType.Field(i)
		keyName := t.Tag.Get("json")
		if keyName == "-" {
			continue
		}
		keyName = keyName[:len(keyName)-10]
		switch t.Type.Kind() {
		case reflect.Bool:
			if mv, ok := m[keyName]; ok {
				v.Set(reflect.ValueOf(utils.ParseBool(mv)))
			}
		case reflect.String:
			if t.Type.String() == "template.HTML" {
				if mv, ok := m[keyName]; ok {
					v.Set(reflect.ValueOf(template.HTML(mv)))
				}
				continue
			}
			if keyName == "prefix" {
				keyName = "url_prefix"
			} else if keyName == "index" {
				keyName = "index_url"
			}
			if mv, ok := m[keyName]; ok {
				if keyName == "info_log" || keyName == "error_log" || keyName == "access_log" {
					v.Set(reflect.ValueOf(utils.SetDefault(mv, v.String(), v.String())))
				} else if keyName == "app_id" {
					v.Set(reflect.ValueOf(utils.SetDefault(mv, "", v.String())))
				} else if keyName == "color_scheme" {
					if m["theme"] == "adminlte" {
						v.Set(reflect.ValueOf(mv))
					}
				} else {
					v.Set(reflect.ValueOf(mv))
				}
			}
		case reflect.Int:
			ses, _ := strconv.Atoi(m[keyName])
			if ses != 0 {
				v.Set(reflect.ValueOf(ses))
			}
		case reflect.Struct:
			switch t.Type.String() {
			case "config.PageAnimation":
				c.Animation.Type = m["animation_type"]
				c.Animation.Duration = utils.ParseFloat32(m["animation_duration"])
				c.Animation.Delay = utils.ParseFloat32(m["animation_delay"])
			case "config.Logger":
				c.Logger.Encoder.Encoding = m["logger_encoder_encoding"]
				loggerLevel, _ := strconv.Atoi(m["logger_level"])
				c.Logger.Level = int8(loggerLevel)

				if c.Logger.Encoder.Encoding == "json" {
					c.Logger.Encoder.TimeKey = m["logger_encoder_time_key"]
					c.Logger.Encoder.LevelKey = m["logger_encoder_level_key"]
					c.Logger.Encoder.NameKey = m["logger_encoder_name_key"]
					c.Logger.Encoder.CallerKey = m["logger_encoder_caller_key"]
					c.Logger.Encoder.MessageKey = m["logger_encoder_message_key"]
					c.Logger.Encoder.StacktraceKey = m["logger_encoder_stacktrace_key"]
					c.Logger.Encoder.Level = m["logger_encoder_level"]
					c.Logger.Encoder.Time = m["logger_encoder_time"]
					c.Logger.Encoder.Duration = m["logger_encoder_duration"]
					c.Logger.Encoder.Caller = m["logger_encoder_caller"]
				}

				initLogger(c)
			case "config.FileUploadEngine":
				c.FileUploadEngine = GetFileUploadEngineFromJSON(m["file_upload_engine"])
			}
		case reflect.Map:
			if t.Type.String() == "config.ExtraInfo" && m["extra"] != "" {
				var extra = make(map[string]interface{})
				_ = utils.JsonUnmarshal([]byte(m["extra"]), &extra)
				c.Extra = extra
			}
		}
	}

	return nil
}

// EraseSens erase sensitive info.
func (c *Config) EraseSens() *Config {
	for key := range c.Databases {
		c.Databases[key] = Database{ Driver: c.Databases[key].Driver }
	}
	return c
}

var (
	_global = new(Config)
	count   uint32
)

func SetDefault(cfg *Config) *Config {
	cfg.Title = utils.SetDefault(cfg.Title, "", "Flower")
	cfg.LoginTitle = utils.SetDefault(cfg.LoginTitle, "", "Flower")
	cfg.Logo = template.HTML(utils.SetDefault(string(cfg.Logo), "", "<b>Flower</b>"))
	cfg.MiniLogo = template.HTML(utils.SetDefault(string(cfg.MiniLogo), "", "<b>F</b>"))
	cfg.Theme = utils.SetDefault(cfg.Theme, "", "adminlte")
	cfg.IndexUrl = utils.SetDefault(cfg.IndexUrl, "", "/info/manager")
	cfg.LoginUrl = utils.SetDefault(cfg.LoginUrl, "", "/login")
	cfg.AuthUserTable = utils.SetDefault(cfg.AuthUserTable, "", "goadmin_users")
	if cfg.Theme == "adminlte" {
		cfg.ColorScheme = utils.SetDefault(cfg.ColorScheme, "", "skin-black")
	}
	cfg.AssetRootPath = utils.SetDefault(cfg.AssetRootPath, "", "./public/")
	cfg.AssetRootPath = filepath.ToSlash(cfg.AssetRootPath)
	cfg.FileUploadEngine.Name = utils.SetDefault(cfg.FileUploadEngine.Name, "", "local")
	cfg.Env = utils.SetDefault(cfg.Env, "", EnvProd)
	if cfg.SessionLifeTime == 0 {
		// default two hours
		cfg.SessionLifeTime = 7200
	}
	cfg.SetupPrefix()
	cfg.URLFormat = cfg.URLFormat.SetDefault()
	return cfg
}

// Initialize initialize the config.
func Initialize(cfg *Config) *Config {
	if atomic.AddUint32(&count, 1) != 1 {
		panic("can not initialize config twice")
	}
	initLogger(SetDefault(cfg))
	_global = cfg
	return _global
}

func initLogger(cfg *Config) {
	logger.InitWithConfig(logger.Config{
		InfoLogOff:         cfg.InfoLogOff,
		ErrorLogOff:        cfg.ErrorLogOff,
		AccessLogOff:       cfg.AccessLogOff,
		SqlLogOpen:         cfg.SqlLog,
		AccessAssetsLogOff: cfg.AccessAssetsLogOff,
		Encode: logger.EncoderCfg{
			TimeKey:       cfg.Logger.Encoder.TimeKey,
			LevelKey:      cfg.Logger.Encoder.LevelKey,
			NameKey:       cfg.Logger.Encoder.NameKey,
			CallerKey:     cfg.Logger.Encoder.CallerKey,
			MessageKey:    cfg.Logger.Encoder.MessageKey,
			StacktraceKey: cfg.Logger.Encoder.StacktraceKey,
			Level:         cfg.Logger.Encoder.Level,
			Time:          cfg.Logger.Encoder.Time,
			Duration:      cfg.Logger.Encoder.Duration,
			Caller:        cfg.Logger.Encoder.Caller,
			Encoding:      cfg.Logger.Encoder.Encoding,
		},
		Debug: cfg.Debug,
		Level: cfg.Logger.Level,
	})
}

// AssertPrefix return the prefix of assert.
func AssertPrefix() string {
	return _global.AssertPrefix()
}

// GetIndexURL get the index url with prefix.
func GetIndexURL() string {
	return _global.GetIndexURL()
}

// IsProductionEnvironment check the environment if it is production.
func IsProductionEnvironment() bool {
	return _global.IsProductionEnvironment()
}

// IsNotProductionEnvironment check the environment if it is not production.
func IsNotProductionEnvironment() bool {
	return _global.IsNotProductionEnvironment()
}

// URLRemovePrefix remove prefix from the given url.
func URLRemovePrefix(url string) string {
	return _global.URLRemovePrefix(url)
}

func Url(suffix string) string {
	return _global.Url(suffix)
}

func GetURLFormats() URLFormat {
	return _global.URLFormat
}

// Prefix return the prefix.
func Prefix() string {
	return _global.prefix
}

// PrefixFixSlash return the prefix fix the slash error.
func PrefixFixSlash() string {
	return _global.PrefixFixSlash()
}

// Get gets the config.
func Get() *Config {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Copy().EraseSens()
}

// Getter methods
// ============================

func GetDatabases() DatabaseList {
	list := make(DatabaseList, len(_global.Databases))
	for k := range _global.Databases {
		list[k] = Database{
			Driver:     _global.Databases[k].Driver,
			DriverMode: _global.Databases[k].DriverMode,
		}
	}
	return list
}

func GetDomain() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Domain
}

func GetLanguage() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Language
}

func GetUrlPrefix() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.UrlPrefix
}

func GetOpenAdminApi() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.OpenAdminApi
}

func GetOperationLogOff() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.OperationLogOff
}

func GetCustom500HTML() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Custom500HTML
}

func GetCustom404HTML() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Custom404HTML
}

func GetCustom403HTML() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Custom403HTML
}

func GetTheme() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Theme
}

func GetStore() Store {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Store
}

func GetTitle() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Title
}

func GetAssetRootPath() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.AssetRootPath
}

func GetLogo() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Logo
}

func GetSiteOff() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.SiteOff
}

func GetMiniLogo() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.MiniLogo
}

func GetIndexUrl() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.IndexUrl
}

func GetLoginUrl() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.LoginUrl
}

func GetDebug() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Debug
}

func GetEnv() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Env
}

func GetSqlLog() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.SqlLog
}

func GetAccessLogOff() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.AccessLogOff
}

func GetInfoLogOff() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.InfoLogOff
}

func GetErrorLogOff() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.ErrorLogOff
}

func GetColorScheme() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.ColorScheme
}

func GetSessionLifeTime() int {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.SessionLifeTime
}

func GetAssetUrl() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.AssetUrl
}

func GetFileUploadEngine() FileUploadEngine {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.FileUploadEngine
}

func GetCustomHeadHtml() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.CustomHeadHtml
}

func GetCustomFootHtml() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.CustomFootHtml
}

func GetFooterInfo() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.FooterInfo
}

func GetLoginTitle() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.LoginTitle
}

func GetLoginLogo() template.HTML {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.LoginLogo
}

func GetAuthUserTable() string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.AuthUserTable
}

func GetExtra() map[string]interface{} {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Extra
}

func GetAnimation() PageAnimation {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.Animation
}

func GetNoLimitLoginIP() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.NoLimitLoginIP
}

func GetHideVisitorUserCenterEntrance() bool {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.HideVisitorUserCenterEntrance
}

func GetExcludeThemeComponents() []string {
	//_global.lock.RLock()
	//defer _global.lock.RUnlock()
	return _global.ExcludeThemeComponents
}

type Service struct {
	C *Config
}

func (s *Service) Name() string {
	return "config"
}

func SrvWithConfig(c *Config) *Service {
	return &Service{c}
}

func GetService(s interface{}) *Config {
	if srv, ok := s.(*Service); ok {
		return srv.C
	}
	panic("wrong service")
}
