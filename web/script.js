// ============================================
// L'ESTAMITECH — Script (Safari-compatible)
// ============================================

(function() {
    'use strict';

    // --- Mobile menu toggle ---
    var navToggle = document.querySelector('.nav__toggle');
    var navLinks = document.querySelector('.nav__links');

    if (navToggle && navLinks) {
        navToggle.addEventListener('click', function() {
            var open = navToggle.classList.toggle('active');
            navLinks.classList.toggle('active');
            navToggle.setAttribute('aria-expanded', open ? 'true' : 'false');
        });

        var links = navLinks.querySelectorAll('a');
        for (var i = 0; i < links.length; i++) {
            links[i].addEventListener('click', function() {
                navToggle.classList.remove('active');
                navLinks.classList.remove('active');
                navToggle.setAttribute('aria-expanded', 'false');
            });
        }
    }

    // --- Navbar scroll effect ---
    var nav = document.querySelector('.nav');
    if (nav) {
        window.addEventListener('scroll', function() {
            if (window.scrollY > 40) {
                nav.classList.add('scrolled');
            } else {
                nav.classList.remove('scrolled');
            }
        }, false);
    }

    // --- Back to top ---
    var backToTop = document.getElementById('back-to-top');
    if (backToTop) {
        window.addEventListener('scroll', function() {
            if (window.scrollY > 400) {
                backToTop.classList.add('visible');
            } else {
                backToTop.classList.remove('visible');
            }
        }, false);

        backToTop.addEventListener('click', function(e) {
            e.preventDefault();
            window.scrollTo({ top: 0, behavior: 'smooth' });
        });
    }

    // --- Smooth anchor scrolling ---
    var anchors = document.querySelectorAll('a[href^="#"]');
    for (var j = 0; j < anchors.length; j++) {
        anchors[j].addEventListener('click', function(e) {
            var href = this.getAttribute('href');
            if (href === '#') return;

            var target = document.querySelector(href);
            if (target) {
                e.preventDefault();
                var navEl = document.querySelector('.nav');
                var navHeight = navEl ? navEl.offsetHeight : 0;
                var top = target.getBoundingClientRect().top + window.scrollY - navHeight;
                window.scrollTo({ top: top, behavior: 'smooth' });
            }
        });
    }

    // --- Scroll reveal (IntersectionObserver) ---
    if ('IntersectionObserver' in window) {
        document.documentElement.classList.add('reveal-ready');

        var revealElements = document.querySelectorAll('.reveal');
        if (revealElements.length > 0) {
            var observer = new IntersectionObserver(function(entries) {
                for (var k = 0; k < entries.length; k++) {
                    if (entries[k].isIntersecting) {
                        entries[k].target.classList.add('revealed');
                        observer.unobserve(entries[k].target);
                    }
                }
            }, { threshold: 0.08, rootMargin: '0px 0px -40px 0px' });

            for (var m = 0; m < revealElements.length; m++) {
                observer.observe(revealElements[m]);
            }
        }
    }

    // --- Set current year ---
    var yearEl = document.getElementById('current-year');
    if (yearEl) {
        yearEl.textContent = new Date().getFullYear();
    }
})();

// --- RSS copy (global, called from onclick) ---
function copyRSSLink() {
    var rssLink = 'https://feeds.zencastr.com/f/bOMlUWx6.rss';

    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(rssLink).then(function() {
            showToast('Lien RSS copié dans le presse-papier !');
        }).catch(function() {
            fallbackCopy(rssLink);
        });
    } else {
        fallbackCopy(rssLink);
    }
}

function fallbackCopy(text) {
    var ta = document.createElement('textarea');
    ta.value = text;
    ta.style.position = 'fixed';
    ta.style.top = '0';
    ta.style.left = '0';
    ta.style.width = '1px';
    ta.style.height = '1px';
    ta.style.opacity = '0';
    document.body.appendChild(ta);
    ta.focus();
    ta.select();

    try {
        var ok = document.execCommand('copy');
        if (ok) {
            showToast('Lien RSS copié dans le presse-papier !');
        } else {
            showToast('Impossible de copier le lien RSS');
        }
    } catch (e) {
        showToast('Impossible de copier le lien RSS');
    }

    document.body.removeChild(ta);
}

function showToast(message) {
    var existing = document.querySelector('.toast');
    if (existing) existing.remove();

    var toast = document.createElement('div');
    toast.className = 'toast';
    toast.setAttribute('role', 'status'); // annonce le message aux lecteurs d'écran (aria-live implicite)
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(function() {
        toast.classList.add('show');
    }, 20);

    setTimeout(function() {
        toast.classList.remove('show');
        setTimeout(function() { toast.remove(); }, 350);
    }, 3000);
}