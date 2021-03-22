package main

import (
	"encoding/json"
	"fmt"
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
	Action string `json:"action"`
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
		isLive map[userID]bool = make(map[userID]bool)
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
			// Current category
			category helix.Game = c.getGameByID(stream.GameID)
			action   string     = "live"
		)

		// Mark user is live
		isLive[userID(stream.UserID)] = true

		// New follow, now live or diffrent start
		if status, isAdded = stor.ChannelStatus[userID(stream.UserID)]; !isAdded || !status.IsLive && stream.StartedAt != status.Start {
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

		// Update channelStatus in storage
		status.StreamID = stream.ID
		status.Username = stream.UserName

		status.Title = stream.Title
		status.Thumbnail = stream.ThumbnailURL
		status.Category = category
		status.End = nil
		stor.ChannelStatus[userID(stream.UserID)] = status

		message := plugin.Message{
			Message: fmt.Sprintf(`**%s**  
Category: %s  
Started: %s (%s)  
[WATCH](https://twitch.tv/%s)`,
				stream.Title,
				category.Name,
				timeFormat(stream.StartedAt.Local()), now.Sub(stream.StartedAt).Round(time.Second),
				status.UsernameLower()),
			Priority: c.config.Priority,
			Extras:   make(map[string]interface{}),
		}

		if categoryChanged {
			message.Title = fmt.Sprintf("Twitch: %s changed category", stream.UserName)
		} else {
			message.Title = fmt.Sprintf("Twitch: %s is live", stream.UserName)
		}
		message.Extras["twitch::metadata"] = status.getMetadata(action)
		message.Extras["client::display"] = messageDisplay{"text/markdown"}

		c.msgHandler.SendMessage(message)

	}

	// Update live status to offline
	for userID, status := range stor.ChannelStatus {
		if !isLive[userID] {
			// Channel was previously online
			if status.IsLive {
				status.setOffline()
				stor.ChannelStatus[userID] = status
				// Notify when stream goes offline
				if c.config.OnOffline {
					message := plugin.Message{
						Priority: c.config.Priority,
						Title:    fmt.Sprintf("Twitch: %s offline", status.Username),
						Extras:   make(map[string]interface{}),
					}
					message.Extras["twitch::metadata"] = status.getMetadata("offline")
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
