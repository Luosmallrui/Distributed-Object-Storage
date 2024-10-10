package controller

import (
	"distributed-object-storage/pkg/db/dao"
	"distributed-object-storage/service"
	"distributed-object-storage/svc"
	"distributed-object-storage/types"
	"fmt"
	"github.com/gin-contrib/cors"
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
	g.Use(cors.Default())
	g.GET("/object", service.DataHandlerWrapper(ctrl.GetObjectMetadata))
	g.GET("/object/list", service.DataHandlerWrapper(ctrl.ListObjectMetadata))
	g.GET("/bucket/list", service.DataHandlerWrapper(ctrl.ListBucket))
	g.POST("/bucket/:name", service.NoDataHandlerWrapper(ctrl.CreateBucket))
	g.DELETE("/bucket/:name", service.NoDataHandlerWrapper(ctrl.DeleteBucket))

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
	return ctrl.MetadataNodeSvc.GetObjectMetadata(ctx, options.BucketName, options.ObjectName)
}

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

func (ctrl *MetadataNodeController) ListBucket(ctx *gin.Context) (interface{}, error) {
	options := types.ListBucketReq{}
	if err := ctx.ShouldBindQuery(&options); err != nil {
		return nil, fmt.Errorf("invaild query parameter: %v", err)
	}
	return ctrl.MetadataNodeSvc.ListBuckets(ctx, options.Prefix, options.MaxKeys)
}

func (ctrl *MetadataNodeController) CreateBucket(ctx *gin.Context) error {
	bucketName := ctx.Param("name")
	if bucketName == "" {
		return fmt.Errorf("invalid path param, %s is blank", bucketName)
	}
	return ctrl.MetadataNodeSvc.CreateBucket(ctx, bucketName)
}

func (ctrl *MetadataNodeController) DeleteBucket(ctx *gin.Context) error {
	bucketName := ctx.Param("name")
	if bucketName == "" {
		return fmt.Errorf("invalid path param, %s is blank", bucketName)
	}
	return ctrl.MetadataNodeSvc.DeleteBucket(ctx, bucketName)
}
