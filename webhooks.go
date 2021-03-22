package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterWebhook implements plugin.Webhooker.
func (c *Plugin) RegisterWebhook(basePath string, g *gin.RouterGroup) {
	// Dump storage
	g.GET("/storage", func(con *gin.Context) {
		con.Header("content-type", "application/json")
		storageBytes, err := c.storageHandler.Load()
		if err != nil {
			con.Writer.Write([]byte(err.Error()))
			con.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		con.Writer.Write(storageBytes)
	})

	// Force manual fetch
	g.GET("/fetch", func(con *gin.Context) {
		err := c.fetch()
		if err != nil {
			c.tickerStop()
			con.Writer.Write([]byte(err.Error()))
			con.Writer.WriteHeader(http.StatusInternalServerError)
		}
	})
}
