// Nếu đã đăng nhập, đẩy thẳng vào Sổ tay
if (localStorage.getItem('user')) {
    window.location.replace('../student_handbook.html');
}

const signUpButton = document.getElementById('signUp');
const signInButton = document.getElementById('signIn');
const container = document.getElementById('container');

signUpButton.addEventListener('click', () => { container.classList.add("right-panel-active"); hideAllErrors(); });
signInButton.addEventListener('click', () => { container.classList.remove("right-panel-active"); hideAllErrors(); });

const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');
const loginError = document.getElementById('login-error-msg');
const regError = document.getElementById('reg-error-msg');

function showError(element, msg) { element.textContent = msg; element.style.display = 'block'; }
function hideAllErrors() { loginError.style.display = 'none'; regError.style.display = 'none'; }

// HÀM ĐÃ ĐƯỢC CẬP NHẬT: Cho phép dùng mọi loại email hợp lệ
function isValidEmail(email) { 
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return re.test(email.trim());
}

function toggleButtonState(button, isLoading, defaultText) {
    if (isLoading) {
        button.disabled = true; button.textContent = 'Đang xử lý...'; button.style.opacity = '0.7'; button.style.cursor = 'not-allowed';
    } else {
        button.disabled = false; button.textContent = defaultText; button.style.opacity = '1'; button.style.cursor = 'pointer';
    }
}

// Xử lý Đăng Nhập
loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    hideAllErrors();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    const submitBtn = e.target.querySelector('.btn-submit');

    if (!isValidEmail(email)) { showError(loginError, 'Vui lòng nhập định dạng email hợp lệ!'); return; }

    toggleButtonState(submitBtn, true); 
    try {
        const res = await fetch('/api/auth/login', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (res.ok) {
            localStorage.setItem('user', JSON.stringify(data));
            window.location.replace('../student_handbook.html'); 
        } else {
            showError(loginError, data.error || 'Sai tài khoản hoặc mật khẩu!');
        }
    } catch (err) { showError(loginError, 'Lỗi kết nối máy chủ!'); } 
    finally { toggleButtonState(submitBtn, false, 'Đăng Nhập'); }
});

// Xử lý Đăng Ký
registerForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    hideAllErrors();
    const name = document.getElementById('reg-name').value;
    const email = document.getElementById('reg-email').value;
    const password = document.getElementById('reg-password').value;
    const submitBtn = e.target.querySelector('.btn-submit');

    if (!isValidEmail(email)) { showError(regError, 'Vui lòng nhập định dạng email hợp lệ!'); return; }

    toggleButtonState(submitBtn, true); 
    try {
        const res = await fetch('/api/auth/register', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password, full_name: name })
        });
        const data = await res.json();
        if (res.ok) {
            alert('Đăng ký thành công!');
            signInButton.click(); 
            registerForm.reset(); 
        } else { showError(regError, data.error || 'Có lỗi xảy ra khi đăng ký!'); }
    } catch (err) { showError(regError, 'Lỗi kết nối máy chủ!'); } 
    finally { toggleButtonState(submitBtn, false, 'Đăng Ký'); }
});

// Xử lý Google
async function handleGoogleLogin(response) {
    hideAllErrors();
    const jwtToken = response.credential;
    try {
        const res = await fetch('/api/auth/google', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ credential: jwtToken })
        });
        const data = await res.json();
        if (res.ok) {
            localStorage.setItem('user', JSON.stringify(data));
            window.location.replace('../student_handbook.html');
        } else { showError(loginError, data.error || 'Lỗi đăng nhập Google.'); }
    } catch(err) { showError(loginError, 'Lỗi kết nối máy chủ!'); }
}