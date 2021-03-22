package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gotify/plugin-api"
	"github.com/nicklaw5/helix"
)

type messageDisplay struct {
	ContentType string `json:"contentType"`
}

// getGameByID gets game by id
func (c *Plugin) getGameByID(id string) helix.Game {
	resp, err := c.api.GetGames(&helix.GamesParams{
		IDs: []string{id},
	})

	if err != nil || len(resp.Data.Games) == 0 {
		return helix.Game{
			ID:        "0",
			Name:      "Unknown",
			BoxArtURL: "",
		}
	}

	return resp.Data.Games[0]
}

type messageMetadata struct {
	*channelStatus
	UserName string `json:"user_name"`
	Action   string `json:"action"`
}

func (c *Plugin) fetch() error {
	resp, err := c.api.GetStreams(&helix.StreamsParams{
		UserLogins: c.config.Follow,
	})
	if err != nil {
		return err
	}

	var (
		stor storage
		// Mark live channels
		isLive map[string]bool = make(map[string]bool)
	)

	// Load storage
	storageBytes, err := c.storageHandler.Load()
	err = json.Unmarshal(storageBytes, &stor)
	now := time.Now()

	for _, stream := range resp.Data.Streams {
		var (
			// Is category changed
			categoryChanged bool
			status          channelStatus
			// Does stream already exist in storage
			isAdded bool
			// Username in lower
			username string = strings.ToLower(stream.UserName)
			// Current category
			category helix.Game = c.getGameByID(stream.GameID)
			action   string     = "live"
		)

		// Mark user is live
		isLive[username] = true

		// New follow, now live or diffrent start
		if status, isAdded = stor.StreamStatus[username]; !isAdded || !status.IsLive && stream.StartedAt != status.Start {
			status.UserID = stream.UserID
			status.Start = stream.StartedAt
			status.IsLive = true

		} else if c.config.OnCategoryChange && isAdded && status.Category.ID != category.ID {
			// Notify about category change
			categoryChanged = true
			action = "category"
		} else {
			// Channel is already live
			continue
		}

		// Update stream info in storage
		status.Title = stream.Title
		status.ThumbnailURL = stream.ThumbnailURL
		status.Category = category
		status.End = nil
		stor.StreamStatus[username] = status

		message := plugin.Message{
			Message: fmt.Sprintf(`**%s**  
Category: %s  
Started: %s (%s)  
[WATCH](https://twitch.tv/%s)`,
				stream.Title,
				category.Name,
				timeFormat(stream.StartedAt.Local()), now.Sub(stream.StartedAt).Round(time.Second),
				username),
			Priority: c.config.Priority,
			Extras:   make(map[string]interface{}),
		}

		if categoryChanged {
			message.Title = fmt.Sprintf("Twitch: %s changed category", stream.UserName)
		} else {
			message.Title = fmt.Sprintf("Twitch: %s is live", stream.UserName)
		}
		message.Extras["twitch::metadata"] = status.GetMetadata(username, action)
		message.Extras["client::display"] = messageDisplay{"text/markdown"}

		c.msgHandler.SendMessage(message)

	}

	// Update live status to offline
	for _, username := range c.config.Follow {
		if !isLive[username] {
			// Channel was previously online
			if status := stor.StreamStatus[username]; status.IsLive {
				status.SetOffline()
				stor.StreamStatus[username] = status
				// Notify when stream goes offline
				if c.config.OnOffline {
					message := plugin.Message{
						Priority: c.config.Priority,
						Title:    fmt.Sprintf("Twitch: %s offline", username),
						Extras:   make(map[string]interface{}),
					}
					message.Extras["twitch::metadata"] = status.GetMetadata(username, "offline")
					message.Extras["client::display"] = messageDisplay{"text/markdown"}
					c.msgHandler.SendMessage(message)
				}
			}
		}
	}

	// Update storage
	if newStorage, err := json.Marshal(stor); err == nil {
		c.storageHandler.Save(newStorage)
	}

	return nil
}
