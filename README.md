# L'Estamitech

[![Website](https://img.shields.io/badge/Website-estamitech.fr-5d1692)](https://estamitech.fr)
[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://golang.org/)
[![Google App Engine](https://img.shields.io/badge/Deployed%20on-Google%20App%20Engine-4285F4?logo=google-cloud)](https://cloud.google.com/appengine)

> **La Tech du Nord** - Un podcast par Ludovic Borie qui met en avant les projets et initiatives tech de la région Nord de la France.

## 📖 À propos

L'Estamitech est un podcast dédié à la tech du Nord de la France. Chaque épisode met en avant un projet ou une initiative, permettant de discuter tech dans la bonne humeur et sans langue de bois.

Le site web [estamitech.fr](https://estamitech.fr) sert de vitrine pour le podcast, proposant :
- Une présentation du podcast et de son créateur
- La liste des épisodes avec pages individuelles optimisées SEO
- Une carte interactive des recommandations des invités
- Des liens vers toutes les plateformes d'écoute
- Un partage social optimisé avec miniatures d'épisodes

## 🚀 Technologies utilisées

- **Backend** : Go 1.23 avec templates HTML/CSS
- **Server-Side Rendering** : Pages générées côté serveur avec cache RSS (30s)
- **Frontend** : HTML5, CSS3, JavaScript vanilla minimal
- **Hébergement** : Google App Engine
- **RSS** : Intégration avec Zencastr pour les épisodes
- **Cartes** : Framacarte pour les recommandations

## 📁 Structure du projet

```
estamitech/
├── main.go              # Serveur HTTP principal avec cache RSS
├── go.mod               # Dépendances Go
├── app.yaml             # Configuration Google App Engine
├── Makefile             # Scripts de développement et déploiement
├── static/              # Assets statiques
│   ├── LogoEstamitech.jpg
│   ├── favicon.ico
│   └── spotify-podcast-badge-blk-grn-165x40.png
└── web/                 # Templates et assets web
    ├── index.gohtml     # Template page d'accueil
    ├── episode.gohtml   # Template page épisode
    ├── script.js        # JavaScript minimal
    └── style.css        # Styles CSS optimisés
```

## 🛠️ Installation et développement

### Prérequis

- Go 1.23 ou supérieur
- Google Cloud SDK (pour le déploiement)
- Un compte Google Cloud Platform

### Développement local

1. **Cloner le projet**
```bash
git clone https://github.com/lborie/estamitech.git
cd estamitech
```

2. **Lancer le serveur de développement**
```bash
# Méthode 1 : avec Go directement
go run main.go

# Méthode 2 : avec l'émulateur App Engine
make dev-run
```

Le site sera accessible sur `http://localhost:8080`

### Déploiement

Le déploiement se fait sur Google App Engine :

```bash
# Configurer les variables d'environnement
export ESTAMITECH_ACCOUNT_ID="votre-compte@gmail.com"
export ESTAMITECH_PROJECT_ID="votre-project-id"

# Déployer
make deploy
```

## 🔧 Endpoints et fonctionnalités

Le serveur expose plusieurs endpoints :

- `GET /` - Page d'accueil générée côté serveur avec liste des épisodes
- `GET /episode/{id}` - Page individuelle d'épisode avec métadonnées Open Graph
- `GET /sitemap.xml` - Sitemap XML automatique pour le SEO

### Fonctionnalités principales

- **Server-Side Rendering** : Pages HTML générées côté serveur avec données RSS
- **Cache intelligent** : Cache RSS de 30 secondes avec gestion de la concurrence
- **URLs propres** : `/episode/{id}` au lieu de query parameters
- **SEO optimisé** : Métadonnées Open Graph, Twitter Cards et sitemap XML automatique
- **Partage social** : Miniatures et descriptions spécifiques par épisode
- **Indexation automatique** : Sitemap XML généré dynamiquement avec tous les épisodes

### Sitemap SEO

L'endpoint `/sitemap.xml` génère automatiquement un sitemap XML conforme aux standards :

- **Page d'accueil** : Priorité 1.0, mise à jour quotidienne
- **Pages d'épisodes** : Priorité 0.8, mise à jour mensuelle  
- **Dates de modification** : Basées sur les dates de publication des épisodes
- **URLs canoniques** : Format `/episode/{id}` pour tous les épisodes
- **Cache intelligent** : Utilise le même cache RSS de 30s pour les performances

## 📝 Variables d'environnement

Pour le déploiement, configurez :

```bash
export ESTAMITECH_ACCOUNT_ID="votre-compte-gcloud"
export ESTAMITECH_PROJECT_ID="votre-project-id-gcp"
export GOOGLE_CLOUD_SDK="/chemin/vers/google-cloud-sdk"
```

## 🤝 Participation au podcast

N'hésitez pas à me contacter pour discuter d'un projet ou d'une initiative tech dans le Nord de la France. Vous pouvez m'envoyer un message via les réseaux sociaux ou par email.

## 📱 Contact

- **Créateur** : Ludovic Borie
- **LinkedIn** : [ludovicborie](https://www.linkedin.com/in/ludovicborie/)
- **Mastodon** : [@bodul@piaille.fr](https://piaille.fr/@bodul)
- **Bluesky** : [@b0dul](https://bsky.app/profile/b0dul.bsky.social)

---

*L'Estamitech - La Tech du Nord* 🎙️