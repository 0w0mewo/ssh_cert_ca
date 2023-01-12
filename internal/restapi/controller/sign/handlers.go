package sign

import (
	"time"

	"github.com/0w0mewo/ssh_cert_ca/internal/model"
	"github.com/0w0mewo/ssh_cert_ca/internal/restapi/controller"
	"github.com/0w0mewo/ssh_cert_ca/pkg/service"
	"github.com/0w0mewo/ssh_cert_ca/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (r *Router) GetCAPublickey(c *fiber.Ctx) error {
	var req SignRequest
	err := c.ParamsParser(&req)
	if err != nil {
		return err
	}

	ct, err := model.ParseCertType(req.Role)
	if err != nil {
		return err
	}

	signer, err := r.getCAServiceByCertType(ct)
	if err != nil {
		return err
	}

	return c.JSON(controller.NewCommonRespWithData(signer.PublicKeyAsAuthKeyStr()))

}

// TODO: get a list of reovked certs key id and base64 encoded ssh key revoke list(KRL) file
func (r *Router) GetRevoked(c *fiber.Ctx) error {
	var req RevokeRequest

	// parse request
	err := c.ParamsParser(&req)
	if err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return err
	}

	ct, err := model.ParseCertType(req.Role)
	if err != nil {
		return err
	}

	signer, err := r.getCAServiceByCertType(ct)
	if err != nil {
		return err
	}

	return c.JSON(controller.NewCommonRespWithData(signer.GetPresentRevokedListBase64()))
}

// TODO: mark cert revoked on DB
func (r *Router) Revoke(c *fiber.Ctx) error {
	var req RevokeRequest

	err := c.ParamsParser(&req)
	if err != nil {
		return err
	}

	if err := req.Validate(); err != nil {
		return err
	}

	ct, err := model.ParseCertType(req.Role)
	if err != nil {
		return err
	}

	signer, err := r.getCAServiceByCertType(ct)
	if err != nil {
		return err
	}

	err = signer.Revoke(req.KeyId)
	if err != nil {
		return err
	}

	return c.JSON(controller.NewCommonRespWithData(nil))
}

// TODO: save signed certs info to DB
func (r *Router) Sign(c *fiber.Ctx) error {
	var req SignRequest

	// parse request
	err := c.QueryParser(&req)
	if err != nil {
		return err
	}
	err = c.ParamsParser(&req)
	if err != nil {
		return err
	}

	// validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// get public key from body
	pubkey, err := utils.ParseSSHPublicKey(c.Body())
	if err != nil {
		return err
	}

	// parse cert type
	ct, err := model.ParseCertType(req.Role)
	if err != nil {
		return err
	}

	signer, err := r.getCAServiceByCertType(ct)
	if err != nil {
		return err
	}

	// sign
	cert, err := signer.Sign(pubkey, uuid.NewString(), req.SplitedSignTo(), time.Duration(req.TTL))
	if err != nil {
		return err
	}

	return c.JSON(NewCertAsCommonResp(cert))
}

func (r *Router) getCAServiceByCertType(role model.RoleType) (*service.SSHCertCAService, error) {
	var signer *service.SSHCertCAService

	if role == model.CertTypeHost {
		signer = r.hostca
	} else if role == model.CerTypeUser {
		signer = r.userca
	} else {
		return nil, errUnknownRole
	}

	return signer, nil
}
