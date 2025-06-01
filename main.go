package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const estamitechRss = "https://feeds.zencastr.com/f/bOMlUWx6.rss"

type Rss struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	Image       struct {
		Href string `xml:"href,attr"`
	} `xml:"itunes_image"`
	Enclosure struct {
		URL  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
	GUID string `xml:"guid"`
}

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "web/index.html")
	})
	http.HandleFunc("/devlille", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "web/devlille.html")
	})
	http.HandleFunc("/episode", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "web/episode.html")
	})
	http.HandleFunc("/rss", func(writer http.ResponseWriter, request *http.Request) {
		response, err := http.DefaultClient.Get(estamitechRss)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(response.Body)

		respBytes, err := io.ReadAll(response.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp := string(respBytes)
		resp = strings.ReplaceAll(resp, "itunes:", "itunes_")
		var rss Rss
		err = xml.Unmarshal([]byte(resp), &rss)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Marshal rss to json
		itemsJson, err := json.Marshal(rss.Channel.Items)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		// Access cors policy
		//writer.Header().Set("Access-Control-Allow-Origin", "*")
		//writer.Header().Set("Access-Control-Allow-Methods", "GET,OPTIONS")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(itemsJson)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	_ = http.ListenAndServe(":"+port, nil)
}
