/**
 * DesignAI Dashboard JavaScript - 仪表板功能模块
 */

class DashboardManager {
    constructor() {
        this.portfolioData = [];
        this.minioConfigs = [];
        this.editingMinioConfig = null;
        this.currentSection = 'dashboard';
        
        this.init();
    }

    init() {
        this.bindEvents();
        this.initSidebar();
        
        // 检查认证状态
        if (!AuthManager.checkAuth()) {
            return;
        }
        
        this.updateUserInfo();
        this.checkAdminAccess();
        this.showSection('dashboard');
    }

    bindEvents() {
        // 导航项点击事件
        document.addEventListener('click', (e) => {
            const navItem = e.target.closest('.nav-item[data-section]');
            if (navItem) {
                e.preventDefault();
                const section = navItem.getAttribute('data-section');
                this.showSection(section);
            }
        });

        // 表单提交事件
        this.bindFormEvents();
        
        // 移动端菜单切换
        const mobileMenuBtn = document.querySelector('.mobile-menu-btn');
        if (mobileMenuBtn) {
            mobileMenuBtn.addEventListener('click', () => this.toggleSidebar());
        }

        // 主题切换
        const themeToggle = document.querySelector('.theme-toggle');
        if (themeToggle) {
            themeToggle.addEventListener('click', () => ThemeManager.toggle());
        }

        // 窗口大小变化处理
        window.addEventListener('resize', () => {
            if (window.innerWidth > 768) {
                document.getElementById('sidebar')?.classList.remove('open');
            }
        });
    }

    bindFormEvents() {
        // 创建作品表单
        const createForm = document.getElementById('createPortfolioForm');
        if (createForm) {
            createForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const formData = new FormData(createForm);
                this.handleCreatePortfolio(formData);
            });
        }

        // 编辑作品表单
        const editForm = document.getElementById('editPortfolioForm');
        if (editForm) {
            editForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const formData = new FormData(editForm);
                this.handleUpdatePortfolio(formData);
            });
        }

        // 个人资料表单
        const profileForm = document.getElementById('profileForm');
        if (profileForm) {
            profileForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const formData = new FormData(profileForm);
                this.handleUpdateProfile(formData);
            });
        }

        // MinIO配置表单
        const minioForm = document.getElementById('minioConfigForm');
        if (minioForm) {
            minioForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const formData = new FormData(minioForm);
                this.handleMinioConfigSave(formData);
            });
        }

        // 状态过滤器
        const statusFilter = document.getElementById('statusFilter');
        if (statusFilter) {
            statusFilter.addEventListener('change', (e) => {
                const status = e.target.value;
                const filteredPortfolios = status 
                    ? this.portfolioData.filter(p => p.status === status)
                    : this.portfolioData;
                this.renderPortfolioList('allPortfolios', filteredPortfolios);
            });
        }
    }

    initSidebar() {
        // 设置用户头像点击事件
        const userProfile = document.querySelector('.user-profile');
        if (userProfile) {
            userProfile.addEventListener('click', () => this.showSection('profile'));
        }
    }

    toggleSidebar() {
        const sidebar = document.getElementById('sidebar');
        if (sidebar) {
            sidebar.classList.toggle('open');
        }
    }

    updateUserInfo() {
        const user = AuthManager.getCurrentUser();
        if (!user) return;
        
        const userName = document.getElementById('userName');
        const userStatus = document.getElementById('userStatus');
        const userAvatar = document.getElementById('userAvatar');
        
        if (userName) userName.textContent = user.username;
        if (userStatus) userStatus.textContent = user.role === 'admin' ? '管理员' : '用户';
        if (userAvatar) userAvatar.textContent = user.username.charAt(0).toUpperCase();
    }

    checkAdminAccess() {
        const adminSection = document.getElementById('adminSection');
        if (adminSection) {
            adminSection.style.display = AuthManager.isAdmin() ? 'block' : 'none';
        }
    }

    showSection(sectionName) {
        this.currentSection = sectionName;
        
        // 更新导航项激活状态
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        
        const activeNavItem = document.querySelector(`[data-section="${sectionName}"]`);
        if (activeNavItem) {
            activeNavItem.classList.add('active');
        }

        // 隐藏所有内容区域
        document.querySelectorAll('.content-section').forEach(section => {
            section.style.display = 'none';
        });

        // 显示选定的内容区域
        const targetSection = document.getElementById(`${sectionName}-section`);
        if (targetSection) {
            targetSection.style.display = 'block';
            targetSection.classList.add('fade-in');
        }

        // 更新页面标题
        const titles = {
            'dashboard': '仪表板',
            'my-portfolios': '我的作品',
            'create-portfolio': '创建作品',
            'edit-portfolio': '编辑作品',
            'profile': '个人资料',
            'settings': '设置',
            'user-management': '用户管理',
            'portfolio-review': '作品审核',
            'system-settings': '系统设置',
            'minio-settings': 'MinIO设置'
        };
        
        const pageTitle = document.getElementById('pageTitle');
        if (pageTitle) {
            pageTitle.textContent = titles[sectionName] || '仪表板';
        }

        // 根据页面加载相应数据
        this.loadSectionData(sectionName);

        // 关闭移动端侧边栏
        if (window.innerWidth <= 768) {
            document.getElementById('sidebar')?.classList.remove('open');
        }
    }

    async loadSectionData(sectionName) {
        try {
            switch (sectionName) {
                case 'dashboard':
                    await this.loadDashboardData();
                    break;
                case 'my-portfolios':
                    await this.loadMyPortfolios();
                    break;
                case 'profile':
                    await this.loadProfile();
                    break;
                case 'user-management':
                    if (AuthManager.isAdmin()) {
                        await this.loadUsers();
                    }
                    break;
                case 'portfolio-review':
                    if (AuthManager.isAdmin()) {
                        await this.loadAllPortfolios();
                    }
                    break;
                case 'minio-settings':
                    if (AuthManager.isAdmin()) {
                        await this.loadMinioConfigs();
                    }
                    break;
                case 'system-settings':
                    if (AuthManager.isAdmin()) {
                        await this.loadAdminSettings();
                    }
                    break;
            }
        } catch (error) {
            console.error(`Failed to load section data for ${sectionName}:`, error);
            NotificationManager.error('加载数据失败');
        }
    }

    // 仪表板数据加载
    async loadDashboardData() {
        try {
            await this.loadMyPortfolios();
            
            // 计算统计数据
            const totalPortfolios = this.portfolioData.length;
            const publishedPortfolios = this.portfolioData.filter(p => p.status === 'published').length;
            const draftCount = this.portfolioData.filter(p => p.status === 'draft').length;
            const totalViews = this.portfolioData.reduce((sum, p) => sum + (p.views || 0), 0);
            const totalLikes = this.portfolioData.reduce((sum, p) => sum + (p.likes || 0), 0);

            // 更新统计卡片
            this.updateStatsCards({
                totalPortfolios,
                publishedPortfolios,
                draftCount,
                totalViews,
                totalLikes
            });

            // 显示最近作品
            const recentPortfolios = this.portfolioData.slice(0, 5);
            this.renderPortfolioList('recentPortfolios', recentPortfolios);

            // 生成活动数据
            this.renderRecentActivity();

        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            NotificationManager.error('加载仪表板数据失败');
        }
    }

    updateStatsCards(stats) {
        const elements = {
            totalPortfolios: document.getElementById('totalPortfolios'),
            publishedPortfolios: document.getElementById('publishedPortfolios'),
            draftCount: document.getElementById('draftCount'),
            totalViews: document.getElementById('totalViews'),
            totalLikes: document.getElementById('totalLikes')
        };

        for (const [key, element] of Object.entries(elements)) {
            if (element && stats[key] !== undefined) {
                element.textContent = stats[key];
            }
        }
    }

    // 作品相关方法
    async loadMyPortfolios() {
        try {
            const result = await apiClient.request('/my-portfolios');
            this.portfolioData = result.data || [];
            
            if (this.currentSection === 'my-portfolios') {
                this.renderPortfolioList('allPortfolios', this.portfolioData);
            }
            
            return this.portfolioData;
        } catch (error) {
            console.error('Failed to load portfolios:', error);
            NotificationManager.error('加载作品失败');
            return [];
        }
    }

    renderPortfolioList(containerId, portfolios) {
        const container = document.getElementById(containerId);
        if (!container) return;
        
        if (!portfolios || portfolios.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">🎨</div>
                    <h3 class="empty-title">暂无作品</h3>
                    <p class="empty-description">您还没有创建任何作品</p>
                    <button class="btn btn-primary" onclick="dashboardManager.showSection('create-portfolio')">
                        <span>➕</span>
                        <span>创建第一个作品</span>
                    </button>
                </div>
            `;
            return;
        }

        container.innerHTML = portfolios.map(portfolio => {
            const statusClass = {
                'published': 'status-success',
                'draft': 'status-warning',
                'rejected': 'status-error'
            }[portfolio.status] || 'status-warning';

            const statusText = {
                'published': '已发布',
                'draft': '草稿',
                'rejected': '已拒绝'
            }[portfolio.status] || '草稿';

            return `
                <div class="portfolio-item" onclick="dashboardManager.editPortfolio('${portfolio.id}')">
                    <div class="portfolio-image">
                        ${portfolio.imageUrl ? `<img src="${portfolio.imageUrl}" alt="${portfolio.title}" style="width:100%;height:100%;object-fit:cover;border-radius:10px;">` : '🎨'}
                    </div>
                    <div class="portfolio-info">
                        <h4 class="portfolio-title">${portfolio.title}</h4>
                        <div class="portfolio-meta">
                            <span>📅 ${Utils.formatDate(portfolio.createdAt)}</span>
                            <span>👁️ ${portfolio.views || 0}</span>
                            <span>❤️ ${portfolio.likes || 0}</span>
                            <span class="status-badge ${statusClass}">${statusText}</span>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    }

    renderRecentActivity() {
        const activities = [
            { icon: '🎨', text: '创建了新作品', time: '2小时前' },
            { icon: '❤️', text: '作品获得了新的点赞', time: '5小时前' },
            { icon: '👁️', text: '作品浏览量增加', time: '1天前' },
            { icon: '✅', text: '作品通过审核并发布', time: '2天前' }
        ];

        const container = document.getElementById('recentActivity');
        if (container) {
            container.innerHTML = activities.map(activity => `
                <div class="activity-item">
                    <div class="activity-icon">${activity.icon}</div>
                    <div class="activity-content">
                        <div class="activity-text">${activity.text}</div>
                        <div class="activity-time">${activity.time}</div>
                    </div>
                </div>
            `).join('');
        }
    }

    // 表单处理方法
    async handleCreatePortfolio(formData) {
        try {
            const portfolioData = {
                title: formData.get('title'),
                author: AuthManager.getCurrentUser().username,
                description: formData.get('description'),
                content: formData.get('content'),
                category: formData.get('category'),
                tags: formData.get('tags').split(',').map(tag => tag.trim()).filter(tag => tag),
                imageUrl: formData.get('imageUrl'),
                aiLevel: formData.get('aiLevel'),
                status: 'draft' // 默认草稿状态
            };

            await apiClient.request('/portfolios', {
                method: 'POST',
                body: JSON.stringify(portfolioData)
            });

            NotificationManager.success('作品创建成功！');
            document.getElementById('createPortfolioForm').reset();
            
            await this.loadMyPortfolios();
            this.showSection('my-portfolios');

        } catch (error) {
            console.error('Failed to create portfolio:', error);
            NotificationManager.error('创建作品失败：' + error.message);
        }
    }

    async handleUpdatePortfolio(formData) {
        try {
            const portfolioId = formData.get('id');
            const updateData = {
                title: formData.get('title'),
                author: AuthManager.getCurrentUser().username,
                description: formData.get('description'),
                content: formData.get('content'),
                category: formData.get('category'),
                tags: formData.get('tags').split(',').map(tag => tag.trim()).filter(tag => tag),
                imageUrl: formData.get('imageUrl'),
                aiLevel: formData.get('aiLevel'),
                status: formData.get('status')
            };

            await apiClient.request(`/portfolios/${portfolioId}`, {
                method: 'PUT',
                body: JSON.stringify(updateData)
            });

            NotificationManager.success('作品更新成功！');
            
            await this.loadMyPortfolios();
            this.showSection('my-portfolios');

        } catch (error) {
            console.error('Failed to update portfolio:', error);
            NotificationManager.error('更新作品失败：' + error.message);
        }
    }

    async editPortfolio(portfolioId) {
        try {
            const result = await apiClient.request(`/portfolios/${portfolioId}`);
            const portfolio = result.data;
            
            // 保存编辑中的作品ID，用于版本管理
            this.editingPortfolioId = portfolioId;
            
            // 填充表单数据
            this.fillEditForm(portfolio);
            this.showSection('edit-portfolio');
            
            // 显示版本管理部分并加载版本列表
            const versionManagement = document.getElementById('versionManagement');
            if (versionManagement) {
                versionManagement.style.display = 'block';
                loadPortfolioVersions(portfolioId);
            }
            
        } catch (error) {
            console.error('Failed to load portfolio for editing:', error);
            NotificationManager.error('加载作品数据失败，请稍后重试');
        }
    }

    fillEditForm(portfolio) {
        const fields = {
            'editPortfolioId': portfolio.id,
            'editTitle': portfolio.title,
            'editCategory': portfolio.category,
            'editDescription': portfolio.description,
            'editContent': portfolio.content,
            'editAiLevel': portfolio.aiLevel,
            'editImageUrl': portfolio.imageUrl,
            'editStatus': portfolio.status
        };

        for (const [fieldId, value] of Object.entries(fields)) {
            const element = document.getElementById(fieldId);
            if (element && value !== undefined) {
                element.value = value || '';
            }
        }

        // 处理标签
        const tagsElement = document.getElementById('editTags');
        if (tagsElement && portfolio.tags) {
            if (Array.isArray(portfolio.tags)) {
                tagsElement.value = portfolio.tags.join(', ');
            } else if (typeof portfolio.tags === 'string') {
                try {
                    const tags = JSON.parse(portfolio.tags);
                    tagsElement.value = Array.isArray(tags) ? tags.join(', ') : portfolio.tags;
                } catch {
                    tagsElement.value = portfolio.tags;
                }
            }
        }
        
        // 处理HTML内容
        const htmlContentElement = document.getElementById('editHtmlContent');
        if (htmlContentElement && portfolio.activeVersion && portfolio.activeVersion.htmlContent) {
            htmlContentElement.value = portfolio.activeVersion.htmlContent;
        }
        
        // 设置HTML编辑器切换事件
        this.setupHtmlEditorToggle('Edit');
    }

    setupHtmlEditorToggle(suffix = '') {
        const toggleBtn = document.getElementById(`toggleHtmlEditor${suffix}`);
        const previewBtn = document.getElementById(`previewHtml${suffix}`);
        const container = document.getElementById(`htmlEditorContainer${suffix}`);
        const editor = document.getElementById(`${suffix === 'Edit' ? 'editHtmlContent' : 'htmlContent'}`);
        const preview = document.getElementById(`htmlPreview${suffix}`);
        
        if (toggleBtn && container) {
            // 移除旧的事件监听器
            const newToggleBtn = toggleBtn.cloneNode(true);
            toggleBtn.parentNode.replaceChild(newToggleBtn, toggleBtn);
            
            newToggleBtn.addEventListener('click', function() {
                const isVisible = container.style.display !== 'none';
                container.style.display = isVisible ? 'none' : 'block';
                newToggleBtn.innerHTML = isVisible ? '<span>📝</span> 显示HTML编辑器' : '<span>🔽</span> 隐藏HTML编辑器';
                if (previewBtn) {
                    previewBtn.style.display = isVisible ? 'none' : 'inline-flex';
                }
            });
        }
        
        if (previewBtn && editor && preview) {
            // 移除旧的事件监听器
            const newPreviewBtn = previewBtn.cloneNode(true);
            previewBtn.parentNode.replaceChild(newPreviewBtn, previewBtn);
            
            newPreviewBtn.addEventListener('click', function() {
                const htmlContent = editor.value || '<p>请先输入HTML代码</p>';
                const blob = new Blob([htmlContent], { type: 'text/html' });
                const url = URL.createObjectURL(blob);
                preview.src = url;
                
                // 清理旧的URL
                setTimeout(() => URL.revokeObjectURL(url), 1000);
            });
        }
    }

    async handleUpdateProfile(formData) {
        try {
            const updateData = {
                username: formData.get('username'),
                avatar: formData.get('avatar'),
                bio: formData.get('bio')
            };

            const result = await apiClient.request('/profile', {
                method: 'PUT',
                body: JSON.stringify(updateData)
            });

            // 更新本地存储的用户信息
            const updatedUser = { ...AuthManager.getCurrentUser(), ...result.data };
            localStorage.setItem('user', JSON.stringify(updatedUser));
            currentUser = updatedUser;
            this.updateUserInfo();

            NotificationManager.success('资料更新成功！');

        } catch (error) {
            console.error('Failed to update profile:', error);
            NotificationManager.error('更新资料失败：' + error.message);
        }
    }

    async loadProfile() {
        try {
            const result = await apiClient.request('/profile');
            const user = result.data;

            const fields = {
                'profileUsername': user.username,
                'profileEmail': user.email,
                'profileAvatar': user.avatar,
                'profileBio': user.bio
            };

            for (const [fieldId, value] of Object.entries(fields)) {
                const element = document.getElementById(fieldId);
                if (element) {
                    element.value = value || '';
                }
            }

        } catch (error) {
            console.error('Failed to load profile:', error);
            NotificationManager.error('加载资料失败');
        }
    }

    // MinIO配置管理方法将在minioManager中实现
    async loadMinioConfigs() {
        // 委托给MinIOManager处理
        if (window.minioManager) {
            return window.minioManager.loadMinioConfigs();
        }
    }

    // 用户管理和作品审核方法
    async loadUsers() {
        if (!AuthManager.isAdmin()) return;
        
        try {
            const result = await apiClient.request('/admin/users');
            const users = result.data || [];
            
            this.updateUserStats(users);
            this.renderUserList(users);
            
        } catch (error) {
            console.error('Failed to load users:', error);
            NotificationManager.error('加载用户列表失败');
        }
    }

    updateUserStats(users) {
        const stats = {
            totalUsersCount: users.length,
            pendingUsersCount: users.filter(u => u.status === 'pending').length,
            approvedUsersCount: users.filter(u => u.status === 'approved').length
        };

        for (const [id, count] of Object.entries(stats)) {
            const element = document.getElementById(id);
            if (element) element.textContent = count;
        }
    }

    renderUserList(users) {
        const container = document.getElementById('usersList');
        if (!container) return;
        
        if (!users || users.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">👥</div>
                    <h3 class="empty-title">暂无用户数据</h3>
                    <p class="empty-description">当前系统中没有用户记录</p>
                </div>
            `;
            return;
        }

        container.innerHTML = `
            <div class="table-container">
                <table class="data-table">
                    <thead>
                        <tr>
                            <th>用户名</th>
                            <th>邮箱</th>
                            <th>角色</th>
                            <th>状态</th>
                            <th>注册时间</th>
                            <th>操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${users.map(user => {
                            const statusClass = {
                                'approved': 'status-success',
                                'pending': 'status-warning', 
                                'rejected': 'status-error'
                            }[user.status] || 'status-warning';
                            
                            const statusText = {
                                'approved': '已通过',
                                'pending': '待审核',
                                'rejected': '已拒绝'
                            }[user.status] || '待审核';

                            return `
                                <tr>
                                    <td>
                                        <div style="display: flex; align-items: center; gap: 0.5rem;">
                                            <div class="user-avatar" style="width: 32px; height: 32px; background: var(--primary-color); border-radius: 50%; display: flex; align-items: center; justify-content: center; color: white; font-size: 0.8rem; font-weight: bold;">
                                                ${user.username.charAt(0).toUpperCase()}
                                            </div>
                                            <span>${user.username}</span>
                                        </div>
                                    </td>
                                    <td>${user.email}</td>
                                    <td>
                                        <span class="role-badge ${user.role === 'admin' ? 'role-admin' : 'role-user'}">
                                            ${user.role === 'admin' ? '管理员' : '用户'}
                                        </span>
                                    </td>
                                    <td><span class="status-badge ${statusClass}">${statusText}</span></td>
                                    <td>${Utils.formatDate(user.createdAt)}</td>
                                    <td>
                                        <div class="action-buttons">
                                            ${user.status === 'pending' ? `
                                                <button class="btn btn-small btn-success" onclick="dashboardManager.updateUserStatus('${user.id}', 'approved')">
                                                    <span>✅</span>
                                                    <span>通过</span>
                                                </button>
                                                <button class="btn btn-small btn-danger" onclick="dashboardManager.updateUserStatus('${user.id}', 'rejected')">
                                                    <span>❌</span>
                                                    <span>拒绝</span>
                                                </button>
                                            ` : user.status === 'approved' ? `
                                                <button class="btn btn-small btn-warning" onclick="dashboardManager.updateUserStatus('${user.id}', 'rejected')">
                                                    <span>🚫</span>
                                                    <span>禁用</span>
                                                </button>
                                            ` : `
                                                <button class="btn btn-small btn-success" onclick="dashboardManager.updateUserStatus('${user.id}', 'approved')">
                                                    <span>✅</span>
                                                    <span>启用</span>
                                                </button>
                                            `}
                                            <button class="btn btn-small btn-secondary" onclick="dashboardManager.resetUserPassword('${user.id}')">
                                                <span>🔑</span>
                                                <span>重置密码</span>
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            `;
                        }).join('')}
                    </tbody>
                </table>
            </div>
        `;
    }

    async loadAllPortfolios() {
        if (!AuthManager.isAdmin()) return;
        
        try {
            const result = await apiClient.request('/admin/portfolios');
            const portfolios = result.data || [];
            
            this.renderPortfolioReviewList(portfolios);
            
        } catch (error) {
            console.error('Failed to load all portfolios:', error);
            NotificationManager.error('加载作品列表失败');
        }
    }

    renderPortfolioReviewList(portfolios) {
        const container = document.getElementById('reviewPortfoliosList');
        if (!container) return;
        
        if (!portfolios || portfolios.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">🎨</div>
                    <h3 class="empty-title">暂无待审核作品</h3>
                    <p class="empty-description">当前没有需要审核的作品</p>
                </div>
            `;
            return;
        }

        container.innerHTML = `
            <div class="portfolio-review-grid">
                ${portfolios.map(portfolio => {
                    const statusClass = {
                        'published': 'status-success',
                        'draft': 'status-warning',
                        'rejected': 'status-error'
                    }[portfolio.status] || 'status-warning';

                    const statusText = {
                        'published': '已发布',
                        'draft': '待审核',
                        'rejected': '已拒绝'
                    }[portfolio.status] || '待审核';

                    return `
                        <div class="portfolio-review-item">
                            <div class="portfolio-image">
                                ${portfolio.imageUrl ? 
                                    `<img src="${portfolio.imageUrl}" alt="${portfolio.title}" style="width:100%;height:200px;object-fit:cover;border-radius:10px;">` : 
                                    '<div style="width:100%;height:200px;background:var(--bg-secondary);border-radius:10px;display:flex;align-items:center;justify-content:center;font-size:3rem;">🎨</div>'
                                }
                            </div>
                            <div class="portfolio-info">
                                <h4 class="portfolio-title">${portfolio.title}</h4>
                                <p class="portfolio-author">作者: ${portfolio.author}</p>
                                <p class="portfolio-category">分类: ${portfolio.category}</p>
                                <p class="portfolio-description">${portfolio.description || '暂无描述'}</p>
                                <div class="portfolio-meta">
                                    <span>📅 ${Utils.formatDate(portfolio.createdAt)}</span>
                                    <span>👁️ ${portfolio.views || 0}</span>
                                    <span>❤️ ${portfolio.likes || 0}</span>
                                    <span class="status-badge ${statusClass}">${statusText}</span>
                                </div>
                                <div class="portfolio-actions">
                                    ${portfolio.status === 'draft' ? `
                                        <button class="btn btn-small btn-success" onclick="dashboardManager.updatePortfolioStatus('${portfolio.id}', 'published')">
                                            <span>✅</span>
                                            <span>通过审核</span>
                                        </button>
                                        <button class="btn btn-small btn-danger" onclick="dashboardManager.updatePortfolioStatus('${portfolio.id}', 'rejected')">
                                            <span>❌</span>
                                            <span>拒绝审核</span>
                                        </button>
                                    ` : portfolio.status === 'published' ? `
                                        <button class="btn btn-small btn-warning" onclick="dashboardManager.updatePortfolioStatus('${portfolio.id}', 'rejected')">
                                            <span>🚫</span>
                                            <span>下架作品</span>
                                        </button>
                                    ` : `
                                        <button class="btn btn-small btn-success" onclick="dashboardManager.updatePortfolioStatus('${portfolio.id}', 'published')">
                                            <span>✅</span>
                                            <span>重新发布</span>
                                        </button>
                                    `}
                                    <button class="btn btn-small btn-secondary" onclick="dashboardManager.viewPortfolioDetails('${portfolio.id}')">
                                        <span>👁️</span>
                                        <span>查看详情</span>
                                    </button>
                                </div>
                            </div>
                        </div>
                    `;
                }).join('')}
            </div>
        `;
    }

    // 用户管理操作方法
    async updateUserStatus(userId, status) {
        try {
            await apiClient.request(`/admin/users/${userId}`, {
                method: 'PUT',
                body: JSON.stringify({ status })
            });

            NotificationManager.success(`用户状态已更新为: ${status === 'approved' ? '已通过' : '已拒绝'}`);
            await this.loadUsers(); // 重新加载用户列表

        } catch (error) {
            console.error('Failed to update user status:', error);
            NotificationManager.error('更新用户状态失败: ' + error.message);
        }
    }

    async resetUserPassword(userId) {
        if (!confirm('确定要重置该用户的密码吗？新密码将发送到用户邮箱。')) {
            return;
        }

        try {
            const result = await apiClient.request(`/admin/users/${userId}/reset-password`, {
                method: 'POST'
            });

            NotificationManager.success('密码重置成功！新密码已发送到用户邮箱。');

        } catch (error) {
            console.error('Failed to reset user password:', error);
            NotificationManager.error('重置密码失败: ' + error.message);
        }
    }

    // 作品管理操作方法
    async updatePortfolioStatus(portfolioId, status) {
        try {
            await apiClient.request(`/admin/portfolios/${portfolioId}`, {
                method: 'PUT',
                body: JSON.stringify({ status })
            });

            const statusText = {
                'published': '已发布',
                'draft': '草稿',
                'rejected': '已拒绝'
            }[status] || status;

            NotificationManager.success(`作品状态已更新为: ${statusText}`);
            await this.loadAllPortfolios(); // 重新加载作品列表

        } catch (error) {
            console.error('Failed to update portfolio status:', error);
            NotificationManager.error('更新作品状态失败: ' + error.message);
        }
    }

    async viewPortfolioDetails(portfolioId) {
        try {
            const result = await apiClient.request(`/portfolios/${portfolioId}`);
            const portfolio = result.data;
            
            // 创建详情模态框
            const detailsHtml = `
                <div class="portfolio-details">
                    <div class="portfolio-image-large">
                        ${portfolio.imageUrl ? 
                            `<img src="${portfolio.imageUrl}" alt="${portfolio.title}" style="width:100%;max-height:400px;object-fit:cover;border-radius:10px;">` : 
                            '<div style="width:100%;height:200px;background:var(--bg-secondary);border-radius:10px;display:flex;align-items:center;justify-content:center;font-size:4rem;">🎨</div>'
                        }
                    </div>
                    <div class="portfolio-info-detailed">
                        <h3>${portfolio.title}</h3>
                        <p><strong>作者:</strong> ${portfolio.author}</p>
                        <p><strong>分类:</strong> ${portfolio.category}</p>
                        <p><strong>AI参与程度:</strong> ${portfolio.aiLevel}</p>
                        <p><strong>描述:</strong> ${portfolio.description || '暂无描述'}</p>
                        <p><strong>内容:</strong></p>
                        <div style="background:var(--bg-secondary);padding:1rem;border-radius:5px;margin:0.5rem 0;">
                            ${portfolio.content || '暂无详细内容'}
                        </div>
                        <p><strong>标签:</strong> ${Array.isArray(portfolio.tags) ? portfolio.tags.join(', ') : (portfolio.tags || '无标签')}</p>
                        <div class="portfolio-stats">
                            <span>👁️ ${portfolio.views || 0} 浏览</span>
                            <span>❤️ ${portfolio.likes || 0} 点赞</span>
                            <span>📅 ${Utils.formatDate(portfolio.createdAt)}</span>
                        </div>
                    </div>
                </div>
            `;

            // 使用现有的模态框系统或创建新的
            if (window.ModalManager) {
                ModalManager.show({
                    title: '作品详情',
                    content: detailsHtml,
                    size: 'large'
                });
            } else {
                // 简单的alert fallback
                alert(`作品详情:\n标题: ${portfolio.title}\n作者: ${portfolio.author}\n分类: ${portfolio.category}\n描述: ${portfolio.description}`);
            }

        } catch (error) {
            console.error('Failed to load portfolio details:', error);
            NotificationManager.error('加载作品详情失败');
        }
    }

    // 管理员设置相关方法
    async loadAdminSettings() {
        try {
            const response = await apiClient.request('/admin/settings');
            const settings = response.data;

            // 更新界面状态
            const userApprovalCheckbox = document.getElementById('userApprovalRequired');
            const portfolioApprovalCheckbox = document.getElementById('portfolioApprovalRequired');

            if (userApprovalCheckbox) {
                userApprovalCheckbox.checked = settings.userApprovalRequired;
            }

            if (portfolioApprovalCheckbox) {
                portfolioApprovalCheckbox.checked = settings.portfolioApprovalRequired;
            }

        } catch (error) {
            console.error('Failed to load admin settings:', error);
            NotificationManager.error('加载管理员设置失败');
        }
    }

    async saveAdminSettings() {
        try {
            const userApprovalRequired = document.getElementById('userApprovalRequired').checked;
            const portfolioApprovalRequired = document.getElementById('portfolioApprovalRequired').checked;

            const data = {
                userApprovalRequired,
                portfolioApprovalRequired
            };

            await apiClient.request('/admin/settings', {
                method: 'PUT',
                body: JSON.stringify(data)
            });

            NotificationManager.success('管理员设置已保存');

        } catch (error) {
            console.error('Failed to save admin settings:', error);
            NotificationManager.error('保存管理员设置失败');
        }
    }
}

// 全局变量，供其他脚本使用
let dashboardManager;

// 全局函数，供HTML模板调用
window.saveAdminSettings = function() {
    if (dashboardManager && dashboardManager.saveAdminSettings) {
        dashboardManager.saveAdminSettings();
    } else {
        console.error('Dashboard manager not available');
        NotificationManager.error('系统未初始化完成，请刷新页面');
    }
};

// HTML编辑器相关功能
window.toggleCodeEditor = function() {
    const editor = document.getElementById('htmlContentEditor');
    const toggleText = document.getElementById('editorToggleText');
    const preview = document.getElementById('htmlPreview');
    
    if (editor.style.display === 'none') {
        editor.style.display = 'block';
        toggleText.textContent = '📝 隐藏代码编辑器';
        preview.style.display = 'none';
    } else {
        editor.style.display = 'none';
        toggleText.textContent = '📝 显示代码编辑器';
    }
};

window.previewHTML = function() {
    const editor = document.getElementById('htmlContentEditor');
    const preview = document.getElementById('htmlPreview');
    const frame = document.getElementById('previewFrame');
    
    if (preview.style.display === 'none') {
        const htmlContent = editor.value || '<p style="text-align: center; color: #666; padding: 2rem;">请先输入HTML代码</p>';
        
        // 更新iframe内容
        const blob = new Blob([htmlContent], { type: 'text/html' });
        const url = URL.createObjectURL(blob);
        frame.src = url;
        
        preview.style.display = 'block';
        editor.style.display = 'none';
        
        // 清理旧的URL
        frame.onload = function() {
            setTimeout(() => URL.revokeObjectURL(url), 1000);
        };
    } else {
        preview.style.display = 'none';
    }
};

// 版本管理功能
window.showCreateVersionModal = function() {
    const modal = document.getElementById('versionModal');
    const form = document.getElementById('versionForm');
    const title = document.getElementById('versionModalTitle');
    
    // 重置表单
    form.reset();
    document.getElementById('versionId').value = '';
    document.getElementById('versionPortfolioId').value = dashboardManager.editingPortfolioId || '';
    
    // 设置模态框标题
    title.textContent = '新建版本';
    document.getElementById('versionSubmitIcon').textContent = '➕';
    document.getElementById('versionSubmitText').textContent = '创建版本';
    
    modal.classList.add('show');
};

window.closeVersionModal = function() {
    const modal = document.getElementById('versionModal');
    modal.classList.remove('show');
};

window.editVersion = function(portfolioId, versionId) {
    const modal = document.getElementById('versionModal');
    const form = document.getElementById('versionForm');
    const title = document.getElementById('versionModalTitle');
    
    // 获取版本数据
    fetch(`/api/v1/portfolios/${portfolioId}/versions/${versionId}`, {
        headers: { 'Authorization': `Bearer ${AuthManager.getToken()}` }
    })
    .then(response => response.json())
    .then(result => {
        const version = result.data;
        
        // 填充表单
        document.getElementById('versionId').value = version.id;
        document.getElementById('versionPortfolioId').value = portfolioId;
        document.getElementById('versionTitle').value = version.title;
        document.getElementById('versionDescription').value = version.description || '';
        document.getElementById('versionHtmlContent').value = version.htmlContent;
        document.getElementById('versionChangeLog').value = version.changeLog || '';
        document.getElementById('versionIsActive').checked = version.isActive;
        
        // 设置模态框标题
        title.textContent = '编辑版本';
        document.getElementById('versionSubmitIcon').textContent = '💾';
        document.getElementById('versionSubmitText').textContent = '保存更改';
        
        modal.classList.add('show');
    })
    .catch(error => {
        console.error('获取版本数据失败:', error);
        UIHelper.showToast('获取版本数据失败', 'error');
    });
};

window.deleteVersion = function(portfolioId, versionId, versionTitle) {
    if (!confirm(`确定要删除版本 "${versionTitle}" 吗？此操作无法撤销。`)) {
        return;
    }
    
    fetch(`/api/v1/portfolios/${portfolioId}/versions/${versionId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${AuthManager.getToken()}` }
    })
    .then(response => response.json())
    .then(result => {
        UIHelper.showToast('版本删除成功', 'success');
        loadPortfolioVersions(portfolioId);
    })
    .catch(error => {
        console.error('删除版本失败:', error);
        UIHelper.showToast('删除版本失败', 'error');
    });
};

window.setActiveVersion = function(portfolioId, versionId) {
    fetch(`/api/v1/portfolios/${portfolioId}/versions/${versionId}/activate`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${AuthManager.getToken()}` }
    })
    .then(response => response.json())
    .then(result => {
        UIHelper.showToast('已设置为活跃版本', 'success');
        loadPortfolioVersions(portfolioId);
    })
    .catch(error => {
        console.error('设置活跃版本失败:', error);
        UIHelper.showToast('设置活跃版本失败', 'error');
    });
};

// 加载作品版本列表
function loadPortfolioVersions(portfolioId) {
    fetch(`/api/v1/portfolios/${portfolioId}/versions`, {
        headers: { 'Authorization': `Bearer ${AuthManager.getToken()}` }
    })
    .then(response => response.json())
    .then(result => {
        renderVersionsList(result.data || []);
    })
    .catch(error => {
        console.error('加载版本列表失败:', error);
        document.getElementById('versionsList').innerHTML = '<p style="color: var(--text-secondary); text-align: center; padding: 2rem;">加载版本列表失败</p>';
    });
}

// 渲染版本列表
function renderVersionsList(versions) {
    const container = document.getElementById('versionsList');
    
    if (!versions.length) {
        container.innerHTML = '<p style="color: var(--text-secondary); text-align: center; padding: 2rem;">暂无版本</p>';
        return;
    }
    
    const html = versions.map(version => `
        <div class="version-item ${version.isActive ? 'active' : ''}" style="border: 1px solid var(--border-color); border-radius: 8px; padding: 1rem; margin-bottom: 1rem; background: var(--card-bg);">
            <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 0.5rem;">
                <div style="flex: 1;">
                    <div style="display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.25rem;">
                        <h4 style="margin: 0; color: var(--text-primary);">${version.version}</h4>
                        ${version.isActive ? '<span class="badge badge-success">活跃</span>' : ''}
                    </div>
                    <p style="margin: 0; font-weight: 600; color: var(--text-primary);">${version.title}</p>
                    ${version.description ? `<p style="margin: 0.5rem 0 0; color: var(--text-secondary); font-size: 0.9rem;">${version.description}</p>` : ''}
                </div>
                <div style="display: flex; gap: 0.5rem;">
                    <button class="btn btn-sm btn-secondary" onclick="editVersion('${version.portfolioId}', '${version.id}')" title="编辑版本">
                        <span>✏️</span>
                    </button>
                    ${!version.isActive ? `<button class="btn btn-sm btn-primary" onclick="setActiveVersion('${version.portfolioId}', '${version.id}')" title="设为活跃版本">
                        <span>✨</span>
                    </button>` : ''}
                    ${!version.isActive ? `<button class="btn btn-sm btn-danger" onclick="deleteVersion('${version.portfolioId}', '${version.id}', '${version.title.replace(/'/g, "\\'")}')" title="删除版本">
                        <span>🗑️</span>
                    </button>` : ''}
                </div>
            </div>
            <div style="font-size: 0.85rem; color: var(--text-secondary);">
                <span>创建时间：${new Date(version.createdAt).toLocaleDateString()}</span>
                ${version.changeLog ? ` • 更新日志：${version.changeLog}` : ''}
            </div>
        </div>
    `).join('');
    
    container.innerHTML = html;
}

// 版本表单提交事件
document.addEventListener('DOMContentLoaded', function() {
    const versionForm = document.getElementById('versionForm');
    if (versionForm) {
        versionForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const portfolioId = document.getElementById('versionPortfolioId').value;
            const versionId = document.getElementById('versionId').value;
            const isEdit = !!versionId;
            
            const data = {
                title: formData.get('title'),
                description: formData.get('description'),
                htmlContent: formData.get('htmlContent'),
                changeLog: formData.get('changeLog'),
                isActive: document.getElementById('versionIsActive').checked
            };
            
            const url = isEdit 
                ? `/api/v1/portfolios/${portfolioId}/versions/${versionId}`
                : `/api/v1/portfolios/${portfolioId}/versions`;
            const method = isEdit ? 'PUT' : 'POST';
            
            fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${AuthManager.getToken()}`
                },
                body: JSON.stringify(data)
            })
            .then(response => response.json())
            .then(result => {
                UIHelper.showToast(isEdit ? '版本更新成功' : '版本创建成功', 'success');
                closeVersionModal();
                loadPortfolioVersions(portfolioId);
            })
            .catch(error => {
                console.error('版本操作失败:', error);
                UIHelper.showToast(isEdit ? '版本更新失败' : '版本创建失败', 'error');
            });
        });
    }
});

// DOM加载完成后初始化仪表板
document.addEventListener('DOMContentLoaded', function() {
    dashboardManager = new DashboardManager();
    window.dashboardManager = dashboardManager;
    
    console.log('Dashboard initialized successfully');
});