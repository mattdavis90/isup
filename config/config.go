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
}

func (c *Config) Check() error {
	log.Info().Msg("Checking config")

	alerterNames := make([]string, 0, len(c.Alerters))
	for n, a := range c.Alerters {
		err := a.Check()
		if err != nil {
			return err
		}
		alerterNames = append(alerterNames, n)
	}

	err := c.Schedule.Check(alerterNames)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) Reload() error {
	data, err := ioutil.ReadFile(c.filename)
	if err != nil {
		return err
	}

	var cfg Config
	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		return err
	}

	err = cfg.Check()
	if err != nil {
		return err
	}

	c.Schedule = cfg.Schedule
	c.Alerters = cfg.Alerters
	return nil
}

func LoadConfig(filename string) (*Config, error) {
	cfg := Config{
		filename: filename,
	}

	err := cfg.Reload()
	return &cfg, err
}
