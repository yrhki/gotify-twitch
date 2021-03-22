package main

import (
	"time"

	"github.com/nicklaw5/helix"
)

type channelStatus struct {
	UserID       string     `json:"user_id"`
	Category     helix.Game `json:"category"`
	IsLive       bool       `json:"live"`
	Start        time.Time  `json:"start"`
	Title        string     `json:"title"`
	ThumbnailURL string     `json:"thumbnail_url"`
	End          *time.Time `json:"end"`
}

// Set Stream offline
func (stat *channelStatus) SetOffline() {
	stat.IsLive = false
	end := time.Now().UTC().Round(time.Second)
	stat.End = &end
}

// Convert StreamStatus to messageMetadata
func (stat *channelStatus) GetMetadata(username, action string) messageMetadata {
	out := messageMetadata{
		channelStatus: stat,
		UserName:      username,
		Action:        action,
	}

	return out
}

// Storage for plugin
type storage struct {
	StreamStatus map[string]channelStatus `json:"streams_status"`
}
