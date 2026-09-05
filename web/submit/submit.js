
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

document.addEventListener('DOMContentLoaded',()=>{
  initViewportMode();
  initSyncedHeader();
  const form=document.getElementById('submitPostForm');
  const select=document.getElementById('categorySelect');
  let fields=[...document.querySelectorAll('[data-field]')];
  const dynamicFields=document.getElementById('dynamicFields');
  let customValueCache={};
  try{
    const raw=form?.dataset?.customValuesB64||'';
    if(raw){const binary=atob(raw);const bytes=Uint8Array.from(binary,ch=>ch.charCodeAt(0));customValueCache=JSON.parse(new TextDecoder().decode(bytes)||'{}')||{};}
  }catch(e){customValueCache={};}
  const list=document.getElementById('positionsList');
  const tpl=document.getElementById('positionTemplate');
  const picker=document.getElementById('categoryPicker');
  const trigger=document.getElementById('categoryTrigger');
  const menu=document.getElementById('categoryMenu');
  const selectedText=document.getElementById('selectedCategoryText');
  const optionButtons=[...document.querySelectorAll('.category-option')];
  const previewModal=document.getElementById('postPreviewModal');
  const documentInput=form?.querySelector('[data-document-input]');
  const documentPreviewBtn=document.getElementById('previewDocumentBtn');
  const documentFileName=document.getElementById('documentFileName');
  const documentPreviewModal=document.getElementById('documentPreviewModal');
  const documentPreviewStage=document.getElementById('documentPreviewStage');
  const documentPreviewName=document.getElementById('documentPreviewName');
  const documentPreviewType=document.getElementById('documentPreviewType');
  let documentObjectURL='';

  function applyFieldLabel(el, configItem){
    if(!el) return;
    let label=null;
    for(const child of el.children){
      if(child.tagName==='LABEL'){ label=child; break; }
    }
    if(!label && el.dataset.field==='positions') label=el.querySelector('.positions-head label');
    if(!label) label=el.querySelector('label');
    if(!label) return;

    let textNode=[...label.childNodes].find(node=>node.nodeType===Node.TEXT_NODE && node.textContent.trim());
    if(!el.dataset.defaultFieldLabel){
      el.dataset.defaultFieldLabel=(textNode?.textContent||'').trim();
    }
    const next=String(configItem?.label||el.dataset.defaultFieldLabel||'').trim();
    if(!next) return;
    if(!textNode){
      textNode=document.createTextNode(next);
      label.insertBefore(textNode,label.firstChild);
    }else{
      textNode.textContent=next;
    }
  }

  function isCustomConfig(c){return !!c && String(c.key||'').startsWith('custom_');}
  function syncCustomValueCache(){
    form?.querySelectorAll('[data-custom-field="1"] [name]').forEach(control=>{customValueCache[control.name]=control.value||'';});
  }
  function renderCustomFields(config){
    if(!dynamicFields)return;
    syncCustomValueCache();
    dynamicFields.querySelectorAll('[data-custom-field="1"]').forEach(el=>el.remove());
    (config||[]).filter(c=>c?.enabled&&isCustomConfig(c)).sort((a,b)=>(a.order||999)-(b.order||999)).forEach(c=>{
      const section=document.createElement('section');
      const type=String(c.type||'text').toLowerCase();
      section.className=`form-block ${type==='textarea'?'':'two-col'} custom-form-block`;
      section.dataset.field=c.key;
      section.dataset.customField='1';
      const label=document.createElement('label');
      label.append(document.createTextNode(String(c.label||'Trường tùy chỉnh')));
      const mark=document.createElement('span');mark.className='required-mark';label.appendChild(mark);
      let control;
      if(type==='textarea'){
        control=document.createElement('textarea');control.rows=5;control.placeholder=`Nhập ${String(c.label||'nội dung').toLowerCase()}`;
      }else{
        control=document.createElement('input');
        control.type=['number','email','url','date'].includes(type)?type:'text';
        if(type==='url')control.placeholder='https://...';
        else if(type==='email')control.placeholder='name@example.com';
        else if(type==='number')control.placeholder='0';
        else if(type!=='date')control.placeholder=`Nhập ${String(c.label||'thông tin').toLowerCase()}`;
      }
      control.name=c.key;control.value=String(customValueCache[c.key]||'');control.dataset.requiredTarget='';
      label.appendChild(control);section.appendChild(label);dynamicFields.appendChild(section);
    });
    fields=[...document.querySelectorAll('[data-field]')];
  }

  async function syncFilterSuggestions(config){
    const categoryID=Number(select?.value)||0;
    document.querySelectorAll('datalist[data-dynamic-filter-list]').forEach(x=>x.remove());
    for(const c of (config||[])){
      if(!c?.enabled||(!c?.filterable&&!c?.suggest_values))continue;
      const control=form?.querySelector(`[name="${c.key}"]`);
      if(!control)continue;
      control.dataset.filterField=c.key;
      control.dataset.allowCustom=c.allow_custom?'1':'0';
      let values=[];
      if(c.suggest_values&&categoryID){
        try{
          const res=await fetch(`/filter-options?category_id=${categoryID}&field=${encodeURIComponent(c.key)}`,{headers:{Accept:'application/json'}});
          const data=await res.json(); if(data?.ok&&Array.isArray(data.values))values=data.values;
        }catch(e){}
      }
      if(control.tagName==='INPUT'){
        const id=`filter-values-${c.key}`;
        const dl=document.createElement('datalist');dl.id=id;dl.dataset.dynamicFilterList='1';
        values.forEach(v=>{const o=document.createElement('option');o.value=v;dl.appendChild(o)});
        document.body.appendChild(dl);control.setAttribute('list',id);
        let hint=control.closest('label')?.querySelector('.filter-data-hint');
        if(!hint){hint=document.createElement('small');hint.className='filter-data-hint';control.insertAdjacentElement('afterend',hint)}
        hint.textContent=c.allow_custom?'Có thể chọn gợi ý hoặc nhập giá trị mới. Giá trị mới sẽ chờ quản trị viên duyệt.':'Chọn đúng nhóm dữ liệu đã được duyệt.';
      }
    }
  }

  function applyConfig(){
    const opt=select?.options[select.selectedIndex];
    const allowedFormats=String(opt?.dataset?.documentFormats||'pdf,doc,docx,xls,xlsx,ppt,pptx,txt,zip').split(',').map(v=>v.trim().toLowerCase()).filter(Boolean);
    const docInput=form?.querySelector('[data-document-input]');
    if(docInput){
      docInput.accept=allowedFormats.map(v=>'.'+v).join(',');
      const hint=form.querySelector('[data-document-format-hint]');
      if(hint) hint.textContent=allowedFormats.map(v=>v.toUpperCase()).join(', ');
    }
    let config=[];
    try{
      const raw=opt?.dataset?.configB64||'';
      if(raw){
        const binary=atob(raw);
        const bytes=Uint8Array.from(binary,ch=>ch.charCodeAt(0));
        config=JSON.parse(new TextDecoder().decode(bytes));
      }
    }catch(e){config=[]}
    renderCustomFields(config);
    const map=new Map(config.map(x=>[x.key,x]));
    fields.forEach(el=>{
      const c=map.get(el.dataset.field); const show=!!(c&&c.enabled);
      applyFieldLabel(el,c);
      el.hidden=!show; el.style.order=c?.order||999;
      el.querySelectorAll('[data-required-target]').forEach(i=>{
        const hasExisting=i.type==='file' && !!(i.dataset.existing||'').trim();
        i.required=show&&!!c.required&&!hasExisting;
        if(!show) i.setCustomValidity('');
      });
      const mark=el.querySelector('.required-mark'); if(mark)mark.textContent=show&&c?.required?' *':'';
    });
    const pos=map.get('positions'); list?.querySelectorAll('input[name="position_name[]"]').forEach(i=>i.required=!!(pos?.enabled&&pos?.required));
    syncFilterSuggestions(config);
  }
  function syncPicker(){
    const opt=select?.options[select.selectedIndex];
    const value=select?.value||'';
    if(selectedText)selectedText.textContent=value?(opt?.textContent?.trim()||'Chọn một chuyên mục'):'Chọn một chuyên mục';
    optionButtons.forEach(x=>x.classList.toggle('selected',x.dataset.value===value));
  }
  function closeCategoryMenu(){if(menu)menu.hidden=true;if(trigger){trigger.classList.remove('open');trigger.setAttribute('aria-expanded','false')}}
  function openCategoryMenu(){if(menu)menu.hidden=false;if(trigger){trigger.classList.add('open');trigger.setAttribute('aria-expanded','true')}}
  trigger?.addEventListener('click',()=>{menu?.hidden?openCategoryMenu():closeCategoryMenu()});
  optionButtons.forEach(btn=>btn.addEventListener('click',()=>{select.value=btn.dataset.value||'';select.dispatchEvent(new Event('change',{bubbles:true}));closeCategoryMenu()}));
  document.addEventListener('click',e=>{if(picker&&!picker.contains(e.target))closeCategoryMenu()});
  document.addEventListener('keydown',e=>{if(e.key==='Escape'){closeCategoryMenu();closePreview();closeDocumentPreview()}});
  select?.addEventListener('change',()=>{applyConfig();syncPicker()});
  applyConfig(); syncPicker();

  function renumber(){
    [...list.children].forEach((c,i)=>{
      const index=i+1;
      const n=c.querySelector('.position-number');
      if(n)n.textContent=index;
      const labelInput=c.querySelector('input[name="position_label[]"]');
      if(labelInput){
        const fallback=`Vị trí tuyển dụng số ${index}`;
        labelInput.placeholder=fallback;
        if(!labelInput.value.trim()) labelInput.dataset.fallback=fallback;
      }
      const rm=c.querySelector('.remove-position');
      if(rm)rm.style.display=list.children.length===1?'none':'inline-flex'
    });
    applyConfig()
  }
  document.getElementById('addPosition')?.addEventListener('click',()=>{list.appendChild(tpl.content.firstElementChild.cloneNode(true));renumber()});
  list?.addEventListener('click',e=>{const b=e.target.closest('.remove-position');if(!b)return;if(list.children.length>1){b.closest('.position-card').remove();renumber()}});renumber();

  document.querySelectorAll('.editor-toolbar button').forEach(b=>b.addEventListener('click',()=>{const box=b.closest('.editor-shell,.mini-editor');box?.querySelector('textarea')?.focus()}));

  function field(name){return form?.querySelector(`[name="${name}"]`)?.value?.trim()||''}
  function openPreview(){
    const title=field('title')||'Tiêu đề bài viết';
    const content=field('content')||'Nội dung bài viết sẽ hiển thị tại đây.';
    const cat=select?.value?(select.options[select.selectedIndex]?.textContent?.trim()||'Chuyên mục'):'Chưa chọn chuyên mục';
    document.getElementById('previewPostTitle').textContent=title;
    document.getElementById('previewContent').textContent=content;
    document.getElementById('previewCategory').textContent=cat;
    const extra=[];
    const deadline=field('deadline'); if(deadline)extra.push(`<div><b>Hạn cuối</b><span>${deadline}</span></div>`);
    const website=field('website'); if(website)extra.push(`<div><b>Website</b><span>${website}</span></div>`);
    const fanpage=field('fanpage'); if(fanpage)extra.push(`<div><b>Fanpage</b><span>${fanpage}</span></div>`);
    const cv=field('cv_email'); if(cv)extra.push(`<div><b>Email nhận CV</b><span>${cv}</span></div>`);
    [['organization','Đơn vị / tổ chức'],['contact_name','Người liên hệ'],['contact_phone','Số điện thoại'],['location','Địa điểm'],['salary_range','Mức lương / hỗ trợ'],['application_link','Link đăng ký'],['event_time','Thời gian'],['audience','Đối tượng'],['tags','Từ khóa'],['source','Nguồn'],['school','Trường học'],['document_type','Loại tài liệu'],['subject','Môn học'],['academic_year','Niên khóa']].forEach(([key,label])=>{const value=field(key);if(value)extra.push(`<div><b>${label}</b><span>${value}</span></div>`)});
    form?.querySelectorAll('[data-custom-field="1"]').forEach(block=>{const control=block.querySelector('[name]');const value=control?.value?.trim()||'';const label=block.querySelector('label')?.childNodes?.[0]?.textContent?.trim()||'Thông tin';if(value)extra.push(`<div><b>${escapePreviewText(label)}</b><span>${escapePreviewText(value)}</span></div>`)});
    document.getElementById('previewExtra').innerHTML=extra.join('');
    const imageBox=document.getElementById('previewImage'); const file=form?.querySelector('[name="image"]')?.files?.[0];
    const existing=imageBox?.dataset.existingImage||'';
    if(file){const reader=new FileReader();reader.onload=()=>{imageBox.innerHTML=`<img src="${reader.result}" alt="Ảnh xem trước">`};reader.readAsDataURL(file)}
    else if(existing){imageBox.innerHTML=`<img src="${existing}" alt="Ảnh đã lưu">`}
    else imageBox.innerHTML='<span>Ảnh bài viết sẽ hiển thị tại đây</span>';
    previewModal.hidden=false; document.body.classList.add('modal-open');
  }
  function closePreview(){if(previewModal){previewModal.hidden=true;document.body.classList.remove('modal-open')}}
  document.getElementById('previewPostBtn')?.addEventListener('click',openPreview);
  previewModal?.querySelectorAll('[data-close-preview]').forEach(x=>x.addEventListener('click',closePreview));

  function escapePreviewText(value){return String(value??'').replace(/[&<>"']/g,ch=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch]));}
  function closeDocumentPreview(){
    if(documentPreviewModal) documentPreviewModal.hidden=true;
    if(documentObjectURL){URL.revokeObjectURL(documentObjectURL);documentObjectURL='';}
    document.body.classList.remove('modal-open');
  }
  function syncDocumentPreviewButton(){
    const file=documentInput?.files?.[0];
    if(documentPreviewBtn) documentPreviewBtn.hidden=!file;
    if(documentFileName) documentFileName.textContent=file?file.name:'';
  }
  async function openDocumentPreview(){
    const file=documentInput?.files?.[0];
    if(!file||!documentPreviewModal||!documentPreviewStage)return;
    const ext=(file.name.includes('.')?file.name.split('.').pop():'').toLowerCase();
    if(documentPreviewName)documentPreviewName.textContent=file.name;
    if(documentPreviewType)documentPreviewType.textContent=(ext||'FILE').toUpperCase();
    documentPreviewModal.hidden=false;document.body.classList.add('modal-open');
    documentPreviewStage.innerHTML='<div class="document-preview-loading">Đang tạo bản xem trước…</div>';
    if(documentObjectURL){URL.revokeObjectURL(documentObjectURL);documentObjectURL='';}
    if(ext==='pdf'){
      documentObjectURL=URL.createObjectURL(file);
      documentPreviewStage.innerHTML=`<object data="${documentObjectURL}" type="application/pdf"><div class="document-preview-error">Trình duyệt không hỗ trợ xem PDF trực tiếp.</div></object>`;
      return;
    }
    try{
      const data=new FormData();data.append('document_file',file);
      const response=await fetch('/tai-lieu/xem-truoc-tam',{method:'POST',body:data,headers:{'X-Requested-With':'XMLHttpRequest','Accept':'application/json'}});
      const result=await response.json().catch(()=>({}));
      if(!response.ok||!result.ok)throw new Error(result.message||'Không thể tạo bản xem trước');
      if(result.kind==='pdf'&&result.preview_url){
        documentPreviewStage.innerHTML=`<object data="${encodeURI(result.preview_url)}" type="application/pdf"><div class="document-preview-error">Trình duyệt không hỗ trợ xem PDF trực tiếp.</div></object>`;
      }else{
        const type=escapePreviewText((result.ext||ext||'FILE').toUpperCase());
        const name=escapePreviewText(result.name||file.name||'Tài liệu');
        const message=escapePreviewText(result.message||'Hiển thị bìa minh họa tài liệu.');
        documentPreviewStage.innerHTML=`<div class="document-preview-cover"><div class="document-preview-cover-sheet"><span>${type}</span><b>📄</b></div><strong>${name}</strong><small>${message}</small></div>`;
      }
    }catch(error){
      documentPreviewStage.innerHTML=`<div class="document-preview-error">${escapePreviewText(error.message||'Không thể xem trước tài liệu.')}</div>`;
    }
  }
  documentInput?.addEventListener('change',syncDocumentPreviewButton);
  documentPreviewBtn?.addEventListener('click',openDocumentPreview);
  documentPreviewModal?.querySelectorAll('[data-close-document-preview]').forEach(x=>x.addEventListener('click',closeDocumentPreview));
  syncDocumentPreviewButton();

  form?.querySelectorAll('[data-required-target]').forEach(input=>{
    input.addEventListener('invalid',()=>{
      const block=input.closest('[data-field]');
      if(block?.hidden){ input.required=false; input.setCustomValidity(''); return; }
      block?.classList.add('field-invalid');
      setTimeout(()=>block?.classList.remove('field-invalid'),1800);
    });
    input.addEventListener('input',()=>input.closest('[data-field]')?.classList.remove('field-invalid'));
    input.addEventListener('change',()=>input.closest('[data-field]')?.classList.remove('field-invalid'));
  });

  form?.addEventListener('submit',e=>{
    const action=e.submitter?.value||'publish';
    if(action!=='save_draft'&&!select.value){e.preventDefault();openCategoryMenu();trigger?.focus();return}
    const b=e.submitter; if(b){b.disabled=true;b.dataset.original=b.textContent;b.textContent=action==='save_draft'?'ĐANG LƯU...':'ĐANG ĐĂNG...'}
  });
});


// DACS_V1.22.2 · Validate document extension against selected category.
document.addEventListener('change',(event)=>{
  const input=event.target.closest?.('[data-document-input]');
  if(!input||!input.files?.[0]) return;
  const select=document.getElementById('categorySelect');
  const opt=select?.options[select.selectedIndex];
  const allowed=String(opt?.dataset?.documentFormats||'').split(',').map(v=>v.trim().toLowerCase()).filter(Boolean);
  const name=input.files[0].name||'';
  const ext=(name.includes('.')?name.split('.').pop():'').toLowerCase();
  input.setCustomValidity(allowed.length&& !allowed.includes(ext)?`Chuyên mục này không cho phép tệp .${ext}`:'');
  if(input.validationMessage) input.reportValidity();
});
