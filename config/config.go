package config

import (
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"isup/scheduler"
)

type Config struct {
	filename string
	Schedule scheduler.Schedule
	Alerters map[string]*scheduler.Alerter
	Client   *Client
}

func (c *Config) Check() (*Config, error) {
	log.Info().Msg("Checking config")

	alerterNames := make([]string, 0, len(c.Alerters))
	for n, a := range c.Alerters {
		err := a.Check()
		if err != nil {
			return nil, err
		}
		alerterNames = append(alerterNames, n)
	}

	err := c.Schedule.Check(alerterNames)
	if err != nil {
		return nil, err
	}

	if c.Client == nil {
		c.Client = NewClient()
	} else {
		err = c.Client.Check()
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Config) Reload() (*Config, error) {
	data, err := ioutil.ReadFile(c.filename)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal([]byte(data), cfg)
	if err != nil {
		return nil, err
	}

	cfg, err = cfg.Check()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func LoadConfig(filename string) (*Config, error) {
	cfg := &Config{
		filename: filename,
	}

	return cfg.Reload()
}
