package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const clientID = "5y0akk7vcdxib7yj92vt2vmsh48jrt"
const port = 5483
const html = `
<html>
    <head>
        <title>Token</title>
    </head>
    <body>
        <div id="response"></div>
        <script charset="utf-8">
            document.querySelector("#response").innerText = window.location.hash.split('#')[1].split('&').filter((x) => x.split('=')[0] == "access_token")[0].split('=')[1];
        </script>
    </body>
</html>
`

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	})
	u, err := url.Parse("https://id.twitch.tv/oauth2/authorize")
	if err != nil { log.Fatalln(err) }

	query := u.Query()
	query.Add("client_id", clientID)
	query.Add("redirect_uri", fmt.Sprintf("http://localhost:%d/", port))
	query.Add("response_type", "token")
	u.RawQuery = query.Encode()

	log.Printf("Open: %s\n", u.String())
	log.Fatalln(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
