package controller

import (
	"errors"
	"time"
	"x-ui/web/entity"
	"x-ui/web/service"
	"x-ui/web/session"

	"github.com/gin-gonic/gin"
)

type updateUserForm struct {
	OldUsername string `json:"oldUsername" form:"oldUsername"`
	OldPassword string `json:"oldPassword" form:"oldPassword"`
	NewUsername string `json:"newUsername" form:"newUsername"`
	NewPassword string `json:"newPassword" form:"newPassword"`
}

type SettingController struct {
	settingService service.SettingService
	userService    service.UserService
	panelService   service.PanelService
}

func NewSettingController(g *gin.RouterGroup) *SettingController {
	a := &SettingController{}
	a.initRouter(g)
	return a
}

func (a *SettingController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/setting")

	g.POST("/all", a.getAllSetting)
	g.POST("/defaultSettings", a.getDefaultSettings)
	g.POST("/update", a.updateSetting)
	g.POST("/updateUser", a.updateUser)
	g.POST("/restartPanel", a.restartPanel)
	g.GET("/getDefaultJsonConfig", a.getDefaultJsonConfig)
}

func (a *SettingController) getAllSetting(c *gin.Context) {
	allSetting, err := a.settingService.GetAllSetting()
	if err != nil {
		jsonMsg(c, I18n(c, "pages.setting.toasts.getSetting"), err)
		return
	}
	jsonObj(c, allSetting, nil)
}

func (a *SettingController) getDefaultJsonConfig(c *gin.Context) {
	defaultJsonConfig, err := a.settingService.GetDefaultJsonConfig()
	if err != nil {
		jsonMsg(c, I18n(c, "pages.setting.toasts.getSetting"), err)
		return
	}
	jsonObj(c, defaultJsonConfig, nil)
}

func (a *SettingController) getDefaultSettings(c *gin.Context) {
	expireDiff, err := a.settingService.GetExpireDiff()
	if err != nil {
		jsonMsg(c, I18n(c, "pages.setting.toasts.getSetting"), err)
		return
	}
	trafficDiff, err := a.settingService.GetTrafficDiff()
	if err != nil {
		jsonMsg(c, I18n(c, "pages.setting.toasts.getSetting"), err)
		return
	}
	defaultCert, err := a.settingService.GetCertFile()
	if err != nil {
		jsonMsg(c, I18n(c, "pages.setting.toasts.getSetting"), err)
		return
	}
	defaultKey, err := a.settingService.GetKeyFile()
	if err != nil {
		jsonMsg(c, I18n(c, "pages.setting.toasts.getSetting"), err)
		return
	}
	result := map[string]interface{}{
		"expireDiff":  expireDiff,
		"trafficDiff": trafficDiff,
		"defaultCert": defaultCert,
		"defaultKey":  defaultKey,
	}
	jsonObj(c, result, nil)
}

func (a *SettingController) updateSetting(c *gin.Context) {
	allSetting := &entity.AllSetting{}
	err := c.ShouldBind(allSetting)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.setting.toasts.modifySetting"), err)
		return
	}
	err = a.settingService.UpdateAllSetting(allSetting)
	jsonMsg(c, I18n(c, "pages.setting.toasts.modifySetting"), err)
}

func (a *SettingController) updateUser(c *gin.Context) {
	form := &updateUserForm{}
	err := c.ShouldBind(form)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.setting.toasts.modifySetting"), err)
		return
	}
	user := session.GetLoginUser(c)
	if user.Username != form.OldUsername || user.Password != form.OldPassword {
		jsonMsg(c, I18n(c, "pages.setting.toasts.modifyUser"), errors.New(I18n(c, "pages.setting.toasts.originalUserPassIncorrect")))
		return
	}
	if form.NewUsername == "" || form.NewPassword == "" {
		jsonMsg(c, I18n(c, "pages.setting.toasts.modifyUser"), errors.New(I18n(c, "pages.setting.toasts.userPassMustBeNotEmpty")))
		return
	}
	err = a.userService.UpdateUser(user.Id, form.NewUsername, form.NewPassword)
	if err == nil {
		user.Username = form.NewUsername
		user.Password = form.NewPassword
		session.SetLoginUser(c, user)
	}
	jsonMsg(c, I18n(c, "pages.setting.toasts.modifyUser"), err)
}

func (a *SettingController) restartPanel(c *gin.Context) {
	err := a.panelService.RestartPanel(time.Second * 3)
	jsonMsg(c, I18n(c, "pages.setting.restartPanel"), err)
}
