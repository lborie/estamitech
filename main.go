package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"net/url"
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
	GUID        string `xml:"guid"`
	Author      string `xml:"itunes_author"`
	Duration    string `xml:"itunes_duration"`
	Episode     string `xml:"itunes_episode"`
	Season      string `xml:"itunes_season"`
	Summary     string `xml:"itunes_summary"`
	EpisodeType string `xml:"itunes_episodeType"`
	Explicit    string `xml:"itunes_explicit"`
	Keywords    string `xml:"itunes_keywords"`
	Category    string `xml:"itunes_category"`
	Owner       struct {
		Name  string `xml:"itunes_name"`
		Email string `xml:"itunes_email"`
	} `xml:"itunes_owner"`
}

type EpisodePageData struct {
	Episode       Item
	CleanDesc     string
	FormattedDate string
	PubDateISO    string // date de publication en ISO-8601 (pour JSON-LD / article:published_time)
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
		return rssCache.staleOrError(err)
	}
	defer response.Body.Close()

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return rssCache.staleOrError(err)
	}

	resp := string(respBytes)
	resp = strings.ReplaceAll(resp, "itunes:", "itunes_")

	var rss Rss
	err = xml.Unmarshal([]byte(resp), &rss)
	if err != nil {
		return rssCache.staleOrError(err)
	}

	// Mettre à jour le cache
	rssCache.data = rss.Channel.Items
	rssCache.timestamp = time.Now()

	return rssCache.data, nil
}

// staleOrError renvoie le dernier cache RSS connu (même expiré) quand un
// rafraîchissement échoue, pour garder le site en ligne pendant une panne
// temporaire du flux. À n'appeler qu'avec le verrou d'écriture détenu.
func (c *RSSCache) staleOrError(err error) ([]Item, error) {
	if c.data != nil {
		log.Printf("Échec du rafraîchissement RSS (%v) — service du cache expiré (%d épisodes)", err, len(c.data))
		return c.data, nil
	}
	return nil, err
}

func cleanDescription(desc string) string {
	// Supprimer les balises HTML
	re := regexp.MustCompile(`<[^>]*>`)
	clean := re.ReplaceAllString(desc, "")

	// Tronquer à 160 caractères — en runes (pas en octets) pour ne pas
	// couper un caractère UTF-8 accentué en plein milieu (é, è, à…)
	runes := []rune(clean)
	if len(runes) > 160 {
		clean = string(runes[:157]) + "..."
	}

	return strings.TrimSpace(clean)
}

// writeErrorPage rend une page d'erreur de repli aux couleurs du site, avec le
// bon code HTTP — au lieu de servir un fichier web/index.html qui n'existe pas.
func writeErrorPage(writer http.ResponseWriter, status int, eyebrow, heading, message string) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)
	_, _ = fmt.Fprintf(writer, `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<meta name="robots" content="noindex">
<title>%s — L'EstamiTech</title>
<link rel="icon" type="image/x-icon" href="/static/favicon.ico">
<link rel="stylesheet" href="/style.css">
</head>
<body>
<main class="ep-page" style="text-align:center">
<span class="label">%s</span>
<h1>%s</h1>
<p style="color:var(--text-dim);margin:16px 0 32px">%s</p>
<a href="/" class="ep-link">Retour à l'accueil <span>&rarr;</span></a>
</main>
</body>
</html>`, heading, eyebrow, heading, message)
}

// writeNotFound renvoie un vrai statut 404 (au lieu d'une redirection 303 vers
// l'accueil qui créait un « soft-404 » néfaste pour le référencement).
func writeNotFound(writer http.ResponseWriter) {
	writeErrorPage(writer, http.StatusNotFound, "Erreur 404", "Épisode introuvable",
		"Cet épisode n'existe pas ou n'est plus disponible.")
}

var moisFR = [...]string{"janvier", "février", "mars", "avril", "mai", "juin",
	"juillet", "août", "septembre", "octobre", "novembre", "décembre"}

// formatDateFR formate une date en français long, ex. « 26 avril 2026 »
// (time.Format ne gère pas de locale et produisait des mois en anglais).
func formatDateFR(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), moisFR[t.Month()-1], t.Year())
}

// renderOGImage compose l'image source (carrée) centrée sur le fond 1200x630 de
// la marque, puis l'encode en PNG dans la réponse. Base commune aux images
// Open Graph de la page d'accueil et des épisodes.
func renderOGImage(writer http.ResponseWriter, srcImg image.Image) {
	dc := gg.NewContext(1200, 630)

	bgImg, err := gg.LoadImage("static/estamitech-bckg.png")
	if err != nil {
		log.Printf("Error loading background image: %v", err)
		dc.SetColor(color.RGBA{93, 22, 146, 255}) // #5d1692
		dc.Clear()
	} else {
		dc.DrawImageAnchored(bgImg, 600, 315, 0.5, 0.5)
	}

	imgSize := 500
	x := (1200 - imgSize) / 2
	y := (630 - imgSize) / 2

	bounds := srcImg.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()
	scale := float64(imgSize) / float64(srcWidth)
	if srcHeight > srcWidth {
		scale = float64(imgSize) / float64(srcHeight)
	}

	dc.Push()
	dc.Translate(float64(x+imgSize/2), float64(y+imgSize/2))
	dc.Scale(scale, scale)
	dc.DrawImageAnchored(srcImg, 0, 0, 0.5, 0.5)
	dc.Pop()

	writer.Header().Set("Content-Type", "image/png")
	writer.Header().Set("Cache-Control", "public, max-age=86400")
	if err := dc.EncodePNG(writer); err != nil {
		log.Printf("Error encoding PNG: %v", err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
	}
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
			writeErrorPage(writer, http.StatusServiceUnavailable, "Erreur 503",
				"Service momentanément indisponible",
				"Le flux du podcast est temporairement injoignable, merci de réessayer dans un instant.")
			return
		}

		// Préparer les données pour le template
		var episodeItems []EpisodeItem
		for _, episode := range episodes {
			pubDate, _ := time.Parse(time.RFC1123, episode.PubDate)
			episodeItems = append(episodeItems, EpisodeItem{
				Item:          episode,
				CleanDesc:     cleanDescription(episode.Description),
				FormattedDate: formatDateFR(pubDate),
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
			writeErrorPage(writer, http.StatusInternalServerError, "Erreur",
				"Une erreur est survenue", "Impossible d'afficher la page pour le moment.")
			return
		}

		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(writer, pageData)
		if err != nil {
			// La réponse a déjà commencé à être écrite : on ne peut plus
			// remplacer le statut, on se contente de journaliser.
			log.Printf("Error executing template: %v", err)
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

		renderOGImage(writer, srcImg)
	})
	// Carte Open Graph 1200x630 de la page d'accueil : le logo du podcast composé
	// sur le fond de la marque (URL absolue, dimensions déclarées côté template).
	http.HandleFunc("/og-image", func(writer http.ResponseWriter, request *http.Request) {
		logo, err := gg.LoadImage("static/LogoEstamitech.jpg")
		if err != nil {
			log.Printf("Error loading homepage logo: %v", err)
			http.NotFound(writer, request)
			return
		}
		renderOGImage(writer, logo)
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
			writeNotFound(writer)
			return
		}

		// Préparer les données pour le template
		pubDate, _ := time.Parse(time.RFC1123, episode.PubDate)
		log.Printf("Episode found: %s, published on %s", episode.Title, episode.PubDate)
		pageData := EpisodePageData{
			Episode:       *episode,
			CleanDesc:     cleanDescription(episode.Description),
			FormattedDate: pubDate.Format("02-01-2006"),
			PubDateISO:    pubDate.Format(time.RFC3339),
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

		// lastmod de l'accueil = date du dernier épisode (plus fidèle que « aujourd'hui »)
		homeLastmod := time.Now().Format("2006-01-02")
		if len(episodes) > 0 {
			if d, err := time.Parse(time.RFC1123, episodes[0].PubDate); err == nil {
				homeLastmod = d.Format("2006-01-02")
			}
		}

		// En-tête XML et sitemap
		_, _ = writer.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>https://estamitech.fr/</loc>
		<lastmod>` + homeLastmod + `</lastmod>
		<changefreq>daily</changefreq>
		<priority>1.0</priority>
	</url>`))

		// Ajouter chaque épisode (GUID échappé pour l'URL, par sécurité)
		for _, episode := range episodes {
			pubDate, _ := time.Parse(time.RFC1123, episode.PubDate)
			_, _ = writer.Write([]byte(`
	<url>
		<loc>https://estamitech.fr/episode/` + url.PathEscape(episode.GUID) + `</loc>
		<lastmod>` + pubDate.Format("2006-01-02") + `</lastmod>
		<changefreq>monthly</changefreq>
		<priority>0.8</priority>
	</url>`))
		}

		// Fermeture du sitemap
		_, _ = writer.Write([]byte(`
</urlset>`))
	})

	http.HandleFunc("/robots.txt", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain")
		writer.Write([]byte(`User-agent: *
Allow: /

Sitemap: https://estamitech.fr/sitemap.xml`))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	_ = http.ListenAndServe(":"+port, nil)
}
