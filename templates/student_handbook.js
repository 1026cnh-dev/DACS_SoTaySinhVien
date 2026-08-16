document.addEventListener('DOMContentLoaded', () => {
    const categoryList = document.getElementById('category-list');
    const postList = document.getElementById('post-list');
    const postDetail = document.getElementById('post-detail');
    const searchInput = document.getElementById('search-input');
    const viewList = document.getElementById('view-list');
    const viewDetail = document.getElementById('view-detail');
    const btnBack = document.getElementById('btn-back');
    const currentCategoryTitle = document.getElementById('current-category-title');

    let categories = [];
    let posts = [];
    let selectedCategory = '';
    let searchTerm = '';

    init();
    async function init() {
        await loadCategories();
        await loadPosts();
        checkAuthStatus(); 
    }

    async function loadCategories() {
        try {
            const res = await fetch('/api/categories');
            if(res.ok) {
                categories = await res.json();
                renderCategories();
            }
        } catch (e) { console.error("Lỗi tải chuyên mục", e); }
    }

    function renderCategories() {
        let html = `<li class="category-item active" data-id="">Tất cả bài viết</li>`;
        categories.forEach(cat => {
            html += `<li class="category-item" data-id="${cat.id}">${cat.name}</li>`;
        });
        categoryList.innerHTML = html;

        document.querySelectorAll('.category-item').forEach(item => {
            item.addEventListener('click', (e) => {
                document.querySelectorAll('.category-item').forEach(i => i.classList.remove('active'));
                e.target.classList.add('active');
                selectedCategory = e.target.dataset.id;
                currentCategoryTitle.textContent = e.target.textContent;
                loadPosts();
                switchToView('list');
            });
        });
    }

    async function loadPosts() {
        let url = '/api/posts';
        const params = new URLSearchParams();
        if (selectedCategory) params.append('category_id', selectedCategory); 
        if (searchTerm) params.append('q', searchTerm);
        if(params.toString()) url += '?' + params.toString();

        try {
            const res = await fetch(url);
            if(res.ok) {
                posts = await res.json();
                renderPosts();
            }
        } catch (e) { console.error("Lỗi tải bài viết", e); }
    }

    function renderPosts() {
        if (!posts || posts.length === 0) {
            postList.innerHTML = `
                <div style="grid-column: 1/-1; text-align: center; color: var(--text-muted); padding: 3rem 0;">
                    <i class="fa-regular fa-folder-open fa-3x" style="margin-bottom:1rem;"></i><br>
                    Không có bài viết nào phù hợp.
                </div>`;
            return;
        }

        postList.innerHTML = posts.map(post => `
            <article class="post-card">
                <h3>${post.title}</h3>
                <p>${post.content.substring(0, 80)}...</p>
                <button class="btn-read" data-id="${post.id}">
                    Đọc chi tiết <i class="fa-solid fa-arrow-right" style="font-size:0.8em; margin-left:4px;"></i>
                </button>
            </article>
        `).join('');

        document.querySelectorAll('.btn-read').forEach(btn => {
            btn.addEventListener('click', () => showDetail(btn.dataset.id));
        });
    }

    async function showDetail(id) {
        try {
            const res = await fetch(`/api/posts?id=${id}`); 
            if(res.ok) {
                const post = await res.json();
                postDetail.innerHTML = `
                    <h2>${post.title}</h2>
                    <div class="post-meta"><i class="fa-regular fa-clock"></i> Cập nhật gần đây</div>
                    <div class="post-content">${post.content}</div>
                `;
                switchToView('detail');
            }
        } catch(e) { console.error("Lỗi tải chi tiết", e); }
    }

    function switchToView(view) {
        if (view === 'list') {
            viewList.classList.add('active');
            viewDetail.classList.remove('active');
        } else {
            viewList.classList.remove('active');
            viewDetail.classList.add('active');
            window.scrollTo({top: 0, behavior: 'smooth'}); 
        }
    }

    searchInput.addEventListener('input', (e) => {
        searchTerm = e.target.value.trim();
    });
    btnBack.addEventListener('click', () => { switchToView('list'); });

    // ==========================================
    // LOGIC TÀI KHOẢN 
    // ==========================================
    const postModal = document.getElementById('post-modal');
    const createPostForm = document.getElementById('create-post-form');

    function checkAuthStatus() {
        const user = JSON.parse(localStorage.getItem('user'));
        if (user) {
            document.getElementById('btn-login-page').style.display = 'none';
            document.getElementById('user-profile').style.display = 'flex';
            document.getElementById('display-name').textContent = `Chào, ${user.full_name}`;
            
            if(user.role === 'admin') {
                document.getElementById('btn-show-create').style.display = 'block';
            }
        } else {
            document.getElementById('btn-login-page').style.display = 'block';
            document.getElementById('user-profile').style.display = 'none';
            document.getElementById('btn-show-create').style.display = 'none';
        }
    }

    // Nút chuyển trang đăng nhập
    document.getElementById('btn-login-page').onclick = () => {
        window.location.href = 'auth.html';
    };

    document.getElementById('btn-show-create').onclick = () => {
        postModal.classList.add('active');
        document.getElementById('post-category').innerHTML = categories.map(c => `<option value="${c.id}">${c.name}</option>`).join('');
    };
    document.getElementById('close-post').onclick = () => postModal.classList.remove('active');

    document.getElementById('btn-logout').onclick = () => {
        localStorage.removeItem('user');
        checkAuthStatus();
        window.location.reload();
    };

    createPostForm.onsubmit = async (e) => {
        e.preventDefault();
        const payload = {
            title: document.getElementById('post-title').value,
            category_id: parseInt(document.getElementById('post-category').value),
            content: document.getElementById('post-content').value
        };

        try {
            const res = await fetch('/api/posts', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if(res.ok) {
                alert('Đăng bài thành công!');
                postModal.classList.remove('active');
                createPostForm.reset();
                loadPosts(); 
            }
        } catch (error) { console.error("Lỗi đăng bài", error); }
    };
});
// XỬ LÝ NÚT ĐĂNG XUẤT
const btnLogout = document.getElementById('btnLogout');
if (btnLogout) {
    btnLogout.addEventListener('click', () => {
        // Xóa sạch dữ liệu phiên đăng nhập
        localStorage.removeItem('user');
        // Quay trở lại trang đăng nhập
        window.location.replace('auth/auth.html');
    });
}