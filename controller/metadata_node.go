package controller

import (
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/service"
	"distributed-object-storage/svc"
	"distributed-object-storage/types"
	"fmt"
	"github.com/gin-gonic/gin"
)

type MetadataNodeController struct {
	MetadataNodeSvc *svc.MetadataSvc
}

func NewMetadataNodeController(daoS *dao.S) *MetadataNodeController {
	return &MetadataNodeController{
		MetadataNodeSvc: svc.NewMetadataSvc(daoS),
	}
}

func (ctrl *MetadataNodeController) RegisterRouter(r gin.IRouter) {
	g := r.Group("/metadata") // middwares.AuthMiddleware()
	g.GET("/object", service.DataHandlerWrapper(ctrl.GetObjectMetadata))

}

func (ctrl *MetadataNodeController) GetObjectMetadata(ctx *gin.Context) (interface{}, error) {
	options := types.GetObjectMetadataReq{}
	if err := ctx.ShouldBindQuery(&options); err != nil {
		return nil, fmt.Errorf("invaild query parameter: %v", err)
	}
	if options.ObjectName == "" {
		return nil, fmt.Errorf("empty object name")
	}
	if options.BucketName == "" {
		return nil, fmt.Errorf("empty bucket name")
	}
	return ctrl.MetadataNodeSvc.GetObjectMetadata(ctx, options.ObjectName, options.BucketName)
}
