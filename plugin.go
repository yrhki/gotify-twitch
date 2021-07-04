package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gotify/plugin-api"
	"github.com/nicklaw5/helix"
)

// GetGotifyPluginInfo returns gotify plugin info.
func GetGotifyPluginInfo() plugin.Info {
	return plugin.Info{
		ModulePath:  "github.com/yrhki/gotify-twitch",
		Version:     "1.0.0",
		Author:      "yrhki",
		Website:     "https://github.com/yrhki/gotify-twitch",
		Description: "Twitch notifications for gotify",
		License:     "MIT",
		Name:        "Gotify Twitch",
	}
}

// Plugin is the gotify plugin instance.
type Plugin struct {
	userCtx plugin.UserContext
	// User config
	config *Config
	// Twitch API
	api *helix.Client
	// Is plugin enabled
	enabled        bool
	msgHandler     plugin.MessageHandler
	storageHandler plugin.StorageHandler
	// Wait for ticker to finnish
	waitGroup *sync.WaitGroup
	ticker    *time.Ticker
	// Signal to stop ticker
	stop chan struct{}
}

// Enable enables the plugin.
func (c *Plugin) Enable() error {
	err := c.config.Valid()
	if err != nil {
		return err
	}

	// Create API instance
	api, err := helix.NewClient(&helix.Options{
		ClientID:       c.config.ClientID,
		AppAccessToken: c.config.Token,
		HTTPClient: &http.Client{
			Timeout: time.Hour * 1,
		},
	})

	// Test authentication
	resp, err := api.GetStreams(&helix.StreamsParams{})
	if err != nil {
		return err
	}
	if resp.ErrorStatus != 0 {
		return fmt.Errorf("Twitch error: %s", resp.ErrorMessage)
	}

	c.api = api
	c.tickerStart()
	c.enabled = true
	return nil
}

// Disable disables the plugin.
func (c *Plugin) Disable() error {
	c.tickerStop()
	c.enabled = false
	return nil
}

// SetStorageHandler implements plugin.StorageHandler
func (c *Plugin) SetStorageHandler(h plugin.StorageHandler) {
	c.storageHandler = h

	// Init for storage
	var stor storage
	b, _ := c.storageHandler.Load()
	json.Unmarshal(b, &stor)

	if stor.ChannelStatus == nil || stor.Version < storageVersion {
		stor.ChannelStatus = make(map[userID]channelStatus)
		stor.Version = storageVersion
		if newStorage, err := json.Marshal(stor); err == nil {
			c.storageHandler.Save(newStorage)
		}
	}
}

// SetMessageHandler implements plugin.StorageHandler
func (c *Plugin) SetMessageHandler(h plugin.MessageHandler) {
	c.msgHandler = h
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &Plugin{userCtx: ctx}
}

func main() {
	panic("this should be built as go plugin")
}
