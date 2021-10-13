package models

import (
	"database/sql"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"net/url"
	"strconv"
	"strings"
)

// UserModel is user model structure.
type UserModel struct {
	Base `json:"-"`

	Id            int64             `json:"id"`
	Name          string            `json:"name"`
	UserName      string            `json:"user_name"`
	Password      string            `json:"password"`
	Email         string            `json:"email"`
	Avatar        string            `json:"avatar"`
	Disabled      string            `json:"disabled"`
	Root          string            `json:"root"`
	Permissions   []PermissionModel `json:"permissions"`
	MenuIds       []int64           `json:"menu_ids"`
	Roles         []RoleModel       `json:"role"`
	Level         string            `json:"level"`
	LevelName     string            `json:"level_name"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	cacheReplacer *strings.Replacer
}

// User return a default user model.
func User() UserModel {
	return UserModel{ Base: Base{ TableName: config.GetAuthUserTable() }}
}

// UserWithId return a default user model of given id.
func UserWithId(id string) UserModel {
	idInt, _ := strconv.ParseInt(id, 10, 64)
	return UserModel{ Base: Base{ TableName: config.GetAuthUserTable() }, Id: idInt }
}

func (t UserModel) SetConn(con db.Connection) UserModel {
	t.Conn = con
	return t
}

func (t UserModel) WithTx(tx *sql.Tx) UserModel {
	t.Tx = tx
	return t
}

// Find return a default user model of given id.
func (t UserModel) Find(id interface{}) UserModel {
	item, _ := t.Table(t.TableName).Find(id)
	return t.MapToModel(item)
}

// FindByUserName return a default user model of given name.
func (t UserModel) FindByUserName(username interface{}) UserModel {
	item, _ := t.Table(t.TableName).Where("username", "=", username).First()
	return t.MapToModel(item)
}

// IsEmpty check the user model is empty or not.
func (t UserModel) IsEmpty() bool {
	return t.Id == 0
}

// HasMenu check the user has visitable menu or not.
func (t UserModel) HasMenu() bool {
	return len(t.MenuIds) != 0 || t.IsSuperAdmin()
}

// IsSuperAdmin check the user model is super admin or not.
func (t UserModel) IsSuperAdmin() bool {
	if t.IsRootAdmin() { return true }
	for _, perm := range t.Permissions {
		if len(perm.HttpPath) > 0 && perm.HttpPath[0] == "*" && perm.HttpMethod[0] == "" {
			return true
		}
	}
	return false
}

func (t UserModel) GetCheckPermissionByUrlMethod(path, method string) string {
	if !t.CheckPermissionByUrlMethod(path, method, nil) {
		return ""
	}
	return path
}

func (t UserModel) IsVisitor() bool {
	return !t.CheckPermissionByUrlMethod(config.Url("/info/normal_manager"), "GET", nil)
}

func (t UserModel) HideUserCenterEntrance() bool {
	return t.IsVisitor() && config.GetHideVisitorUserCenterEntrance()
}

func (t UserModel) Template(str string) string {
	if t.cacheReplacer == nil {
		t.cacheReplacer = strings.NewReplacer("{{.AuthId}}", strconv.Itoa(int(t.Id)),
			"{{.AuthName}}", t.Name, "{{.AuthUserName}}", t.UserName)
	}
	return t.cacheReplacer.Replace(str)
}

func (t UserModel) CheckPermissionByUrlMethod(path, method string, formParams url.Values) bool {
	if t.IsSuperAdmin() {
		return true
	}

	// path, _ = url.PathUnescape(path)
	if path == "" {
		return false
	}

	if utils.IsLogoutUrl(path) {
		return true
	}

	if path != "/" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	path = utils.PkReplacer.Replace(path)

	path, params := getParam(path)
	for key, value := range formParams {
		if len(value) > 0 {
			if params == nil {
				params = url.Values{ key: []string{ value[0] }}
			} else {
				params.Add(key, value[0])
			}
		}
	}

	//if t.isMySettingRequest(method, path, params) {
	//	return true
	//}

	for _, v := range t.Permissions {
		if v.HttpMethod[0] == "" || inMethodArr(v.HttpMethod, method) {
			if v.HttpPath[0] == "*" {
				return true
			}

			for _, httpPath := range v.HttpPath {
				matchPath := config.Url(t.Template(httpPath))
				matchPath, matchParams := getParam(matchPath)

				if matchPath == path {
					if t.checkParam(params, matchParams) {
						return true
					}
				}

				rex, err := utils.CachedRex(normMatchPath(matchPath))
				if err != nil {
					logger.Error("CheckPermissions error: ", err)
					continue
				}

				if rex.FindString(path) == path {
					if t.checkParam(params, matchParams) {
						return true
					}
				}
			}
		}
	}

	return false
}

func getParam(u string) (string, url.Values) {
	if p := strings.IndexByte(u, '?'); p >= 0 {
		var m url.Values
		if p < len(u) - 1 {
			m, _ = url.ParseQuery(u[p + 1:])
		}
		return u[:p], m
	}
	return u, nil
	/*m := make(url.Values)
	urr := strings.Split(u, "?")
	if len(urr) > 1 {
		m, _ = url.ParseQuery(urr[1])
	}
	return urr[0], m*/
}

func (t UserModel) checkParam(src, comp url.Values) bool {
	if len(comp) == 0 {
		return true
	}
	if len(src) == 0 {
		return false
	}
	for key, value := range comp {
		v, ok := src[key]
		if !ok {
			return false
		}
		if len(value) == 0 {
			continue
		}
		if len(v) == 0 {
			return false
		}
		for i, e := range v {
			if e != t.Template(value[i]) {
				return false
			}
		}
	}
	return true
}

func inMethodArr(arr []string, str string) bool {
	for _, method := range arr {
		if strings.EqualFold(method, str) {
			return true
		}
	}
	return false
}

// UpdateAvatar update the avatar of user.
func (t UserModel) ReleaseConn() UserModel {
	t.Conn = nil
	return t
}

// UpdateAvatar update the avatar of user.
func (t UserModel) UpdateAvatar(avatar string) {
	t.Avatar = avatar
}

// WithRoles query the role info of the user.
func (t UserModel) WithRoles() UserModel {
	roleModel, _ := t.Table("goadmin_role_users").
		LeftJoin("goadmin_roles", "goadmin_roles.id", "=", "goadmin_role_users.role_id").
		Where("user_id", "=", t.Id).
		Select("goadmin_roles.id", "goadmin_roles.name", "goadmin_roles.slug", "goadmin_roles.created_at", "goadmin_roles.updated_at").
		All()

	for _, role := range roleModel {
		t.Roles = append(t.Roles, Role().MapToModel(role))
	}

	if len(t.Roles) > 0 {
		r := t.Roles[0]
		t.Level     = r.Slug
		t.LevelName = r.Name
	}

	return t
}

func (t UserModel) GetAllRoleId() []interface{} {
	ids := make([]interface{}, len(t.Roles))
	for i, role := range t.Roles {
		ids[i] = role.Id
	}
	return ids
}

// WithPermissions query the permission info of the user.
func (t UserModel) WithPermissions() UserModel {
	var permissions []map[string]interface{}
	var err         error

	roleIds := t.GetAllRoleId()
	if len(roleIds) > 0 {
		permissions, err = t.Table("goadmin_role_permissions").
			LeftJoin("goadmin_permissions", "goadmin_permissions.id", "=", "goadmin_role_permissions.permission_id").
			WhereIn("role_id", roleIds).
			Select("goadmin_permissions.http_method", "goadmin_permissions.http_path",
				"goadmin_permissions.id", "goadmin_permissions.name", "goadmin_permissions.slug",
				"goadmin_permissions.created_at", "goadmin_permissions.updated_at").
			All()
		if err != nil {
			logger.Errorf("cannot retrieve role permissions (related to user %s): %v", t.UserName, err)
		}
	}

	userPermissions, err := t.Table("goadmin_user_permissions").
		LeftJoin("goadmin_permissions", "goadmin_permissions.id", "=", "goadmin_user_permissions.permission_id").
		Where("user_id", "=", t.Id).
		Select("goadmin_permissions.http_method", "goadmin_permissions.http_path",
			"goadmin_permissions.id", "goadmin_permissions.name", "goadmin_permissions.slug",
			"goadmin_permissions.created_at", "goadmin_permissions.updated_at").
		All()
	if err != nil {
		logger.Errorf("cannot retrieve user permissions (related to user %s): %v", t.UserName, err)
	}

	permissions = append(permissions, userPermissions...)
	t.Permissions = make([]PermissionModel, 0, len(permissions))

	for _, perm := range permissions {
		permId := perm["id"]
		exist  := false
		for _, p := range t.Permissions {
			if p.Id == permId {
				exist = true
				break
			}
		}
		if exist { continue }
		t.Permissions = append(t.Permissions, Permission().MapToModel(perm))
	}

	return t
}

// WithMenus query the menu info of the user.
func (t UserModel) WithMenus() UserModel {
	var menuIdsModel []map[string]interface{}
	var err error

	if t.IsSuperAdmin() {
		menuIdsModel, err = t.Table("goadmin_role_menu").
			LeftJoin("goadmin_menu", "goadmin_menu.id", "=", "goadmin_role_menu.menu_id").
			Select("menu_id", "parent_id").
			All()
	} else {
		rolesId := t.GetAllRoleId()
		if len(rolesId) > 0 {
			menuIdsModel, err = t.Table("goadmin_role_menu").
				LeftJoin("goadmin_menu", "goadmin_menu.id", "=", "goadmin_role_menu.menu_id").
				WhereIn("goadmin_role_menu.role_id", rolesId).
				Select("menu_id", "parent_id").
				All()
		}
	}

	if err != nil {
		logger.Errorf("cannot retrieve menu entries (related to user '%s'): %v", t.UserName, err)
	}

	menuIds := make([]int64, 0, len(menuIdsModel))

	for _, m := range menuIdsModel {
		if parentId, _ := m["parent_id"].(int64); parentId != 0 {
			for _, p := range menuIdsModel {
				if p["menu_id"].(int64) == parentId {
					menuIds = append(menuIds, m["menu_id"].(int64))
					break
				}
			}
		} else {
			menuIds = append(menuIds, m["menu_id"].(int64))
		}
	}

	t.MenuIds = menuIds
	return t
}

// New create a user model.
func (t UserModel) New(username, password, name, email, disabled, root, avatar string) (UserModel, error) {
	disabled = normUserDisabled(disabled)
	root     = normUserRoot(root)

	id, err := t.Table(t.TableName).Insert(dialect.H{
		"username" : username,
		"password" : password,
		"name"     : name,
		"email"    : email,
		"disabled" : disabled,
		"root"     : root,
		"avatar"   : avatar,
	})

	t.Id = id
	t.UserName = username
	t.Password = password
	t.Name = name
	t.Email = email
	t.Disabled = disabled
	t.Root = root
	t.Avatar = avatar

	return t, err
}

// Update update the user model.
func (t UserModel) Update(username, password, name, email, disabled, root, avatar string, isUpdateAvatar bool) (int64, error) {
	fieldValues := dialect.H{
		"username"  : username,
		"updated_at": utils.NowStr(),
	}
	if name != "" {
		fieldValues["name"] = name
	}
	if email != "" {
		fieldValues["email"] = email
	}
	if disabled != "" {
		fieldValues["disabled"] = normUserDisabled(disabled)
	}
	if root != "" {
		fieldValues["root"] = normUserRoot(root)
	}
	if avatar != "" || isUpdateAvatar {
		fieldValues["avatar"] = avatar
	}
	if password != "" {
		fieldValues["password"] = password
	}
	return t.Table(t.TableName).
		Where("id", "=", t.Id).
		Update(fieldValues)
}

// UpdatePwd update the password of the user model.
func (t UserModel) UpdatePwd(password string) UserModel {
	_, _ = t.Table(t.TableName).
		Where("id", "=", t.Id).
		Update(dialect.H{ "password": password })
	t.Password = password
	return t
}

// CheckRole check the role of the user model.
func (t UserModel) CheckRoleId(roleId string) bool {
	checkRole, _ := t.Table("goadmin_role_users").
		Where("role_id", "=", roleId).
		Where("user_id", "=", t.Id).
		First()
	return checkRole != nil
}

// DeleteRoles delete all the roles of the user model.
func (t UserModel) DeleteRoles() error {
	return t.Table("goadmin_role_users").
		Where("user_id", "=", t.Id).
		Delete()
}

// AddRole add a role of the user model.
func (t UserModel) AddRole(roleId string) (int64, error) {
	if roleId != "" {
		if !t.CheckRoleId(roleId) {
			return t.Table("goadmin_role_users").
				Insert(dialect.H{
					"role_id": roleId,
					"user_id": t.Id,
				})
		}
	}
	return 0, nil
}

// CheckRole check the role of the user.
func (t UserModel) CheckRole(slug string) bool {
	for _, role := range t.Roles {
		if role.Slug == slug {
			return true
		}
	}
	return false
}

// CheckPermission check the permission of the user.
func (t UserModel) CheckPermissionById(permissionId string) bool {
	checkPermission, _ := t.Table("goadmin_user_permissions").
		Where("permission_id", "=", permissionId).
		Where("user_id", "=", t.Id).
		First()
	return checkPermission != nil
}

// CheckPermission check the permission of the user.
func (t UserModel) CheckPermission(permission string) bool {
	for _, perm := range t.Permissions {
		if perm.Slug == permission {
			return true
		}
	}
	return false
}

// DeletePermissions delete all the permissions of the user model.
func (t UserModel) DeletePermissions() error {
	return t.Table("goadmin_user_permissions").
		Where("user_id", "=", t.Id).
		Delete()
}

// AddPermission add a permission of the user model.
func (t UserModel) AddPermission(permissionId string) (int64, error) {
	if permissionId != "" {
		if !t.CheckPermissionById(permissionId) {
			return t.Table("goadmin_user_permissions").
				Insert(dialect.H{
					"permission_id": permissionId,
					"user_id"      : t.Id,
				})
		}
	}
	return 0, nil
}

// MapToModel get the user model from given map.
func (t UserModel) MapToModel(m map[string]interface{}) UserModel {
	t.Id, _ = m["id"].(int64)
	t.Name, _ = m["name"].(string)
	t.UserName, _ = m["username"].(string)
	t.Password, _ = m["password"].(string)
	t.Email, _ = m["email"].(string)
	t.Disabled, _ = m["disabled"].(string)
	t.Root, _ = m["root"].(string)
	t.Avatar, _ = m["avatar"].(string)
	t.CreatedAt, _ = m["created_at"].(string)
	t.UpdatedAt, _ = m["updated_at"].(string)
	return t
}
