package sign

import (
	"log"

	"github.com/0w0mewo/ssh_cert_ca/internal/config"
	"github.com/0w0mewo/ssh_cert_ca/internal/model"
	"github.com/0w0mewo/ssh_cert_ca/internal/restapi/controller"
	"github.com/0w0mewo/ssh_cert_ca/pkg/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/keyauth/v2"
)

func init() {
	controller.RegisterController(&Router{})
}

type Router struct {
	userca *service.SSHCertCAService
	hostca *service.SSHCertCAService
}

func (r *Router) RegisterToPath(attchedTo *fiber.App) {
	var err error
	if r.userca == nil {
		r.userca, err = service.NewSSHCertCAService(config.Cfg.DBconfig.Driver, config.Cfg.DBconfig.DSN,
			config.Cfg.UserCA.PrivateKeyPath, "", model.CerTypeUser)
		if err != nil {
			panic(err)
		}
	}

	if r.hostca == nil {
		r.hostca, err = service.NewSSHCertCAService(config.Cfg.DBconfig.Driver, config.Cfg.DBconfig.DSN,
			config.Cfg.HostCA.PrivateKeyPath, "", model.CertTypeHost)
		if err != nil {
			panic(err)
		}
	}

	grp := attchedTo.Group("/ca")

	grp.Use(keyauth.New(keyauth.Config{
		Validator: func(c *fiber.Ctx, s string) (bool, error) {
			if s == config.Cfg.AuthKey {
				return true, nil
			}

			return false, errInvalidAuthKey
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			log.Printf("auth success from %s: %s %s", c.IP(), c.Method(), c.Path())

			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("auth fail from %s: %s %s", c.IP(), c.Method(), c.Path())

			return c.JSON(controller.CommonResp{
				Code:   -1,
				ErrMsg: err.Error(),
				Data:   nil,
			})
		},
	}))

	// routes
	{
		grp.Post("/sign/:role", r.Sign)
		grp.Get("/capubkey/:role", r.GetCAPublickey)
		grp.Delete("/revoke/:role/:keyid", r.Revoke)
		grp.Get("/getrevoked/:role", r.GetRevoked)
	}

}

func (r *Router) Close() {

}
