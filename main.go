package main

import (
	"encoding/xml"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
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

type EpisodePageData struct {
	Episode       Item
	CleanDesc     string
	FormattedDate string
	HTMLDesc      template.HTML
}

type IndexPageData struct {
	Episodes    []EpisodeItem
	CurrentYear int
}

type EpisodeItem struct {
	Item
	CleanDesc     string
	FormattedDate string
}

func getRSSData() ([]Item, error) {
	response, err := http.DefaultClient.Get(estamitechRss)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	resp := string(respBytes)
	resp = strings.ReplaceAll(resp, "itunes:", "itunes_")

	var rss Rss
	err = xml.Unmarshal([]byte(resp), &rss)
	if err != nil {
		return nil, err
	}

	return rss.Channel.Items, nil
}

func cleanDescription(desc string) string {
	// Supprimer les balises HTML
	re := regexp.MustCompile(`<[^>]*>`)
	clean := re.ReplaceAllString(desc, "")

	// Tronquer à 160 caractères pour les meta tags
	if len(clean) > 160 {
		clean = clean[:157] + "..."
	}

	return strings.TrimSpace(clean)
}

func findEpisodeByID(episodes []Item, episodeID string) *Item {
	for _, episode := range episodes {
		if episode.GUID == episodeID || episode.Link == episodeID {
			return &episode
		}
	}
	return nil
}

func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// Récupérer les données RSS
		episodes, err := getRSSData()
		if err != nil {
			log.Printf("Error getting RSS data: %v", err)
			http.ServeFile(writer, request, "web/index.html")
			return
		}

		// Préparer les données pour le template
		var episodeItems []EpisodeItem
		for _, episode := range episodes {
			pubDate, _ := time.Parse(time.RFC1123, episode.PubDate)
			episodeItems = append(episodeItems, EpisodeItem{
				Item:          episode,
				CleanDesc:     cleanDescription(episode.Description),
				FormattedDate: pubDate.Format("2 January 2006"),
			})
		}

		pageData := IndexPageData{
			Episodes:    episodeItems,
			CurrentYear: time.Now().Year(),
		}

		// Charger et exécuter le template
		tmpl, err := template.ParseFiles("web/index.gohtml")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.ServeFile(writer, request, "web/index.html")
			return
		}

		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(writer, pageData)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.ServeFile(writer, request, "web/index.html")
			return
		}
	})
	http.HandleFunc("/episode/", func(writer http.ResponseWriter, request *http.Request) {
		// Extraire l'ID de l'épisode depuis le path
		path := strings.TrimPrefix(request.URL.Path, "/episode/")
		if path == "" {
			http.Redirect(writer, request, "/", http.StatusSeeOther)
			return
		}
		episodeID := path

		// Récupérer les données RSS
		episodes, err := getRSSData()
		if err != nil {
			log.Printf("Error getting RSS data: %v", err)
			http.Redirect(writer, request, "/", http.StatusSeeOther)
			return
		}

		// Trouver l'épisode
		episode := findEpisodeByID(episodes, episodeID)
		if episode == nil {
			http.Redirect(writer, request, "/", http.StatusSeeOther)
			return
		}

		// Préparer les données pour le template
		pubDate, _ := time.Parse(time.RFC1123, episode.PubDate)
		log.Printf("Episode found: %s, published on %s", episode.Title, episode.PubDate)
		pageData := EpisodePageData{
			Episode:       *episode,
			CleanDesc:     cleanDescription(episode.Description),
			FormattedDate: pubDate.Format("02-01-2006"),
			HTMLDesc:      template.HTML(episode.Description),
		}

		// Charger et exécuter le template
		tmpl, err := template.ParseFiles("web/episode.gohtml")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Redirect(writer, request, "/", http.StatusSeeOther)
			return
		}

		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(writer, pageData)
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Redirect(writer, request, "/", http.StatusSeeOther)
			return
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	_ = http.ListenAndServe(":"+port, nil)
}
