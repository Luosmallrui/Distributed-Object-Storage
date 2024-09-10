package controller

import (
	"distributed-object-storage/errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
)

func ParseBody(ctx *gin.Context, obj interface{}) error {
	err := ctx.BindJSON(obj)
	if err != nil {
		if b, _ := io.ReadAll(ctx.Request.Body); err != nil {
			fmt.Printf("Body内容: %v", string(b))
		}
		return fmt.Errorf("%w body格式错误: %v", errors.ErrBadRequest, err)
	}
	return nil
}
