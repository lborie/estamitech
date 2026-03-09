// ============================================
// L'ESTAMITECH — Script
// ============================================

// --- Mobile menu toggle ---
const navToggle = document.querySelector('.nav__toggle');
const navLinks = document.querySelector('.nav__links');

if (navToggle && navLinks) {
    navToggle.addEventListener('click', () => {
        navToggle.classList.toggle('active');
        navLinks.classList.toggle('active');
    });

    // Close menu on link click
    navLinks.querySelectorAll('a').forEach(link => {
        link.addEventListener('click', () => {
            navToggle.classList.remove('active');
            navLinks.classList.remove('active');
        });
    });
}

// --- Navbar scroll effect ---
const nav = document.querySelector('.nav');
if (nav) {
    window.addEventListener('scroll', () => {
        nav.classList.toggle('scrolled', window.scrollY > 40);
    }, { passive: true });
}

// --- Back to top ---
const backToTop = document.getElementById('back-to-top');
if (backToTop) {
    window.addEventListener('scroll', () => {
        backToTop.classList.toggle('visible', window.scrollY > 400);
    }, { passive: true });

    backToTop.addEventListener('click', (e) => {
        e.preventDefault();
        window.scrollTo({ top: 0, behavior: 'smooth' });
    });
}

// --- Smooth anchor scrolling ---
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
        const href = this.getAttribute('href');
        if (href === '#') return;

        const target = document.querySelector(href);
        if (target) {
            e.preventDefault();
            const navHeight = document.querySelector('.nav')?.offsetHeight || 0;
            const top = target.getBoundingClientRect().top + window.scrollY - navHeight;
            window.scrollTo({ top, behavior: 'smooth' });
        }
    });
});

// --- Scroll reveal (IntersectionObserver) ---
const revealElements = document.querySelectorAll('.reveal');
if (revealElements.length > 0) {
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('revealed');
                observer.unobserve(entry.target);
            }
        });
    }, { threshold: 0.08, rootMargin: '0px 0px -40px 0px' });

    revealElements.forEach(el => observer.observe(el));
}

// --- Set current year ---
const yearEl = document.getElementById('current-year');
if (yearEl) {
    yearEl.textContent = new Date().getFullYear();
}

// --- RSS copy ---
function copyRSSLink() {
    const rssLink = 'https://feeds.zencastr.com/f/bOMlUWx6.rss';

    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(rssLink).then(() => {
            showToast('Lien RSS copié dans le presse-papier !');
        }).catch(() => {
            fallbackCopy(rssLink);
        });
    } else {
        fallbackCopy(rssLink);
    }
}

function fallbackCopy(text) {
    const ta = document.createElement('textarea');
    ta.value = text;
    Object.assign(ta.style, {
        position: 'fixed', top: 0, left: 0,
        width: '1px', height: '1px', opacity: 0,
    });
    document.body.appendChild(ta);
    ta.focus();
    ta.select();

    try {
        document.execCommand('copy')
            ? showToast('Lien RSS copié dans le presse-papier !')
            : showToast('Impossible de copier le lien RSS');
    } catch {
        showToast('Impossible de copier le lien RSS');
    }

    document.body.removeChild(ta);
}

function showToast(message) {
    document.querySelector('.toast')?.remove();

    const toast = document.createElement('div');
    toast.className = 'toast';
    toast.textContent = message;
    document.body.appendChild(toast);

    requestAnimationFrame(() => {
        requestAnimationFrame(() => toast.classList.add('show'));
    });

    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.remove(), 350);
    }, 3000);
}