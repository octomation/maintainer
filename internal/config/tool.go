package config

import (
	"sync"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"go.octolab.org/toolkit/config"
)

type Git struct {
	// Keys match viper's flat env/flag keys (GIT_REMOTE → git_remote). Since
	// viper 1.20 reads these as flat keys, the squashed mapstructure tag must
	// use the full key, not the bare "remote"/"token".
	Remote string `mapstructure:"git_remote"`
}

type GitHub struct {
	Token config.Secret `mapstructure:"github_token"`
}

type Tool struct {
	FS     afero.Fs `mapstructure:"-"`
	Git    `mapstructure:"git,squash"`
	GitHub `mapstructure:"github,squash"`

	lazy   sync.Once
	config *viper.Viper
}

func (cnf *Tool) init() *Tool {
	// No file is read: the working directory is not a source of
	// configuration, so the token comes from the environment only.
	cnf.lazy.Do(func() { cnf.config = viper.New() })
	return cnf
}

func (cnf *Tool) Bind(bind func(*viper.Viper) error) error {
	return bind(cnf.init().config)
}

func (cnf *Tool) Load(fs afero.Fs, bindings ...func(*viper.Viper) error) error {
	v := cnf.init().config
	cnf.FS = fs

	for _, bind := range bindings {
		if err := bind(v); err != nil {
			return err
		}
	}
	return v.Unmarshal(cnf)
}
