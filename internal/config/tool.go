package config

import (
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
	FS     afero.Fs `mapstructure:"-"`
	Git    `mapstructure:"git,squash"`
	GitHub `mapstructure:"github,squash"`

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
	cnf.FS = fs

	for _, bind := range bindings {
		if err := bind(v); err != nil {
			return err
		}
	}
	unsafe.Ignore(v.ReadInConfig())

	return v.Unmarshal(cnf)
}
