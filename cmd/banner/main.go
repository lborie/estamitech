// Commande banner : génère les bannières de chaîne de L'EstamiTech à partir des
// mêmes ressources de marque que les miniatures Open Graph du site — le fond
// étoilé, l'icône du chaudron et la police Outfit.
//
//	go run ./cmd/banner                      # les deux formats
//	go run ./cmd/banner -format youtube      # un seul
//	go run ./cmd/banner -guides -o /tmp/     # avec la zone sûre matérialisée
//
// Chaque format porte ses propres constantes de mise en page (en pixels
// absolus) : elles sont calées sur ses dimensions et sur sa zone sûre, et ne
// se transposent pas d'un format à l'autre.
package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/fogleman/gg"
)

const (
	bgPath      = "static/estamitech-bckg.png"
	iconPath    = "static/estamitech-icon.png"
	fontBlack   = "static/fonts/Outfit-ExtraBold.ttf"
	fontRegular = "static/fonts/Outfit-Regular.ttf"

	wordLight = "L’ESTAMI"
	wordBold  = "TECH"
	tagline   = "La Tech du Nord"
)

// Couleurs de marque échantillonnées sur le logo officiel.
//
// Toute couleur non opaque est déclarée en color.NRGBA, jamais en color.RGBA :
// gg traite les couleurs comme pré-multipliées par l'alpha et les additionne
// dans un uint32 (patternPainter.Paint). Un color.RGBA{0x0d, 0x04, 0x1f, 0x00}
// n'est pas pré-multiplié valide — son RGBA() rend du chromatique sous un alpha
// nul, ce qui fait déborder l'accumulateur là où le fond est clair et retourne
// des pixels aberrants. NRGBA.RGBA() pré-multiplie correctement, ce qui garantit
// cr <= ca et donc une somme bornée par 65535², sous la limite de l'uint32.
var (
	magenta  = color.NRGBA{0xc3, 0x1d, 0xbd, 0xff}
	deepPurp = color.NRGBA{0x5d, 0x16, 0x92, 0xff}
	scrimCol = color.NRGBA{0x0d, 0x04, 0x1f, 0xff}
)

// spec décrit un format de bannière : ses dimensions, sa zone sûre et la mise
// en page du lockup qui y tient.
type spec struct {
	name string
	file string
	w, h int

	// safeW/safeH : région centrale garantie visible sur tous les appareils.
	// Zéro = pas de contrainte connue (toute l'image est visible).
	safeW, safeH float64

	iconSize     float64 // côté de l'icône du chaudron
	gap          float64 // espace entre l'icône et le bloc texte
	wordmarkSize float64
	taglineSize  float64
	tracking     float64 // interlettrage du wordmark, calqué sur le logo
	scrimSpan    float64 // rayon du voile central, en fraction de la largeur
}

var specs = map[string]spec{
	// Bannière de profil Twitch. Twitch rogne les bords selon la largeur de
	// fenêtre : pas de zone sûre publiée, on garde le lockup au centre.
	"twitch": {
		name: "Twitch", file: "static/twitch-banner.png",
		w: 1200, h: 480,
		iconSize: 272, gap: 50, wordmarkSize: 72, taglineSize: 32,
		tracking: 5, scrimSpan: 0.525,
	},
	// Bannière de chaîne YouTube. 2048x1152 est le minimum recommandé par
	// YouTube ; la zone sûre pour le texte et les logos y fait 1235x338,
	// centrée — tout ce qui déborde est rogné sur mobile.
	"youtube": {
		name: "YouTube", file: "static/youtube-banner.png",
		w: 2048, h: 1152,
		safeW: 1235, safeH: 338,
		iconSize: 262, gap: 56, wordmarkSize: 94, taglineSize: 40,
		tracking: 6.5, scrimSpan: 0.40,
	},
}

func main() {
	format := flag.String("format", "", "twitch|youtube (vide = les deux)")
	outDir := flag.String("o", "", "répertoire de sortie (vide = static/)")
	guides := flag.Bool("guides", false, "matérialiser la zone sûre (vérification, pas pour publication)")
	flag.Parse()

	var todo []spec
	switch *format {
	case "":
		todo = []spec{specs["twitch"], specs["youtube"]}
	default:
		s, ok := specs[*format]
		if !ok {
			log.Fatalf("format inconnu %q (twitch|youtube)", *format)
		}
		todo = []spec{s}
	}

	for _, s := range todo {
		path := s.file
		if *outDir != "" {
			path = filepath.Join(*outDir, filepath.Base(s.file))
		}
		if err := render(s, path, *guides); err != nil {
			log.Fatalf("%s : %v", s.name, err)
		}
	}
}

func render(s spec, path string, guides bool) error {
	dc := gg.NewContext(s.w, s.h)
	cx, cy := float64(s.w)/2, float64(s.h)/2

	drawBackground(dc, s, cx, cy)
	drawScrim(dc, s, cx, cy)
	if err := drawLockup(dc, s, cx, cy); err != nil {
		return err
	}
	if guides {
		drawGuides(dc, s, cx, cy)
	}

	if err := dc.SavePNG(path); err != nil {
		return fmt.Errorf("écriture de %s : %w", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	log.Printf("%-8s %-30s %dx%d  %.2f Mo", s.name, path, s.w, s.h, float64(info.Size())/(1<<20))
	return nil
}

// drawBackground reprend le fond étoilé 1200x630 du site en le redimensionnant
// « cover » : mise à l'échelle au plus juste pour remplir le cadre, le surplus
// étant rogné. On ne déforme donc jamais le dégradé ni les étincelles.
func drawBackground(dc *gg.Context, s spec, cx, cy float64) {
	bg, err := gg.LoadImage(bgPath)
	if err != nil {
		log.Printf("fond introuvable (%v) — repli sur l'aplat de marque", err)
		dc.SetColor(deepPurp)
		dc.Clear()
		return
	}
	b := bg.Bounds()
	scale := max(float64(s.w)/float64(b.Dx()), float64(s.h)/float64(b.Dy()))

	dc.Push()
	dc.Translate(cx, cy)
	dc.Scale(scale, scale)
	dc.DrawImageAnchored(bg, 0, 0, 0.5, 0.5)
	dc.Pop()
}

// drawScrim assombrit doucement le centre : le fond porte une étincelle magenta
// vive juste derrière le bloc de texte, qui mangerait la lisibilité du wordmark.
func drawScrim(dc *gg.Context, s spec, cx, cy float64) {
	r := float64(s.w) * s.scrimSpan
	grad := gg.NewRadialGradient(cx, cy, 0, cx, cy, r)
	grad.AddColorStop(0, withAlpha(scrimCol, 0x8c))
	grad.AddColorStop(0.55, withAlpha(scrimCol, 0x5e))
	grad.AddColorStop(1, withAlpha(scrimCol, 0x00))
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, float64(s.w), float64(s.h))
	dc.Fill()
}

// drawLockup compose le bloc icône + wordmark + baseline, centré pour tenir
// dans la zone sûre du format.
func drawLockup(dc *gg.Context, s spec, cx, cy float64) error {
	icon, err := gg.LoadImage(iconPath)
	if err != nil {
		return fmt.Errorf("icône introuvable : %w", err)
	}

	// Mesurer le wordmark avant de placer quoi que ce soit : la largeur du bloc
	// détermine l'origine gauche de l'ensemble.
	if err := dc.LoadFontFace(fontBlack, s.wordmarkSize); err != nil {
		return fmt.Errorf("police %s : %w", fontBlack, err)
	}
	wLight := trackedWidth(dc, wordLight, s.tracking)
	wordmarkW := wLight + trackedWidth(dc, wordBold, s.tracking)

	totalW := s.iconSize + s.gap + wordmarkW
	if s.safeW > 0 {
		if totalW > s.safeW {
			return fmt.Errorf("lockup de %.0f px de large, hors zone sûre de %.0f px", totalW, s.safeW)
		}
		if s.iconSize > s.safeH {
			return fmt.Errorf("icône de %.0f px de haut, hors zone sûre de %.0f px", s.iconSize, s.safeH)
		}
		log.Printf("         zone sûre %.0fx%.0f — lockup %.0fx%.0f (marges %.0f / %.0f px)",
			s.safeW, s.safeH, totalW, s.iconSize,
			(s.safeW-totalW)/2, (s.safeH-s.iconSize)/2)
	}
	left := cx - totalW/2

	// Icône du chaudron, centrée verticalement.
	iconScale := s.iconSize / float64(icon.Bounds().Dx())
	dc.Push()
	dc.Translate(left+s.iconSize/2, cy)
	dc.Scale(iconScale, iconScale)
	dc.DrawImageAnchored(icon, 0, 0, 0.5, 0.5)
	dc.Pop()

	textX := left + s.iconSize + s.gap

	// Wordmark bicolore : « L’ESTAMI » en blanc, « TECH » en magenta.
	baseline := cy + s.wordmarkSize*0.20
	dc.SetColor(color.White)
	drawTracked(dc, wordLight, textX, baseline, s.tracking)
	dc.SetColor(magenta)
	drawTracked(dc, wordBold, textX+wLight, baseline, s.tracking)

	// Baseline « La Tech du Nord », centrée sous le wordmark comme sur le logo.
	if err := dc.LoadFontFace(fontRegular, s.taglineSize); err != nil {
		return fmt.Errorf("police %s : %w", fontRegular, err)
	}
	dc.SetColor(color.NRGBA{0xff, 0xff, 0xff, 0xe8})
	dc.DrawStringAnchored(tagline, textX+wordmarkW/2, baseline+s.taglineSize*1.55, 0.5, 0)
	return nil
}

// drawGuides matérialise la zone sûre pour contrôler visuellement le cadrage.
// Jamais utilisé pour un rendu destiné à la publication.
func drawGuides(dc *gg.Context, s spec, cx, cy float64) {
	if s.safeW == 0 {
		return
	}
	dc.SetColor(color.NRGBA{0xff, 0x2d, 0x55, 0xcc})
	dc.SetLineWidth(3)
	dc.DrawRectangle(cx-s.safeW/2, cy-s.safeH/2, s.safeW, s.safeH)
	dc.Stroke()
}

// trackedWidth mesure une chaîne rendue glyphe par glyphe avec interlettrage.
func trackedWidth(dc *gg.Context, s string, tracking float64) float64 {
	w := 0.0
	for _, r := range s {
		gw, _ := dc.MeasureString(string(r))
		w += gw + tracking
	}
	return w
}

// drawTracked dessine une chaîne glyphe par glyphe : gg n'expose pas de réglage
// d'interlettrage, or le wordmark de la marque est nettement espacé.
func drawTracked(dc *gg.Context, s string, x, y, tracking float64) {
	for _, r := range s {
		dc.DrawString(string(r), x, y)
		gw, _ := dc.MeasureString(string(r))
		x += gw + tracking
	}
}

func withAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}
