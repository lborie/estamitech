// Toggle mobile menu
const menuToggle = document.querySelector('.menu-toggle');
const navbarLinks = document.querySelector('.navbar-links');

menuToggle.addEventListener('click', () => {
    menuToggle.classList.toggle('active');
    navbarLinks.classList.toggle('active');
});

// Set current year in footer
document.getElementById('current-year').textContent = new Date().getFullYear();

// Copy RSS link to clipboard
function copyRSSLink() {
    const rssLink = 'https://feeds.zencastr.com/f/bOMlUWx6.rss';
    
    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(rssLink).then(() => {
            showToast('Lien RSS copié dans le presse-papier!');
        }).catch(() => {
            fallbackCopyTextToClipboard(rssLink);
        });
    } else {
        fallbackCopyTextToClipboard(rssLink);
    }
}

function fallbackCopyTextToClipboard(text) {
    const textArea = document.createElement("textarea");
    textArea.value = text;
    textArea.style.position = "fixed";
    textArea.style.top = 0;
    textArea.style.left = 0;
    textArea.style.width = "2em";
    textArea.style.height = "2em";
    textArea.style.padding = 0;
    textArea.style.border = "none";
    textArea.style.outline = "none";
    textArea.style.boxShadow = "none";
    textArea.style.background = "transparent";
    
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        const successful = document.execCommand('copy');
        if (successful) {
            showToast('Lien RSS copié dans le presse-papier!');
        } else {
            showToast('Impossible de copier le lien RSS');
        }
    } catch (err) {
        showToast('Impossible de copier le lien RSS');
    }
    
    document.body.removeChild(textArea);
}

function showToast(message) {
    const existingToast = document.querySelector('.toast');
    if (existingToast) {
        existingToast.remove();
    }
    
    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.classList.add('show');
    }, 10);
    
    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => {
            toast.remove();
        }, 300);
    }, 3000);
}

// Episode-specific functions
function validateEpisodeId(id) {
    if (!id) return null;
    return id.replace(/[^a-zA-Z0-9\-_.]/g, '');
}

function loadEpisode() {
    const urlParams = new URLSearchParams(window.location.search);
    const rawEpisodeId = urlParams.get('id');
    const episodeId = validateEpisodeId(rawEpisodeId);
    
    if (!episodeId) {
        displayError();
        return;
    }
    
    fetch('/rss')
        .then(response => response.json())
        .then(episodes => {
            const episode = episodes.find(ep => ep.GUID === episodeId || ep.Link === episodeId);
            
            if (episode) {
                displayEpisode(episode);
            } else {
                displayError();
            }
        })
        .catch(error => {
            console.error('Error loading episode:', error);
            displayError();
        });
}

function displayEpisode(episode) {
    document.title = `${episode.Title} - L'ESTAMITECH`;
    
    document.querySelector('meta[property="og:title"]').content = `${episode.Title} - L'ESTAMITECH`;
    document.querySelector('meta[property="og:description"]').content = episode.Description;
    document.querySelector('meta[property="og:image"]').content = episode.Image.Href;
    
    document.getElementById('episode-title').textContent = episode.Title;
    
    const date = new Date(episode.PubDate);
    const formattedDate = date.toLocaleDateString('fr-FR', {
        day: 'numeric',
        month: 'long',
        year: 'numeric'
    });
    document.getElementById('episode-date').innerHTML = `<i class="far fa-calendar-alt"></i> ${formattedDate}`;
    
    document.getElementById('episode-image').src = episode.Image.Href;
    document.getElementById('episode-image').alt = `Miniature de l'épisode - ${episode.Title}`;
    
    if (episode.Enclosure && episode.Enclosure.URL) {
        document.getElementById('episode-audio').src = episode.Enclosure.URL;
    } else {
        document.getElementById('episode-audio').src = episode.Link;
    }
    
    document.getElementById('episode-description').innerHTML = episode.Description;
}

function displayError() {
    document.getElementById('episode-title').textContent = 'Épisode non trouvé';
    document.getElementById('episode-date').innerHTML = '';
    document.getElementById('episode-description').innerHTML = '<p>Désolé, cet épisode n\'a pas pu être chargé.</p>';
}

// Index-specific functions
function initializeIndexPage() {
    // Close the menu when clicking on a link
    document.querySelectorAll('.navbar-links a').forEach(link => {
        link.addEventListener('click', () => {
            menuToggle.classList.remove('active');
            navbarLinks.classList.remove('active');
        });
    });

    // Back to Top button visibility
    const backToTopButton = document.getElementById('back-to-top');
    
    if (backToTopButton) {
        window.addEventListener('scroll', () => {
            if (window.pageYOffset > 300) {
                backToTopButton.classList.add('visible');
            } else {
                backToTopButton.classList.remove('visible');
            }
        });

        backToTopButton.addEventListener('click', (e) => {
            e.preventDefault();
            window.scrollTo({ top: 0, behavior: 'smooth' });
        });
    }

    // Smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            if (this.getAttribute('href') !== '#') {
                e.preventDefault();

                const targetId = this.getAttribute('href');
                const targetElement = document.querySelector(targetId);

                if (targetElement) {
                    const navbarHeight = document.querySelector('.navbar').offsetHeight;
                    const targetPosition = targetElement.getBoundingClientRect().top + window.pageYOffset - navbarHeight;

                    window.scrollTo({
                        top: targetPosition,
                        behavior: 'smooth'
                    });
                }
            }
        });
    });

    // Initialize episodes loading
    parseRSS(function(data) {
        if (data && data.length) {
            displayEpisodes(data);
        } else {
            handleEpisodesError("Aucun épisode trouvé");
        }
    });
}

function parseRSS(callback) {
    const xhr = new XMLHttpRequest();
    xhr.open('GET', `/rss`, true);
    xhr.onreadystatechange = function() {
        if (xhr.readyState === 4) {
            if (xhr.status === 200) {
                try {
                    const data = JSON.parse(xhr.responseText);
                    callback(data);
                } catch (error) {
                    handleEpisodesError("Error parsing response");
                }
            } else {
                handleEpisodesError("Failed to fetch episodes");
            }
        }
    };
    xhr.timeout = 10000;
    xhr.ontimeout = function() {
        handleEpisodesError("Request timed out");
    };
    xhr.onerror = function() {
        handleEpisodesError("Network error occurred");
    };
    xhr.send();
}

function displayEpisodes(episodes) {
    const episodeContainer = document.getElementById('podcast-episodes');
    episodeContainer.innerHTML = '';

    if (!episodes || episodes.length === 0) {
        episodeContainer.innerHTML = `
            <div class="no-episodes">
                <i class="fas fa-podcast"></i>
                <p>Aucun épisode trouvé. Revenez bientôt pour de nouveaux contenus!</p>
            </div>
        `;
        return;
    }

    episodes.forEach(episode => {
        const date = new Date(episode.PubDate);
        const formattedDate = date.toLocaleDateString('fr-FR', {
            day: 'numeric',
            month: 'long',
            year: 'numeric'
        });

        const episodeElement = document.createElement('div');
        episodeElement.classList.add('episode-card');

        const episodeId = episode.GUID || episode.Link;
        const episodePageUrl = `/episode?id=${encodeURIComponent(episodeId)}`;
        
        episodeElement.innerHTML = `
            <div class="episode-image">
                <a href="${episodePageUrl}"><img src="${episode.Image.Href}" alt="Miniature de l'épisode - ${episode.Title}"></a>
            </div>
            <div class="episode-details">
                <h3 class="episode-title"><a href="${episodePageUrl}" style="text-decoration: none; color: inherit;">${episode.Title}</a></h3>
                <div class="episode-meta">
                    <span><i class="far fa-calendar-alt"></i> ${formattedDate}</span>
                </div>
                <div class="episode-description">${episode.Description}</div>
            </div>
        `;

        episodeContainer.appendChild(episodeElement);
    });
}

function handleEpisodesError(message) {
    const episodeContainer = document.getElementById('podcast-episodes');
    episodeContainer.innerHTML = `
        <div class="loading-indicator" style="color: #dc3545;">
            <i class="fas fa-exclamation-triangle" style="animation: none;"></i>
            <p>${message}. Veuillez réessayer plus tard.</p>
        </div>
    `;
}