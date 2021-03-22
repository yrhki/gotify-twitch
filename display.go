package main

import (
	"bytes"
	"encoding/json"
	"net/url"
	"text/template"
)

const tempDisplayChannelTxt = `
{{ if .IsLive }}
### {{.Username}} [LIVE](https://twitch.tv/{{.UsernameLower}})  
**Title**: {{.Title}}  
**Category**: {{.Category.Name}}  
**Start**: {{timeFormat .Start.Local}} ({{.Uptime}})  
[![ThumbnailURL]({{thumbnailSize .Thumbnail}} "Click to watch")](https://twitch.tv/{{.UsernameLower}})  
{{ else }}
### {{.Username}} [OFFLINE](https://twitch.tv/{{.UsernameLower}})  
**Last Stream**  
**Title**: {{.Title}}  
**Category**: {{.Category.Name}}  
**Start**: {{timeFormat .Start.Local}}  
**End**: {{timeFormat .Start.Local}} ({{.Uptime}})  
{{ end }}
`

var tempDisplayChannel = template.Must(template.New("display").Funcs(template.FuncMap{
	"timeFormat": timeFormat,
	"thumbnailSize": thumbnailSize,
}).Parse(tempDisplayChannelTxt))

// GetDisplay implements plugin.Displayer
func (c *Plugin) GetDisplay(location *url.URL) string {
	out := bytes.NewBufferString("# Live Status\n")
	bufLive := new(bytes.Buffer)
	bufOffline := new(bytes.Buffer)

	storBytes, err := c.storageHandler.Load()
	if err != nil {
		return err.Error()
	}

	var stor storage
	err = json.Unmarshal(storBytes, &stor)
	if err != nil {
		return err.Error()
	}

	for _, status := range stor.ChannelStatus {
		if status.IsLive {
			err = tempDisplayChannel.Execute(bufLive, status)
		} else {
			err = tempDisplayChannel.Execute(bufOffline, status)
		}
		if err != nil {
			return err.Error()
		}
	}
	return out.String() + bufLive.String() + bufOffline.String()
}

