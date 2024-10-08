// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"
	"sync"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
)

var (
	EncryptPass      func(algo, pass string) string
	EncryptPassMatch func(pass, hashedPass string) bool
	EncryptPassAlgo  string
)

// Auth get the user model from Context.
func Auth(ctx *context.Context) models.UserModel {
	return ctx.User().(models.UserModel)
}

// Check username and password and return the user model.
func Check(username, password string, conn db.Connection) (models.UserModel, bool) {
	user := models.User().SetConn(conn).FindByUserName(username)
	if user.IsEmpty() || !EncryptPassMatch(password, user.Password) {
		return user, false
	}
	//user.UpdatePwd(EncodePassword([]byte(password)))			// uncomment to enforce security: rewrite new hashed password at each successful access
	return user.WithRoles().WithPermissions().WithMenus(), true
}

// EncodePassword encode the password.
func EncodePassword(pwd string) string {
	return EncryptPass(EncryptPassAlgo, pwd)
}

// SetCookie set the cookie.
func SetCookie(ctx *context.Context, user models.UserModel, conn db.Connection) error {
	ses, err := InitSession(ctx, conn)
	if err != nil { return err }
	return ses.Add("user_id", user.Id)
}

func DefaultCookie() *http.Cookie {
	return &http.Cookie{
		Name    : DefaultCookieKey,
		Path    : "/",
		HttpOnly: true,
		Domain  : config.GetDomain(),
		MaxAge  : -1,
	}
}

// DelCookie delete the cookie from Context.
func DelCookie(ctx *context.Context, conn db.Connection) error {
	ctx.SetCookie(DefaultCookie())
	sess, err := InitSession(ctx, conn)
	if err != nil { return err }
	return sess.Clear()
}

type TokenService struct {
	tokens CSRFToken
	lock   sync.Mutex
	conn   db.Connection
}

func (s *TokenService) Name() string {
	return TokenServiceKey
}

func InitCSRFTokenSrv(conn db.Connection) (string, service.Service) {
	list, err := db.WithDriver(conn).Table("goadmin_session").
		Where("values", "=", "__csrf_token__").
		All()
	if db.CheckError(err, db.QUERY) {
		logger.Error("cannot retrieve csrf tokens from db: ", err)
	}
	tokens := make(CSRFToken, len(list))
	for _, elem := range list {
		tokens[elem["sid"].(string)] = struct{}{}
	}
	return TokenServiceKey, &TokenService{
		tokens: tokens,
		conn  : conn,
	}
}

const (
	TokenServiceKey = "token_csrf_helper"
	ServiceKey      = "auth"
)

func GetTokenService(s interface{}) *TokenService {
	if srv, ok := s.(*TokenService); ok {
		return srv
	}
	panic("wrong service")
}

// AddToken add the token to the CSRFToken.
func (s *TokenService) AddToken() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	tokenStr := modules.Uuid()
	s.tokens[tokenStr] = struct{}{}
	_, err := db.WithDriver(s.conn).Table("goadmin_session").Insert(dialect.H{
		"sid"   : tokenStr,
		"values": "__csrf_token__",
	})
	if db.CheckError(err, db.INSERT) {
		logger.Error("cannot insert csrf token into db: ", err)
	}
	return tokenStr
}

// CheckToken check the given token with tokens in the CSRFToken, if exist return true.
func (s *TokenService) CheckToken(tokenToCheck string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.tokens[tokenToCheck]; ok {
		delete(s.tokens, tokenToCheck)
		err := db.WithDriver(s.conn).Table("goadmin_session").
			Where("sid"   , "=", tokenToCheck).
			Where("values", "=", "__csrf_token__").
			Delete()
		if db.CheckError(err, db.DELETE) {
			logger.Error("cannot delete csrf token from db: ", err)
		}
		return true
	}
	return false
}

// CSRFToken is type of a csrf token list.
type CSRFToken map[string]struct{}

type Processor func(ctx *context.Context) (model models.UserModel, exist bool, msg string)

type Service struct {
	P Processor
}

func (s *Service) Name() string {
	return "auth"
}

func GetService(s interface{}) *Service {
	if srv, ok := s.(*Service); ok {
		return srv
	}
	panic("wrong service")
}

func NewService(processor Processor) *Service {
	return &Service{ P: processor }
}
