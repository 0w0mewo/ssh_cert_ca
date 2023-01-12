package main

import (
	"flag"

	"github.com/0w0mewo/ssh_cert_ca/internal/config"
	"github.com/0w0mewo/ssh_cert_ca/internal/restapi"
	"github.com/0w0mewo/ssh_cert_ca/pkg/utils"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "config.json", "config file path")
	flag.Parse()
}

func main() {
	var err error
	config.Cfg, err = config.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	svr := restapi.NewApiServer()
	go svr.Start(config.Cfg.ListenTo)

	<-utils.WaitForSignal()
	svr.Close()

}
