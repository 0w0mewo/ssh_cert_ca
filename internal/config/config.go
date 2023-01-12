package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/0w0mewo/ssh_cert_ca/pkg/utils"
)

var Cfg *Config

type CAConfig struct {
	PrivateKeyPath string `json:"priva_key_path"`
}

type DBConfig struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

type Config struct {
	HostCA   *CAConfig `json:"host_ca"`
	UserCA   *CAConfig `json:"user_ca"`
	ListenTo string    `json:"listen_to"`
	AuthKey  string    `json:"auth_key"`
	DBconfig *DBConfig `json:"db"`
}

func LoadConfig(fname string) (cfg *Config, err error) {
	// generate default config file if it's not exist
	if !utils.IsFileExist(fname) {
		cfg = &Config{
			HostCA: &CAConfig{
				PrivateKeyPath: "ca_host",
			},
			UserCA: &CAConfig{
				PrivateKeyPath: "ca_user",
			},
			ListenTo: "127.0.0.1:8077",
			AuthKey:  utils.RandomSha1Hex(),
			DBconfig: &DBConfig{
				Driver: "sqlite3",
				DSN:    "file:certs.db?mode=rwc&cache=shared&_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=8000",
			},
		}

		cfgfileBytes, err := json.MarshalIndent(cfg, "", " ")
		if err != nil {
			return nil, err
		}

		err = ioutil.WriteFile(fname, cfgfileBytes, 0644)
		if err != nil {
			return nil, err
		}

		return cfg, nil

	}

	cf, err := os.Open(fname)
	if err != nil {
		return
	}
	defer cf.Close()

	cfg = &Config{}

	err = json.NewDecoder(cf).Decode(cfg)
	if err != nil {
		return
	}

	return

}
