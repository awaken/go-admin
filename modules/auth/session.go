// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
)

const DefaultCookieKey = "_f_"

// NewDBDriver return the default PersistenceDriver.
func newDBDriver(conn db.Connection) *DBDriver {
	return &DBDriver{
		conn:      conn,
		tableName: "goadmin_session",
	}
}

// PersistenceDriver is a driver of storing and getting the session info.
type PersistenceDriver interface {
	Load(string) (map[string]interface{}, error)
	Update(sid string, values map[string]interface{}) error
}

// GetSessionByKey get the session value by key.
func GetSessionByKey(sesKey, key string, conn db.Connection) (interface{}, error) {
	m, err := newDBDriver(conn).Load(sesKey)
	return m[key], err
}

// Session contains info of session.
type Session struct {
	Expires time.Duration
	Cookie  string
	Values  map[string]interface{}
	Driver  PersistenceDriver
	Sid     string
	Context *context.Context
}

// Config wraps the Session info.
type Config struct {
	Expires time.Duration
	Cookie  string
}

// UpdateConfig update the Expires and Cookie of Session.
func (ses *Session) UpdateConfig(config Config) {
	ses.Expires = config.Expires
	ses.Cookie = config.Cookie
}

// Get get the session value.
func (ses *Session) Get(key string) interface{} {
	return ses.Values[key]
}

// Add add the session value of key.
func (ses *Session) Add(key string, value interface{}) error {
	ses.Values[key] = value
	if err := ses.Driver.Update(ses.Sid, ses.Values); err != nil {
		return err
	}
	cookie := http.Cookie{
		Name    : ses.Cookie,
		Value   : ses.Sid,
		MaxAge  : config.GetSessionLifeTime(),
		Expires : time.Now().Add(ses.Expires),
		HttpOnly: true,
		Path    : "/",
		Domain  : config.GetDomain(),
	}
	ses.Context.SetCookie(&cookie)
	return nil
}

// Clear clear a Session.
func (ses *Session) Clear() error {
	//ses.Values = map[string]interface{}{}
	return ses.Driver.Update(ses.Sid, nil)
}

// UseDriver set the driver of the Session.
func (ses *Session) UseDriver(driver PersistenceDriver) {
	ses.Driver = driver
}

func (ses *Session) load(ctx *context.Context) (bool, error) {
	if cookie, err := ctx.Request.Cookie(ses.Cookie); err == nil && cookie.Value != "" {
		ses.Sid = cookie.Value
		valueFromDriver, err := ses.Driver.Load(cookie.Value)
		if err != nil {
			return false, err
		}
		if len(valueFromDriver) > 0 {
			ses.Values = valueFromDriver
		}
		return true, nil
	}
	return false, nil
}

// StartCtx return a Session from the given Context.
func (ses *Session) StartCtx(ctx *context.Context) (*Session, error) {
	ok, err := ses.load(ctx)
	if err != nil { return nil, err }
	if !ok { ses.Sid = modules.Uuid() }
	ses.Context = ctx
	return ses, nil
}

func initSession(conn db.Connection) *Session {
	ses := new(Session)
	ses.UpdateConfig(Config{
		Expires: time.Second * time.Duration(config.GetSessionLifeTime()),
		Cookie:  DefaultCookieKey,
	})

	ses.UseDriver(newDBDriver(conn))
	ses.Values = make(map[string]interface{})

	return ses
}

// InitSession return the default Session.
func InitSession(ctx *context.Context, conn db.Connection) (*Session, error) {
	return initSession(conn).StartCtx(ctx)
}

// LoadSession return the default Session, only if already exists.
func LoadSession(ctx *context.Context, conn db.Connection) (*Session, error) {
	ses := initSession(conn)
	ok, err := ses.load(ctx)
	if err != nil { return nil, err }
	if !ok { return nil, nil }
	return ses, nil
}

// DBDriver is a driver which uses database as a persistence tool.
type DBDriver struct {
	conn      db.Connection
	tableName string
}

// Load implements the PersistenceDriver.Load.
func (driver *DBDriver) Load(sid string) (map[string]interface{}, error) {
	sesModel, err := driver.table().Where("sid", "=", sid).First()
	if db.CheckError(err, db.QUERY) {
		return nil, err
	}

	if sesModel == nil {
		return map[string]interface{}{}, nil
	}

	var values map[string]interface{}
	err = utils.JsonUnmarshal([]byte(sesModel["values"].(string)), &values)
	return values, err
}

func (driver *DBDriver) deleteOverdueSession() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			panic(err)
		}
	}()

	duration   := strconv.Itoa(config.GetSessionLifeTime() + 1000)
	driverName := config.GetDatabases().GetDefault().Driver
	raw        := ""

	switch driverName {
	case db.DriverMysql:
		raw = `unix_timestamp(created_at) < unix_timestamp() - ` + duration
	case db.DriverPostgresql:
		raw = `extract(epoch from now()) - ` + duration + ` > extract(epoch from created_at)`
	case db.DriverMssql:
		raw = `DATEDIFF(second, [created_at], GETDATE()) > ` + duration
	case db.DriverSqlite:
		raw = `strftime('%s', created_at) < strftime('%s', 'now') - ` + duration
	default:
		return
	}

	_ = driver.table().WhereRaw(raw).Delete()
}

// Update implements the PersistenceDriver.Update.
func (driver *DBDriver) Update(sid string, values map[string]interface{}) error {
	go driver.deleteOverdueSession()

	if sid != "" {
		if len(values) == 0 {
			err := driver.table().Where("sid", "=", sid).Delete()
			if db.CheckError(err, db.DELETE) {
				return err
			}
		}
		valuesByte, err := utils.JsonMarshal(values)
		if err != nil {
			return err
		}
		sesValue := string(valuesByte)
		sesModel, _ := driver.table().Where("sid", "=", sid).First()
		if sesModel == nil {
			if !config.GetNoLimitLoginIP() {
				err = driver.table().Where("values", "=", sesValue).Delete()
				if db.CheckError(err, db.DELETE) {
					return err
				}
			}
			_, err := driver.table().Insert(dialect.H{
				"values": sesValue,
				"sid":    sid,
			})
			if db.CheckError(err, db.INSERT) {
				return err
			}
		} else {
			_, err := driver.table().
				Where("sid", "=", sid).
				Update(dialect.H{
					"values": sesValue,
				})
			if db.CheckError(err, db.UPDATE) {
				return err
			}
		}
	}
	return nil
}

func (driver *DBDriver) table() *db.SQL {
	return db.Table(driver.tableName).WithDriver(driver.conn)
}
