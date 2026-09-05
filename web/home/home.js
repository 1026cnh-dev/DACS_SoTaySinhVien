
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
  initQuickVoting();
  initSavedPostButtons();
  initClickablePostCards();

  const searchInputs = document.querySelectorAll('input[name="q"]');
  searchInputs.forEach(input => input.addEventListener('keydown', e => {
    if (e.key === 'Escape') input.value = '';
  }));

  const selects = document.querySelectorAll('.filter-row select');
  selects.forEach(select => {
    select.classList.toggle('selected', select.selectedIndex > 0);
  });

  enhanceFilterSelects();
  initHomeAdRotator();
  initRandomFeedAds();
  initSmartPostFilters();
  initExactCardRatio();
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

function enhanceFilterSelects() {
  const rows = document.querySelectorAll('.filter-row');
  rows.forEach((row) => {
    const existing = row.querySelectorAll('.filter-select');
    existing.forEach((item) => item.remove());

    const selects = Array.from(row.querySelectorAll('select'));
    selects.forEach((select, index) => {
      const wrapper = document.createElement('div');
      wrapper.className = 'filter-select';
      wrapper.dataset.index = String(index);
      if (select.matches('[data-smart-filter]')) wrapper.dataset.smart = 'true';

      const trigger = document.createElement('button');
      trigger.type = 'button';
      trigger.className = 'filter-select-trigger';
      trigger.setAttribute('aria-haspopup', 'listbox');
      trigger.setAttribute('aria-expanded', 'false');

      const label = document.createElement('span');
      label.className = 'filter-select-label';
      const small = document.createElement('small');
      small.textContent = (select.options[0]?.textContent || 'Bộ lọc').trim();
      const value = document.createElement('span');
      value.textContent = getSelectedOptionText(select);
      label.appendChild(small);
      label.appendChild(value);

      const arrow = document.createElement('span');
      arrow.className = 'filter-select-arrow';
      arrow.innerHTML = '<img src="/web/icons/arrow-down.svg" alt="">';
      trigger.appendChild(label);
      trigger.appendChild(arrow);

      const menu = document.createElement('div');
      menu.className = 'filter-select-menu';
      menu.setAttribute('role', 'listbox');

      Array.from(select.options).forEach((option, optIndex) => {
        const item = document.createElement('button');
        item.type = 'button';
        item.className = 'filter-select-option';
        if (optIndex === select.selectedIndex) item.classList.add('active');
        item.dataset.index = String(optIndex);
        item.setAttribute('role', 'option');
        item.setAttribute('aria-selected', optIndex === select.selectedIndex ? 'true' : 'false');

        const main = document.createElement('span');
        main.className = 'filter-option-main';
        const title = document.createElement('b');
        title.textContent = option.textContent.trim();
        main.appendChild(title);
        if (optIndex === 0) {
          const hint = document.createElement('small');
          hint.textContent = 'Chọn tiêu chí lọc phù hợp';
          main.appendChild(hint);
        }

        const check = document.createElement('span');
        check.className = 'filter-option-check';
        check.innerHTML = '<img src="/web/icons/check.svg" alt="">';

        item.appendChild(main);
        item.appendChild(check);
        item.addEventListener('click', () => {
          select.selectedIndex = optIndex;
          select.classList.toggle('selected', select.selectedIndex > 0);
          value.textContent = getSelectedOptionText(select);
          menu.querySelectorAll('.filter-select-option').forEach((btn) => {
            const active = Number(btn.dataset.index) === optIndex;
            btn.classList.toggle('active', active);
            btn.setAttribute('aria-selected', active ? 'true' : 'false');
          });
          wrapper.classList.remove('open');
          trigger.setAttribute('aria-expanded', 'false');
          select.dispatchEvent(new Event('change', { bubbles: true }));
        });
        menu.appendChild(item);
      });

      trigger.addEventListener('click', (event) => {
        event.stopPropagation();
        const opening = !wrapper.classList.contains('open');
        document.querySelectorAll('.filter-select.open').forEach((openItem) => {
          openItem.classList.remove('open');
          const openTrigger = openItem.querySelector('.filter-select-trigger');
          if (openTrigger) openTrigger.setAttribute('aria-expanded', 'false');
        });
        if (opening) {
          wrapper.classList.add('open');
          trigger.setAttribute('aria-expanded', 'true');
        }
      });

      wrapper.appendChild(trigger);
      wrapper.appendChild(menu);
      row.appendChild(wrapper);
    });
  });

  document.addEventListener('click', (event) => {
    if (event.target.closest('.filter-select')) return;
    document.querySelectorAll('.filter-select.open').forEach((openItem) => {
      openItem.classList.remove('open');
      const openTrigger = openItem.querySelector('.filter-select-trigger');
      if (openTrigger) openTrigger.setAttribute('aria-expanded', 'false');
    });
  });
}

function getSelectedOptionText(select) {
  const option = select.options[select.selectedIndex];
  if (!option) return '';
  if (select.selectedIndex === 0) {
    return select.matches('[data-smart-filter]') ? 'Đề xuất' : 'Tất cả';
  }
  return option.textContent.trim();
}


function initClickablePostCards() {
  document.querySelectorAll('.clickable-post-card[data-post-url]').forEach((card) => {
    card.setAttribute('tabindex', '0');
    card.setAttribute('role', 'link');

    const shouldIgnore = (target) => Boolean(target.closest(
      'button, form, input, textarea, select, label, [data-card-no-nav], .comment-count, .read-compact, [data-account-menu]'
    ));

    card.addEventListener('click', (event) => {
      if (shouldIgnore(event.target)) return;
      // Links inside title/image already navigate correctly; let the browser handle them.
      if (event.target.closest('a')) return;
      window.location.href = card.dataset.postUrl;
    });

    card.addEventListener('keydown', (event) => {
      if (event.key !== 'Enter' && event.key !== ' ') return;
      if (shouldIgnore(event.target)) return;
      event.preventDefault();
      window.location.href = card.dataset.postUrl;
    });
  });
}

function initQuickVoting() {
  document.querySelectorAll('form.quick-vote').forEach((form) => {
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
        const data = new FormData(form);
        data.set('value', value);
        const response = await fetch(form.action, {
          method: 'POST',
          body: data,
          headers: {
            'X-Requested-With': 'XMLHttpRequest',
            'Accept': 'application/json'
          }
        });
        const result = await response.json().catch(() => ({}));
        if (response.status === 401 && result.redirect) {
          window.location.href = result.redirect;
          return;
        }
        if (!response.ok || !result.ok) throw new Error(result.message || 'Không thể bình chọn.');

        document.querySelectorAll('form.quick-vote').forEach((sameForm) => {
          const sameID = sameForm.querySelector('input[name="post_id"]')?.value;
          if (sameID !== String(result.post_id)) return;
          const score = sameForm.querySelector('b');
          if (score) {
            score.textContent = String(result.score);
            score.classList.remove('vote-score-pop');
            void score.offsetWidth;
            score.classList.add('vote-score-pop');
          }
          sameForm.querySelectorAll('button[name="value"]').forEach((btn) => {
            btn.classList.toggle('is-selected', Number(btn.value) === Number(result.user_vote));
          });
        });
      } catch (error) {
        form.classList.add('vote-error');
        setTimeout(() => form.classList.remove('vote-error'), 500);
      } finally {
        form.classList.remove('is-voting');
        form.querySelectorAll('button').forEach((btn) => btn.disabled = false);
      }
    });
  });
}

function initSavedPostButtons() {
  document.querySelectorAll('.save-post-btn[data-post-id]').forEach((button) => {
    button.addEventListener('click', async (event) => {
      event.preventDefault();
      event.stopPropagation();
      if (button.disabled) return;
      button.disabled = true;
      try {
        const data = new FormData();
        data.set('post_id', button.dataset.postId || '');
        data.set('next', button.dataset.next || '/');
        const response = await fetch('/post/save', {
          method: 'POST',
          body: data,
          headers: {'X-Requested-With':'XMLHttpRequest','Accept':'application/json'}
        });
        const result = await response.json().catch(() => ({}));
        if (response.status === 401 && result.redirect) {
          window.location.href = result.redirect;
          return;
        }
        if (!response.ok || !result.ok) throw new Error(result.message || 'Không thể lưu bài viết.');
        document.querySelectorAll(`.save-post-btn[data-post-id="${button.dataset.postId}"]`).forEach((same) => {
          same.classList.toggle('saved', Boolean(result.saved));
          const label = same.querySelector('span');
          if (label) label.textContent = result.saved ? 'Đã lưu' : 'Lưu';
          same.title = result.saved ? 'Bỏ lưu bài viết' : 'Lưu bài viết';
        });
      } catch (error) {
        button.classList.add('save-error');
        setTimeout(() => button.classList.remove('save-error'), 500);
      } finally {
        button.disabled = false;
      }
    });
  });
}


// DACS_V1.20.4 · Smart filters use data already rendered with each post card.
function initSmartPostFilters(){
  const smartSelect=document.querySelector('[data-smart-filter]');
  if(!smartSelect)return;
  smartSelect.addEventListener('change',()=>applySmartPostFilter(smartSelect.value));
  applySmartPostFilter(smartSelect.value || 'all');
}
function parseCardDate(value){
  if(!value)return 0;
  const v=String(value).trim();
  const vn=v.match(/^(\d{1,2})\/(\d{1,2})\/(\d{4})$/);
  if(vn)return new Date(Number(vn[3]),Number(vn[2])-1,Number(vn[1])).getTime();
  const iso=Date.parse(v);
  return Number.isFinite(iso)?iso:0;
}
function cardMetric(card,mode){
  const score=Number(card.dataset.score)||0;
  const comments=Number(card.dataset.comments)||0;
  const date=parseCardDate(card.dataset.date);
  const deadline=parseCardDate(card.dataset.deadline);
  if(mode==='newest')return date;
  if(mode==='popular')return comments*4+score*2;
  if(mode==='hot')return score*5+comments*3+date/1e12;
  if(mode==='deadline')return deadline||Number.MAX_SAFE_INTEGER;
  return 0;
}
function applySmartPostFilter(mode){
  const board=document.getElementById('feed-4-cot');
  if(!board)return;
  const cards=[...board.querySelectorAll('.smart-filter-card')];
  const currentCategory=document.querySelector('[data-smart-filter-row]')?.dataset.category||'';
  cards.forEach(card=>card.classList.remove('smart-hidden'));
  let visible=cards.slice();
  if(currentCategory){
    visible=visible.filter(card=>!card.dataset.category||card.dataset.category===currentCategory);
    cards.forEach(card=>{if(card.dataset.category&&card.dataset.category!==currentCategory)card.classList.add('smart-hidden')});
  }
  if(mode==='image'){
    visible=visible.filter(card=>card.dataset.hasImage==='1');
    cards.forEach(card=>{if(!card.classList.contains('smart-hidden')&&card.dataset.hasImage!=='1')card.classList.add('smart-hidden')});
  }else if(mode==='pinned'){
    visible=visible.filter(card=>card.dataset.pinned==='1');
    cards.forEach(card=>{if(!card.classList.contains('smart-hidden')&&card.dataset.pinned!=='1')card.classList.add('smart-hidden')});
  }else if(mode==='deadline'){
    const now=new Date();now.setHours(0,0,0,0);
    visible=visible.filter(card=>{const d=parseCardDate(card.dataset.deadline);return d&&d>=now.getTime();}).sort((a,b)=>cardMetric(a,mode)-cardMetric(b,mode));
    cards.forEach(card=>{const d=parseCardDate(card.dataset.deadline);if(!card.classList.contains('smart-hidden')&&(!d||d<now.getTime()))card.classList.add('smart-hidden')});
  }else if(mode==='popular'){
    visible.sort((a,b)=>cardMetric(b,mode)-cardMetric(a,mode));
  }else if(mode==='hot'){
    visible.sort((a,b)=>cardMetric(b,mode)-cardMetric(a,mode));
  }else if(mode==='newest'){
    visible.sort((a,b)=>cardMetric(b,mode)-cardMetric(a,mode));
  }
  // Reorder only inside each native column so the 4-column layout remains intact.
  ['.pinned-list','.today-two-columns','.hot-ranking'].forEach(selector=>{
    const container=board.querySelector(selector);if(!container)return;
    const local=[...container.querySelectorAll(':scope > .smart-filter-card:not(.smart-hidden)')];
    if(['newest','popular','hot','deadline'].includes(mode)){
      local.sort((a,b)=>mode==='deadline'?cardMetric(a,mode)-cardMetric(b,mode):cardMetric(b,mode)-cardMetric(a,mode)).forEach(card=>container.appendChild(card));
    }
  });
  document.querySelectorAll('.smart-filter-empty').forEach(x=>x.remove());
  const shown=cards.filter(card=>!card.classList.contains('smart-hidden')).length;
  if(!shown&&mode!=='all'){
    const zone=board.querySelector('.today-two-columns')||board;
    const msg=document.createElement('div');msg.className='smart-filter-empty';msg.textContent='Chưa có bài viết phù hợp với bộ lọc này.';zone.appendChild(msg);
  }
  placeRandomFeedAds();
  scheduleExactCardRatio();
}



// DACS_V1.11 · Banner quảng cáo trang chủ luân phiên mỗi 5 giây.
// Chỉ chạy khi cùng vị trí có từ 2 quảng cáo đang bật trở lên.
function initHomeAdRotator(){
  document.querySelectorAll('[data-home-ad-rotator]').forEach((rotator)=>{
    const slides=[...rotator.querySelectorAll('[data-ad-slide]')];
    const dots=[...rotator.querySelectorAll('[data-ad-dot]')];
    if(!slides.length)return;

    let index=Math.max(0,slides.findIndex((slide)=>slide.classList.contains('is-active')));
    let timer=0;
    const interval=Math.max(1000,Number(rotator.dataset.interval)||5000);

    const show=(next)=>{
      index=(next+slides.length)%slides.length;
      slides.forEach((slide,i)=>{
        const active=i===index;
        slide.classList.toggle('is-active',active);
        slide.setAttribute('aria-hidden',active?'false':'true');
        slide.tabIndex=active?0:-1;
      });
      dots.forEach((dot,i)=>dot.classList.toggle('is-active',i===index));
    };

    if(slides.length<2){
      show(0);
      return;
    }

    const stop=()=>{
      if(timer){window.clearInterval(timer);timer=0;}
    };
    const start=()=>{
      stop();
      timer=window.setInterval(()=>show(index+1),interval);
    };

    document.addEventListener('visibilitychange',()=>{
      if(document.hidden)stop(); else start();
    });
    show(index);
    start();
  });
}

// DACS_V1.25.0 · Random in-feed advertisements.
// - Trang chủ: quảng cáo xen kẽ ngẫu nhiên trong cột Ghim và Xu hướng.
// - Trang chuyên mục: quảng cáo chỉ nằm ở cột đầu tiên, vị trí dọc thay đổi mỗi lần tải/lọc.
function shuffleFeedAds(items){
  const result=items.slice();
  for(let i=result.length-1;i>0;i--){
    const j=Math.floor(Math.random()*(i+1));
    [result[i],result[j]]=[result[j],result[i]];
  }
  return result;
}

function randomUniqueSlots(postCount,adCount){
  if(adCount<=0)return [];
  if(postCount<=0)return [0];
  const slots=[];
  for(let i=1;i<=postCount;i++)slots.push(i);
  return shuffleFeedAds(slots).slice(0,Math.min(adCount,slots.length)).sort((a,b)=>a-b);
}

function cloneFeedAd(template,placement){
  const node=template.content.firstElementChild?.cloneNode(true);
  if(!node)return null;
  node.dataset.adPlacement=placement;
  if(placement==='category-main')node.classList.add('category-inline-ad');
  return node;
}

function insertFeedAdsInto(container,templates,placement,count){
  if(!container||!templates.length||count<=0)return;
  const posts=[...container.children].filter((child)=>child.classList.contains('smart-filter-card')&&!child.classList.contains('smart-hidden'));
  if(!posts.length&&placement==='category-main')return;
  const selected=shuffleFeedAds(templates).slice(0,Math.min(count,templates.length));
  const slots=randomUniqueSlots(posts.length,selected.length);
  selected.forEach((template,index)=>{
    const ad=cloneFeedAd(template,placement);
    if(!ad)return;
    const slot=slots[index]??posts.length;
    const target=posts[slot]||null;
    container.insertBefore(ad,target);
  });
}

function placeRandomFeedAds(){
  const board=document.querySelector('.four-column-board');
  if(!board)return;
  document.querySelectorAll('.inline-feed-ad').forEach((ad)=>ad.remove());
  const templates=[...document.querySelectorAll('.feed-ad-template')];
  if(!templates.length)return;

  if(board.classList.contains('category-focused-board')){
    const main=board.querySelector('[data-ad-feed="category-main"]');
    if(!main)return;
    const posts=[...main.querySelectorAll(':scope > .smart-filter-card:not(.smart-hidden)')];
    if(!posts.length)return;
    const count=Math.min(templates.length,4,Math.max(1,Math.ceil(posts.length/15)));
    insertFeedAdsInto(main,templates,'category-main',count);
  }else{
    const pinned=board.querySelector('[data-ad-feed="pinned"]');
    const trending=board.querySelector('[data-ad-feed="trending"]');
    const pinnedPosts=pinned?[...pinned.querySelectorAll(':scope > .smart-filter-card:not(.smart-hidden)')]:[];
    const trendingPosts=trending?[...trending.querySelectorAll(':scope > .smart-filter-card:not(.smart-hidden)')]:[];
    const pinCount=Math.min(templates.length,2,Math.max(1,Math.ceil(Math.max(1,pinnedPosts.length)/3)));
    const trendCount=Math.min(templates.length,2,Math.max(1,Math.ceil(Math.max(1,trendingPosts.length)/3)));
    insertFeedAdsInto(pinned,templates,'pinned',pinCount);
    insertFeedAdsInto(trending,templates,'trending',trendCount);
  }
  scheduleExactCardRatio();
}

function initRandomFeedAds(){
  placeRandomFeedAds();
}

// DACS_V1.24.17 · Exact 2:1 card ratio across all four columns.
// 2 thẻ không ảnh + 1 khoảng cách giữa chúng = đúng 1 thẻ có ảnh.
// Chiều cao đơn vị được tính từ bề rộng cột để bố cục vẫn cân đối khi resize.
let cardRatioFrame=0;
let cardRatioObserver=null;

function clampNumber(value,min,max){
  return Math.max(min,Math.min(max,value));
}

function setExactCardRatio(){
  const board=document.querySelector('.four-column-board');
  if(!board || board.classList.contains('document-category-board'))return;

  // Mobile trở về chiều cao tự nhiên.
  if(window.matchMedia('(max-width: 560px)').matches){
    board.style.removeProperty('--card-unit');
    return;
  }

  const main=board.querySelector('.today-two-columns');
  if(main){
    const visible=[...main.querySelectorAll(':scope > .daily-card:not(.smart-hidden)')];
    const sample=visible[0];
    if(sample){
      const width=sample.getBoundingClientRect().width;
      // Ở bề rộng ~300–360px, unit khoảng 235–275px. Ảnh card = 2*unit + gap.
      const unit=Math.round(clampNumber(width*0.72+24,220,286));
      main.style.setProperty('--card-unit',`${unit}px`);
      board.style.setProperty('--card-unit',`${unit}px`);
    }
  }

  // V1.24.17: cột 1 và 4 dùng chung --card-unit với hai cột giữa.
}

function scheduleExactCardRatio(){
  if(cardRatioFrame)cancelAnimationFrame(cardRatioFrame);
  cardRatioFrame=requestAnimationFrame(()=>{
    cardRatioFrame=0;
    setExactCardRatio();
  });
}

function initExactCardRatio(){
  const board=document.querySelector('.four-column-board');
  if(!board)return;
  if('ResizeObserver' in window){
    cardRatioObserver=new ResizeObserver(scheduleExactCardRatio);
    cardRatioObserver.observe(board);
  }
  window.addEventListener('resize',scheduleExactCardRatio,{passive:true});
  window.addEventListener('load',scheduleExactCardRatio,{once:true});
  if(document.fonts&&document.fonts.ready)document.fonts.ready.then(scheduleExactCardRatio).catch(()=>{});
  scheduleExactCardRatio();
}

// DACS_V1.23.0 · Bộ lọc động lấy trực tiếp từ dữ liệu form chuyên đề.
document.querySelectorAll('[data-dynamic-filter]').forEach((select)=>{
  select.addEventListener('change',()=>select.closest('form')?.requestSubmit());
});

// ===== DACS_V1.24.33 · Adaptive card preview =====
// Tận dụng khoảng trống do tiêu đề ngắn để hiển thị thêm nội dung.
// Khi nội dung không vừa, cắt theo đúng số dòng và nối "… Xem thêm".
let adaptivePreviewFrame=0;

function cardTitleLineCount(card){
  const title=card.querySelector('h3:not(.document-preview-title), h4');
  if(!title)return 2;
  const style=getComputedStyle(title);
  const lineHeight=parseFloat(style.lineHeight)||24;
  const height=title.getBoundingClientRect().height;
  return Math.max(1,Math.min(2,Math.round(height/lineHeight)||1));
}

function previewMaxLines(card){
  const titleLines=cardTitleLineCount(card);
  const hasImage=card.dataset.hasImage==='1';
  // Với ảnh: title 1 dòng => 3 dòng nội dung, title 2 dòng => 2 dòng.
  if(hasImage)return titleLines===1?3:2;
  // Không ảnh: dùng vùng trống tới footer; title càng ngắn càng được thêm dòng.
  return titleLines===1?4:3;
}

function fitCardPreview(preview){
  const card=preview.closest('.daily-card,.compact-feed-card');
  const copy=preview.querySelector('.card-preview-copy');
  const more=preview.querySelector('.card-preview-more');
  if(!card||!copy||!more)return;

  const full=(copy.dataset.fullText||copy.textContent||'').replace(/\s+/g,' ').trim();
  copy.dataset.fullText=full;
  const maxLines=previewMaxLines(card);
  const style=getComputedStyle(preview);
  const lineHeight=parseFloat(style.lineHeight)||18.75;
  const maxHeight=Math.ceil(lineHeight*maxLines)+1;
  preview.style.maxHeight=`${maxHeight}px`;

  copy.textContent=full;
  more.hidden=true;
  if(preview.scrollHeight<=maxHeight+1)return;

  more.hidden=false;
  const words=full.split(' ').filter(Boolean);
  let lo=0,hi=words.length,best=0;
  while(lo<=hi){
    const mid=(lo+hi)>>1;
    copy.textContent=(mid?words.slice(0,mid).join(' '):'')+'… ';
    if(preview.scrollHeight<=maxHeight+1){best=mid;lo=mid+1}else{hi=mid-1}
  }
  copy.textContent=(best?words.slice(0,best).join(' '):'')+'… ';
}

function applyAdaptiveCardPreviews(){
  document.querySelectorAll('[data-card-preview]').forEach(fitCardPreview);
  scheduleExactCardRatio();
}

function scheduleAdaptiveCardPreviews(){
  if(adaptivePreviewFrame)cancelAnimationFrame(adaptivePreviewFrame);
  adaptivePreviewFrame=requestAnimationFrame(()=>{
    adaptivePreviewFrame=0;
    applyAdaptiveCardPreviews();
  });
}

document.addEventListener('DOMContentLoaded',scheduleAdaptiveCardPreviews);
window.addEventListener('load',scheduleAdaptiveCardPreviews,{once:true});
window.addEventListener('resize',scheduleAdaptiveCardPreviews,{passive:true});
if(document.fonts&&document.fonts.ready)document.fonts.ready.then(scheduleAdaptiveCardPreviews).catch(()=>{});
