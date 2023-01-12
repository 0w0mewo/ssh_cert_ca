package sign

import (
	"strings"
	"time"

	"github.com/0w0mewo/ssh_cert_ca/internal/model"
	"github.com/0w0mewo/ssh_cert_ca/internal/restapi/controller"
)

type SignRequest struct {
	Role   string `query:"-" params:"role"`
	SignTo string `query:"signto"`
	TTL    uint64 `query:"ttl"`
}

func (srq SignRequest) SplitedSignTo() []string {
	return strings.Split(srq.SignTo, ",")
}

func (srq *SignRequest) Validate() error {
	if srq.Role == "" || srq.SignTo == "" {
		return errInvalidInput
	}

	if srq.TTL == 0 {
		srq.TTL = uint64(8 * time.Hour)
	}

	return nil

}

type RevokeRequest struct {
	Role  string `params:"role"`
	KeyId string `params:"keyid"`
}

func (rr RevokeRequest) Validate() error {
	if rr.Role == "" || rr.KeyId == "" {
		return errInvalidInput
	}

	return nil
}

func NewCertAsCommonResp(cert model.Cert) *controller.CommonResp {
	return &controller.CommonResp{
		Code:   0,
		ErrMsg: "OK",
		Data:   cert,
	}
}
