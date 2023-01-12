package controller

import (
	"github.com/gofiber/fiber/v2"
)

type Controller interface {
	RegisterToPath(attchedTo *fiber.App)
	Close()
}

type Validatable interface {
	Validate() error
}

type CommonResp struct {
	Code   int    `json:"code"`
	ErrMsg string `json:"errMsg"`
	Data   any    `json:"data"`
}

func NewCommonRespWithData(data any) *CommonResp {
	return &CommonResp{
		Code:   0,
		ErrMsg: "OK",
		Data:   data,
	}

}

var Controllers map[Controller]bool = make(map[Controller]bool)

func RegisterController(c Controller) {
	Controllers[c] = true
}

func GetController() map[Controller]bool {
	return Controllers
}
