package controller

import (
	"fmt"
	"strconv"
	"x-ui/database/model"
	"x-ui/logger"
	"x-ui/web/global"
	"x-ui/web/service"
	"x-ui/web/session"

	"github.com/gin-gonic/gin"
)

type InboundController struct {
	inboundService service.InboundService
	xrayService    service.XrayService
}

func NewInboundController(g *gin.RouterGroup) *InboundController {
	a := &InboundController{}
	a.initRouter(g)
	a.startTask()
	return a
}

func (a *InboundController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/inbound")

	g.POST("/list", a.getInbounds)
	g.POST("/add", a.addInbound)
	g.POST("/del/:id", a.delInbound)
	g.POST("/update/:id", a.updateInbound)
	g.POST("/clientIps/:email", a.getClientIps)
	g.POST("/clearClientIps/:email", a.clearClientIps)
	g.POST("/addClient", a.addInboundClient)
	g.POST("/delClient/:email", a.delInboundClient)
	g.POST("/updateClient/:index", a.updateInboundClient)
	g.POST("/:id/resetClientTraffic/:email", a.resetClientTraffic)
	g.POST("/resetAllTraffics", a.resetAllTraffics)
	g.POST("/resetAllClientTraffics/:id", a.resetAllClientTraffics)

}

func (a *InboundController) startTask() {
	webServer := global.GetWebServer()
	c := webServer.GetCron()
	c.AddFunc("@every 10s", func() {
		if a.xrayService.IsNeedRestartAndSetFalse() {
			err := a.xrayService.RestartXray(false)
			if err != nil {
				logger.Error("restart xray failed:", err)
			}
		}
	})
}

func (a *InboundController) getInbounds(c *gin.Context) {
	user := session.GetLoginUser(c)
	inbounds, err := a.inboundService.GetInbounds(user.Id)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, inbounds, nil)
}
func (a *InboundController) getInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18n(c, "get"), err)
		return
	}
	inbound, err := a.inboundService.GetInbound(id)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.toasts.obtain"), err)
		return
	}
	jsonObj(c, inbound, nil)
}

func (a *InboundController) addInbound(c *gin.Context) {
	inbound := &model.Inbound{}
	err := c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.addTo"), err)
		return
	}
	user := session.GetLoginUser(c)
	inbound.UserId = user.Id
	inbound.Enable = true
	inbound.Tag = fmt.Sprintf("inbound-%v", inbound.Port)
	inbound, err = a.inboundService.AddInbound(inbound)
	jsonMsgObj(c, I18n(c, "pages.inbounds.addTo"), inbound, err)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) delInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18n(c, "delete"), err)
		return
	}
	err = a.inboundService.DelInbound(id)
	jsonMsgObj(c, I18n(c, "delete"), id, err)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) updateInbound(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.revise"), err)
		return
	}
	inbound := &model.Inbound{
		Id: id,
	}
	err = c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.revise"), err)
		return
	}
	inbound, err = a.inboundService.UpdateInbound(inbound)
	jsonMsgObj(c, I18n(c, "pages.inbounds.revise"), inbound, err)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) getClientIps(c *gin.Context) {
	email := c.Param("email")

	ips, err := a.inboundService.GetInboundClientIps(email)
	if err != nil {
		jsonObj(c, "No IP Record", nil)
		return
	}
	jsonObj(c, ips, nil)
}
func (a *InboundController) clearClientIps(c *gin.Context) {
	email := c.Param("email")

	err := a.inboundService.ClearClientIps(email)
	if err != nil {
		jsonMsg(c, "修改", err)
		return
	}
	jsonMsg(c, "Log Cleared", nil)
}
func (a *InboundController) addInboundClient(c *gin.Context) {
	data := &model.Inbound{}
	err := c.ShouldBind(data)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.revise"), err)
		return
	}

	err = a.inboundService.AddInboundClient(data)
	if err != nil {
		jsonMsg(c, "something worng!", err)
		return
	}
	jsonMsg(c, "Client(s) added", nil)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) delInboundClient(c *gin.Context) {
	email := c.Param("email")
	inbound := &model.Inbound{}
	err := c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.revise"), err)
		return
	}

	err = a.inboundService.DelInboundClient(inbound, email)
	if err != nil {
		jsonMsg(c, "something worng!", err)
		return
	}
	jsonMsg(c, "Client deleted", nil)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) updateInboundClient(c *gin.Context) {
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.revise"), err)
		return
	}

	inbound := &model.Inbound{}
	err = c.ShouldBind(inbound)
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.revise"), err)
		return
	}

	err = a.inboundService.UpdateInboundClient(inbound, index)
	if err != nil {
		jsonMsg(c, "something worng!", err)
		return
	}
	jsonMsg(c, "Client updated", nil)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) resetClientTraffic(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.revise"), err)
		return
	}
	email := c.Param("email")

	err = a.inboundService.ResetClientTraffic(id, email)
	if err != nil {
		jsonMsg(c, "something worng!", err)
		return
	}
	jsonMsg(c, "traffic reseted", nil)
	if err == nil {
		a.xrayService.SetToNeedRestart()
	}
}

func (a *InboundController) resetAllTraffics(c *gin.Context) {
	err := a.inboundService.ResetAllTraffics()
	if err != nil {
		jsonMsg(c, "something worng!", err)
		return
	}
	jsonMsg(c, "All traffics reseted", nil)
}

func (a *InboundController) resetAllClientTraffics(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18n(c, "pages.inbounds.revise"), err)
		return
	}

	err = a.inboundService.ResetAllClientTraffics(id)
	if err != nil {
		jsonMsg(c, "something worng!", err)
		return
	}
	jsonMsg(c, "All traffics of client reseted", nil)
}
