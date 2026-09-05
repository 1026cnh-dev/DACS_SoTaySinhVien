const fieldDefs = [
  ['title', 'Tiêu đề', 10],
  ['content', 'Nội dung', 20],
  ['image', 'Ảnh bài viết', 30],
  ['deadline', 'Hạn cuối', 40],
  ['company_logo', 'Logo công ty', 50],
  ['website', 'Website', 60],
  ['fanpage', 'Fanpage', 70],
  ['recruitment_content', 'Nội dung tuyển dụng', 80],
  ['cv_email', 'Email nhận CV', 90],
  ['positions', 'Các vị trí tuyển', 100],
  ['contact_name', 'Người liên hệ', 110],
  ['contact_phone', 'Số điện thoại liên hệ', 120],
  ['organization', 'Đơn vị / tổ chức', 130],
  ['location', 'Địa điểm', 140],
  ['salary_range', 'Mức lương / hỗ trợ', 150],
  ['application_link', 'Liên kết đăng ký / ứng tuyển', 160],
  ['event_time', 'Thời gian diễn ra', 170],
  ['audience', 'Đối tượng tham gia', 180],
  ['tags', 'Từ khóa / thẻ', 190],
  ['source', 'Nguồn thông tin', 200],
  ['document_file', 'File tài liệu học tập', 210],
  ['school', 'Trường học', 220],
  ['document_type', 'Loại tài liệu', 230],
  ['subject', 'Môn học', 240],
  ['academic_year', 'Niên khóa', 250]
];

const builtInFieldKeys = new Set(fieldDefs.map(([key]) => key));
const customFieldTypeMeta = {
  text:{icon:'TXT',type:'Văn bản ngắn'}, textarea:{icon:'¶',type:'Nội dung dài'}, number:{icon:'123',type:'Số'},
  email:{icon:'@',type:'Email'}, url:{icon:'URL',type:'Liên kết'}, date:{icon:'D',type:'Ngày'}
};

const fieldGroups = {
  main: { label: 'Nội dung chính', short: 'Chính', desc: 'Tiêu đề, nội dung và ảnh bài viết.' },
  schedule: { label: 'Thời gian & địa điểm', short: 'Thời gian', desc: 'Hạn cuối, thời gian diễn ra và địa điểm.' },
  contact: { label: 'Liên hệ & tổ chức', short: 'Liên hệ', desc: 'Thông tin tổ chức, liên hệ và các đường dẫn.' },
  job: { label: 'Việc làm & tuyển dụng', short: 'Việc làm', desc: 'Các trường dành cho tin tuyển dụng và vị trí việc làm.' },
  classify: { label: 'Phân loại & tìm kiếm', short: 'Phân loại', desc: 'Đối tượng, từ khóa và nguồn thông tin.' },
  document: { label: 'Tài liệu học tập', short: 'Tài liệu', desc: 'Tệp tài liệu, môn học, loại tài liệu và niên khóa.' },
  custom: { label: 'Trường tự tạo', short: 'Tự tạo', desc: 'Các trường được quản trị viên thêm riêng cho chuyên mục.' }
};

const fieldGroupMap = {
  title:'main', content:'main', image:'main',
  deadline:'schedule', event_time:'schedule', location:'schedule',
  organization:'contact', contact_name:'contact', contact_phone:'contact', website:'contact', fanpage:'contact', application_link:'contact', cv_email:'contact',
  company_logo:'job', recruitment_content:'job', positions:'job', salary_range:'job',
  audience:'classify', tags:'classify', source:'classify',
  document_file:'document', school:'document', document_type:'document', subject:'document', academic_year:'document'
};


// V1.24.27 · mỗi field là một module độc lập. Profile dựa trên slug để
// giao diện Form Builder mang đúng ngữ cảnh của từng chuyên mục mà không
// thay đổi mã kỹ thuật hay cấu trúc dữ liệu đã lưu.
const fieldModuleMeta = {
  title:{icon:'T',type:'Văn bản ngắn'}, content:{icon:'¶',type:'Nội dung dài'}, image:{icon:'IMG',type:'Hình ảnh'},
  deadline:{icon:'D',type:'Ngày hạn'}, company_logo:{icon:'LG',type:'Hình ảnh'}, website:{icon:'WWW',type:'Liên kết'},
  fanpage:{icon:'FB',type:'Liên kết'}, recruitment_content:{icon:'JOB',type:'Nội dung dài'}, cv_email:{icon:'@',type:'Email'},
  positions:{icon:'POS',type:'Danh sách lặp'}, contact_name:{icon:'CN',type:'Văn bản ngắn'}, contact_phone:{icon:'TEL',type:'Số điện thoại'},
  organization:{icon:'ORG',type:'Văn bản ngắn'}, location:{icon:'LOC',type:'Địa điểm'}, salary_range:{icon:'₫',type:'Mức giá'},
  application_link:{icon:'URL',type:'Liên kết'}, event_time:{icon:'TIME',type:'Ngày & giờ'}, audience:{icon:'AUD',type:'Phân loại'},
  tags:{icon:'TAG',type:'Từ khóa'}, source:{icon:'SRC',type:'Nguồn'}, document_file:{icon:'DOC',type:'Tệp tài liệu'},
  school:{icon:'SCH',type:'Danh mục'}, document_type:{icon:'TYPE',type:'Danh mục'}, subject:{icon:'SUB',type:'Danh mục'}, academic_year:{icon:'YEAR',type:'Danh mục'}
};

const categoryModuleProfiles = {
  'hoc-tap': {
    label:'Học tập & tài liệu', tone:'document',
    recommended:['title','content','image','document_file','school','document_type','subject','academic_year','tags','source'],
    groupOrder:['main','document','classify','contact','schedule','job']
  },
  'hoc-bong': {
    label:'Học bổng', tone:'scholarship',
    recommended:['title','content','image','deadline','organization','website','application_link','contact_name','contact_phone','audience','tags','source'],
    groupOrder:['main','schedule','contact','classify','document','job']
  },
  'su-kien': {
    label:'Sự kiện', tone:'event',
    recommended:['title','content','image','event_time','location','organization','contact_name','contact_phone','application_link','audience','tags','source'],
    groupOrder:['main','schedule','contact','classify','job','document']
  },
  'ky-nang': {
    label:'Kỹ năng', tone:'skill',
    recommended:['title','content','image','event_time','location','organization','application_link','audience','tags','source'],
    groupOrder:['main','classify','schedule','contact','job','document']
  },
  'viec-lam': {
    label:'Việc làm & tuyển dụng', tone:'job',
    recommended:['title','content','image','deadline','company_logo','website','fanpage','recruitment_content','cv_email','positions','contact_name','contact_phone','organization','location','salary_range','application_link','tags'],
    groupOrder:['main','job','contact','schedule','classify','document']
  },
  'thong-bao': {
    label:'Thông báo', tone:'notice',
    recommended:['title','content','image','deadline','organization','source','tags'],
    groupOrder:['main','schedule','classify','contact','document','job']
  },
  'confession': {
    label:'Confession', tone:'confession',
    recommended:['title','content','image','tags'],
    groupOrder:['main','classify','schedule','contact','job','document']
  }
};

function categoryModuleProfile(slug='') {
  return categoryModuleProfiles[String(slug || '').trim().toLowerCase()] || {
    label:'Chuyên mục tùy chỉnh', tone:'default',
    recommended:['title','content','image'],
    groupOrder:['main','schedule','contact','job','classify','document']
  };
}

const positionLabels = {
  'home-right': 'Trang chủ · banner quảng cáo lớn bên phải',
  'feed-sponsored': 'Feed · xen kẽ Ghim/Xu hướng và cột 1 chuyên mục',
  'post-left': 'Trang bài viết · quảng cáo bên trái',
  'post-right': 'Trang bài viết · quảng cáo bên phải'
};

const adminToast = document.getElementById('adminToast');
let toastTimer;
function showToast(message, error = false) {
  if (!adminToast) return;
  adminToast.classList.toggle('error', error);
  adminToast.querySelector('b').textContent = error ? 'Không thể lưu' : 'Đã lưu';
  adminToast.querySelector('small').textContent = message || '';
  adminToast.classList.add('show');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => adminToast.classList.remove('show'), 2200);
}

async function postFormJSON(form) {
  const response = await fetch(form.action, {
    method: 'POST',
    body: new FormData(form),
    headers: { 'X-Requested-With': 'XMLHttpRequest', 'Accept': 'application/json' }
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok || !data.ok) throw new Error(data.message || 'Không thể lưu');
  return data;
}

function switchTab(tabName) {
  document.querySelectorAll('.side-link').forEach((btn) => {
    btn.classList.toggle('active', btn.dataset.tab === tabName);
  });
  document.querySelectorAll('.tab-panel').forEach((panel) => {
    panel.classList.toggle('active', panel.id === tabName);
  });
}

document.querySelectorAll('.side-link').forEach((btn) => {
  btn.addEventListener('click', () => switchTab(btn.dataset.tab));
});

const requestedAdminTab = new URLSearchParams(location.search).get('tab');
if (requestedAdminTab && document.getElementById(requestedAdminTab)) switchTab(requestedAdminTab);


// DACS_V1.21.6 · compact create-category popup
const newCategoryModal = document.getElementById('newCategoryModal');
const openNewCategoryModalButton = document.getElementById('openNewCategoryModal');
const closeNewCategoryModalButton = document.getElementById('closeNewCategoryModal');

function openNewCategoryModal() {
  if (!newCategoryModal) return;
  newCategoryModal.classList.add('open');
  document.body.classList.add('modal-lock');
  const body = newCategoryModal.querySelector('.category-create-modal-body');
  if (body) body.scrollTop = 0;
  window.setTimeout(() => newCategoryModal.querySelector('input[name="name"]')?.focus(), 80);
}

function closeNewCategoryModal() {
  if (!newCategoryModal) return;
  newCategoryModal.classList.remove('open');
  syncAdminModalLock();
}

openNewCategoryModalButton?.addEventListener('click', openNewCategoryModal);
closeNewCategoryModalButton?.addEventListener('click', closeNewCategoryModal);

document.querySelectorAll('[data-jump-tab]').forEach((button) => {
  button.addEventListener('click', () => {
    const tab = button.dataset.jumpTab;
    if (!tab || !document.getElementById(tab)) return;
    switchTab(tab);
    document.querySelector('.admin-main')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  });
});


document.querySelectorAll('.ajax-pin-form').forEach((form) => {
  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    const button = form.querySelector('button');
    const hidden = form.querySelector('input[name="pinned"]');
    try {
      const data = await postFormJSON(form);
      hidden.value = data.pinned ? '0' : '1';
      button.textContent = data.pinned ? 'Bỏ ghim' : '📌 Ghim';
      const statusCell = form.closest('tr')?.children[2];
      if (statusCell) statusCell.innerHTML = data.pinned ? '<span class="tag pin-tag">📌 Đã ghim</span>' : '<span class="muted-status">Bình thường</span>';
      showToast(data.message);
    } catch (error) { showToast(error.message, true); }
  });
});

function openPostForm() {
  document.getElementById('postModal')?.classList.add('open');
}
function closePostForm() {
  document.getElementById('postModal')?.classList.remove('open');
}
window.closePostForm = closePostForm;

document.querySelectorAll('.edit-post').forEach((button) => {
  button.addEventListener('click', () => {
    document.getElementById('postId').value = button.dataset.id;
    document.getElementById('postTitle').value = button.dataset.title;
    document.getElementById('postSummary').value = button.dataset.summary;
    document.getElementById('postContent').value = button.dataset.content;
    document.getElementById('postCategory').value = button.dataset.category;
    openPostForm();
  });
});

function closeConfigForm() {
  document.getElementById('configModal')?.classList.remove('open');
}
window.closeConfigForm = closeConfigForm;

const fieldHintMap = {
  title: 'Tiêu đề chính của bài viết.',
  content: 'Nội dung mô tả chi tiết.',
  image: 'Ảnh đại diện hiển thị trên thẻ bài viết.',
  deadline: 'Ngày hết hạn hoặc hạn nộp.',
  company_logo: 'Logo tổ chức hoặc đơn vị đăng bài.',
  website: 'Liên kết website chính thức.',
  fanpage: 'Liên kết fanpage hoặc mạng xã hội.',
  recruitment_content: 'Nội dung tuyển dụng hoặc mô tả yêu cầu.',
  cv_email: 'Email nhận hồ sơ hoặc phản hồi.',
  positions: 'Danh sách vị trí cần tuyển hoặc vai trò.',
  contact_name: 'Tên người liên hệ.',
  contact_phone: 'Số điện thoại liên hệ.',
  organization: 'Tên đơn vị, câu lạc bộ hoặc tổ chức.',
  location: 'Địa điểm diễn ra hoặc làm việc.',
  salary_range: 'Mức hỗ trợ hoặc lương dự kiến.',
  application_link: 'Liên kết biểu mẫu hoặc trang đăng ký.',
  audience: 'Nhóm đối tượng phù hợp.',
  tags: 'Từ khóa để tìm kiếm và phân loại.',
  source: 'Nguồn tham khảo hoặc trích dẫn.',
  event_time: 'Ngày giờ diễn ra sự kiện.',
  document_file: 'Tệp tài liệu đính kèm.',
  school: 'Tên trường học liên quan.',
  document_type: 'Ví dụ: giáo trình, đề cương, bài giảng.',
  subject: 'Tên môn học hoặc học phần.',
  academic_year: 'Niên khóa hoặc năm học áp dụng.'
};

function fieldRow(key, label, order, enabled, required, filterable=false, suggestValues=false, allowCustom=false, categorySlug='', fieldType='', custom=false) {
  const isCustom = custom || !builtInFieldKeys.has(key);
  const normalizedType = isCustom ? (fieldType || 'text') : '';
  const hint = isCustom ? 'Trường tùy chỉnh của riêng chuyên mục này. Có thể đổi tên, sắp xếp hoặc xóa.' : (fieldHintMap[key] || 'Thiết lập cách trường này xuất hiện trên form đăng bài.');
  const groupKey = isCustom ? 'custom' : (fieldGroupMap[key] || 'main');
  const group = fieldGroups[groupKey] || fieldGroups.main;
  const moduleMeta = isCustom ? (customFieldTypeMeta[normalizedType] || customFieldTypeMeta.text) : (fieldModuleMeta[key] || {icon:'MOD',type:'Trường dữ liệu'});
  const profile = categoryModuleProfile(categorySlug);
  const recommended = !isCustom && new Set(profile.recommended || []).has(key);
  const safeKey = escapeHTML(key);
  const safeLabel = escapeHTML(label || key);
  const safeHint = escapeHTML(hint);
  const stateClass = enabled ? 'module-enabled' : 'module-disabled';
  const customHidden = isCustom ? `<input type="hidden" name="custom_field_key" value="${safeKey}"><input type="hidden" name="type_${safeKey}" value="${escapeHTML(normalizedType)}">` : '';
  const deleteButton = isCustom ? '<button type="button" class="field-delete-custom" title="Xóa trường tự tạo" aria-label="Xóa trường">×</button>' : '';
  return `<article class="field-config-row field-module-card draggable-field ${stateClass}${recommended ? ' module-recommended' : ''}${isCustom ? ' module-custom' : ''}" draggable="true" data-key="${safeKey}" data-label="${safeLabel}" data-group="${groupKey}" data-module-type="${escapeHTML(moduleMeta.type)}" data-field-type="${escapeHTML(normalizedType)}" data-custom="${isCustom ? '1' : '0'}">
    ${customHidden}
    <div class="field-module-head">
      <span class="drag-handle" title="Giữ và kéo để đổi thứ tự" aria-label="Kéo module">☰</span>
      <span class="field-module-icon" aria-hidden="true">${escapeHTML(moduleMeta.icon)}</span>
      <div class="field-module-title">
        <label class="field-label-editor"><span>Tên module</span><input class="field-label-input" type="text" name="label_${safeKey}" value="${safeLabel}" maxlength="80" aria-label="Tên hiển thị của trường ${safeLabel}" title="Đổi tên hiển thị; mã kỹ thuật vẫn giữ nguyên"></label>
        <div class="field-module-sub"><span>${escapeHTML(moduleMeta.type)}</span><code>${safeKey}</code>${recommended ? '<b class="module-fit-badge">Phù hợp</b>' : ''}${isCustom ? '<b class="module-custom-badge">Tự tạo</b>' : ''}</div>
      </div>
      <span class="field-group-badge">${escapeHTML(group.short)}</span>${deleteButton}
    </div>
    <p class="field-module-hint">${safeHint}</p>
    <div class="field-config-quick field-module-controls">
      <label class="field-toggle field-toggle-primary"><input type="checkbox" name="enabled_${safeKey}" ${enabled ? 'checked' : ''}><span>Hiện</span></label>
      <label class="field-toggle"><input type="checkbox" name="required_${safeKey}" ${required ? 'checked' : ''}><span>Bắt buộc</span></label>
      <span class="field-move-buttons"><button type="button" class="field-move-up" title="Đưa module lên">↑</button><button type="button" class="field-move-down" title="Đưa module xuống">↓</button></span>
      <input class="order-input" type="hidden" name="order_${safeKey}" value="${order || 999}">
      <span class="order-badge" title="Thứ tự hiển thị">${order || 999}</span>
      <button type="button" class="field-advanced-toggle" aria-expanded="false">Cấu hình <span>⌄</span></button>
    </div>
    <div class="field-config-options field-config-advanced" hidden>
      <label class="field-toggle filter-capability"><input type="checkbox" name="filterable_${safeKey}" ${filterable ? 'checked' : ''}><span>Dùng làm bộ lọc</span></label>
      <label class="field-toggle filter-capability"><input type="checkbox" name="suggest_${safeKey}" ${suggestValues ? 'checked' : ''}><span>Gợi ý dữ liệu đã nhập</span></label>
      <label class="field-toggle filter-capability"><input type="checkbox" name="allow_custom_${safeKey}" ${allowCustom ? 'checked' : ''}><span>Cho tạo giá trị mới</span></label>
    </div>
  </article>`;
}
function defaultConfig() {
  return fieldDefs.map(([key, label, order], index) => ({ key, label, order, enabled: index < 3, required: index < 3, filterable:false, suggest_values:false, allow_custom:false }));
}
function getRows(list) {
  return [...list.querySelectorAll('.draggable-field')];
}
function normalizeOrders(list) {
  getRows(list).forEach((row, index) => {
    const order = (index + 1) * 10;
    row.querySelector('.order-input').value = order;
    row.querySelector('.order-badge').textContent = order;
  });
}
function readConfig(list) {
  return getRows(list).map((row) => ({
    key: row.dataset.key,
    label: (row.querySelector('.field-label-input')?.value || row.dataset.label || row.dataset.key).trim(),
    enabled: row.querySelector(`[name="enabled_${row.dataset.key}"]`).checked,
    required: row.querySelector(`[name="required_${row.dataset.key}"]`).checked,
    filterable: !!row.querySelector(`[name="filterable_${row.dataset.key}"]`)?.checked,
    suggest_values: !!row.querySelector(`[name="suggest_${row.dataset.key}"]`)?.checked,
    allow_custom: !!row.querySelector(`[name="allow_custom_${row.dataset.key}"]`)?.checked,
    type: row.dataset.fieldType || '',
    custom: row.dataset.custom === '1',
    order: Number(row.querySelector('.order-input').value) || 999
  }));
}

function updateFormStudioSummary(list) {
  if (!list || list.id !== 'editFieldConfig') return;
  const cfg = readConfig(list);
  const enabled = cfg.filter((item) => item.enabled).length;
  const required = cfg.filter((item) => item.enabled && item.required).length;
  const enabledEl = document.getElementById('configEnabledCount');
  const requiredEl = document.getElementById('configRequiredCount');
  if (enabledEl) enabledEl.textContent = `${enabled} trường`;
  if (requiredEl) requiredEl.textContent = `${required} trường`;
}

function previewField(config) {
  const req = config.required ? '<em>*</em>' : '';
  const label = escapeHTML(config.label || config.key);
  if (config.key === 'content' || config.key === 'recruitment_content') {
    return `<label class="preview-field"><span>${label}${req}</span><div class="preview-toolbar">B &nbsp; I &nbsp; U &nbsp; ☷ &nbsp; 🔗</div><div class="preview-textarea"></div></label>`;
  }
  if (config.key === 'image' || config.key === 'company_logo' || config.key === 'document_file') {
    return `<label class="preview-field"><span>${label}${req}</span><div class="preview-upload">${config.key === 'document_file' ? '＋ Chọn file PDF / Word / Excel / PowerPoint' : '＋ Chọn tệp ảnh'}</div></label>`;
  }
  if (config.key === 'deadline') {
    return `<label class="preview-field"><span>${label}${req}</span><div class="preview-input">dd/mm/yyyy &nbsp; 📅</div></label>`;
  }
  if (config.key === 'positions') {
    return `<div class="preview-field"><span>${label}${req}</span><div class="preview-position"><b>Vị trí tuyển dụng số 1</b><div class="preview-input">Tên vị trí tuyển dụng</div><div class="preview-two"><div class="preview-input">Tính chất công việc</div><div class="preview-input">Chuyên môn</div></div><button type="button">+ Thêm vị trí khác</button></div></div>`;
  }
  if (config.custom || String(config.key || '').startsWith('custom_')) {
    const type = config.type || 'text';
    if (type === 'textarea') return `<label class="preview-field"><span>${label}${req}</span><div class="preview-textarea"></div></label>`;
    if (type === 'date') return `<label class="preview-field"><span>${label}${req}</span><div class="preview-input">dd/mm/yyyy &nbsp; 📅</div></label>`;
    const placeholders = {email:'name@example.com', url:'https://...', number:'0'};
    return `<label class="preview-field"><span>${label}${req}</span><div class="preview-input">${placeholders[type] || label}</div></label>`;
  }
  const placeholder = {
    title: 'Nhập tiêu đề bài viết',
    website: 'https://...',
    fanpage: 'https://facebook.com/...',
    cv_email: 'example@email.com'
  }[config.key] || label;
  return `<label class="preview-field"><span>${label}${req}</span><div class="preview-input">${placeholder}</div></label>`;
}

function renderPreview(list, preview) {
  if (!list || !preview) return;
  const cfg = readConfig(list)
    .filter((item) => item.enabled)
    .sort((a, b) => a.order - b.order);
  preview.innerHTML = cfg.length ? cfg.map(previewField).join('') : '<div class="preview-empty">Chưa có trường nào được bật.</div>';
}

function syncRequired(row) {
  const key = row.dataset.key;
  const enabled = row.querySelector(`[name="enabled_${key}"]`);
  const required = row.querySelector(`[name="required_${key}"]`);
  const caps=[row.querySelector(`[name="filterable_${key}"]`),row.querySelector(`[name="suggest_${key}"]`),row.querySelector(`[name="allow_custom_${key}"]`)];
  if (!enabled.checked) {
    required.checked = false;
    required.disabled = true;
    caps.forEach(x=>{if(x){x.checked=false;x.disabled=true}});
  } else {
    required.disabled = false;
    caps.forEach(x=>{if(x)x.disabled=false});
  }
}

function setupDesigner(list) {
  if (!list) return;
  let dragging = null;
  list.querySelectorAll('.draggable-field').forEach((row) => {
    syncRequired(row);
    let dragArmed = false;
    const dragHandle = row.querySelector('.drag-handle');
    dragHandle?.addEventListener('pointerdown', () => { dragArmed = true; });
    document.addEventListener('pointerup', () => { dragArmed = false; }, { passive: true });
    row.addEventListener('dragstart', (event) => {
      if (!dragArmed) { event.preventDefault(); return; }
      dragging = row;
      row.classList.add('dragging');
      event.dataTransfer.effectAllowed = 'move';
    });
    row.addEventListener('dragend', () => {
      row.classList.remove('dragging');
      dragging = null;
      dragArmed = false;
      normalizeOrders(list);
      renderPreview(list, list.id === 'newFieldConfig' ? document.getElementById('newFormPreview') : document.getElementById('editFormPreview'));
    });
    row.addEventListener('dragover', (event) => {
      event.preventDefault();
      if (!dragging || dragging === row) return;
      const rect = row.getBoundingClientRect();
      const dy = event.clientY - (rect.top + rect.height / 2);
      const dx = event.clientX - (rect.left + rect.width / 2);
      const after = Math.abs(dy) > rect.height * 0.18 ? dy > 0 : dx > 0;
      list.insertBefore(dragging, after ? row.nextSibling : row);
    });
    row.querySelector('.field-move-up')?.addEventListener('click', () => {
      const rows = getRows(list).filter((item) => !item.hidden);
      const index = rows.indexOf(row);
      const prev = index > 0 ? rows[index - 1] : null;
      if (prev) list.insertBefore(row, prev);
      normalizeOrders(list);
      renderPreview(list, list.id === 'newFieldConfig' ? document.getElementById('newFormPreview') : document.getElementById('editFormPreview'));
    });
    row.querySelector('.field-move-down')?.addEventListener('click', () => {
      const rows = getRows(list).filter((item) => !item.hidden);
      const index = rows.indexOf(row);
      const next = index >= 0 && index < rows.length - 1 ? rows[index + 1] : null;
      if (next) list.insertBefore(row, next.nextSibling);
      normalizeOrders(list);
      renderPreview(list, list.id === 'newFieldConfig' ? document.getElementById('newFormPreview') : document.getElementById('editFormPreview'));
    });
    const labelInput = row.querySelector('.field-label-input');
    labelInput?.addEventListener('input', () => {
      const value = labelInput.value.trim() || row.dataset.key;
      row.dataset.label = value;
      renderPreview(list, list.id === 'newFieldConfig' ? document.getElementById('newFormPreview') : document.getElementById('editFormPreview'));
    });
    row.querySelector('.field-advanced-toggle')?.addEventListener('click', (event) => {
      const button = event.currentTarget;
      const advanced = row.querySelector('.field-config-advanced');
      const open = advanced?.hasAttribute('hidden');
      if (!advanced) return;
      if (open) advanced.removeAttribute('hidden'); else advanced.setAttribute('hidden', '');
      button.setAttribute('aria-expanded', open ? 'true' : 'false');
      row.classList.toggle('advanced-open', open);
    });
    row.querySelectorAll('input[type="checkbox"]').forEach((checkbox) => {
      checkbox.addEventListener('change', () => {
        syncRequired(row);
        const enabledInput = row.querySelector(`[name="enabled_${row.dataset.key}"]`);
        row.classList.toggle('module-enabled', !!enabledInput?.checked);
        row.classList.toggle('module-disabled', !enabledInput?.checked);
        renderPreview(list, list.id === 'newFieldConfig' ? document.getElementById('newFormPreview') : document.getElementById('editFormPreview'));
        updateFormStudioSummary(list);
      });
    });
    row.querySelector('.field-delete-custom')?.addEventListener('click', () => {
      const label = row.querySelector('.field-label-input')?.value?.trim() || 'trường này';
      if (!window.confirm(`Xóa trường tự tạo “${label}”?`)) return;
      const activeGroup = list.dataset.activeFieldGroup || 'all';
      row.remove();
      normalizeOrders(list);
      const cfg = readConfig(list);
      renderDesigner(list, cfg);
      if (typeof list._applyFieldGroupFilter === 'function') list._applyFieldGroupFilter(activeGroup === 'custom' ? 'custom' : activeGroup);
    });
  });
  normalizeOrders(list);
}

function setupFieldGroupFilters(list, initialGroup = 'main') {
  if (!list) return;
  const buttons = [...list.querySelectorAll('.field-group-filter')];
  const summary = list.querySelector('[data-field-filter-summary]');
  const activeTitle = list.querySelector('[data-field-active-title]');
  const activeDesc = list.querySelector('[data-field-active-desc]');

  const apply = (value = initialGroup) => {
    const rows = getRows(list);
    let visibleCount = 0;
    rows.forEach((row) => {
      const enabled = !!row.querySelector(`[name="enabled_${row.dataset.key}"]`)?.checked;
      const visible = value === 'all' || (value === 'enabled' ? enabled : row.dataset.group === value);
      row.classList.toggle('field-filter-hidden', !visible);
      row.hidden = !visible;
      row.setAttribute('aria-hidden', visible ? 'false' : 'true');
      if (visible) visibleCount += 1;
    });
    list.dataset.activeFieldGroup = value;
    buttons.forEach((button) => {
      const active = button.dataset.fieldGroup === value;
      button.classList.toggle('active', active);
      button.setAttribute('aria-pressed', active ? 'true' : 'false');
    });

    const group = fieldGroups[value];
    const label = value === 'all' ? 'Tất cả trường' : value === 'enabled' ? 'Các trường đang dùng' : (group?.label || 'Nhóm trường');
    const desc = value === 'all'
      ? 'Hiển thị toàn bộ trường của form.'
      : value === 'enabled'
        ? 'Chỉ hiển thị những trường đang được bật trong form.'
        : (group?.desc || 'Chỉnh sửa các trường trong nhóm đã chọn.');
    if (summary) summary.textContent = `${visibleCount}/${rows.length} trường`;
    if (activeTitle) activeTitle.textContent = label;
    if (activeDesc) activeDesc.textContent = desc;
  };

  // Presets and checkbox changes reuse the currently selected group.
  list._applyFieldGroupFilter = apply;
  buttons.forEach((button) => button.addEventListener('click', () => apply(button.dataset.fieldGroup || initialGroup)));
  getRows(list).forEach((row) => row.querySelector(`[name="enabled_${row.dataset.key}"]`)?.addEventListener('change', () => {
    apply(list.dataset.activeFieldGroup || initialGroup);
  }));

  const wanted = buttons.some((button) => button.dataset.fieldGroup === initialGroup) ? initialGroup : 'all';
  apply(wanted);
}

let customFieldTargetList = null;
function makeCustomFieldKey() {
  const seed = `${Date.now().toString(36)}_${Math.random().toString(36).slice(2,7)}`.toLowerCase().replace(/[^a-z0-9_]/g,'');
  return `custom_${seed}`;
}
function openCustomFieldModal(list) {
  customFieldTargetList = list;
  const modal = document.getElementById('customFieldModal');
  const form = document.getElementById('customFieldForm');
  form?.reset();
  modal?.classList.add('open');
  document.body.classList.add('modal-lock');
  window.setTimeout(() => document.getElementById('customFieldLabel')?.focus(), 80);
}
function closeCustomFieldModal() {
  document.getElementById('customFieldModal')?.classList.remove('open');
  customFieldTargetList = null;
  if (typeof syncAdminModalLock === 'function') syncAdminModalLock();
}
document.getElementById('closeCustomFieldModal')?.addEventListener('click', closeCustomFieldModal);
document.getElementById('cancelCustomField')?.addEventListener('click', closeCustomFieldModal);
document.getElementById('customFieldForm')?.addEventListener('submit', (event) => {
  event.preventDefault();
  const list = customFieldTargetList;
  if (!list) return;
  const label = String(document.getElementById('customFieldLabel')?.value || '').trim();
  const type = String(document.getElementById('customFieldType')?.value || 'text').trim();
  if (!label) { showToast('Hãy nhập tên trường mới', true); return; }
  const cfg = readConfig(list);
  const nextOrder = Math.max(0, ...cfg.map(item => Number(item.order) || 0)) + 10;
  const item = {key:makeCustomFieldKey(), label, type, custom:true, enabled:true, required:!!document.getElementById('customFieldRequired')?.checked, filterable:false, suggest_values:false, allow_custom:true, order:nextOrder};
  cfg.push(item);
  renderDesigner(list, cfg);
  if (typeof list._applyFieldGroupFilter === 'function') list._applyFieldGroupFilter('custom');
  closeCustomFieldModal();
  showToast('Đã thêm trường mới vào form');
});

function renderDesigner(list, config) {
  if (!list) return;
  const map = new Map((config || []).map((item) => [item.key, item]));
  const categorySlug = String(list.dataset.categorySlug || '').trim();
  const categoryName = String(list.dataset.categoryName || '').trim();
  const profile = categoryModuleProfile(categorySlug);
  const builtIn = fieldDefs.map(([key, label, defaultOrder], index) => {
    const current = map.get(key);
    if (current) return { key, label: String(current.label || label).trim() || label, order: current.order || defaultOrder, enabled: !!current.enabled, required: !!current.required, filterable:!!current.filterable, suggest_values:!!current.suggest_values, allow_custom:!!current.allow_custom, type:'', custom:false };
    return { key, label, order: defaultOrder, enabled: index < 3, required: index < 3, filterable:false, suggest_values:false, allow_custom:false, type:'', custom:false };
  });
  const custom = (config || []).filter(item => item && !builtInFieldKeys.has(item.key) && String(item.key || '').startsWith('custom_')).map((item, index) => ({
    key:String(item.key), label:String(item.label || 'Trường tùy chỉnh').trim() || 'Trường tùy chỉnh', order:Number(item.order) || 1000 + index*10, enabled:item.enabled !== false, required:!!item.required, filterable:!!item.filterable, suggest_values:!!item.suggest_values, allow_custom:item.allow_custom !== false, type:item.type || 'text', custom:true
  }));
  const ordered = [...builtIn, ...custom].sort((a,b)=>(a.order||999)-(b.order||999));

  const groupCounts = Object.keys(fieldGroups).reduce((acc, key) => { acc[key] = 0; return acc; }, {});
  ordered.forEach((item) => { const group = item.custom ? 'custom' : (fieldGroupMap[item.key] || 'main'); groupCounts[group] = (groupCounts[group] || 0) + 1; });
  const enabledCount = ordered.filter((item) => item.enabled).length;
  const recommendedSet = new Set(profile.recommended || []);
  const recommendedCount = ordered.filter((item) => !item.custom && recommendedSet.has(item.key)).length;
  const groupOrder = [...new Set([...(profile.groupOrder || []), 'custom', ...Object.keys(fieldGroups)])];
  const groupButtons = groupOrder.map((key) => {
    const group = fieldGroups[key];
    if (!group) return '';
    return `<button type="button" class="field-group-filter" data-field-group="${key}" aria-pressed="false" title="${escapeHTML(group.label)}"><span>${escapeHTML(group.short)}</span><b>${groupCounts[key] || 0}</b></button>`;
  }).join('');
  const contextLabel = categoryName || profile.label;
  list.dataset.profileTone = profile.tone || 'default';
  list.innerHTML = `<div class="field-builder-toolbar module-builder-toolbar" data-profile-tone="${escapeHTML(profile.tone || 'default')}">
    <div class="module-toolbar-head">
      <div><span class="module-context-kicker">${escapeHTML(contextLabel)}</span><strong>Module form đang sử dụng</strong><small>Có thể dùng module có sẵn hoặc tự tạo thêm trường riêng cho chuyên mục.</small></div>
      <div class="module-toolbar-actions"><button type="button" class="add-custom-field-btn" data-add-custom-field><span>＋</span> Thêm trường</button><div class="module-toolbar-stats"><span><b>${enabledCount}</b> đang dùng</span><span><b>${custom.length}</b> tự tạo</span><span class="field-filter-summary" data-field-filter-summary>${enabledCount}/${ordered.length}</span></div></div>
    </div>
    <div class="field-group-filters"><button type="button" class="field-group-filter field-group-filter-primary" data-field-group="enabled" aria-pressed="false" title="Các module đang dùng"><span>Đang dùng</span><b>${enabledCount}</b></button>${groupButtons}<button type="button" class="field-group-filter field-group-filter-muted" data-field-group="all" aria-pressed="false" title="Hiển thị toàn bộ module"><span>Tất cả</span><b>${ordered.length}</b></button></div>
  </div>` + ordered.map((item) => fieldRow(item.key, item.label, item.order, item.enabled, item.required, item.filterable, item.suggest_values, item.allow_custom, categorySlug, item.type, item.custom)).join('');
  list.querySelector('[data-add-custom-field]')?.addEventListener('click', () => openCustomFieldModal(list));
  setupDesigner(list);
  setupFieldGroupFilters(list, enabledCount ? 'enabled' : (profile.groupOrder?.[0] || 'main'));
  renderPreview(list, list.id === 'newFieldConfig' ? document.getElementById('newFormPreview') : document.getElementById('editFormPreview'));
  updateFormStudioSummary(list);
}
const formPresets = {
  basic: ['title','content','image','tags','source'],
  job: ['title','content','image','deadline','company_logo','website','fanpage','recruitment_content','cv_email','positions','contact_name','contact_phone','organization','location','salary_range','application_link','tags'],
  event: ['title','content','image','event_time','location','organization','contact_name','contact_phone','application_link','audience','tags','source'],
  scholarship: ['title','content','image','deadline','organization','website','application_link','contact_name','contact_phone','audience','tags','source'],
  document: ['title','content','image','document_file','school','document_type','subject','academic_year','tags','source']
};

function applyFieldPreset(list, presetName) {
  if (!list || !formPresets[presetName]) return;
  const enabledKeys = new Set(formPresets[presetName]);
  getRows(list).forEach((row) => {
    const key = row.dataset.key;
    const enabled = row.querySelector(`[name="enabled_${key}"]`);
    const required = row.querySelector(`[name="required_${key}"]`);
    if (!enabled || !required) return;
    enabled.checked = enabledKeys.has(key);
    required.checked = ['title','content'].includes(key) || (presetName === 'job' && ['image','cv_email'].includes(key));
    const filterable=row.querySelector(`[name="filterable_${key}"]`);
    const suggest=row.querySelector(`[name="suggest_${key}"]`);
    const custom=row.querySelector(`[name="allow_custom_${key}"]`);
    const docFilter=presetName==='document' && ['school','document_type','subject','academic_year'].includes(key);
    if(filterable)filterable.checked=docFilter;
    if(suggest)suggest.checked=docFilter;
    if(custom)custom.checked=docFilter;
    syncRequired(row);
  });
  normalizeOrders(list);
  renderPreview(list, list.id === 'newFieldConfig' ? document.getElementById('newFormPreview') : document.getElementById('editFormPreview'));
  if (typeof list._applyFieldGroupFilter === 'function') list._applyFieldGroupFilter(list.dataset.activeFieldGroup || 'all');
  updateFormStudioSummary(list);
  showToast('Đã áp dụng mẫu form');
}

document.querySelectorAll('.field-preset').forEach((button) => {
  button.addEventListener('click', () => applyFieldPreset(document.getElementById(button.dataset.list), button.dataset.preset));
});

function syncCompanyCardSetting(toggle, select) {
  if (!toggle || !select) return;
  select.disabled = !toggle.checked;
  select.closest('label')?.classList.toggle('is-disabled', !toggle.checked);
}

const newList = document.getElementById('newFieldConfig');
if (newList) { newList.dataset.categorySlug = ''; newList.dataset.categoryName = 'Chuyên mục mới'; renderDesigner(newList, defaultConfig()); }
const newShowCompanyCard = document.getElementById('newShowCompanyCard');
const newCompanyCardStyle = document.getElementById('newCompanyCardStyle');
newShowCompanyCard?.addEventListener('change', () => syncCompanyCardSetting(newShowCompanyCard, newCompanyCardStyle));
syncCompanyCardSetting(newShowCompanyCard, newCompanyCardStyle);

function updateAudienceGroup(group, changedInput) {
  if (!group) return;
  const inputs = [...group.querySelectorAll('input[name="audience_scope"]')];
  const publicInput = inputs.find((input) => input.value === 'public');
  if (changedInput?.value === 'public' && changedInput.checked) {
    inputs.forEach((input) => { if (input !== publicInput) input.checked = false; });
  } else if (changedInput && changedInput.value !== 'public' && changedInput.checked && publicInput) {
    publicInput.checked = false;
  }
  let selected = inputs.filter((input) => input.checked);
  if (!selected.length && publicInput) {
    publicInput.checked = true;
    selected = [publicInput];
  }
  group.querySelectorAll('.audience-choice').forEach((card) => {
    const input = card.querySelector('input[name="audience_scope"]');
    card.classList.toggle('selected', Boolean(input?.checked));
  });
  const count = group.querySelector('[data-audience-count]');
  if (count) count.textContent = `${selected.length} nhóm`;
}

function setAudienceGroupValues(group, raw) {
  if (!group) return;
  const values = new Set(String(raw || 'public').split(',').map((v) => v.trim()).filter(Boolean));
  group.querySelectorAll('input[name="audience_scope"]').forEach((input) => {
    input.checked = values.has(input.value);
  });
  updateAudienceGroup(group);
}

document.querySelectorAll('[data-audience-group]').forEach((group) => {
  group.querySelectorAll('input[name="audience_scope"]').forEach((input) => {
    input.addEventListener('change', () => updateAudienceGroup(group, input));
  });
  updateAudienceGroup(group);
});

function decodeCategoryConfig(button) {
  const raw = button?.dataset?.configB64 || '';
  if (!raw) return [];
  try {
    const bytes = Uint8Array.from(atob(raw), (char) => char.charCodeAt(0));
    const json = new TextDecoder('utf-8').decode(bytes);
    return JSON.parse(json || '[]');
  } catch (error) {
    console.warn('Không thể đọc cấu hình chuyên mục:', error);
    return [];
  }
}


function syncDocumentFormatPanel(form) {
  if (!form) return;
  const style = form.querySelector('input[name="post_card_style"]:checked')?.value || 'normal';
  const panel = form.querySelector('[data-document-format-settings]');
  if (!panel) return;
  panel.classList.toggle('is-disabled', style !== 'document');
}
document.querySelectorAll('form').forEach((form) => {
  form.querySelectorAll('input[name="post_card_style"]').forEach((radio) => radio.addEventListener('change', () => syncDocumentFormatPanel(form)));
  if (form.querySelector('input[name="post_card_style"]')) syncDocumentFormatPanel(form);
});

// DACS_V1.24.25 · Quick category settings outside Form Module Studio.
const categoryQuickModal = document.getElementById('categoryQuickModal');
const categoryQuickForm = document.getElementById('categoryQuickForm');
const categoryQuickBody = document.getElementById('categoryQuickBody');
const categoryQuickTitle = document.getElementById('categoryQuickTitle');
const categoryQuickDesc = document.getElementById('categoryQuickDesc');
const categoryQuickId = document.getElementById('categoryQuickId');
const categoryQuickField = document.getElementById('categoryQuickField');
let categoryQuickRow = null;

function categoryAudienceHTML(raw) {
  const values = new Set(String(raw || 'public').split(',').map((v) => v.trim()).filter(Boolean));
  const choices = [
    ['public','◎','Mọi người','Ai cũng được xem nội dung'],
    ['members','●','Thành viên','Đăng nhập để xem nội dung'],
    ['students','🎓','Sinh viên','Hồ sơ Sinh viên mới được xem'],
    ['same_school','⌂','Cùng trường','Cần hồ sơ Sinh viên và thông tin trường'],
    ['admin','◆','Quản trị','Chỉ quản trị viên được xem nội dung']
  ];
  return `<div class="quick-audience-grid" data-quick-audience>${choices.map(([value,icon,label,desc]) => `<label class="quick-audience-choice"><input type="checkbox" name="audience_scope" value="${value}" ${values.has(value) ? 'checked' : ''}><span>${icon}</span><div><b>${label}</b><small>${desc}</small></div><i>✓</i></label>`).join('')}</div>`;
}

function categoryCardStyleHTML(value, documentFormats = '') {
  const current = value || 'normal';
  const selectedFormats = new Set(String(documentFormats || '').split(',').map((v) => v.trim()).filter(Boolean));
  const choices = [
    ['normal','Thẻ bình thường','Bố cục tiêu chuẩn cho nội dung tổng hợp'],
    ['horizontal','Thẻ ngang','Ưu tiên ảnh và nội dung theo chiều ngang'],
    ['document','Thẻ tài liệu','Tối ưu cho tài liệu học tập và tệp đính kèm']
  ];
  const formats = ['pdf','doc','docx','xls','xlsx','ppt','pptx','txt','zip'];
  return `<div class="quick-card-style-grid">${choices.map(([key,label,desc]) => `<label class="quick-card-style-choice"><input type="radio" name="post_card_style" value="${key}" ${current === key ? 'checked' : ''}><span class="quick-style-visual quick-style-${key}"><i></i><i></i><i></i></span><div><b>${label}</b><small>${desc}</small></div><em>✓</em></label>`).join('')}</div><div class="quick-document-formats" data-quick-document-formats><div><strong>Định dạng tài liệu</strong><small>Chỉ áp dụng khi chọn “Thẻ tài liệu”.</small></div><div class="quick-format-grid">${formats.map((format) => `<label><input type="checkbox" name="document_formats" value="${format}" ${selectedFormats.has(format) ? 'checked' : ''}><span>${format.toUpperCase()}</span></label>`).join('')}</div></div>`;
}

function categoryCompanyCardHTML(show, style) {
  const enabled = String(show) === 'true';
  const current = style || 'full';
  return `<div class="quick-company-settings"><label class="quick-switch-row"><input type="checkbox" name="show_company_card" ${enabled ? 'checked' : ''}><span><b>Hiển thị khối thông tin phụ</b><small>Hiện thẻ tóm tắt ở trang chi tiết bài viết khi chuyên mục hỗ trợ.</small></span><i></i></label><label class="quick-select-field"><span>Kiểu hiển thị</span><select name="company_card_style"><option value="full" ${current === 'full' ? 'selected' : ''}>Thẻ đầy đủ</option><option value="horizontal" ${current === 'horizontal' ? 'selected' : ''}>Thẻ ngang</option><option value="vertical" ${current === 'vertical' ? 'selected' : ''}>Thẻ dọc</option></select></label></div>`;
}

function openCategoryQuickEditor(row, field) {
  if (!row || !categoryQuickModal || !categoryQuickBody || !categoryQuickForm) return;
  categoryQuickRow = row;
  const id = row.dataset.categoryId || '';
  const name = row.dataset.categoryName || '';
  const slug = row.dataset.categorySlug || '';
  const audience = row.dataset.audienceScope || 'public';
  const style = row.dataset.postCardStyle || 'normal';
  const documentFormats = row.dataset.documentFormats || '';
  const showCompanyCard = row.dataset.showCompanyCard || 'false';
  const companyCardStyle = row.dataset.companyCardStyle || 'full';
  if (categoryQuickId) categoryQuickId.value = id;
  if (categoryQuickField) categoryQuickField.value = field;
  if (field === 'name') {
    categoryQuickTitle.textContent = 'Đổi tên chuyên mục';
    categoryQuickDesc.textContent = 'Tên mới cập nhật ngay trên thanh chuyên mục. Slug kỹ thuật được giữ nguyên để không làm hỏng liên kết.';
    categoryQuickBody.innerHTML = `<label class="quick-name-field"><span>Tên chuyên mục</span><input name="name" value="${escapeHTML(name)}" maxlength="80" required autofocus><small>Slug cố định: <code>${escapeHTML(slug)}</code></small></label>`;
  } else if (field === 'audience') {
    categoryQuickTitle.textContent = 'Hiển thị chuyên mục cho ai?';
    categoryQuickDesc.textContent = 'Chuyên mục vẫn hiện với mọi người; lựa chọn này chỉ giới hạn quyền xem nội dung bên trong.';
    categoryQuickBody.innerHTML = categoryAudienceHTML(audience);
    const group = categoryQuickBody.querySelector('[data-quick-audience]');
    const sync = (changed) => {
      const inputs = [...group.querySelectorAll('input[name="audience_scope"]')];
      const publicInput = inputs.find((input) => input.value === 'public');
      if (changed?.value === 'public' && changed.checked) inputs.forEach((input) => { if (input !== publicInput) input.checked = false; });
      if (changed && changed.value !== 'public' && changed.checked && publicInput) publicInput.checked = false;
      if (!inputs.some((input) => input.checked) && publicInput) publicInput.checked = true;
      group.querySelectorAll('.quick-audience-choice').forEach((card) => card.classList.toggle('selected', !!card.querySelector('input')?.checked));
    };
    group.querySelectorAll('input').forEach((input) => input.addEventListener('change', () => sync(input)));
    sync();
  } else if (field === 'post_card_style') {
    categoryQuickTitle.textContent = 'Kiểu thẻ bài viết';
    categoryQuickDesc.textContent = 'Chọn cách bài viết của chuyên mục hiển thị trên bảng tin. Form điền bài được cấu hình riêng trong Form Module Studio.';
    categoryQuickBody.innerHTML = categoryCardStyleHTML(style, documentFormats);
    const syncFormatPanel = () => {
      const selectedStyle = categoryQuickBody.querySelector('input[name="post_card_style"]:checked')?.value || 'normal';
      categoryQuickBody.querySelector('[data-quick-document-formats]')?.classList.toggle('active', selectedStyle === 'document');
    };
    categoryQuickBody.querySelectorAll('input[name="post_card_style"]').forEach((input) => input.addEventListener('change', syncFormatPanel));
    syncFormatPanel();
  } else {
    categoryQuickTitle.textContent = 'Khối thông tin phụ';
    categoryQuickDesc.textContent = 'Bật/tắt thẻ tóm tắt bên cạnh nội dung chi tiết và chọn cách hiển thị. Cấu hình này độc lập với Form Module Studio.';
    categoryQuickBody.innerHTML = categoryCompanyCardHTML(showCompanyCard, companyCardStyle);
  }
  categoryQuickModal.classList.add('open');
  document.body.classList.add('modal-lock');
  window.setTimeout(() => categoryQuickBody.querySelector('input:not([type="hidden"])')?.focus(), 80);
}

function closeCategoryQuickEditor() {
  categoryQuickModal?.classList.remove('open');
  categoryQuickRow = null;
  syncAdminModalLock();
}

document.querySelectorAll('[data-category-edit]').forEach((button) => button.addEventListener('click', (event) => {
  event.stopPropagation();
  openCategoryQuickEditor(button.closest('.category-sort-row'), button.dataset.categoryEdit);
}));
document.getElementById('closeCategoryQuickModal')?.addEventListener('click', closeCategoryQuickEditor);
document.getElementById('cancelCategoryQuick')?.addEventListener('click', closeCategoryQuickEditor);

function renderAudienceBadges(raw) {
  const values = new Set(String(raw || 'public').split(',').map((v) => v.trim()).filter(Boolean));
  if (values.has('public')) return '<span class="audience-badge audience-public">Mọi người</span>';
  const labels = {members:'Thành viên',students:'Sinh viên',same_school:'Cùng trường',admin:'Quản trị'};
  return [...values].filter((v) => labels[v]).map((v) => `<span class="audience-badge audience-${v}">${labels[v]}</span>`).join('') || '<span class="audience-badge audience-public">Mọi người</span>';
}
function cardStyleLabel(value) { return value === 'horizontal' ? 'Thẻ ngang' : value === 'document' ? 'Thẻ tài liệu' : 'Thẻ bình thường'; }

categoryQuickForm?.addEventListener('submit', async (event) => {
  event.preventDefault();
  const button = categoryQuickForm.querySelector('button[type="submit"]');
  const quickField = categoryQuickField?.value || '';
  if (quickField === 'post_card_style' && categoryQuickForm.querySelector('input[name="post_card_style"]:checked')?.value === 'document' && !categoryQuickForm.querySelector('input[name="document_formats"]:checked')) {
    showToast('Hãy chọn ít nhất một định dạng tài liệu.', true);
    return;
  }
  try {
    if (button) { button.disabled = true; button.textContent = 'Đang lưu...'; }
    const data = await postFormJSON(categoryQuickForm);
    const cat = data.category || {};
    if (!categoryQuickRow) throw new Error('Không xác định được dòng chuyên mục cần cập nhật');
    const name = String(cat.Name ?? cat.name ?? categoryQuickRow.dataset.categoryName ?? '');
    const audience = String(cat.AudienceScope ?? cat.audienceScope ?? categoryQuickRow.dataset.audienceScope ?? 'public');
    const style = String(cat.PostCardStyle ?? cat.postCardStyle ?? categoryQuickRow.dataset.postCardStyle ?? 'normal');
    const showCompany = Boolean(cat.ShowCompanyCard ?? cat.showCompanyCard ?? (categoryQuickRow.dataset.showCompanyCard === 'true'));
    const companyStyle = String(cat.CompanyCardStyle ?? cat.companyCardStyle ?? categoryQuickRow.dataset.companyCardStyle ?? 'full');
    const formats = String(cat.DocumentFormats ?? cat.documentFormats ?? categoryQuickRow.dataset.documentFormats ?? '');
    categoryQuickRow.dataset.categoryName = name;
    categoryQuickRow.dataset.audienceScope = audience;
    categoryQuickRow.dataset.postCardStyle = style;
    categoryQuickRow.dataset.showCompanyCard = showCompany ? 'true' : 'false';
    categoryQuickRow.dataset.companyCardStyle = companyStyle;
    categoryQuickRow.dataset.documentFormats = formats;
    const nameText = categoryQuickRow.querySelector('[data-category-name-text]');
    if (nameText) nameText.textContent = name;
    const audienceView = categoryQuickRow.querySelector('[data-category-audience-view]');
    if (audienceView) audienceView.innerHTML = renderAudienceBadges(audience);
    const styleView = categoryQuickRow.querySelector('[data-category-card-style-view]');
    if (styleView) { styleView.textContent = cardStyleLabel(style); styleView.className = `tag card-style-badge card-style-${style}`; styleView.setAttribute('data-category-card-style-view',''); }
    const companyView = categoryQuickRow.querySelector('[data-category-company-view]');
    if (companyView) companyView.innerHTML = showCompany ? `<span class="tag">Có thẻ</span><span class="tag ghost-tag">${escapeHTML(companyStyle)}</span>` : '<span class="muted-status">Không hiển thị</span>';
    const configButton = categoryQuickRow.querySelector('.config-category');
    if (configButton) { configButton.dataset.name = name; configButton.dataset.audienceScope = audience; configButton.dataset.postCardStyle = style; configButton.dataset.showCompanyCard = showCompany ? 'true' : 'false'; configButton.dataset.companyCardStyle = companyStyle; configButton.dataset.documentFormats = formats; }
    showToast(data.message || 'Đã cập nhật chuyên mục');
    closeCategoryQuickEditor();
  } catch (error) {
    showToast(error.message, true);
  } finally {
    if (button) { button.disabled = false; button.textContent = 'Lưu thay đổi'; }
  }
});

document.querySelectorAll('.config-category').forEach((button) => {
  button.addEventListener('click', () => {
    const config = decodeCategoryConfig(button);
    const row = button.closest('.category-sort-row');
    const rowID = row?.dataset.categoryId || '';
    const categoryID = String(button.dataset.id || rowID || '').trim();
    const categoryName = String(row?.dataset.categoryName || button.dataset.name || '').trim();
    const categorySlug = String(row?.dataset.categorySlug || button.dataset.slug || '').trim();
    const idInput = document.getElementById('configCategoryId');
    const fallbackInput = document.getElementById('configCategoryIdFallback');
    const slugInput = document.getElementById('configCategorySlug');
    if (!/^\d+$/.test(categoryID) || Number(categoryID) <= 0) {
      showToast('Không xác định được ID chuyên mục. Hãy tải lại trang quản trị.', true);
      return;
    }
    if (idInput) idInput.value = categoryID;
    if (fallbackInput) fallbackInput.value = categoryID;
    if (slugInput) slugInput.value = categorySlug;
    const title = document.getElementById('configTitle');
    const subtitle = document.getElementById('configSubtitle');
    const nameChip = document.getElementById('configCategoryNameChip');
    const slugChip = document.getElementById('configCategorySlugChip');
    if (title) title.textContent = 'Thiết kế form · ' + categoryName;
    if (subtitle) subtitle.textContent = 'Chỉ cấu hình các trường người dùng phải điền khi đăng bài vào chuyên mục này.';
    if (nameChip) nameChip.textContent = categoryName || '—';
    if (slugChip) slugChip.textContent = categorySlug || '—';
    const editFieldList = document.getElementById('editFieldConfig');
    if (editFieldList) { editFieldList.dataset.categorySlug = categorySlug; editFieldList.dataset.categoryName = categoryName; }
    renderDesigner(editFieldList, config);
    document.getElementById('configModal')?.classList.add('open');
    document.body.classList.add('modal-lock');
  });
});

const configForm = document.getElementById('configForm');
configForm?.addEventListener('submit', async (event) => {
  event.preventDefault();
  const idInput = document.getElementById('configCategoryId');
  const fallbackInput = document.getElementById('configCategoryIdFallback');
  const slugInput = document.getElementById('configCategorySlug');
  const categoryID = String(idInput?.value || fallbackInput?.value || '').trim();
  const categorySlug = String(slugInput?.value || '').trim();
  if ((!/^\d+$/.test(categoryID) || Number(categoryID) <= 0) && !categorySlug) {
    showToast('Không xác định được chuyên mục. Hãy đóng cửa sổ và mở lại Cấu hình form.', true);
    return;
  }
  if (idInput && categoryID) idInput.value = categoryID;
  if (fallbackInput && categoryID) fallbackInput.value = categoryID;
  const editList = document.getElementById('editFieldConfig');
  if (editList) normalizeOrders(editList);
  const submitButton = configForm.querySelector('button[type="submit"], button.primary');
  try {
    if (submitButton) {
      submitButton.disabled = true;
      submitButton.dataset.originalText ||= submitButton.textContent;
      submitButton.textContent = 'Đang lưu...';
    }

    // Category configuration contains no file uploads. Send it as URL-encoded
    // data so Go's normal form parser receives every repeated checkbox/order field.
    const formData = new FormData(configForm);
    if (categoryID) { formData.set('id', categoryID); formData.set('category_id', categoryID); }
    if (categorySlug) formData.set('category_slug', categorySlug);
    const body = new URLSearchParams();
    for (const [key, value] of formData.entries()) body.append(key, String(value));

    const target = new URL(configForm.action, window.location.href);
    if (categoryID) target.searchParams.set('id', categoryID);
    if (categorySlug) target.searchParams.set('category_slug', categorySlug);
    const response = await fetch(target.toString(), {
      method: 'POST',
      body,
      headers: {
        'X-Requested-With': 'XMLHttpRequest',
        'Accept': 'application/json',
        'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8'
      },
      credentials: 'same-origin'
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok || !data.ok) throw new Error(data.message || 'Không thể lưu cấu hình chuyên mục');
    if (!data.category || Number(data.category.ID || data.category.id || 0) !== Number(categoryID)) {
      throw new Error('Máy chủ chưa xác nhận dữ liệu chuyên mục vừa lưu');
    }
    showToast(data.message || 'Đã lưu cấu hình chuyên mục');
    window.setTimeout(() => {
      window.location.href = '/admin?tab=categories&msg=' + encodeURIComponent(data.message || 'Đã cập nhật cấu hình chuyên mục');
    }, 250);
  } catch (error) {
    showToast(error.message, true);
    if (submitButton) {
      submitButton.disabled = false;
      submitButton.textContent = submitButton.dataset.originalText || 'Lưu cấu hình';
    }
  }
});

const previewModal = document.getElementById('previewModal');
const previewModalBody = document.getElementById('previewModalBody');
const previewModalTitle = document.getElementById('previewModalTitle');
function openPreviewModal(listId, title) {
  const list = document.getElementById(listId);
  if (!list || !previewModal || !previewModalBody) return;
  normalizeOrders(list);
  renderPreview(list, previewModalBody);
  if (previewModalTitle) previewModalTitle.textContent = title || 'Xem trước form';
  previewModal.classList.add('open');
  document.body.classList.add('modal-lock');
}
function closePreviewModal() {
  previewModal?.classList.remove('open');
  syncAdminModalLock();
}
document.querySelectorAll('.preview-form-btn[data-list]').forEach((btn) => {
  btn.addEventListener('click', () => openPreviewModal(btn.dataset.list, btn.dataset.title));
});
document.getElementById('closePreviewModal')?.addEventListener('click', closePreviewModal);
document.getElementById('closePreviewButton')?.addEventListener('click', closePreviewModal);

const adForm = document.getElementById('adForm');
const adReset = document.getElementById('adReset');
const adImageInput = document.getElementById('adImage');
const adPreviewButton = document.getElementById('adPreviewButton');
const adPreviewModal = document.getElementById('adPreviewModal');
const adPreviewCanvas = document.getElementById('adPreviewCanvas');
const adPreviewTitle = document.getElementById('adPreviewTitle');
const adPreviewPositionLabel = document.getElementById('adPreviewPositionLabel');
const adPositionInput = document.getElementById('adPosition');
const adPositionPicker = document.getElementById('adPositionPicker');
const adSortable = document.getElementById('adSortable');
const adLivePreviewCanvas = document.getElementById('adLivePreviewCanvas');
const adLivePreviewPosition = document.getElementById('adLivePreviewPosition');
const adFormTitle = document.getElementById('adFormTitle');
const adFormHint = document.getElementById('adFormHint');
const adSaveButton = document.getElementById('adSaveButton');
const adImageEditNote = document.getElementById('adImageEditNote');
const adImageEditor = document.getElementById('adImageEditor');
const adImageEditorStatus = document.getElementById('adImageEditorStatus');
const adImageRatioLabel = document.getElementById('adImageRatioLabel');
const adImageZoom = document.getElementById('adImageZoom');
const adImageZoomValue = document.getElementById('adImageZoomValue');
const adImageCenter = document.getElementById('adImageCenter');
const adImageResetCrop = document.getElementById('adImageResetCrop');
const adImageApplyCrop = document.getElementById('adImageApplyCrop');
let currentAdImagePreview = '';
let adjustedAdImageBlob = null;
const adImageAdjustment = { zoom: 1, x: 50, y: 50, dirty: false, baseSource: '', originalSource: '' };

function adCropSpec(position) {
  if (position === 'home-right') return { width: 1600, height: 700, label: 'Khung 16:7' };
  if (position === 'feed-sponsored') return { width: 1200, height: 675, label: 'Khung 16:9' };
  return { width: 1200, height: 900, label: 'Khung 4:3' };
}

function clampAdImageValue(value, min, max) { return Math.min(max, Math.max(min, value)); }

function resetAdImageAdjustment(source = '', keepOriginal = false) {
  adImageAdjustment.zoom = 1;
  adImageAdjustment.x = 50;
  adImageAdjustment.y = 50;
  adImageAdjustment.dirty = false;
  adImageAdjustment.baseSource = source || '';
  if (!keepOriginal) adImageAdjustment.originalSource = source || '';
  adjustedAdImageBlob = null;
  if (adImageZoom) adImageZoom.value = '1';
  syncAdImageEditor();
}

function syncAdImageEditor(message = '') {
  const hasImage = Boolean(currentAdImagePreview || adImageAdjustment.baseSource);
  const spec = adCropSpec(adPositionInput?.value || 'home-right');
  if (adImageRatioLabel) adImageRatioLabel.textContent = spec.label;
  if (adImageZoom) adImageZoom.disabled = !hasImage;
  if (adImageCenter) adImageCenter.disabled = !hasImage;
  if (adImageResetCrop) adImageResetCrop.disabled = !hasImage;
  if (adImageApplyCrop) adImageApplyCrop.disabled = !hasImage;
  if (adImageZoomValue) adImageZoomValue.textContent = Math.round(adImageAdjustment.zoom * 100) + '%';
  if (adImageEditor) adImageEditor.classList.toggle('is-ready', hasImage);
  if (adImageEditorStatus) {
    if (message) adImageEditorStatus.textContent = message;
    else if (!hasImage) adImageEditorStatus.textContent = 'Chọn hoặc tải ảnh quảng cáo để bắt đầu chỉnh.';
    else if (adImageAdjustment.dirty) adImageEditorStatus.textContent = 'Ảnh đang được căn lại · bấm Áp dụng ảnh đã chỉnh trước khi lưu.';
    else if (adjustedAdImageBlob) adImageEditorStatus.textContent = 'Ảnh đã được chỉnh và sẽ được dùng khi lưu quảng cáo.';
    else adImageEditorStatus.textContent = 'Kéo ảnh trực tiếp trong khung hoặc dùng thanh Thu phóng.';
  }
}

function applyAdImageAdjustmentToPreview() {
  const image = adLivePreviewCanvas?.querySelector('.ad-editable-preview-image');
  if (!image) { syncAdImageEditor(); return; }
  image.style.objectPosition = `${adImageAdjustment.x}% ${adImageAdjustment.y}%`;
  image.style.transform = `scale(${adImageAdjustment.zoom})`;
  image.style.transformOrigin = `${adImageAdjustment.x}% ${adImageAdjustment.y}%`;
  image.classList.toggle('is-adjusted', adImageAdjustment.dirty || adImageAdjustment.zoom !== 1 || adImageAdjustment.x !== 50 || adImageAdjustment.y !== 50);
  syncAdImageEditor();
}

function bindAdImageDrag() {
  const image = adLivePreviewCanvas?.querySelector('.ad-editable-preview-image');
  if (!image) return;
  image.setAttribute('draggable', 'false');
  let dragging = false;
  let startClientX = 0;
  let startClientY = 0;
  let startX = 50;
  let startY = 50;
  const stop = () => { dragging = false; image.classList.remove('is-dragging'); };
  image.addEventListener('pointerdown', (event) => {
    if (event.button !== undefined && event.button !== 0) return;
    dragging = true;
    startClientX = event.clientX;
    startClientY = event.clientY;
    startX = adImageAdjustment.x;
    startY = adImageAdjustment.y;
    image.classList.add('is-dragging');
    try { image.setPointerCapture(event.pointerId); } catch (_) {}
    event.preventDefault();
  });
  image.addEventListener('pointermove', (event) => {
    if (!dragging) return;
    const rect = image.parentElement?.getBoundingClientRect();
    if (!rect?.width || !rect?.height) return;
    const factor = Math.max(1, adImageAdjustment.zoom);
    adImageAdjustment.x = clampAdImageValue(startX - ((event.clientX - startClientX) / rect.width) * 100 / factor, 0, 100);
    adImageAdjustment.y = clampAdImageValue(startY - ((event.clientY - startClientY) / rect.height) * 100 / factor, 0, 100);
    adImageAdjustment.dirty = true;
    applyAdImageAdjustmentToPreview();
    event.preventDefault();
  });
  image.addEventListener('pointerup', stop);
  image.addEventListener('pointercancel', stop);
}

function loadAdEditorImage(source) {
  return new Promise((resolve, reject) => {
    if (!source) { reject(new Error('Chưa có ảnh để chỉnh')); return; }
    const image = new Image();
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error('Không thể đọc ảnh quảng cáo để chỉnh'));
    image.src = source;
  });
}

function canvasToAdBlob(canvas) {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => blob ? resolve(blob) : reject(new Error('Không thể tạo ảnh quảng cáo đã chỉnh')), 'image/webp', 0.92);
  });
}

function blobToDataURL(blob) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(typeof reader.result === 'string' ? reader.result : '');
    reader.onerror = () => reject(new Error('Không thể đọc ảnh đã chỉnh'));
    reader.readAsDataURL(blob);
  });
}

async function applyAdjustedAdImage({ silent = false } = {}) {
  const source = adImageAdjustment.baseSource || currentAdImagePreview;
  if (!source) throw new Error('Chưa có ảnh để chỉnh');
  const image = await loadAdEditorImage(source);
  const spec = adCropSpec(adPositionInput?.value || 'home-right');
  const canvas = document.createElement('canvas');
  canvas.width = spec.width; canvas.height = spec.height;
  const ctx = canvas.getContext('2d');
  if (!ctx) throw new Error('Trình duyệt không hỗ trợ chỉnh ảnh');
  const fit = Math.max(spec.width / image.naturalWidth, spec.height / image.naturalHeight);
  const scale = fit * adImageAdjustment.zoom;
  const drawWidth = image.naturalWidth * scale;
  const drawHeight = image.naturalHeight * scale;
  const overflowX = Math.max(0, drawWidth - spec.width);
  const overflowY = Math.max(0, drawHeight - spec.height);
  const drawX = -overflowX * (adImageAdjustment.x / 100);
  const drawY = -overflowY * (adImageAdjustment.y / 100);
  ctx.drawImage(image, drawX, drawY, drawWidth, drawHeight);
  const blob = await canvasToAdBlob(canvas);
  adjustedAdImageBlob = blob;
  currentAdImagePreview = await blobToDataURL(blob);
  adImageAdjustment.baseSource = currentAdImagePreview;
  adImageAdjustment.zoom = 1; adImageAdjustment.x = 50; adImageAdjustment.y = 50; adImageAdjustment.dirty = false;
  if (adImageZoom) adImageZoom.value = '1';
  renderLiveAdPreview();
  syncAdImageEditor('Đã áp dụng khung ảnh mới · ảnh này sẽ được tải lên khi bạn bấm Lưu.');
  if (!silent) showToast('Đã áp dụng ảnh đã chỉnh');
  return blob;
}

async function postAdFormJSON(form) {
  if (adImageAdjustment.dirty) await applyAdjustedAdImage({ silent: true });
  const formData = new FormData(form);
  if (adjustedAdImageBlob) {
    formData.delete('ad_image');
    formData.append('ad_image', adjustedAdImageBlob, 'quang-cao-da-chinh.webp');
  }
  const response = await fetch(form.action, {
    method: 'POST',
    body: formData,
    headers: { 'X-Requested-With': 'XMLHttpRequest', 'Accept': 'application/json' }
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok || !data.ok) throw new Error(data.message || 'Không thể lưu quảng cáo');
  return data;
}

function setAdPosition(position) {
  const value = positionLabels[position] ? position : 'home-right';
  if (adPositionInput) adPositionInput.value = value;
  adPositionPicker?.querySelectorAll('.ad-position-option').forEach((button) => button.classList.toggle('active', button.dataset.position === value));
  renderLiveAdPreview();
  syncAdImageEditor();
}

adPositionPicker?.querySelectorAll('.ad-position-option').forEach((button) => {
  button.addEventListener('click', () => setAdPosition(button.dataset.position));
});

function resetAdForm() {
  if (!adForm) return;
  adForm.reset();
  document.getElementById('adId').value = '';
  document.getElementById('adOrder').value = '10';
  document.getElementById('adActive').checked = true;
  if (adImageInput) adImageInput.value = '';
  currentAdImagePreview = '';
  resetAdImageAdjustment('');
  if (adFormTitle) adFormTitle.textContent = 'Thêm quảng cáo';
  if (adFormHint) adFormHint.textContent = 'Tạo quảng cáo mới. Khi bấm Sửa ở danh sách, form này sẽ nạp lại toàn bộ dữ liệu quảng cáo.';
  if (adSaveButton) adSaveButton.textContent = 'Lưu quảng cáo';
  if (adImageEditNote) adImageEditNote.classList.remove('editing');
  setAdPosition('home-right');
}

function openModal(modal) { modal?.classList.add('open'); document.body.classList.add('modal-lock'); }
function closeModal(modal) { modal?.classList.remove('open'); syncAdminModalLock(); }

function escapeHTML(value) {
  return String(value ?? '').replace(/[&<>'"]/g, (char) => ({'&':'&amp;','<':'&lt;','>':'&gt;',"'":'&#39;','"':'&quot;'}[char]));
}

function adPreviewMarkup(data) {
  const imagePart = data.image ? `<div class="ad-media"><img class="ad-editable-preview-image" src="${escapeHTML(data.image)}" alt="${escapeHTML(data.title || 'Ảnh quảng cáo')}"></div>` : '<div class="ad-media ad-media-empty">ẢNH QUẢNG CÁO</div>';
  const safeTitle = escapeHTML(data.title || 'Tiêu đề quảng cáo');
  const safeDescription = escapeHTML(data.description || 'Mô tả quảng cáo sẽ hiển thị tại đây.');
  const link = escapeHTML(data.link || '#');
  const position = data.position || 'home-right';
  let cardClass = 'preview-ad-card preview-ad-home-right';
  if (position === 'feed-sponsored') cardClass = 'preview-ad-card preview-ad-feed';
  if (position === 'post-left' || position === 'post-right') cardClass = 'preview-ad-card preview-ad-sidebar';
  return `<div class="preview-slot-shell ${position}"><a class="${cardClass}" href="${link}" target="_blank" rel="noopener">${imagePart}<div class="ad-copy"><span class="ad-chip">Quảng cáo</span><strong>${safeTitle}</strong><p>${safeDescription}</p><span class="ad-cta">Xem thêm</span></div></a></div>`;
}

function renderLiveAdPreview() {
  if (!adLivePreviewCanvas) return;
  const data = previewDataFromForm();
  adLivePreviewCanvas.innerHTML = adPreviewMarkup(data);
  if (adLivePreviewPosition) adLivePreviewPosition.textContent = positionLabels[data.position] || data.position;
  applyAdImageAdjustmentToPreview();
  bindAdImageDrag();
}

function renderAdPreview(data) {
  if (!adPreviewCanvas) return;
  adPreviewCanvas.innerHTML = adPreviewMarkup(data);
  const position = data.position || 'home-right';
  if (adPreviewTitle) adPreviewTitle.textContent = 'Xem trước quảng cáo · ' + (data.title || 'Quảng cáo');
  if (adPreviewPositionLabel) adPreviewPositionLabel.textContent = positionLabels[position] || position;
  openModal(adPreviewModal);
}

function previewDataFromForm() {
  return {
    title: document.getElementById('adTitle').value.trim(),
    description: document.getElementById('adDescription').value.trim(),
    position: adPositionInput?.value || 'home-right',
    link: document.getElementById('adLink').value.trim(),
    image: currentAdImagePreview || ''
  };
}

function previewAdFromForm() {
  renderAdPreview(previewDataFromForm());
}

adImageInput?.addEventListener('change', () => {
  if (!adImageInput.files?.[0]) return;
  const reader = new FileReader();
  reader.onload = () => {
    currentAdImagePreview = typeof reader.result === 'string' ? reader.result : '';
    resetAdImageAdjustment(currentAdImagePreview);
    renderLiveAdPreview();
    syncAdImageEditor('Ảnh mới đã sẵn sàng · kéo ảnh hoặc thu phóng nếu cần.');
  };
  reader.readAsDataURL(adImageInput.files[0]);
});
adImageZoom?.addEventListener('input', () => {
  adImageAdjustment.zoom = Number(adImageZoom.value) || 1;
  adImageAdjustment.dirty = true;
  applyAdImageAdjustmentToPreview();
});
adImageCenter?.addEventListener('click', () => {
  adImageAdjustment.x = 50; adImageAdjustment.y = 50; adImageAdjustment.dirty = true;
  applyAdImageAdjustmentToPreview();
});
adImageResetCrop?.addEventListener('click', () => {
  const original = adImageAdjustment.originalSource || currentAdImagePreview;
  currentAdImagePreview = original || '';
  resetAdImageAdjustment(currentAdImagePreview);
  renderLiveAdPreview();
  syncAdImageEditor('Đã hoàn tác về ảnh trước khi chỉnh.');
});
adImageApplyCrop?.addEventListener('click', async () => {
  if (adImageApplyCrop) adImageApplyCrop.disabled = true;
  try { await applyAdjustedAdImage(); }
  catch (error) { showToast(error.message, true); }
  finally { syncAdImageEditor(); }
});
adLivePreviewCanvas?.addEventListener('click', (event) => { if (event.target.closest('a')) event.preventDefault(); });

adPreviewButton?.addEventListener('click', previewAdFromForm);
['adTitle','adDescription','adLink'].forEach((id) => document.getElementById(id)?.addEventListener('input', renderLiveAdPreview));
document.getElementById('adActive')?.addEventListener('change', renderLiveAdPreview);
adReset?.addEventListener('click', () => { resetAdForm(); renderLiveAdPreview(); });

function adRowHTML(ad, index) {
  const active = ad.Active ?? ad.active;
  const id = ad.ID ?? ad.id;
  const title = ad.Title ?? ad.title ?? '';
  const description = ad.Description ?? ad.description ?? '';
  const position = ad.Position ?? ad.position ?? '';
  const link = ad.LinkURL ?? ad.link_url ?? '';
  const image = ad.ImageURL ?? ad.image_url ?? '';
  return `<tr class="ad-sort-row" draggable="true" data-ad-id="${id}">
    <td class="drag-cell"><span class="ad-drag-handle" title="Kéo để sắp xếp">☰</span></td>
    <td><strong>${escapeHTML(title)}</strong><small>${escapeHTML(description)}</small></td>
    <td><span class="position-code">${escapeHTML(positionLabels[position] || position)}</span></td>
    <td>${active ? '<span class="live-status">Đang bật</span>' : '<span class="muted-status">Đã tắt</span>'}</td>
    <td><span class="order-badge">${(index + 1) * 10}</span></td>
    <td class="actions ad-row-actions">
      <button type="button" class="ghost preview-ad-row" data-title="${escapeHTML(title)}" data-description="${escapeHTML(description)}" data-position="${escapeHTML(position)}" data-link="${escapeHTML(link)}" data-image="${escapeHTML(image)}">Xem trước</button>
      <button type="button" class="ghost edit-ad" data-id="${id}" data-title="${escapeHTML(title)}" data-description="${escapeHTML(description)}" data-position="${escapeHTML(position)}" data-link="${escapeHTML(link)}" data-order="${(index + 1) * 10}" data-active="${active}" data-image="${escapeHTML(image)}">Sửa</button>
      <form class="ajax-delete-ad" method="post" action="/admin/ad/delete"><input type="hidden" name="id" value="${id}"><button class="danger-link">Xóa</button></form>
    </td></tr>`;
}

function renderAdTable(ads) {
  if (!adSortable) return;
  adSortable.innerHTML = ads?.length ? ads.map(adRowHTML).join('') : '<tr class="ad-empty-row"><td colspan="6">Chưa có quảng cáo.</td></tr>';
  setupAdSorting();
}

adForm?.addEventListener('submit', async (event) => {
  event.preventDefault();
  const submit = adForm.querySelector('button[type="submit"], button.primary');
  if (submit) submit.disabled = true;
  try {
    const data = await postAdFormJSON(adForm);
    renderAdTable(data.ads || []);
    showToast(data.message);
    resetAdForm();
  } catch (error) { showToast(error.message, true); }
  finally { if (submit) submit.disabled = false; }
});

document.addEventListener('click', async (event) => {
  const edit = event.target.closest('.edit-ad');
  if (edit) {
    document.getElementById('adId').value = edit.dataset.id;
    document.getElementById('adTitle').value = edit.dataset.title;
    document.getElementById('adDescription').value = edit.dataset.description;
    setAdPosition(edit.dataset.position);
    document.getElementById('adLink').value = edit.dataset.link;
    document.getElementById('adOrder').value = edit.dataset.order || '10';
    document.getElementById('adActive').checked = edit.dataset.active === 'true';
    currentAdImagePreview = edit.dataset.image || '';
    resetAdImageAdjustment(currentAdImagePreview);
    if (adImageInput) adImageInput.value = '';
    if (adFormTitle) adFormTitle.textContent = `Chỉnh sửa quảng cáo #${edit.dataset.id}`;
    if (adFormHint) adFormHint.textContent = 'Bạn có thể sửa tiêu đề, nội dung, ảnh, liên kết, vị trí và trạng thái. Không chọn ảnh mới thì hệ thống giữ ảnh hiện tại.';
    if (adSaveButton) adSaveButton.textContent = 'Lưu thay đổi';
    if (adImageEditNote) adImageEditNote.classList.add('editing');
    renderLiveAdPreview();
    switchTab('ads'); window.scrollTo({ top: 0, behavior: 'smooth' });
    return;
  }
  const preview = event.target.closest('.preview-ad-row');
  if (preview) {
    renderAdPreview({ title: preview.dataset.title, description: preview.dataset.description, position: preview.dataset.position, link: preview.dataset.link, image: preview.dataset.image || '' });
  }
});

document.addEventListener('submit', async (event) => {
  const form = event.target.closest('.ajax-delete-ad');
  if (!form) return;
  event.preventDefault();
  if (!confirm('Xóa quảng cáo này?')) return;
  try { const data = await postFormJSON(form); renderAdTable(data.ads || []); showToast(data.message); }
  catch (error) { showToast(error.message, true); }
});

let draggedCategoryRow = null;
const categorySortable = document.querySelector('.category-sort-row')?.closest('tbody');
async function saveCategoryOrder() {
  if (!categorySortable) return;
  const ids = [...categorySortable.querySelectorAll('.category-sort-row')].map((item) => Number(item.dataset.categoryId)).filter(Boolean);
  const response = await fetch('/admin/category/reorder', {
    method: 'POST',
    headers: {'Content-Type':'application/json','X-Requested-With':'XMLHttpRequest','Accept':'application/json'},
    body: JSON.stringify({ ids })
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok || !data.ok) throw new Error(data.message || 'Không thể lưu thứ tự chuyên mục');
  showToast(data.message || 'Đã lưu thứ tự chuyên mục');
}
function setupCategorySorting() {
  if (!categorySortable) return;
  categorySortable.querySelectorAll('.category-sort-row').forEach((row) => {
    const dragCell = row.querySelector('.category-drag-cell');
    const dragHandle = row.querySelector('.category-drag-handle');
    let categoryDragArmed = false;
    dragHandle?.addEventListener('pointerdown', () => { categoryDragArmed = true; });
    document.addEventListener('pointerup', () => { categoryDragArmed = false; }, { passive: true });
    if (dragCell && !dragCell.querySelector('.category-move-buttons')) {
      const controls = document.createElement('span');
      controls.className = 'category-move-buttons';
      controls.innerHTML = '<button type="button" class="category-move-up" title="Đưa chuyên mục lên">↑</button><button type="button" class="category-move-down" title="Đưa chuyên mục xuống">↓</button>';
      dragCell.appendChild(controls);
      controls.querySelector('.category-move-up').addEventListener('click', async () => { const prev=row.previousElementSibling; if(prev){categorySortable.insertBefore(row,prev); try{await saveCategoryOrder()}catch(e){showToast(e.message,true)}} });
      controls.querySelector('.category-move-down').addEventListener('click', async () => { const next=row.nextElementSibling; if(next){categorySortable.insertBefore(next,row); try{await saveCategoryOrder()}catch(e){showToast(e.message,true)}} });
    }
    row.addEventListener('dragstart', (event) => {
      if (!categoryDragArmed) { event.preventDefault(); return; }
      draggedCategoryRow = row;
      row.classList.add('dragging');
      event.dataTransfer.effectAllowed = 'move';
    });
    row.addEventListener('dragover', (event) => {
      event.preventDefault();
      if (!draggedCategoryRow || draggedCategoryRow === row) return;
      const rect = row.getBoundingClientRect();
      const after = event.clientY > rect.top + rect.height / 2;
      categorySortable.insertBefore(draggedCategoryRow, after ? row.nextSibling : row);
    });
    row.addEventListener('dragend', async () => {
      row.classList.remove('dragging');
      draggedCategoryRow = null;
      categoryDragArmed = false;
      try { await saveCategoryOrder(); } catch (error) { showToast(error.message, true); }
    });
  });
}
setupCategorySorting();

let draggedAdRow = null;
function setupAdSorting() {
  if (!adSortable) return;
  adSortable.querySelectorAll('.ad-sort-row').forEach((row) => {
    row.addEventListener('dragstart', (event) => { draggedAdRow = row; row.classList.add('dragging'); event.dataTransfer.effectAllowed = 'move'; });
    row.addEventListener('dragend', async () => {
      row.classList.remove('dragging'); draggedAdRow = null;
      const ids = [...adSortable.querySelectorAll('.ad-sort-row')].map((item) => Number(item.dataset.adId)).filter(Boolean);
      try {
        const response = await fetch('/admin/ad/reorder', { method: 'POST', headers: {'Content-Type':'application/json','X-Requested-With':'XMLHttpRequest','Accept':'application/json'}, body: JSON.stringify({ ids }) });
        const data = await response.json().catch(() => ({}));
        if (!response.ok || !data.ok) throw new Error(data.message || 'Không thể lưu thứ tự');
        renderAdTable(data.ads || []); showToast(data.message);
      } catch (error) { showToast(error.message, true); }
    });
    row.addEventListener('dragover', (event) => {
      event.preventDefault(); if (!draggedAdRow || draggedAdRow === row) return;
      const rect = row.getBoundingClientRect(); const after = event.clientY > rect.top + rect.height / 2;
      adSortable.insertBefore(draggedAdRow, after ? row.nextSibling : row);
    });
  });
}
setupAdSorting();
renderLiveAdPreview();

document.getElementById('closeAdPreviewModal')?.addEventListener('click', () => closeModal(adPreviewModal));
document.getElementById('closeAdPreviewButton')?.addEventListener('click', () => closeModal(adPreviewModal));

document.querySelectorAll('.modal').forEach((modal) => {
  modal.addEventListener('click', (event) => { if (event.target === modal) { modal.classList.remove('open'); syncAdminModalLock(); } });
});

document.addEventListener('keydown', (event) => {
  if (event.key !== 'Escape') return;
  if (previewModal?.classList.contains('open')) { closePreviewModal(); return; }
  if (adPreviewModal?.classList.contains('open')) { closeModal(adPreviewModal); return; }
  if (categoryQuickModal?.classList.contains('open')) { closeCategoryQuickEditor(); return; }
  if (document.getElementById('configModal')?.classList.contains('open')) { closeConfigForm(); return; }
  if (newCategoryModal?.classList.contains('open')) closeNewCategoryModal();
});

// DACS_V1.24.20 · verification actions use native POST + redirect.
// This intentionally avoids depending on event.submitter/AJAX so Approve/Reject works reliably.

// DACS_V1.21.7 · account moderation actions
function syncAdminModalLock() {
  const hasOpen = Boolean(document.querySelector('.modal.open'));
  document.body.classList.toggle('modal-lock', hasOpen);
}

document.querySelectorAll('.account-lock-trigger').forEach((button) => {
  button.addEventListener('click', (event) => {
    event.stopPropagation();
    if (button.disabled) return;
    const menu = button.closest('.account-lock-menu');
    const willOpen = !menu?.classList.contains('open');
    document.querySelectorAll('.account-lock-menu.open').forEach((item) => item.classList.remove('open'));
    if (willOpen) menu?.classList.add('open');
  });
});

document.addEventListener('click', (event) => {
  if (event.target.closest('.account-lock-menu')) return;
  document.querySelectorAll('.account-lock-menu.open').forEach((item) => item.classList.remove('open'));
});


// DACS_V1.22.2 · Require at least one document format for document-card categories.
document.addEventListener('submit', (event) => {
  const form = event.target;
  if (!(form instanceof HTMLFormElement)) return;
  if (!form.matches('#newCategoryForm,#configForm')) return;
  const style = form.querySelector('input[name="post_card_style"]:checked')?.value || 'normal';
  if (style !== 'document') return;
  if (!form.querySelector('input[name="document_formats"]:checked')) {
    event.preventDefault();
    event.stopImmediatePropagation();
    showToast('Hãy chọn ít nhất một định dạng tài liệu được phép đăng.', true);
  }
}, true);
