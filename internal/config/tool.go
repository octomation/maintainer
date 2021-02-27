package config

import (
	"strings"
	"sync"

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

	lazy   sync.Once
	config *viper.Viper
}

func (cnf *Tool) init() *Tool {
	cnf.lazy.Do(func() {
		v := viper.New()
		v.SetConfigFile(".env")
		cnf.config = v
	})
	return cnf
}

func (cnf *Tool) Bind(bind func(*viper.Viper) error) error {
	return bind(cnf.init().config)
}

func (cnf *Tool) Load(fs afero.Fs, bindings ...func(*viper.Viper) error) error {
	v := cnf.init().config

	v.SetFs(fs)
	for _, bind := range bindings {
		if err := bind(v); err != nil {
			return err
		}
	}
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
