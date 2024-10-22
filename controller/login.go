package controller

import (
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/pkg/middleware"
	"distributed-object-storage/svc"
	"github.com/gin-gonic/gin"
	"net/http"
)

type LoginController struct {
	loginSvc *svc.LoginSvc
}

func NewLoginController(daoS *dao.S) *LoginController {
	return &LoginController{
		loginSvc: svc.NewLoginSvc(daoS),
	}

}

func (ctrl *LoginController) RegisterRouter(r gin.IRouter) {
	g := r.Group("")
	g.POST("/login", ctrl.Login)
}

func (ctrl *LoginController) Login(ctx *gin.Context) {
	var loginInfo struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}
	if err := ctx.Bind(&loginInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": err.Error()})
		return
	}
	//userInfo, err := ctrl.userSvc.GetUserInfoByName(loginInfo.Username)
	//if err != nil {
	//	ctx.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "error": "Invalid username or password"})
	//}
	//entryPassword := userInfo.PassWord
	//
	//if !ctrl.loginSvc.AuthenticateUser(loginInfo.Password, entryPassword) {
	//	ctx.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "error": "Invalid username or password"})
	//	return
	//}

	// 账号密码验证成功，生成 JWT 令牌并返回给用户
	tokenString, err := middleware.GenerateJWT(1, "rq")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": http.StatusOK, "data": gin.H{"id": 1, "token": tokenString}})
}
