package middlewares

import (
	"encoding/json"
	"net/http"

	"github.com/ego-component/egorm"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/cetus/pkg/kauth"
	"github.com/gotomicro/ego/core/econf"
	"github.com/gotomicro/ego/core/elog"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/clickvisual/clickvisual/api/internal/invoker"
	"github.com/clickvisual/clickvisual/api/internal/pkg/component/core"
	"github.com/clickvisual/clickvisual/api/internal/pkg/model/db"
	"github.com/clickvisual/clickvisual/api/internal/service/permission"
)

func AuthChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch {
		case !isNotLogin(c):
		case !isNotAuthProxy(c):
		case !isNotAnonymousUser(c):
		default:
			appURL, _, _ := kauth.ParseAppAndSubURL(econf.GetString("app.rootURL"))
			c.JSON(http.StatusOK, core.Res{Code: 302, Data: appURL + "user/login"})
			c.Abort()
			return
		}
	}
}

func DangerPasswordChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		tmp, _ := json.Marshal(user)
		u := db.User{}
		_ = json.Unmarshal(tmp, &u)
		// 如果是 clickvisual 这个默认用户，需要提示用户修改密码
		if u.Username == "clickvisual" {
			userTmp, _ := db.UserInfo(u.ID)
			isDefault := bcrypt.CompareHashAndPassword([]byte(userTmp.Password), []byte("c37de4f875d7f764d27cd57dccfa0e56"))
			if isDefault == nil {
				c.JSON(http.StatusOK, core.Res{Code: 12001, Data: "【风险提示】点击右上角用户名修改默认密码 / Click on the upper right corner to change the default password", Msg: "【风险提示】点击右上角用户名修改默认密码 / Click on the upper right corner to change the default password"})
				c.Abort()
				return
			}
		}
	}
}

func RootChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := core.ContextUser(c)
		if user == nil {
			c.JSON(http.StatusOK, core.Res{Code: 1, Data: "session user is nil", Msg: ""})
			c.Abort()
			return
		}
		u := db.User{}
		userBytes, _ := json.Marshal(user)
		if _ = json.Unmarshal(userBytes, &u); u.Username == "" {
			c.JSON(http.StatusOK, core.Res{Code: 1, Data: "user content is empty: " + string(userBytes), Msg: "administrator privileges are required to access this api"})
			c.Abort()
			return
		}
		if err := permission.Manager.IsRootUser(u.Uid); err != nil {
			c.JSON(http.StatusOK, core.Res{Code: 1, Data: "IsRootUser: " + err.Error(), Msg: "administrator privileges are required to access this api"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func isNotLogin(c *gin.Context) bool {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		return true
	}
	u := db.User{}
	userBytes, _ := json.Marshal(user)
	if _ = json.Unmarshal(userBytes, &u); u.Username == "" {
		return true
	}
	ctxUser := &core.User{Uid: int64(u.ID), Nickname: u.Nickname, Username: u.Username, Avatar: u.Avatar, Email: u.Email}
	c.Set(core.UserContextKey, ctxUser)
	c.Next()
	return false
}

func isNotAnonymousUser(c *gin.Context) bool {
	if !econf.GetBool("auth.anonymous.enabled") {
		return true
	}
	u := &db.User{Username: "anonymous", Nickname: "anonymous", BaseModel: db.BaseModel{ID: 999999}}
	ctxUser := &core.User{Uid: int64(u.ID), Nickname: u.Nickname, Username: u.Username, Avatar: u.Avatar, Email: u.Email}
	c.Set(core.UserContextKey, ctxUser)
	c.Next()
	return false
}

func isNotAuthProxy(c *gin.Context) bool {
	username := c.GetHeader(econf.GetString("auth.proxy.headerName"))
	// Bail if auth proxy is not enabled
	if !econf.GetBool("auth.proxy.enabled") {
		return true
	}
	// If there is no header - we can't move forward
	if username == "" {
		return true
	}
	// User login
	conds := egorm.Conds{}
	conds["username"] = username
	u, err := db.UserInfoX(conds)
	if err != nil && !errors.Is(err, egorm.ErrRecordNotFound) {
		elog.Error("isNotAuthProxy", elog.String("step", "UserInfoX"), elog.String("username", username), elog.String("error", err.Error()))
		return true
	}
	if u.ID == 0 {
		nickName := c.GetHeader(econf.GetString("auth.proxy.headerNickName"))
		if nickName == "" {
			nickName = username
		}
		u = db.User{Username: username, Nickname: nickName, Access: "auth.proxy"}
		err = db.UserCreate(invoker.Db, &u)
		if err != nil {
			elog.Error("isNotAuthProxy", elog.String("step", "UserCreate"), elog.String("username", username), elog.String("error", err.Error()))
			return true
		}
	}
	// root？
	if econf.GetString("auth.proxy.rootTokenValue") != "" && c.GetHeader(econf.GetString("auth.proxy.rootTokenKey")) == econf.GetString("auth.proxy.rootTokenValue") {
		errRoot := permission.Manager.IsRootUser(u.ID)
		elog.Debug("isNotAuthProxy", elog.Any("errRoot", errRoot), elog.Any("u.ID", u.ID))
		if errRoot != nil {
			elog.Debug("isNotAuthProxy", elog.String("step", "rootUpdate"), elog.Any("user", u))
			roots := permission.Manager.GetRootUsersId()
			roots = append(roots, u.ID)
			permission.Manager.GrantRootUsers(roots)
		}
	}
	elog.Debug("isNotAuthProxy", elog.String("step", "finish"), elog.Any("user", u))
	ctxUser := &core.User{Uid: int64(u.ID), Nickname: u.Nickname, Username: u.Username, Avatar: u.Avatar, Email: u.Email}
	c.Set(core.UserContextKey, ctxUser)
	c.Next()
	return false
}
