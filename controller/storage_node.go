package controller

import (
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/service"
	"distributed-object-storage/svc"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

type StorageNodeController struct {
	MetadataNodeSvc *svc.MetadataSvc
	StorageNodeSvc  *svc.StorageNodeSvc
}

func NewStorageNodeController(daoS *dao.S) *StorageNodeController {
	return &StorageNodeController{
		MetadataNodeSvc: svc.NewMetadataSvc(daoS),
		StorageNodeSvc:  svc.NewStorageNodeSvc(daoS),
	}
}

func (ctrl *StorageNodeController) RegisterRouter(r gin.IRouter) {
	g := r.Group("/storage") // middwares.AuthMiddleware()
	g.POST("/", service.DataHandlerWrapper(ctrl.PutObject))
}

func (ctrl *StorageNodeController) PutObject(ctx *gin.Context) (interface{}, error) {
	bucket := ctx.PostForm("bucket_name")
	object := ctx.PostForm("object_name")
	if bucket == "" || object == "" {
		return nil, errors.New("bucket_name and object_name required in parameter")
	}
	file, size, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("form file: %v", err)
	}
	metadata := make(map[string]string)
	metadata["size"] = strconv.FormatInt(size.Size, 10)
	return ctrl.StorageNodeSvc.PutObject(ctx, bucket, object, file, metadata)
}
