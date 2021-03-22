package main

import (
	"sync"
	"time"

	"github.com/gotify/plugin-api"
)

func (c *Plugin) tickerStart() {
	c.waitGroup = &sync.WaitGroup{}
	c.stop = make(chan struct{})
	c.ticker = time.NewTicker(time.Duration(c.config.Interval) * time.Minute)

	c.waitGroup.Add(1)
	go func() {
		defer c.waitGroup.Done()
		for {
			select {
			case <-c.stop:
				return
			case <-c.ticker.C:
				err := c.fetch()
				if err != nil {
					c.msgHandler.SendMessage(plugin.Message{
						Title:   "Gotify Twitch Error",
						Message: err.Error(),
					})
					return
				}
			}
		}
	}()
}

func (c *Plugin) tickerStop() {
	if c.enabled {
		c.ticker.Stop()
		close(c.stop)
		c.waitGroup.Wait()
	}
}
