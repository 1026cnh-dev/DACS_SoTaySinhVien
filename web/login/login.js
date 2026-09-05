document.addEventListener('DOMContentLoaded', () => {
  setupAuthSwitcher();
  setupPasswordToggles();
  setupPhoneAccountCheck();
});

function setupAuthSwitcher() {
  const body = document.body;
  const tabs = [...document.querySelectorAll('[data-auth-mode]')];
  const views = [...document.querySelectorAll('[data-auth-view]')];
  let switchTimer = 0;
  const stage = document.querySelector('.auth-stage');

  const currentMode = () => body.classList.contains('auth-mode-register') ? 'register' : 'login';
  const syncStageHeight = (mode = currentMode(), animate = true) => {
    if (!stage) return;
    const view = views.find((item) => item.dataset.authView === mode);
    if (!view) return;
    if (!animate) stage.style.transition = 'none';
    stage.style.height = Math.ceil(view.scrollHeight) + 'px';
    if (!animate) requestAnimationFrame(() => { stage.style.transition = ''; });
  };

  syncStageHeight(currentMode(), false);
  if ('ResizeObserver' in window) {
    const observer = new ResizeObserver(() => syncStageHeight(currentMode()));
    views.forEach((view) => observer.observe(view));
  }
  window.addEventListener('resize', () => syncStageHeight(currentMode(), false));

  const setMode = (mode, options = {}) => {
    if (mode !== 'login' && mode !== 'register') return;
    const previous = currentMode();
    if (previous === mode && !options.force) return;

    window.clearTimeout(switchTimer);
    body.classList.add('auth-switching');
    body.classList.toggle('auth-mode-login', mode === 'login');
    body.classList.toggle('auth-mode-register', mode === 'register');

    tabs.forEach((item) => {
      const active = item.dataset.authMode === mode;
      if (item.classList.contains('auth-tab')) {
        item.classList.toggle('active', active);
        item.setAttribute('aria-selected', active ? 'true' : 'false');
      }
    });
    views.forEach((view) => view.setAttribute('aria-hidden', view.dataset.authView === mode ? 'false' : 'true'));
    syncStageHeight(mode);

    document.querySelectorAll('.alert').forEach((alert) => { alert.hidden = true; });
    const googleError = document.getElementById('googleError');
    if (googleError) googleError.hidden = true;

    if (options.updateHistory !== false) {
      const link = tabs.find((item) => item.dataset.authMode === mode && item.getAttribute('href'));
      if (link) history.pushState({authMode: mode}, '', link.getAttribute('href'));
    }

    switchTimer = window.setTimeout(() => {
      syncStageHeight(mode);
      body.classList.remove('auth-switching');
      const focusTarget = mode === 'register' ? document.getElementById('phone') : document.getElementById('login');
      if (options.focus !== false && focusTarget) focusTarget.focus({preventScroll:true});
    }, 390);
  };

  tabs.forEach((link) => {
    link.addEventListener('click', (event) => {
      if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
      const mode = link.dataset.authMode;
      if (!mode) return;
      event.preventDefault();
      setMode(mode);
    });
  });

  window.addEventListener('popstate', () => {
    const mode = location.pathname.startsWith('/register') ? 'register' : 'login';
    setMode(mode, {updateHistory:false, focus:false, force:true});
  });
}

function setupPasswordToggles() {
  document.querySelectorAll('[data-password-target]').forEach((toggle) => {
    const input = document.getElementById(toggle.dataset.passwordTarget || '');
    if (!input) return;
    toggle.addEventListener('click', () => {
      const hidden = input.type === 'password';
      input.type = hidden ? 'text' : 'password';
      toggle.setAttribute('aria-label', hidden ? 'Ẩn mật khẩu' : 'Hiện mật khẩu');
      const icon = toggle.querySelector('img');
      if (icon) icon.src = hidden ? '/web/icons/eye-off.svg' : '/web/icons/eye.svg';
    });
  });
}

async function handleGoogleCredential(response) {
  const errorBox = document.getElementById('googleError');
  try {
    const res = await fetch('/auth/google', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({credential: response.credential, next: (document.getElementById('googleWrap')?.dataset.next || '')})
    });
    const data = await res.json();
    if (!res.ok || !data.ok) throw new Error(data.message || 'Không thể đăng nhập bằng Google');
    window.location.href = data.redirect || '/';
  } catch (err) {
    if (errorBox) {
      errorBox.hidden = false;
      errorBox.textContent = err.message;
    }
  }
}
window.handleGoogleCredential = handleGoogleCredential;

function normalizePhoneForCheck(value) {
  const digits = String(value || '').replace(/\D/g, '');
  if (digits.startsWith('84') && digits.length >= 11) return '0' + digits.slice(2);
  return digits;
}

function setupPhoneAccountCheck() {
  const registerPhone = document.getElementById('phone');
  const registerStatus = document.getElementById('phoneAccountStatus');
  const registerSubmit = document.querySelector('form[action="/register"] .submit-btn');
  const loginInput = document.getElementById('login');
  const loginStatus = document.getElementById('phoneLoginStatus');
  let registerTimer = 0;
  let loginTimer = 0;

  const renderStatus = (box, state, message) => {
    if (!box) return;
    box.hidden = !message;
    box.className = 'phone-account-status' + (state ? ' ' + state : '');
    box.textContent = message || '';
    requestAnimationFrame(() => {
      const mode = document.body.classList.contains('auth-mode-register') ? 'register' : 'login';
      const stage = document.querySelector('.auth-stage');
      const view = document.querySelector(`[data-auth-view="${mode}"]`);
      if (stage && view) stage.style.height = Math.ceil(view.scrollHeight) + 'px';
    });
  };

  const check = async (phone, box, mode) => {
    try {
      const response = await fetch('/auth/phone/check?phone=' + encodeURIComponent(phone), {headers:{'Accept':'application/json'}});
      const data = await response.json().catch(() => ({}));
      if (!response.ok || !data.ok) {
        renderStatus(box, 'invalid', data.message || 'Số điện thoại chưa hợp lệ');
        if (mode === 'register' && registerSubmit) registerSubmit.disabled = true;
        return;
      }
      if (data.exists) {
        renderStatus(box, 'exists', mode === 'register' ? 'Đã có tài khoản với số này — hãy chuyển sang Đăng nhập.' : 'Đã tìm thấy tài khoản. Bạn có thể nhập mật khẩu.');
        if (mode === 'register' && registerSubmit) registerSubmit.disabled = true;
      } else {
        renderStatus(box, mode === 'register' ? 'available' : 'not-found', mode === 'register' ? 'Số điện thoại chưa được sử dụng, có thể tạo tài khoản.' : 'Chưa có tài khoản với số điện thoại này.');
        if (mode === 'register' && registerSubmit) registerSubmit.disabled = false;
      }
    } catch (_) {
      renderStatus(box, '', '');
      if (mode === 'register' && registerSubmit) registerSubmit.disabled = false;
    }
  };

  registerPhone?.addEventListener('input', () => {
    clearTimeout(registerTimer);
    const phone = normalizePhoneForCheck(registerPhone.value);
    if (registerSubmit) registerSubmit.disabled = false;
    if (phone.length < 9) { renderStatus(registerStatus, '', ''); return; }
    renderStatus(registerStatus, 'checking', 'Đang kiểm tra tài khoản...');
    registerTimer = window.setTimeout(() => check(phone, registerStatus, 'register'), 320);
  });

  loginInput?.addEventListener('input', () => {
    clearTimeout(loginTimer);
    const raw = loginInput.value.trim();
    const phone = normalizePhoneForCheck(raw);
    if (!/^\+?[0-9 .()-]+$/.test(raw) || phone.length < 9) { renderStatus(loginStatus, '', ''); return; }
    renderStatus(loginStatus, 'checking', 'Đang kiểm tra số điện thoại...');
    loginTimer = window.setTimeout(() => check(phone, loginStatus, 'login'), 320);
  });
}
