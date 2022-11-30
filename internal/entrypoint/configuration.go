package entrypoint

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type config struct {
	ClientID     string `yaml:"clientID"`
	Username     string `yaml:"username"`
	RedirectPort string `yaml:"redirectPort"`
	LineSpacing  int    `yaml:"lineSpacing"`
	Token        string `yaml:"token"`
	hd           string
}

func newConfig(hd string) (*config, error) {
	f, err := os.ReadFile(filepath.Join(hd, ".ttchat", "config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	conf := &config{hd: hd}
	err = yaml.Unmarshal(f, conf)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return conf, nil
}

func (c *config) save() error {
	b, err := yaml.Marshal(&c)
	if err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	err = os.WriteFile(filepath.Join(c.hd, ".ttchat", "config.yaml"), b, 0644)
	if err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	return nil
}
