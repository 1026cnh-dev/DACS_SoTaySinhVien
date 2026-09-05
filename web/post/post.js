
function initViewportMode(){
  if(window.__DACS_SHARED_HEADER__) return;
  const mq=window.matchMedia('(max-width: 820px)');
  const header=document.querySelector('.site-header');
  const nav=document.querySelector('.category-nav');
  const marker=document.createComment('category-nav-origin');
  if(nav&&nav.parentNode) nav.parentNode.insertBefore(marker,nav);
  const apply=()=>{
    const mobile=mq.matches;
    document.documentElement.setAttribute('data-viewport-mode',mobile?'mobile':'desktop');
    document.body.classList.toggle('is-mobile-view',mobile);
    document.body.classList.toggle('is-desktop-view',!mobile);
    if(nav){
      if(mobile){
        if(nav.parentNode!==document.body) document.body.appendChild(nav);
      }else if(marker.parentNode){
        marker.parentNode.insertBefore(nav,marker.nextSibling);
      }else if(header){
        header.appendChild(nav);
      }
    }
  };
  apply();
  if(typeof mq.addEventListener==='function')mq.addEventListener('change',apply);
  else if(typeof mq.addListener==='function')mq.addListener(apply);
  window.addEventListener('resize',apply,{passive:true});
}
document.addEventListener('DOMContentLoaded', () => {
  initViewportMode();
  initThemeToggle();
  initAccountMenu();
  initReplyToggles();
  initShareButton();
  initPostVoting();
  initPostMoreMenu();
  initReportModal();
});

function initAccountMenu() {
  if(window.__DACS_SHARED_HEADER__) return;
  const menus = document.querySelectorAll('[data-account-menu]');
  menus.forEach((menu) => {
    const trigger = menu.querySelector('[data-account-trigger]');
    if (!trigger) return;
    trigger.addEventListener('click', (event) => {
      event.stopPropagation();
      const isOpen = menu.classList.contains('open');
      document.querySelectorAll('[data-account-menu].open').forEach((openMenu) => {
        openMenu.classList.remove('open');
        const openTrigger = openMenu.querySelector('[data-account-trigger]');
        if (openTrigger) openTrigger.setAttribute('aria-expanded', 'false');
      });
      if (!isOpen) {
        menu.classList.add('open');
        trigger.setAttribute('aria-expanded', 'true');
      }
    });
  });

  document.addEventListener('click', (event) => {
    if (event.target.closest('[data-account-menu]')) return;
    document.querySelectorAll('[data-account-menu].open').forEach((menu) => {
      menu.classList.remove('open');
      const trigger = menu.querySelector('[data-account-trigger]');
      if (trigger) trigger.setAttribute('aria-expanded', 'false');
    });
  });
}

function initThemeToggle() {
  if(window.__DACS_SHARED_HEADER__) return;
  const key = 'sotaysinhvien-theme';
  const stored = localStorage.getItem(key);
  applyTheme(stored === 'dark' ? 'dark' : 'light');

  document.querySelectorAll('[data-theme-toggle]').forEach((btn) => {
    btn.addEventListener('click', (event) => {
      event.preventDefault();
      event.stopPropagation();
      const next = document.body.classList.contains('theme-dark') ? 'light' : 'dark';
      applyTheme(next);
    });
  });

  function applyTheme(theme) {
    const dark = theme === 'dark';
    document.body.classList.toggle('theme-dark', dark);
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem(key, theme);
    document.querySelectorAll('[data-theme-toggle]').forEach((btn) => {
      btn.classList.toggle('active', dark);
      btn.setAttribute('aria-pressed', dark ? 'true' : 'false');
      const label = btn.querySelector('[data-theme-label]');
      if (label) label.textContent = dark ? 'Đổi sang sáng' : 'Đổi sang tối';
    });
  }
}

function initReplyToggles() {
  document.querySelectorAll('.reply-toggle').forEach((btn) => {
    btn.addEventListener('click', () => {
      const id = btn.getAttribute('data-reply-target');
      const el = document.getElementById(id);
      if (!el) return;
      const willOpen = el.classList.contains('hidden');
      document.querySelectorAll('.reply-form').forEach((form) => { if (form !== el) form.classList.add('hidden'); });
      el.classList.toggle('hidden');
      if (willOpen) {
        const textarea = el.querySelector('textarea');
        const replyName = btn.getAttribute('data-reply-name') || '';
        const mention = replyName ? `@${replyName} ` : '';
        if (textarea) {
          if (!textarea.value.trim() && mention) textarea.value = mention;
          textarea.focus();
          const pos = textarea.value.length;
          textarea.setSelectionRange(pos, pos);
        }
      }
    });
  });
}

function initShareButton() {
  const shareBtn = document.querySelector('.share-btn');
  if (!shareBtn) return;
  const defaultHTML = shareBtn.innerHTML;
  shareBtn.addEventListener('click', async () => {
    const url = location.origin + shareBtn.dataset.url;
    try {
      if (navigator.share) {
        await navigator.share({ url });
      } else if (navigator.clipboard) {
        await navigator.clipboard.writeText(url);
        shareBtn.innerHTML = '<img class="inline-icon" src="/web/icons/check.svg" alt=""> Đã sao chép';
        setTimeout(() => shareBtn.innerHTML = defaultHTML, 1600);
      }
    } catch (e) {}
  });
}

function initPostVoting() {
  document.querySelectorAll('form.vote-pill').forEach((form) => {
    form.addEventListener('submit', async (event) => {
      event.preventDefault();
      const submitter = event.submitter;
      if (!submitter || form.classList.contains('is-voting')) return;
      const postID = form.querySelector('input[name="post_id"]')?.value;
      const value = submitter.value;
      if (!postID || !value) return;
      form.classList.add('is-voting');
      form.querySelectorAll('button').forEach((btn) => btn.disabled = true);
      try {
        const body = new FormData(form);
        body.set('value', value);
        const res = await fetch(form.action, {
          method: 'POST',
          body,
          headers: { 'X-Requested-With': 'XMLHttpRequest', 'Accept': 'application/json' }
        });
        const data = await res.json().catch(() => ({}));
        if (res.status === 401 && data.redirect) {
          location.href = data.redirect;
          return;
        }
        if (!res.ok || !data.ok) throw new Error(data.message || 'Không thể bình chọn');
        const score = form.querySelector('.vote-score') || form.querySelector('span');
        if (score) {
          score.textContent = String(data.score);
          score.classList.remove('vote-score-pop');
          void score.offsetWidth;
          score.classList.add('vote-score-pop');
        }
        form.querySelectorAll('button[name="value"]').forEach((btn) => btn.classList.toggle('is-selected', Number(btn.value) === Number(data.user_vote)));
      } catch (err) {
        form.classList.add('vote-error');
        setTimeout(() => form.classList.remove('vote-error'), 500);
      } finally {
        form.classList.remove('is-voting');
        form.querySelectorAll('button').forEach((btn) => btn.disabled = false);
      }
    });
  });
}


function initPostMoreMenu() {
  const closeMenu = (menu) => {
    if (!menu) return;
    menu.classList.remove('open');
    const trigger = menu.querySelector('[data-post-more-trigger]');
    const dropdown = menu.querySelector('[data-post-more-dropdown]');
    trigger?.setAttribute('aria-expanded', 'false');
    if (dropdown) {
      dropdown.style.removeProperty('top');
      dropdown.style.removeProperty('left');
      dropdown.style.removeProperty('right');
      dropdown.style.removeProperty('max-height');
    }
  };

  const placeMenu = (menu) => {
    const trigger = menu.querySelector('[data-post-more-trigger]');
    const dropdown = menu.querySelector('[data-post-more-dropdown]');
    if (!trigger || !dropdown) return;

    const rect = trigger.getBoundingClientRect();
    const viewportWidth = document.documentElement.clientWidth;
    const viewportHeight = window.innerHeight;
    const width = Math.min(260, Math.max(210, viewportWidth - 24));
    const gap = 10;
    const edge = 12;

    let left = rect.right - width;
    left = Math.max(edge, Math.min(left, viewportWidth - width - edge));

    // Prefer opening below the trigger. If the menu would run off-screen,
    // position it above the trigger instead.
    const estimatedHeight = Math.min(dropdown.scrollHeight || 240, 360);
    let top = rect.bottom + gap;
    if (top + estimatedHeight > viewportHeight - edge) {
      top = Math.max(edge, rect.top - estimatedHeight - gap);
    }

    dropdown.style.width = `${width}px`;
    dropdown.style.left = `${Math.round(left)}px`;
    dropdown.style.right = 'auto';
    dropdown.style.top = `${Math.round(top)}px`;
    dropdown.style.maxHeight = `${Math.max(160, viewportHeight - top - edge)}px`;
  };

  document.querySelectorAll('[data-post-more]').forEach((menu) => {
    const trigger = menu.querySelector('[data-post-more-trigger]');
    if (!trigger) return;
    trigger.addEventListener('click', (event) => {
      event.preventDefault();
      event.stopPropagation();

      const wasOpen = menu.classList.contains('open');
      document.querySelectorAll('[data-post-more].open').forEach((other) => closeMenu(other));
      if (wasOpen) return;

      menu.classList.add('open');
      trigger.setAttribute('aria-expanded', 'true');
      placeMenu(menu);
      requestAnimationFrame(() => placeMenu(menu));
    });
  });

  document.addEventListener('click', (event) => {
    if (event.target.closest('[data-post-more-dropdown]')) return;
    document.querySelectorAll('[data-post-more].open').forEach((menu) => closeMenu(menu));
  });

  // A fixed-position dropdown should never remain detached from its trigger.
  window.addEventListener('resize', () => {
    document.querySelectorAll('[data-post-more].open').forEach((menu) => placeMenu(menu));
  }, { passive: true });
  window.addEventListener('scroll', () => {
    document.querySelectorAll('[data-post-more].open').forEach((menu) => closeMenu(menu));
  }, { passive: true, capture: true });
}

function initReportModal() {
  const modal = document.getElementById('reportPostModal');
  if (!modal) return;
  const open = () => { modal.classList.add('open'); modal.setAttribute('aria-hidden','false'); document.body.classList.add('modal-lock'); };
  const close = () => { modal.classList.remove('open'); modal.setAttribute('aria-hidden','true'); document.body.classList.remove('modal-lock'); };
  document.querySelectorAll('[data-open-report]').forEach((btn) => btn.addEventListener('click', (event) => {
    event.preventDefault();
    document.querySelectorAll('[data-post-more].open').forEach((menu) => menu.classList.remove('open'));
    open();
  }));
  modal.querySelectorAll('[data-close-report]').forEach((btn) => btn.addEventListener('click', close));
  modal.addEventListener('click', (event) => { if (event.target === modal) close(); });
}
