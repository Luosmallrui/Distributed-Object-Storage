package service

import (
	"distributed-object-storage/errors"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type NoDataHandler = func(ctx *gin.Context) error
type DataHandler = func(ctx *gin.Context) (interface{}, error)
type Response struct {
	// 返回Code码，成功为200，为了mock方便,swagger的min和max都是200
	Code int `json:"status,omitempty"`
	// 返回数据，如果有数据的话
	Data interface{} `json:"data,omitempty"`
	// 错误描述
	Msg string `json:"msg,omitempty"`

	// Reference returns the reference document which maybe useful to solve this error.
	Reference string `json:"reference,omitempty"`
}

// stack represents a stack of program counters.
type stack []uintptr
type withCode struct {
	err   error
	code  int
	cause error
	*stack
}

func (r *Response) wrapWithErr(err error) int {
	if err == nil {
		return http.StatusOK
	}

	coder := errors.ParseCoder(err)

	if coder.Code() == 500 {
		r.Code = errors.ErrorToHTTPCode(err)
		r.Msg = err.Error()
		return http.StatusOK
	}

	r.Code = coder.Code()
	r.Msg = fmt.Sprintf("%s: %s", coder.String(), err.Error())
	r.Reference = coder.Reference()

	return coder.HTTPStatus()
}

func handleJSONResp(ctx *gin.Context, httpStatus int, resp Response) {
	_, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err)
	}
	//logs.Debug("resp:%s", string(data))

	ctx.JSON(httpStatus, resp)
}

func DataHandlerWrapper(handler DataHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var resp Response
		httpStatus := http.StatusOK
		ret, err := handler(ctx)
		if err != nil {
			httpStatus = resp.wrapWithErr(err)
		} else {
			resp.Code = http.StatusOK
			resp.Data = ret
		}
		handleJSONResp(ctx, httpStatus, resp)
	}
}
func NoDataHandlerWrapper(handler NoDataHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var resp Response
		err := handler(ctx)
		httpStatus := http.StatusOK
		if err != nil {
			httpStatus = resp.wrapWithErr(err)
		} else {
			resp.Code = http.StatusOK
		}
		handleJSONResp(ctx, httpStatus, resp)
	}
}
