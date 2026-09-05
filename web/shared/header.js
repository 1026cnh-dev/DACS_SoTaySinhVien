window.__DACS_SHARED_HEADER__ = true;
// DACS_V1.12 · robust pinned category navigation for desktop, tablet and mobile
(function(){
  function syncHeaderOffset(){
    const nav=document.querySelector('[data-shared-category-nav]');
    if(!nav)return;
    document.body.classList.add('has-shared-header');
    const height=Math.ceil(nav.getBoundingClientRect().height || nav.offsetHeight || 48);
    document.documentElement.style.setProperty('--shared-header-offset',height+'px');
    document.documentElement.style.setProperty('--shared-category-nav-height',height+'px');
  }
  function initPinnedCategoryNav(){
    const nav=document.querySelector('[data-shared-category-nav]');
    if(!nav)return;
    let anchor=nav.previousElementSibling;
    if(!anchor || !anchor.classList.contains('category-nav-pin-anchor')){
      anchor=document.createElement('div');
      anchor.className='category-nav-pin-anchor';
      anchor.setAttribute('aria-hidden','true');
      nav.parentNode.insertBefore(anchor,nav);
    }
    let ticking=false;
    const update=()=>{
      ticking=false;
      const shouldPin=anchor.getBoundingClientRect().top<=0;
      document.body.classList.toggle('category-nav-pinned',shouldPin);
      anchor.style.height=shouldPin?Math.ceil(nav.getBoundingClientRect().height || nav.offsetHeight || 48)+'px':'0px';
      syncHeaderOffset();
    };
    const requestUpdate=()=>{if(!ticking){ticking=true;requestAnimationFrame(update);}};
    update();
    window.addEventListener('scroll',requestUpdate,{passive:true});
    window.addEventListener('resize',requestUpdate,{passive:true});
    window.addEventListener('orientationchange',requestUpdate,{passive:true});
    if('ResizeObserver' in window)new ResizeObserver(requestUpdate).observe(nav);
  }
  function initViewportMode(){
    const mq=window.matchMedia('(max-width: 820px)');
    const nav=document.querySelector('[data-shared-category-nav]');
    if(!nav)return;
    const apply=()=>{
      const mobile=mq.matches;
      document.documentElement.setAttribute('data-viewport-mode',mobile?'mobile':'desktop');
      document.body.classList.toggle('is-mobile-view',mobile);
      document.body.classList.toggle('is-desktop-view',!mobile);
      requestAnimationFrame(syncHeaderOffset);
    };
    apply();
    if(typeof mq.addEventListener==='function')mq.addEventListener('change',apply);else if(typeof mq.addListener==='function')mq.addListener(apply);
  }
  function setMenuOpen(menu,open){
    menu.classList.toggle('open',open);
    menu.querySelector('[data-account-trigger]')?.setAttribute('aria-expanded',open?'true':'false');
    document.body.classList.toggle('shared-account-menu-open',!!document.querySelector('[data-account-menu].open'));
  }
  function initAccountMenu(){
    document.querySelectorAll('[data-account-menu]').forEach(menu=>{
      const trigger=menu.querySelector('[data-account-trigger]');if(!trigger)return;
      trigger.addEventListener('click',e=>{
        e.preventDefault();e.stopPropagation();const open=!menu.classList.contains('open');
        document.querySelectorAll('[data-account-menu].open').forEach(x=>setMenuOpen(x,false));
        if(open)setMenuOpen(menu,true);
      });
    });
    document.addEventListener('click',e=>{
      if(!e.target.closest('[data-account-menu]'))document.querySelectorAll('[data-account-menu].open').forEach(x=>setMenuOpen(x,false));
    });
  }
  function initJumpToMainSearch(){
    const input=document.querySelector('[data-main-search-input]');
    const mast=document.querySelector('[data-shared-site-header] .masthead');
    document.querySelectorAll('[data-jump-main-search]').forEach(btn=>{
      btn.addEventListener('click',e=>{
        e.preventDefault();
        document.querySelectorAll('[data-account-menu].open').forEach(x=>setMenuOpen(x,false));
        const target=mast||input;
        if(!target)return;
        target.scrollIntoView({behavior:'smooth',block:'start'});
        const focusInput=()=>{
          input?.focus({preventScroll:true});
          if(input && typeof input.setSelectionRange==='function'){
            const n=input.value.length; input.setSelectionRange(n,n);
          }
        };
        window.setTimeout(focusInput,260);
      });
    });
  }
  function initCondensedHeader(){
    const mast=document.querySelector('[data-shared-site-header] .masthead');
    if(!mast)return;
    const update=()=>{
      if(window.matchMedia('(max-width: 820px)').matches){document.body.classList.remove('shared-header-condensed');return;}
      const bottom=mast.getBoundingClientRect().bottom;
      document.body.classList.toggle('shared-header-condensed',bottom<=8);
    };
    update();window.addEventListener('scroll',update,{passive:true});window.addEventListener('resize',update,{passive:true});
  }
  function initTheme(){
    const key='sotaysinhvien-theme';
    const apply=theme=>{const dark=theme==='dark';document.body.classList.toggle('theme-dark',dark);document.documentElement.setAttribute('data-theme',theme);document.querySelectorAll('[data-theme-toggle]').forEach(btn=>{btn.classList.toggle('active',dark);btn.setAttribute('aria-pressed',dark?'true':'false');const label=btn.querySelector('[data-theme-label]');if(label)label.textContent=dark?'Đổi sang sáng':'Đổi sang tối'})};
    apply(localStorage.getItem(key)==='dark'?'dark':'light');
    document.querySelectorAll('[data-theme-toggle]').forEach(btn=>btn.addEventListener('click',e=>{e.preventDefault();e.stopPropagation();const next=document.body.classList.contains('theme-dark')?'light':'dark';localStorage.setItem(key,next);apply(next)}));
  }
  document.addEventListener('DOMContentLoaded',()=>{initViewportMode();initAccountMenu();initJumpToMainSearch();initCondensedHeader();initTheme();initPinnedCategoryNav();syncHeaderOffset();window.addEventListener('resize',syncHeaderOffset,{passive:true});});
})();
