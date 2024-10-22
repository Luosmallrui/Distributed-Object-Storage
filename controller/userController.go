package controller

import (
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/service"
	"distributed-object-storage/svc"
	"distributed-object-storage/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type UserController struct {
	userSvc *svc.UserSvc
}

func NewUserController(daoS *dao.S) *UserController {
	return &UserController{
		userSvc: svc.NewUserSvc(daoS),
	}
}

func (ctrl *UserController) RegisterRouter(r gin.IRouter) {
	g := r.Group("/user") // middwares.AuthMiddleware()
	g.GET("/list", service.DataHandlerWrapper(ctrl.ListAllUser))
	g.POST("/info/:id", service.DataHandlerWrapper(ctrl.UserInfo))
}

// ListAllUser 得到所有的用户信息列表
func (ctrl *UserController) ListAllUser(ctx *gin.Context) (interface{}, error) {
	userList, err := ctrl.userSvc.FindAllUser()
	if err != nil {
		return err.Error(), err
	}
	return userList, err
}

// UserInfo 给用户Id返回它的所有信息
func (ctrl *UserController) UserInfo(ctx *gin.Context) (interface{}, error) {
	id := ctx.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
	}
	userinfo, err := ctrl.userSvc.GetUserInfoByID(uint(userID))
	if err != nil {
		return err.Error(), err
	}
	userMetaData := &types.UserMetaData{
		Id:       userinfo.Id,
		UserName: userinfo.UserName,
	}
	return userMetaData, err
}
