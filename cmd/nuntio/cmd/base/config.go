package base

import (
	"os"
	"path"

	"github.com/mitchellh/mapstructure"
	"github.com/rigdev/rig/pkg/uuid"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Auth struct {
	UserID       uuid.UUID `yaml:"user_id,omitempty"`
	AccessToken  string    `yaml:"access_token,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
}

type Context struct {
	Name    string `yaml:"name"`
	Service string `yaml:"service"`
	User    string `yaml:"user"`
	Project struct {
		ProjectID    uuid.UUID `yaml:"project_id"`
		ProjectToken string    `yaml:"project_token"`
	} `yaml:"project"`

	service *Service
	auth    *Auth
}

type Service struct {
	Name   string `yaml:"name,omitempty"`
	Server string `yaml:"server,omitempty"`
}

type User struct {
	Name string `yaml:"name,omitempty"`
	Auth *Auth  `yaml:"auth,omitempty"`
}

type Config struct {
	Contexts []*Context `yaml:"contexts"`

	Services []*Service `yaml:"services"`

	Users []*User `yaml:"users"`

	CurrentContext string `yaml:"current_context"`

	filePath string
}

func (cfg *Config) Context() *Context {
	for _, c := range cfg.Contexts {
		if c.Name == cfg.CurrentContext {
			return c
		}
	}

	return nil
}

func (cfg *Config) Auth() *Auth {
	c := cfg.Context()
	if c == nil {
		return nil
	}

	for _, u := range cfg.Users {
		if u.Name == c.User {
			return u.Auth
		}
	}
	return nil
}

func (cfg *Config) Service() *Service {
	c := cfg.Context()
	if c == nil {
		return nil
	}

	for _, cl := range cfg.Services {
		if cl.Name == c.Service {
			return cl
		}
	}
	return nil
}

func (cfg Config) Save() error {
	bs, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cfg.filePath, bs, 0o600); err != nil {
		return err
	}

	return nil
}

func NewConfig(cfgPath string) (*Config, error) {
	cfg := &Config{
		filePath: cfgPath,
	}

	if cfg.filePath != "" {
		viper.AddConfigPath(cfg.filePath)
	} else {
		p, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}

		cfg.filePath = path.Join(p, "rig", "config.yaml")

		if _, err := os.Stat(cfg.filePath); os.IsNotExist(err) {
			if err := os.MkdirAll(path.Dir(cfg.filePath), 0o775); err != nil {
				return nil, err
			}

			if err := cfg.Save(); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	}

	_, err := os.ReadFile(cfg.filePath)
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(cfg.filePath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg, viper.DecodeHook(
		uuid.MapstructureDecodeFunc(),
	), func(cfg *mapstructure.DecoderConfig) {
		cfg.TagName = "yaml"
	}); err != nil {
		return nil, err
	}

	return cfg, nil
}
