package controller

import (
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/service"
	"distributed-object-storage/svc"
	"distributed-object-storage/types"
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
	g.POST("/upload", service.DataHandlerWrapper(ctrl.PutObject))
	g.GET("/object", service.DataHandlerWrapper(ctrl.GetObject))
	g.DELETE("/", service.NoDataHandlerWrapper(ctrl.DeleteObject))
}

// PutObject 上传文件
// @Summary 上传文件
// @Description 根据 bucket_name 和 object_name 上传文件
// @Tags storage
// @Accept multipart/form-data
// @Produce json
// @Param bucket_name formData string true "Bucket Name"
// @Param object_name formData string true "Object Name"
// @Param file formData file true "File to upload"
// @Success 200 {string} string "成功返回上传的文件ETag"
// @Failure 400   "错误响应"
// @Router /storage/upload [POST]
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

// GetObject 获取文件信息
// @Summary 获取文件信息
// @Description 根据 bucket_name 和 object_name 查询文件信息
// @Tags storage
// @Accept json
// @Produce json
// @Param  types.GetObjectMetadataReq query  types.GetObjectMetadataReq true "返回文件的相关信息"
// @Success 200 {object} object "成功返回上传的文件信息"
// @Failure 400
// @Router /storage/object [GET]
func (ctrl *StorageNodeController) GetObject(ctx *gin.Context) (interface{}, error) {
	req := types.GetObjectMetadataReq{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, fmt.Errorf("invaild query parameter: %v", err)
	}
	res := types.GetObjectReq{BucketName: req.BucketName}
	ioReader, objectInfo, err := ctrl.StorageNodeSvc.GetObject(ctx, req.BucketName, req.ObjectName)
	res.ObjectInfo = objectInfo
	res.FileReader = ioReader
	return res, err
}

// DeleteObject 删除文件
// @Summary 删除文件
// @Description 根据 bucket_name 和object_name 删除文件
// @Tags storage
// @Accept json
// @Produce json
// @Param  types.GetObjectMetadataReq query  types.GetObjectMetadataReq true "Bucket Name"
// @Success 200
// @Failure 400
// @Router /storage [DELETE]
func (ctrl *StorageNodeController) DeleteObject(ctx *gin.Context) error {
	req := types.GetObjectMetadataReq{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return fmt.Errorf("invaild query parameter: %v", err)
	}
	return ctrl.StorageNodeSvc.DeleteObject(ctx, req.BucketName, req.ObjectName)
}
