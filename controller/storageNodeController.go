package controller

import (
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/service"
	"distributed-object-storage/svc"
	"distributed-object-storage/types"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
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
	//g.POST("/download", service.DataHandlerWrapper(ctrl.MetadataNodeSvc))
	g.GET("/object", ctrl.GetObject)
	g.DELETE("/delete", service.NoDataHandlerWrapper(ctrl.DeleteObject))
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
	req := types.UploadReq{}
	if err := ctx.BindJSON(&req); err != nil {
		return nil, fmt.Errorf("invaild query parameter: %v", err)
	}
	if req.BucketName == "" || req.FilePath == "" {
		return nil, errors.New("bucket_name and object_name required in parameter")
	}
	return ctrl.StorageNodeSvc.PutObject(ctx, req.BucketName, req.ObjectName, req.FilePath)
}

// GetObject 下载分文
// @Summary 获取文件信息
// @Description 根据 bucket_name 和 object_name 查询文件信息
// @Tags storage
// @Accept json
// @Produce json
// @Param  types.GetObjectMetadataReq query  types.GetObjectMetadataReq true "返回文件的相关信息"
// @Success 200 {object} object "成功返回上传的文件信息"
// @Failure 400
// @Router /storage/object [GET]
func (ctrl *StorageNodeController) GetObject(ctx *gin.Context) {
	req := types.GetObjectMetadataReq{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	res := types.GetObjectReq{BucketName: req.BucketName}
	ioReader, objectInfo, err := ctrl.StorageNodeSvc.GetObject(ctx, req.BucketName, req.ObjectName)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	res.ObjectInfo = objectInfo
	res.FileReader = ioReader
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", objectInfo.Name))
	//ctx.Header("Content-Length", fmt.Sprintf("%d", objectInfo.Size))
	ctx.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, ioReader)
		if err != nil {
			// 处理错误
			ctx.Error(err)
			return false
		}
		return false // 复制完成后返回 false 结束流
	})
	return
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
// @Router /storage/delete [DELETE]
func (ctrl *StorageNodeController) DeleteObject(ctx *gin.Context) error {
	req := types.GetObjectMetadataReq{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return fmt.Errorf("invaild query parameter: %v", err)
	}
	return ctrl.StorageNodeSvc.DeleteObject(ctx, req.BucketName, req.ObjectName)
}