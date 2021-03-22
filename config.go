package main

import (
	"errors"
	"strings"

	"github.com/nicklaw5/helix"
)

// Config is user plugin configuration
type Config struct {
	ClientID         string   `yaml:"clientID"`
	Token            string   `yaml:"token"`
	Priority         int      `yaml:"priority"`
	Interval         int64    `yaml:"interval"`
	OnCategoryChange bool     `yaml:"on_category_change"`
	OnOffline        bool     `yaml:"on_offline"`
	Follow           []string `yaml:"follow"`
}

// DefaultConfig implements plugin.Configurer
func (c *Plugin) DefaultConfig() interface{} {
	return &Config{
		ClientID:         "5y0akk7vcdxib7yj92vt2vmsh48jrt",
		Token:            "token",
		Priority:         4,
		Interval:         5,
		OnCategoryChange: true,
		OnOffline:        true,
		Follow:           []string{"channel1", "channel2", "channel3"},
	}
}

// Valid checks if plugin configuration is valid
func (c *Config) Valid() error {
	if c == nil {
		return errors.New("Config not found")
	}

	if c.ClientID == "" || c.Token == "" {
		return errors.New("clientid or token empty")
	}

	return nil
}

func (c Config) isFollow(username string) bool {
	for _, follow := range c.Follow {
		if strings.ToLower(follow) == strings.ToLower(username) {
			return true
		}
	}
	return false
}

// ValidateAndSetConfig implements  plugin.Configurer
func (c *Plugin) ValidateAndSetConfig(config interface{}) error {
	var err error
	newConfig := config.(*Config)

	if c.enabled {
		// Get new helix instance if credentials are changed
		if newConfig.ClientID != c.config.ClientID || newConfig.Token != c.config.Token {
			c.api, err = helix.NewClient(&helix.Options{
				ClientID:       newConfig.ClientID,
				AppAccessToken: newConfig.Token,
			})
			if err != nil {
				return err
			}
		}
		// Restart ticker if interval is changed
		if newConfig.Interval != c.config.Interval {
			c.tickerStop()
			c.tickerStart()
		}
	}

	c.config = newConfig
	return nil
}
