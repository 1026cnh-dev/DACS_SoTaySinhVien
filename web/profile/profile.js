document.addEventListener('DOMContentLoaded',()=>{
  const input=document.getElementById('avatarInput');
  const preview=document.getElementById('avatarPreview');
  if(input&&preview){
    input.addEventListener('change',()=>{
      const file=input.files?.[0];
      if(!file)return;
      if(file.size>8*1024*1024){alert('Ảnh đại diện tối đa 8 MB.');input.value='';return;}
      if(!['image/jpeg','image/png','image/webp'].includes(file.type)){alert('Chỉ hỗ trợ JPG, PNG hoặc WEBP.');input.value='';return;}
      const disclosure=document.querySelector('.profile-edit-disclosure');
      if(disclosure)disclosure.open=true;
      const url=URL.createObjectURL(file);
      preview.innerHTML='';
      const img=document.createElement('img');
      img.src=url;img.alt='Xem trước ảnh đại diện';
      preview.appendChild(img);
    });
  }


  // DACS V1.6 · autocomplete trường gọn: alias/mã chỉ dùng tìm kiếm, mỗi trường chỉ render một lựa chọn.
  const profileForm=document.getElementById('profileForm');
  const schoolInput=document.getElementById('schoolInput');
  const majorInput=document.getElementById('majorInput');
  const schoolSuggestions=document.getElementById('schoolSuggestions');
  const majorOptions=document.getElementById('majorOptions');
  const educationState={loaded:false,schools:[],majors:[],schoolMap:new Map(),majorMap:new Map(),visibleSchools:[],activeSchoolIndex:-1};

  function educationKey(value){
    let out=String(value||'').trim().toLowerCase();
    try{out=out.normalize('NFD').replace(/[\u0300-\u036f]/g,'');}catch(_){/* old browser fallback */}
    out=out.replace(/đ/g,'d');
    out=out.replace(/\btp\.?\s*ho\s*chi\s*minh\b/g,'tp hcm').replace(/\bthanh\s*pho\s*ho\s*chi\s*minh\b/g,'tp hcm');
    out=out.replace(/^\s*dh\s+/,'dai hoc ').replace(/^\s*hv\s+/,'hoc vien ').replace(/^\s*cd\s+/,'cao dang ');
    return out.replace(/[^a-z0-9]+/g,' ').trim().replace(/\s+/g,' ');
  }
  function compactSchoolKey(value){
    let key=educationKey(value);
    for(const prefix of ['truong dai hoc ','dai hoc ','hoc vien ','truong cao dang ','cao dang ','truong ']){
      if(key.startsWith(prefix)){key=key.slice(prefix.length).trim();break;}
    }
    return key;
  }
  function registerSchoolKey(raw,school){
    for(const key of [educationKey(raw),compactSchoolKey(raw)]){
      if(!key)continue;
      if(!educationState.schoolMap.has(key))educationState.schoolMap.set(key,school);
      else if(educationState.schoolMap.get(key)?.name!==school.name)educationState.schoolMap.set(key,null);
    }
  }
  function resolveSchool(value){
    const full=educationState.schoolMap.get(educationKey(value));
    if(full)return full;
    const compact=educationState.schoolMap.get(compactSchoolKey(value));
    return compact||null;
  }
  function schoolDisplayName(name){
    return String(name||'').replace(/\s+-\s+Đại học Quốc gia\s+/gi,' - ĐHQG ');
  }
  function schoolTerms(school){
    return [school.name,school.code,...(school.aliases||[])].filter(Boolean).flatMap(value=>{
      const full=educationKey(value);
      const compact=compactSchoolKey(value);
      return full===compact?[full]:[full,compact];
    }).filter(Boolean);
  }
  function schoolMatchScore(school,query){
    const q=educationKey(query);
    const cq=compactSchoolKey(query);
    if(!q)return 999;
    let best=999;
    for(const term of schoolTerms(school)){
      if(term===q||term===cq)best=Math.min(best,0);
      else if(term.startsWith(q)||term.startsWith(cq))best=Math.min(best,1);
      else if(term.includes(q)||term.includes(cq))best=Math.min(best,2);
    }
    return best;
  }
  function closeSchoolSuggestions(){
    if(!schoolSuggestions)return;
    schoolSuggestions.hidden=true;
    schoolSuggestions.replaceChildren();
    educationState.visibleSchools=[];
    educationState.activeSchoolIndex=-1;
    schoolInput?.setAttribute('aria-expanded','false');
    schoolInput?.removeAttribute('aria-activedescendant');
  }
  function setActiveSchool(index){
    if(!schoolSuggestions||!educationState.visibleSchools.length)return;
    const count=educationState.visibleSchools.length;
    educationState.activeSchoolIndex=(index+count)%count;
    [...schoolSuggestions.querySelectorAll('.education-suggestion')].forEach((item,i)=>{
      const active=i===educationState.activeSchoolIndex;
      item.classList.toggle('is-active',active);
      item.setAttribute('aria-selected',active?'true':'false');
      if(active){
        schoolInput?.setAttribute('aria-activedescendant',item.id);
        item.scrollIntoView({block:'nearest'});
      }
    });
  }
  function chooseSchool(school){
    if(!schoolInput||!school)return;
    schoolInput.value=school.name;
    schoolInput.setCustomValidity('');
    closeSchoolSuggestions();
  }
  function renderSchoolSuggestions(){
    if(!schoolInput||!schoolSuggestions||!educationState.loaded)return;
    const raw=schoolInput.value.trim();
    if(!raw){closeSchoolSuggestions();return;}
    const q=educationKey(raw);
    const ranked=educationState.schools
      .map(school=>({school,score:schoolMatchScore(school,q)}))
      .filter(item=>item.score<999)
      .sort((a,b)=>a.score-b.score||a.school.name.localeCompare(b.school.name,'vi'));
    const bestScore=ranked[0]?.score;
    const matches=ranked
      .filter(item=>bestScore===0?item.score===0:true)
      .slice(0,8)
      .map(item=>item.school);
    if(!matches.length){closeSchoolSuggestions();return;}

    educationState.visibleSchools=matches;
    educationState.activeSchoolIndex=-1;
    const frag=document.createDocumentFragment();
    matches.forEach((school,index)=>{
      const item=document.createElement('span');
      item.className='education-suggestion';
      item.id=`schoolSuggestion${index}`;
      item.setAttribute('role','option');
      item.setAttribute('aria-selected','false');
      item.textContent=schoolDisplayName(school.name);
      item.addEventListener('pointerdown',event=>{
        event.preventDefault();
        chooseSchool(school);
        schoolInput.focus();
      });
      frag.appendChild(item);
    });
    schoolSuggestions.replaceChildren(frag);
    schoolSuggestions.hidden=false;
    schoolInput.setAttribute('aria-expanded','true');
  }
  function canonicalizeSchool(showError=false){
    if(!schoolInput||!educationState.loaded)return true;
    const raw=schoolInput.value.trim();
    if(!raw){schoolInput.setCustomValidity('');closeSchoolSuggestions();return true;}
    const school=resolveSchool(raw);
    if(school){chooseSchool(school);return true;}
    if(showError)schoolInput.setCustomValidity('Vui lòng chọn trường từ danh sách gợi ý.');
    else schoolInput.setCustomValidity('');
    return false;
  }

  async function loadEducationOptions(){
    if(!schoolInput&&!majorInput)return;
    try{
      const response=await fetch('/api/education-options',{headers:{Accept:'application/json'}});
      if(!response.ok)throw new Error('Không tải được danh mục giáo dục');
      const data=await response.json();
      educationState.schools=Array.isArray(data.schools)?data.schools:[];
      educationState.majors=Array.isArray(data.majors)?data.majors:[];
      educationState.schoolMap.clear();
      educationState.majorMap.clear();

      educationState.schools.forEach(school=>{
        registerSchoolKey(school.name,school);
        if(school.code)registerSchoolKey(school.code,school);
        (school.aliases||[]).forEach(alias=>registerSchoolKey(alias,school));
      });
      if(majorOptions){
        const frag=document.createDocumentFragment();
        educationState.majors.forEach(major=>{
          educationState.majorMap.set(educationKey(major),major);
          const option=document.createElement('option');
          option.value=major;
          frag.appendChild(option);
        });
        majorOptions.replaceChildren(frag);
      }
      educationState.loaded=true;
      canonicalizeSchool(false);
      if(majorInput){
        const canonical=educationState.majorMap.get(educationKey(majorInput.value));
        if(canonical)majorInput.value=canonical;
      }
    }catch(err){
      console.warn(err);
      // Không khóa form nếu API nội bộ tạm lỗi; backend vẫn kiểm tra dữ liệu khi lưu.
    }
  }

  schoolInput?.addEventListener('input',()=>{
    schoolInput.setCustomValidity('');
    renderSchoolSuggestions();
  });
  schoolInput?.addEventListener('focus',renderSchoolSuggestions);
  schoolInput?.addEventListener('keydown',event=>{
    if(event.key==='ArrowDown'){
      if(schoolSuggestions?.hidden)renderSchoolSuggestions();
      if(educationState.visibleSchools.length){event.preventDefault();setActiveSchool(educationState.activeSchoolIndex+1);}
    }else if(event.key==='ArrowUp'&&educationState.visibleSchools.length){
      event.preventDefault();setActiveSchool(educationState.activeSchoolIndex-1);
    }else if(event.key==='Enter'&&educationState.activeSchoolIndex>=0){
      event.preventDefault();chooseSchool(educationState.visibleSchools[educationState.activeSchoolIndex]);
    }else if(event.key==='Escape'){
      closeSchoolSuggestions();
    }
  });
  schoolInput?.addEventListener('blur',()=>{
    setTimeout(()=>{canonicalizeSchool(false);closeSchoolSuggestions();},0);
  });
  document.addEventListener('pointerdown',event=>{
    if(!event.target.closest('.education-autocomplete-shell'))closeSchoolSuggestions();
  });
  majorInput?.addEventListener('change',()=>{
    const canonical=educationState.majorMap.get(educationKey(majorInput.value));
    if(canonical)majorInput.value=canonical;
  });
  profileForm?.addEventListener('submit',(event)=>{
    if(educationState.loaded&&!canonicalizeSchool(true)){
      event.preventDefault();
      closeSchoolSuggestions();
      schoolInput?.reportValidity();
      schoolInput?.focus();
    }
  });
  loadEducationOptions();

  const profilePhoneInput=document.getElementById('profilePhoneInput');
  const phoneVerifyButton=document.getElementById('phoneVerifyButton');
  const phoneVerifyState=document.getElementById('phoneVerifyState');
  function phoneDigits(value){return String(value||'').replace(/\D/g,'');}
  function syncPhoneVerifyButton(){
    if(!profilePhoneInput||!phoneVerifyButton)return;
    const digits=phoneDigits(profilePhoneInput.value);
    const hasUsablePhone=digits.length>=9&&digits.length<=12;
    phoneVerifyButton.classList.toggle('is-hidden',!hasUsablePhone);
    if(phoneVerifyState){
      phoneVerifyState.classList.toggle('ready',hasUsablePhone);
      phoneVerifyState.textContent=hasUsablePhone?'Có thể xác thực':'Chưa xác thực';
    }
  }
  profilePhoneInput?.addEventListener('input',syncPhoneVerifyButton);
  syncPhoneVerifyButton();

  const verificationForm=document.querySelector('.verification-form');
  const typeInputs=verificationForm?[...verificationForm.querySelectorAll('input[name="type"]')]:[];
  const profileRoleInput=document.getElementById('profileRoleInput');
  const rolePanels=[...document.querySelectorAll('[data-profile-role-panel]')];
  const previewBox=document.querySelector('[data-role-preview]');
  const previewIcon=previewBox?.querySelector('img');
  const previewTitle=previewBox?.querySelector('[data-role-preview-title]');
  const previewText=previewBox?.querySelector('[data-role-preview-text]');

  const roleUI={
    student:{title:'Thông tin Sinh viên',text:'Trường/Đại học · Ngành học · Mã số sinh viên được điền ở cột bên trái.',icon:'/web/icons/graduation-cap.svg'},
    employer:{title:'Thông tin Nhà tuyển dụng',text:'Công ty · Mã số thuế · Người đại diện · Website/Fanpage được điền ở cột bên trái.',icon:'/web/icons/briefcase-business.svg'},
    landlord:{title:'Thông tin Chủ trọ',text:'Tên cơ sở · Địa chỉ khu trọ · Số điện thoại · Thông tin giấy tờ được điền ở cột bên trái.',icon:'/web/icons/house.svg'}
  };

  function selectedRole(){
    return verificationForm?.querySelector('input[name="type"]:checked')?.value || profileRoleInput?.value || 'student';
  }

  function syncRoleUI(){
    const role=selectedRole();
    if(profileRoleInput)profileRoleInput.value=role;
    rolePanels.forEach(panel=>panel.classList.toggle('is-hidden',panel.dataset.profileRolePanel!==role));
    const ui=roleUI[role]||roleUI.student;
    if(previewIcon)previewIcon.src=ui.icon;
    if(previewTitle)previewTitle.textContent=ui.title;
    if(previewText)previewText.textContent=ui.text;
  }

  typeInputs.forEach(radio=>radio.addEventListener('change',()=>{
    syncRoleUI();
    const disclosure=document.querySelector('.profile-edit-disclosure');
    if(disclosure)disclosure.open=true;
  }));
  syncRoleUI();
});
