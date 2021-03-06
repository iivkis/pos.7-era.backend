package myservice

import "github.com/gin-gonic/gin"

type Response struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

func NewResponse(c *gin.Context, code int, data interface{}) {
	var obj Response
	if data == nil {
		data = struct{}{}
	}
	obj.Data = data

	if _, ok := data.(*serviceError); !ok {
		obj.Status = true
	}

	c.JSON(code, obj)
}
