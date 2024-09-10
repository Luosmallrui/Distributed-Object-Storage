package config

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	OssConfig OssConfig `yaml:"oss_config" json:"oss_config"`
}

type OssConfig struct {
	AK       string `yaml:"ak,omitempty"`
	SK       string `yaml:"sk,omitempty"`
	Endpoint string `yaml:"endpoint,omitempty"`
}

func InitConfig(cli *cli.Context) *Config {
	//configPath := cli.String("configPath")
	//profile := cli.String("profile")

	p := fmt.Sprintf("./config/config-dev.yaml")
	fmt.Println(p, 11)

	c, err := NewByFile(p)
	if err != nil {
		panic(err)
	}
	return c
}

func NewByFile(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("read config file failed,%v", err)
	}
	ret := new(Config)
	err = yaml.Unmarshal(data, ret)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal,%v", err)
	}
	//logs.Info("config: %s", string(data))
	return ret, nil
}
