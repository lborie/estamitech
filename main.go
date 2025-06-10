package main

import (
	"encoding/xml"
	"html/template"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fogleman/gg"
)

const estamitechRss = "https://feeds.zencastr.com/f/bOMlUWx6.rss"

// Cache structure
type RSSCache struct {
	data      []Item
	timestamp time.Time
	mutex     sync.RWMutex
}

var rssCache = &RSSCache{}

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
	rssCache.mutex.RLock()
	// Vérifier si le cache est valide (moins de 30 secondes)
	if time.Since(rssCache.timestamp) < 30*time.Second && rssCache.data != nil {
		data := rssCache.data
		rssCache.mutex.RUnlock()
		return data, nil
	}
	rssCache.mutex.RUnlock()

	// Cache expiré ou vide, récupérer les données
	rssCache.mutex.Lock()
	defer rssCache.mutex.Unlock()

	// Double vérification après avoir acquis le verrou d'écriture
	if time.Since(rssCache.timestamp) < 30*time.Second && rssCache.data != nil {
		return rssCache.data, nil
	}

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

	// Mettre à jour le cache
	rssCache.data = rss.Channel.Items
	rssCache.timestamp = time.Now()

	return rssCache.data, nil
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
	http.HandleFunc("/og-image/", func(writer http.ResponseWriter, request *http.Request) {
		// Extraire l'ID de l'épisode depuis le path
		path := strings.TrimPrefix(request.URL.Path, "/og-image/")
		if path == "" {
			http.NotFound(writer, request)
			return
		}
		episodeID := path

		// Récupérer les données RSS
		episodes, err := getRSSData()
		if err != nil {
			log.Printf("Error getting RSS data: %v", err)
			http.NotFound(writer, request)
			return
		}

		// Trouver l'épisode
		episode := findEpisodeByID(episodes, episodeID)
		if episode == nil {
			http.NotFound(writer, request)
			return
		}

		// Télécharger l'image originale
		resp, err := http.Get(episode.Image.Href)
		if err != nil {
			log.Printf("Error downloading image: %v", err)
			http.NotFound(writer, request)
			return
		}
		defer resp.Body.Close()

		// Décoder l'image
		srcImg, _, err := image.Decode(resp.Body)
		if err != nil {
			log.Printf("Error decoding image: %v", err)
			http.NotFound(writer, request)
			return
		}

		// Créer un nouveau contexte de dessin 1200x630
		dc := gg.NewContext(1200, 630)

		// Définir la couleur de fond (violet foncé comme votre site)
		dc.SetColor(color.RGBA{93, 22, 146, 255}) // #5d1692
		dc.Clear()

		// Calculer la taille pour l'image carrée centrée (500x500 pour garder de la marge)
		imgSize := 500
		x := (1200 - imgSize) / 2
		y := (630 - imgSize) / 2

		// Calculer le facteur de redimensionnement
		bounds := srcImg.Bounds()
		srcWidth := bounds.Dx()
		srcHeight := bounds.Dy()
		scale := float64(imgSize) / float64(srcWidth)
		if srcHeight > srcWidth {
			scale = float64(imgSize) / float64(srcHeight)
		}

		// Redimensionner et dessiner l'image centrée
		dc.Push()
		dc.Translate(float64(x+imgSize/2), float64(y+imgSize/2))
		dc.Scale(scale, scale)
		dc.DrawImageAnchored(srcImg, 0, 0, 0.5, 0.5)
		dc.Pop()

		// Encoder et envoyer l'image
		writer.Header().Set("Content-Type", "image/png")
		writer.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
		err = dc.EncodePNG(writer)
		if err != nil {
			log.Printf("Error encoding PNG: %v", err)
			http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
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
	http.HandleFunc("/sitemap.xml", func(writer http.ResponseWriter, request *http.Request) {
		episodes, err := getRSSData()
		if err != nil {
			log.Printf("Error getting RSS data for sitemap: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
		writer.WriteHeader(http.StatusOK)

		// En-tête XML et sitemap
		_, _ = writer.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>https://estamitech.fr/</loc>
		<lastmod>` + time.Now().Format("2006-01-02") + `</lastmod>
		<changefreq>daily</changefreq>
		<priority>1.0</priority>
	</url>`))

		// Ajouter chaque épisode
		for _, episode := range episodes {
			pubDate, _ := time.Parse(time.RFC1123, episode.PubDate)
			_, _ = writer.Write([]byte(`
	<url>
		<loc>https://estamitech.fr/episode/` + episode.GUID + `</loc>
		<lastmod>` + pubDate.Format("2006-01-02") + `</lastmod>
		<changefreq>monthly</changefreq>
		<priority>0.8</priority>
	</url>`))
		}

		// Fermeture du sitemap
		_, _ = writer.Write([]byte(`
</urlset>`))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	_ = http.ListenAndServe(":"+port, nil)
}
