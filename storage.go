package main

import (
	"strings"
	"time"

	"github.com/nicklaw5/helix"
)

type userID string

type channelStatus struct {
	Category  helix.Game `json:"category"`
	IsLive    bool       `json:"live"`
	Start     time.Time  `json:"start"`
	Title     string     `json:"title"`
	Thumbnail string     `json:"thumbnail"`
	End       *time.Time `json:"end"`
	StreamID  string     `json:"stream_id"`
	Username  string     `json:"username"`
}

func (status *channelStatus) username() string {
	return strings.ToLower(status.Username)
}

// Set Stream offline
func (status *channelStatus) setOffline() {
	status.IsLive = false
	end := time.Now().UTC().Round(time.Second)
	status.End = &end
}

// Convert StreamStatus to messageMetadata
func (status *channelStatus) getMetadata(action string) messageMetadata {
	out := messageMetadata{
		channelStatus: status,
		Action:        action,
	}

	return out
}

const storageVersion = 3

// Storage for plugin
type storage struct {
	ChannelStatus map[userID]channelStatus `json:"channel_status"`
	Version       uint                     `json:"version"`
}
