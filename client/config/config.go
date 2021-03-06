package config

import (
	"os"
	"p2pNat/client/common"

	"gopkg.in/yaml.v2"
)

func ReadCfg(path string) (*common.Config, error) {
	conf := &common.Config{}
	if f, err := os.Open(path); err != nil {
		return nil, err
	} else {
		yaml.NewDecoder(f).Decode(conf)
	}
	return conf, nil
}
