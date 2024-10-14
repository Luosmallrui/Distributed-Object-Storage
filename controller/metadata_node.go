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
	g.GET("/object/list", service.DataHandlerWrapper(ctrl.ListObjectMetadata))
	g.GET("/bucket/list", service.DataHandlerWrapper(ctrl.ListBucket))
	g.POST("/bucket/:name", service.NoDataHandlerWrapper(ctrl.CreateBucket))
	g.DELETE("/bucket/:name", service.NoDataHandlerWrapper(ctrl.DeleteBucket))

}

// GetObjectMetadata 获取对象元数据信息
// @Summary 获取对象元数据信息
// @Description 根据 bucket_name 和 object_name 查询对象元数据信息
// @Tags metadata
// @Accept json
// @Produce json
// @Param  types.GetObjectMetadataReq query  types.GetObjectMetadataReq true "Bucket Name"
// @Success 200 {object} types.ObjectMetadata
// @Failure 400
// @Router /metadata/object [get]
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
	return ctrl.MetadataNodeSvc.GetObjectMetadata(ctx, options.BucketName, options.ObjectName)
}

// ListObjectMetadata 获取对象元数据列表
// @Summary 获取对象元数据列表
// @Description 根据 bucket_name、prefix 和 max_keys 查询对象元数据
// @Tags metadata
// @Accept json
// @Produce json
// @Param bucket_name query string true "Bucket Name"
// @Param prefix query string false "Prefix"
// @Param max_keys query int false "Maximum number of keys to return"
// @Success 200 {array} types.ObjectInfo
// @Failure 400
// @Router /metadata/object/list [get]
func (ctrl *MetadataNodeController) ListObjectMetadata(ctx *gin.Context) (interface{}, error) {
	options := types.ListObjectMetadataReq{}
	if err := ctx.ShouldBindQuery(&options); err != nil {
		return nil, fmt.Errorf("invaild query parameter: %v", err)
	}
	if options.BucketName == "" {
		return nil, fmt.Errorf("empty bucket name")
	}
	return ctrl.MetadataNodeSvc.ListObjects(ctx, options.BucketName, options.Prefix, options.MaxKeys)
}

// ListBucket 获取Bucket列表
// @Summary 获取Bucket列表
// @Description 根据 prefix 和 max_keys 查询Bucket
// @Tags metadata
// @Accept json
// @Produce json
// @Param prefix query string false "Prefix"
// @Param max_keys query int false "Maximum number of keys to return"
// @Success 200 {array} types.BucketInfo
// @Failure 400
// @Router /metadata/bucket/list [get]
func (ctrl *MetadataNodeController) ListBucket(ctx *gin.Context) (interface{}, error) {
	options := types.ListBucketReq{}
	if err := ctx.ShouldBindQuery(&options); err != nil {
		return nil, fmt.Errorf("invaild query parameter: %v", err)
	}
	return ctrl.MetadataNodeSvc.ListBuckets(ctx, options.Prefix, options.MaxKeys)
}

// CreateBucket 创建Bucket
// @Summary 创建Bucket
// @Description 根据 name 创建Bucket
// @Tags metadata
// @Accept json
// @Produce json
// @Param name path string true "Bucket名字"
// @Success 200
// @Failure 400
// @Router /metadata/:name [POST]
func (ctrl *MetadataNodeController) CreateBucket(ctx *gin.Context) error {
	bucketName := ctx.Param("name")
	if bucketName == "" {
		return fmt.Errorf("invalid path param, %s is blank", bucketName)
	}
	return ctrl.MetadataNodeSvc.CreateBucket(ctx, bucketName)
}

// DeleteBucket 删除Bucket
// @Summary 删除Bucket
// @Description 根据 name 删除Bucket
// @Tags metadata
// @Accept json
// @Produce json
// @Param name path string true "Bucket名字"
// @Success 200
// @Failure 400
// @Router /metadata/:name [DELETE]
func (ctrl *MetadataNodeController) DeleteBucket(ctx *gin.Context) error {
	bucketName := ctx.Param("name")
	if bucketName == "" {
		return fmt.Errorf("invalid path param, %s is blank", bucketName)
	}
	return ctrl.MetadataNodeSvc.DeleteBucket(ctx, bucketName)
}
