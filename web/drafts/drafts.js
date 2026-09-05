
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

function initSyncedHeader(){
  if(window.__DACS_SHARED_HEADER__) return;
  const menus=document.querySelectorAll('[data-account-menu]');
  menus.forEach(menu=>{const trigger=menu.querySelector('[data-account-trigger]'); if(!trigger)return; trigger.addEventListener('click',e=>{e.stopPropagation(); const open=!menu.classList.contains('open'); document.querySelectorAll('[data-account-menu].open').forEach(x=>x.classList.remove('open')); menu.classList.toggle('open',open); trigger.setAttribute('aria-expanded',open?'true':'false');});});
  document.addEventListener('click',e=>{if(e.target.closest('[data-account-menu]'))return;document.querySelectorAll('[data-account-menu].open').forEach(x=>x.classList.remove('open'));});
  const key='sotaysinhvien-theme';
  const apply=t=>{document.body.classList.toggle('theme-dark',t==='dark');document.documentElement.setAttribute('data-theme',t);document.querySelectorAll('[data-theme-toggle]').forEach(btn=>{btn.classList.toggle('active',t==='dark');const l=btn.querySelector('[data-theme-label]');if(l)l.textContent=t==='dark'?'Đổi sang sáng':'Đổi sang tối';});};
  apply(localStorage.getItem(key)==='dark'?'dark':'light');
  document.querySelectorAll('[data-theme-toggle]').forEach(btn=>btn.addEventListener('click',e=>{e.preventDefault();e.stopPropagation();const next=document.body.classList.contains('theme-dark')?'light':'dark';localStorage.setItem(key,next);apply(next);}));
}

document.addEventListener('DOMContentLoaded',()=>{initViewportMode();initSyncedHeader();});
