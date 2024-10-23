package controller

import (
	"crypto/md5"
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/service"
	"distributed-object-storage/svc"
	"distributed-object-storage/types"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
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
	g.POST("/upload", ctrl.PutObject)
	g.GET("/object", ctrl.GetObject)
	g.POST("/pause/:uploadId", handlePause)
	g.POST("/resume/:uploadId", handleResume)
	g.POST("/cancel/:uploadId", handleCancel)
	g.GET("/status/:uploadId", handleStatus)
	g.DELETE("/delete", service.NoDataHandlerWrapper(ctrl.DeleteObject))
}

func handleResume(c *gin.Context) {
	uploadID := c.Param("uploadId")
	types.UploadTasks.RLock()
	status, exists := types.UploadTasks.Tasks[uploadID]
	types.UploadTasks.RUnlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload task not found"})
		return
	}
	status.Mutex.Lock()
	status.IsPaused = false
	status.Mutex.Unlock()
	c.JSON(http.StatusOK, gin.H{"message": "Upload resumed"})
}

func handleStatus(c *gin.Context) {
	uploadID := c.Param("uploadId")

	types.UploadTasks.RLock()
	status, exists := types.UploadTasks.Tasks[uploadID]
	types.UploadTasks.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload task not found"})
		return
	}

	status.Mutex.Lock()
	response := gin.H{
		"is_paused":       status.IsPaused,
		"is_canceled":     status.IsCanceled,
		"current_part":    status.CurrentPart,
		"completed_parts": len(status.CompletedParts),
	}
	status.Mutex.Unlock()
	c.JSON(http.StatusOK, response)
}

func handleCancel(c *gin.Context) {
	uploadID := c.Param("uploadId")

	types.UploadTasks.RLock()
	status, exists := types.UploadTasks.Tasks[uploadID]
	types.UploadTasks.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload task not found"})
		return
	}

	status.Mutex.Lock()
	status.IsCanceled = true
	status.Mutex.Unlock()

	// 清理上传任务
	types.UploadTasks.Lock()
	delete(types.UploadTasks.Tasks, uploadID)
	types.UploadTasks.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "Upload canceled"})
}
func handlePause(c *gin.Context) {
	uploadID := c.Param("uploadId")

	types.UploadTasks.RLock()
	status, exists := types.UploadTasks.Tasks[uploadID]
	types.UploadTasks.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upload task not found"})
		return
	}

	status.Mutex.Lock()
	status.IsPaused = true
	status.Mutex.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": "Upload paused"})
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
func (ctrl *StorageNodeController) PutObject(ctx *gin.Context) {
	bucketName := ctx.PostForm("bucket_name")
	objectName := ctx.PostForm("object_name")

	if bucketName == "" || objectName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bucketName or objectName is empty"})
		return
	}

	// 生成 uploadId
	hash := md5.Sum([]byte(objectName))
	uploadId := fmt.Sprintf("%s-%d", hex.EncodeToString(hash[:])[:8], time.Now().UnixNano())

	uploadStatus := &types.UploadStatus{
		UploadID:   uploadId,
		IsPaused:   false,
		IsCanceled: false,
	}

	types.UploadTasks.Lock()
	types.UploadTasks.Tasks[uploadStatus.UploadID] = uploadStatus
	types.UploadTasks.Unlock()

	// 立即返回响应
	ctx.JSON(http.StatusOK, gin.H{
		"upload_id": uploadStatus.UploadID,
		"message":   "Upload started",
	})

	// 在 goroutine 中打开和处理文件
	go func() {
		err := ctx.Request.ParseMultipartForm(32 << 20) // 32MB buffer
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 获取文件头信息
		header, err := ctx.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		file, err := header.Open()
		if err != nil {
			types.UploadTasks.Lock()
			if task, ok := types.UploadTasks.Tasks[uploadStatus.UploadID]; ok {
				task.Error = err
			}
			types.UploadTasks.Unlock()
			return
		}
		defer file.Close()

		_, err = ctrl.StorageNodeSvc.PutObject(ctx, bucketName, objectName, file, header.Size, uploadStatus.UploadID)
		if err != nil {
			types.UploadTasks.Lock()
			if task, ok := types.UploadTasks.Tasks[uploadStatus.UploadID]; ok {
				task.Error = err
			}
			types.UploadTasks.Unlock()
		}
	}()
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
