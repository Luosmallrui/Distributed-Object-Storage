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
	// 获取上传的文件
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("failed to get file from request: %v", err)
	}
	defer file.Close() // 确保文件流在函数结束时关闭

	// 获取文件的大小和文件名
	fileSize := header.Size
	//fileName := header.Filename
	// 从请求中获取其他表单字段，比如 bucket_name 和 object_name
	bucketName := ctx.PostForm("bucket_name")
	objectName := ctx.PostForm("object_name")

	// 验证必填字段
	if bucketName == "" || objectName == "" {
		return nil, errors.New("bucket_name and object_name required in parameter")
	}

	// 调用存储服务，上传文件
	return ctrl.StorageNodeSvc.PutObject(ctx, bucketName, objectName, file, fileSize)
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
	ctx.Header("Content-Length", fmt.Sprintf("%d", objectInfo.Size))
	ctx.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, ioReader)
		if err != nil {
			ctx.Error(err)
			return false
		}
		return true // 复制完成后返回 false 结束流
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
