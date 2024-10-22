package controller

import (
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/pkg/middleware"
	"distributed-object-storage/svc"
	"distributed-object-storage/types"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthController struct {
	authSvc *svc.AuthSvc
	userSvc *svc.UserSvc
}

func NewAuthController(daoS *dao.S) *AuthController {
	return &AuthController{
		authSvc: svc.NewAuthSvc(daoS),
		userSvc: svc.NewUserSvc(daoS),
	}

}

func (ctrl *AuthController) RegisterRouter(r gin.IRouter) {
	g := r.Group("")
	g.POST("/login", ctrl.Login)
	g.POST("/register", ctrl.Register)
}

func (ctrl *AuthController) Login(ctx *gin.Context) {
	var loginInfo struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}
	if err := ctx.Bind(&loginInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": err.Error()})
		return
	}
	userInfo, err := ctrl.userSvc.GetUserInfoByName(loginInfo.Username)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "error": "Invalid username or password"})
		return
	}

	if !ctrl.authSvc.AuthenticateUser(loginInfo.Password, userInfo.PassWord) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "error": "Invalid username 1or password"})
		return
	}

	// 账号密码验证成功，生成 JWT 令牌并返回给用户
	tokenString, err := middleware.GenerateJWT(userInfo.Id, "rq")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (ctrl *AuthController) Register(ctx *gin.Context) {
	var registerInfo struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&registerInfo); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"status": http.StatusBadRequest, "error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	existingUser, err := ctrl.userSvc.GetUserInfoByName(registerInfo.Username)
	if existingUser != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusConflict, gin.H{"status": http.StatusConflict, "error": "user already exists"})
		return
	}

	// 创建新用户
	userInfo := &types.UserMetaData{
		UserName: registerInfo.Username,
		Password: registerInfo.Password,
	}

	err = ctrl.userSvc.CreateUser(ctx.Request.Context(), userInfo)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Failed to create user"})
		return
	}

	// 生成 JWT 令牌
	tokenString, err := middleware.GenerateJWT(uint(int(userInfo.Id)), userInfo.UserName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"status": http.StatusInternalServerError, "error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": http.StatusCreated, "data": gin.H{"id": userInfo.Id, "username": userInfo.UserName, "token": tokenString}})
}
