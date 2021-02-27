package config

import (
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"go.octolab.org/toolkit/config"
	"go.octolab.org/unsafe"
)

type Git struct {
	Remote string `mapstructure:"remote"`
}

type GitHub struct {
	Token config.Secret `mapstructure:"token"`
}

type Tool struct {
	Git    `mapstructure:"git,prefix"`
	GitHub `mapstructure:"github,prefix"`
}

func (cnf *Tool) Load(fs afero.Fs, bindings ...func(*Tool, *viper.Viper) error) error {
	v := viper.New()
	for _, bind := range bindings {
		if err := bind(cnf, v); err != nil {
			return err
		}
	}
	v.SetFs(fs)
	v.SetConfigFile(".env")
	unsafe.Ignore(v.ReadInConfig())

	// workaround for nested structs and prefixes
	keys := v.AllKeys()
	for _, prefix := range []string{"git", "github"} {
		for _, key := range keys {
			if !strings.HasPrefix(key, prefix) {
				continue
			}
			oldPrefix, newPrefix := prefix+"_", prefix+"."
			newKey := strings.Replace(key, oldPrefix, newPrefix, 1)
			v.Set(newKey, v.Get(key))
		}
	}

	return v.Unmarshal(cnf)
}
