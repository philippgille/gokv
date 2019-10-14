package main

import (
	"github.com/spf13/viper"
)

type Config struct {
	Implementation string
	Encoding       string
	Options        map[string]string
}

func readConfig(path string) (*Config, error) {
	v := viper.New()
	if len(path) > 0 {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("gokv")
		v.AddConfigPath(".")
	}
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return &Config{
		Implementation: v.GetString("implementation"),
		Encoding:       v.GetString("encoding"),
		Options:        v.GetStringMapString("options"),
	}, nil
}
