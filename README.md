# L'Estamitech

[![Website](https://img.shields.io/badge/Website-estamitech.fr-5d1692)](https://estamitech.fr)
[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://golang.org/)
[![Google App Engine](https://img.shields.io/badge/Deployed%20on-Google%20App%20Engine-4285F4?logo=google-cloud)](https://cloud.google.com/appengine)

> **La Tech du Nord** - Un podcast par Ludovic Borie qui met en avant les projets et initiatives tech de la région Nord de la France.

## 📖 À propos

L'Estamitech est un podcast dédié à la tech du Nord de la France. Chaque épisode met en avant un projet ou une initiative, permettant de discuter tech dans la bonne humeur et sans langue de bois.

Le site web [estamitech.fr](https://estamitech.fr) sert de vitrine pour le podcast, proposant :
- Une présentation du podcast et de son créateur
- La liste des épisodes récents
- Une carte interactive des recommandations des invités
- Des liens vers toutes les plateformes d'écoute

## 🚀 Technologies utilisées

- **Backend** : Go 1.23
- **Frontend** : HTML5, CSS3, JavaScript vanilla
- **Hébergement** : Google App Engine
- **RSS** : Intégration avec Zencastr pour les épisodes
- **Cartes** : Framacarte pour les recommandations

## 📁 Structure du projet

```
estamitech/
├── main.go              # Serveur HTTP principal
├── go.mod               # Dépendances Go
├── app.yaml             # Configuration Google App Engine
├── Makefile             # Scripts de développement et déploiement
├── static/              # Assets statiques
│   ├── LogoEstamitech.jpg
│   ├── favicon.ico
│   └── spotify-podcast-badge-blk-grn-165x40.png
└── web/                 # Fichiers web
    ├── index.html       # Page principale
    └── style.css        # Styles CSS
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

## 🔧 API Endpoints

Le serveur expose plusieurs endpoints :

- `GET /` - Page d'accueil du podcast
- `GET /rss` - API JSON pour récupérer les épisodes du podcast

### Endpoint RSS

L'endpoint `/rss` récupère le flux RSS depuis Zencastr et le convertit en JSON :

```json
[
  {
    "Title": "Titre de l'épisode",
    "PubDate": "Date de publication",
    "Description": "Description de l'épisode",
    "Image": {
      "Href": "URL de l'image"
    }
  }
]
```

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