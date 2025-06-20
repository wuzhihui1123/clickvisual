package initialize

import (
	"github.com/clickvisual/clickvisual/api/internal/invoker"
	"github.com/clickvisual/clickvisual/api/internal/pkg/component/core"
	"github.com/clickvisual/clickvisual/api/internal/pkg/model/db"
	"github.com/clickvisual/clickvisual/api/internal/service/install"
	"github.com/clickvisual/clickvisual/api/internal/service/permission"
)

// IsInstall Determine whether the installation process is required
// @Tags         INSTILL
func IsInstall(c *core.Context) {
	var u db.User
	err := invoker.Db.Table(db.TableNameUser).Select("id, username").Limit(1).First(&u).Error
	if err != nil {
		c.JSONOK(0)
		return
	}
	if u.ID == 0 {
		c.JSONOK(0)
		return
	}
	c.JSONOK(1)
}

// Install Perform the installation process
// @Tags         INSTILL
func Install(c *core.Context) {
	err := install.Install()
	if err != nil {
		c.JSONE(1, err.Error(), nil)
		return
	}
	c.JSONOK("install finish")
}

// @Tags         INSTILL
func Migration(c *core.Context) {
	defaultAdminId := 1
	roots := permission.Manager.GetRootUsersId()
	isExist := false
	for _, val := range roots {
		if val == defaultAdminId {
			isExist = true
		}
	}
	if !isExist {
		roots = append(roots, defaultAdminId)
	}
	permission.Manager.GrantRootUsers(roots)
	if err := permission.Manager.IsRootUser(c.Uid()); err != nil {
		c.JSONE(1, "permission verification failed", err)
		return
	}
	err := install.Migration()
	if err != nil {
		c.JSONE(1, err.Error(), nil)
		return
	}
	c.JSONOK("migration finish")
}
