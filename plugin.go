package main

import (
	"encoding/json"
	"fmt"
	"net/url"
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
	b, err := c.storageHandler.Load()
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &stor)
	if err != nil {
		panic(err)
	}

	if stor.StreamStatus == nil {
		stor.StreamStatus = make(map[string]channelStatus)
		if newStorage, err := json.Marshal(stor); err == nil {
			c.storageHandler.Save(newStorage)
		}
	}
}

// SetMessageHandler implements plugin.StorageHandler
func (c *Plugin) SetMessageHandler(h plugin.MessageHandler) {
	c.msgHandler = h
}

// GetDisplay implements plugin.Displayer
func (c *Plugin) GetDisplay(location *url.URL) string {
	var (
		stor          storage
		output        = "# Live Status\n"
		outputLive    string
		outputOffline string
	)

	storageBytes, err := c.storageHandler.Load()
	if err != nil {
		return err.Error()
	}
	err = json.Unmarshal(storageBytes, &stor)
	if err != nil {
		return err.Error()
	}

	now := time.Now()
	for _, username := range c.config.Follow {
		if status, exists := stor.StreamStatus[username]; exists {
			if status.IsLive {
				outputLive += fmt.Sprintf(`### %s [LIVE](https://twitch.tv/%s)  
**Title**: %s  
**Category**: %s  
**Start**: %s (%s)  
[![ThumbnailURL](%s "Click to watch")](https://twitch.tv/%s)  
`,
					username,
					username,
					status.Title,
					status.Category.Name,
					timeFormat(status.Start.Local()), now.Sub(status.Start).Round(time.Second),
					thumbnailSize(status.ThumbnailURL), username)
			} else {
				outputOffline += fmt.Sprintf(`### %s [OFFLINE](https://twitch.tv/%s)  
**Last Stream**  
**Title**: %s  
**Category**: %s  
**Start**: %s  
**End**: %s (%s)  
`,
					username,
					username,
					status.Title,
					status.Category.Name,
					timeFormat(status.Start.Local()), timeFormat(status.End.Local()), status.End.Sub(status.Start).Round(time.Second))
			}
		}
	}
	return output + outputLive + outputOffline
}

// NewGotifyPluginInstance creates a plugin instance for a user context.
func NewGotifyPluginInstance(ctx plugin.UserContext) plugin.Plugin {
	return &Plugin{userCtx: ctx}
}

func main() {
	panic("this should be built as go plugin")
}
