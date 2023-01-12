package restapi

import (
	"errors"
	"log"
	"time"

	"github.com/0w0mewo/ssh_cert_ca/internal/restapi/controller"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

type ApiServer struct {
	router *fiber.App
}

func NewApiServer() *ApiServer {
	svr := &ApiServer{
		router: fiber.New(
			fiber.Config{
				ServerHeader:          "miao",
				AppName:               "ssh cert ca",
				ErrorHandler:          httpErrHandlerfunc,
				DisableStartupMessage: true,
				Prefork:               false,
			},
		),
	}

	return svr
}

func (as *ApiServer) init() {
	// middleware
	as.router.Use(limiter.New(limiter.Config{}))

	// register controllers
	for c := range controller.GetController() {
		c.RegisterToPath(as.router)
	}

}

func (as *ApiServer) Start(address string) {
	as.init()

	log.Println("server start at ", address)

	// start
	err := as.router.Listen(address)
	if err != nil {
		log.Println(err)
		return
	}

}

func (as *ApiServer) Close() {
	// shutdown all controllers
	for c := range controller.GetController() {
		c.Close()
	}

	// shutdown server
	err := as.router.ShutdownWithTimeout(4 * time.Second)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("server stop")

}

func httpErrHandlerfunc(c *fiber.Ctx, err error) error {
	var em *controller.CommonResp

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		em = &controller.CommonResp{
			Code:   e.Code,
			ErrMsg: e.Message,
		}
	} else { // internal functional errors
		em = &controller.CommonResp{
			Code:   -1,
			ErrMsg: err.Error(),
		}
	}
	// Return status code with error message
	return c.JSON(em)
}
