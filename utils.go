package main

import (
	"strings"
	"time"
)

func thumbnailSize(url string) string {
	url = strings.Replace(url, "{width}", "400", 1)
	url = strings.Replace(url, "{height}", "225", 1)
	return url
}

func timeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
