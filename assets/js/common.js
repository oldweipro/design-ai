/**
 * DesignAI Common JavaScript - 公共功能模块
 */

// 全局变量
let currentUser = null;
let currentTheme = 'light';
const API_BASE_URL = '/api/v1';

// API请求封装
class ApiClient {
    constructor() {
        this.baseURL = API_BASE_URL;
    }

    // 获取认证token
    getAuthToken() {
        return localStorage.getItem('authToken');
    }

    // 获取认证请求头
    getAuthHeaders() {
        const token = this.getAuthToken();
        return token ? { 'Authorization': `Bearer ${token}` } : {};
    }

    // API请求封装
    async request(url, options = {}) {
        try {
            const response = await fetch(this.baseURL + url, {
                headers: {
                    'Content-Type': 'application/json',
                    ...this.getAuthHeaders(),
                    ...options.headers
                },
                ...options
            });

            if (response.status === 401) {
                AuthManager.logout();
                return;
            }

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || `HTTP error! status: ${response.status}`);
            }

            return data;
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }
}

// 认证管理器
class AuthManager {
    static checkAuth() {
        const token = localStorage.getItem('authToken');
        const userStr = localStorage.getItem('user');
        
        if (!token || !userStr) {
            window.location.href = '/auth';
            return false;
        }
        
        try {
            currentUser = JSON.parse(userStr);
            return true;
        } catch (error) {
            console.error('Failed to parse user data:', error);
            AuthManager.logout();
            return false;
        }
    }

    static logout() {
        if (confirm('确定要退出登录吗？')) {
            localStorage.removeItem('authToken');
            localStorage.removeItem('user');
            NotificationManager.info('已退出登录');
            setTimeout(() => {
                window.location.href = '/auth';
            }, 1000);
        }
    }

    static getCurrentUser() {
        return currentUser;
    }

    static isAdmin() {
        return currentUser && currentUser.role === 'admin';
    }
}

// 主题管理器
class ThemeManager {
    static init() {
        const savedTheme = localStorage.getItem('theme');
        if (savedTheme === 'dark') {
            ThemeManager.setTheme('dark');
        }
    }

    static setTheme(theme) {
        const body = document.body;
        const themeIcon = document.getElementById('themeIcon');

        if (theme === 'dark') {
            body.setAttribute('data-theme', 'dark');
            if (themeIcon) themeIcon.textContent = '☀️';
            currentTheme = 'dark';
            localStorage.setItem('theme', 'dark');
        } else {
            body.removeAttribute('data-theme');
            if (themeIcon) themeIcon.textContent = '🌙';
            currentTheme = 'light';
            localStorage.setItem('theme', 'light');
        }
    }

    static toggle() {
        ThemeManager.setTheme(currentTheme === 'light' ? 'dark' : 'light');
    }

    static getCurrentTheme() {
        return currentTheme;
    }
}

// 通知管理器
class NotificationManager {
    static show(message, type = 'success', duration = 3000) {
        const toast = document.createElement('div');
        const bgColor = {
            'success': 'var(--success-color)',
            'error': 'var(--error-color)',
            'warning': 'var(--warning-color)',
            'info': 'var(--info-color)'
        }[type] || 'var(--info-color)';
        
        toast.style.cssText = `
            position: fixed;
            top: 100px;
            right: 2rem;
            background: ${bgColor};
            color: white;
            padding: 1rem 2rem;
            border-radius: 10px;
            box-shadow: 0 10px 30px var(--shadow-medium);
            z-index: 3000;
            animation: slideInRight 0.3s ease;
            backdrop-filter: blur(10px);
            max-width: 300px;
            word-wrap: break-word;
            font-weight: 500;
        `;
        
        toast.textContent = message;
        document.body.appendChild(toast);

        setTimeout(() => {
            toast.style.animation = 'slideOutRight 0.3s ease';
            setTimeout(() => {
                if (document.body.contains(toast)) {
                    document.body.removeChild(toast);
                }
            }, 300);
        }, duration);
    }

    static success(message) {
        NotificationManager.show(message, 'success');
    }

    static error(message) {
        NotificationManager.show(message, 'error');
    }

    static warning(message) {
        NotificationManager.show(message, 'warning');
    }

    static info(message) {
        NotificationManager.show(message, 'info');
    }
}

// 模态框管理器
class ModalManager {
    static show(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'block';
            document.body.style.overflow = 'hidden';
            
            // 添加ESC键关闭功能
            const closeOnEscape = (e) => {
                if (e.key === 'Escape') {
                    ModalManager.hide(modalId);
                    document.removeEventListener('keydown', closeOnEscape);
                }
            };
            document.addEventListener('keydown', closeOnEscape);
            
            // 点击外部关闭
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    ModalManager.hide(modalId);
                }
            });
        }
    }

    static hide(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.style.display = 'none';
            document.body.style.overflow = '';
        }
    }

    static hideAll() {
        const modals = document.querySelectorAll('.modal');
        modals.forEach(modal => {
            modal.style.display = 'none';
        });
        document.body.style.overflow = '';
    }
}

// 表单验证器
class FormValidator {
    static validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    }

    static validatePassword(password) {
        // 至少8位，包含字母和数字
        return password.length >= 8 && /[a-zA-Z]/.test(password) && /\d/.test(password);
    }

    static validateURL(url) {
        try {
            new URL(url);
            return true;
        } catch {
            return false;
        }
    }

    static validateRequired(value) {
        return value !== null && value !== undefined && value.trim() !== '';
    }

    static validateForm(formData, rules) {
        const errors = {};
        
        for (const [field, rule] of Object.entries(rules)) {
            const value = formData.get ? formData.get(field) : formData[field];
            
            if (rule.required && !FormValidator.validateRequired(value)) {
                errors[field] = rule.messages?.required || `${field} 是必填项`;
                continue;
            }
            
            if (value && rule.type === 'email' && !FormValidator.validateEmail(value)) {
                errors[field] = rule.messages?.invalid || '邮箱格式不正确';
            }
            
            if (value && rule.type === 'password' && !FormValidator.validatePassword(value)) {
                errors[field] = rule.messages?.invalid || '密码至少8位，包含字母和数字';
            }
            
            if (value && rule.type === 'url' && !FormValidator.validateURL(value)) {
                errors[field] = rule.messages?.invalid || 'URL格式不正确';
            }
            
            if (value && rule.minLength && value.length < rule.minLength) {
                errors[field] = rule.messages?.minLength || `最少${rule.minLength}个字符`;
            }
            
            if (value && rule.maxLength && value.length > rule.maxLength) {
                errors[field] = rule.messages?.maxLength || `最多${rule.maxLength}个字符`;
            }
        }
        
        return Object.keys(errors).length === 0 ? null : errors;
    }
}

// 工具函数
class Utils {
    // 防抖函数
    static debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    // 节流函数
    static throttle(func, limit) {
        let inThrottle;
        return function(...args) {
            if (!inThrottle) {
                func.apply(this, args);
                inThrottle = true;
                setTimeout(() => inThrottle = false, limit);
            }
        };
    }

    // 格式化日期
    static formatDate(date, format = 'YYYY-MM-DD') {
        const d = new Date(date);
        const year = d.getFullYear();
        const month = String(d.getMonth() + 1).padStart(2, '0');
        const day = String(d.getDate()).padStart(2, '0');
        const hours = String(d.getHours()).padStart(2, '0');
        const minutes = String(d.getMinutes()).padStart(2, '0');
        
        return format
            .replace('YYYY', year)
            .replace('MM', month)
            .replace('DD', day)
            .replace('HH', hours)
            .replace('mm', minutes);
    }

    // 格式化文件大小
    static formatFileSize(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    // 复制到剪贴板
    static async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            NotificationManager.success('已复制到剪贴板');
        } catch (err) {
            // 降级方案
            const textArea = document.createElement('textarea');
            textArea.value = text;
            document.body.appendChild(textArea);
            textArea.select();
            document.execCommand('copy');
            document.body.removeChild(textArea);
            NotificationManager.success('已复制到剪贴板');
        }
    }

    // 生成随机ID
    static generateId(length = 8) {
        const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let result = '';
        for (let i = 0; i < length; i++) {
            result += chars.charAt(Math.floor(Math.random() * chars.length));
        }
        return result;
    }

    // 深拷贝对象
    static deepClone(obj) {
        if (obj === null || typeof obj !== 'object') return obj;
        if (obj instanceof Date) return new Date(obj.getTime());
        if (obj instanceof Array) return obj.map(item => Utils.deepClone(item));
        if (typeof obj === 'object') {
            const clonedObj = {};
            for (const key in obj) {
                if (obj.hasOwnProperty(key)) {
                    clonedObj[key] = Utils.deepClone(obj[key]);
                }
            }
            return clonedObj;
        }
    }

    // 查询参数解析
    static parseQuery(queryString = window.location.search) {
        const params = new URLSearchParams(queryString);
        const result = {};
        for (const [key, value] of params) {
            result[key] = value;
        }
        return result;
    }

    // 构建查询字符串
    static buildQuery(params) {
        const searchParams = new URLSearchParams();
        for (const [key, value] of Object.entries(params)) {
            if (value !== null && value !== undefined && value !== '') {
                searchParams.append(key, value);
            }
        }
        return searchParams.toString();
    }
}

// 加载管理器
class LoadingManager {
    static show(element, text = '加载中...') {
        if (typeof element === 'string') {
            element = document.getElementById(element);
        }
        
        if (!element) return;
        
        const loadingHtml = `
            <div class="loading-overlay" style="
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: rgba(255, 255, 255, 0.8);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 999;
            ">
                <div class="loading">
                    <div class="loading-dot"></div>
                    <div class="loading-dot"></div>
                    <div class="loading-dot"></div>
                    <span style="margin-left: 1rem;">${text}</span>
                </div>
            </div>
        `;
        
        element.style.position = 'relative';
        element.insertAdjacentHTML('beforeend', loadingHtml);
    }

    static hide(element) {
        if (typeof element === 'string') {
            element = document.getElementById(element);
        }
        
        if (!element) return;
        
        const loadingOverlay = element.querySelector('.loading-overlay');
        if (loadingOverlay) {
            loadingOverlay.remove();
        }
    }
}

// 事件委托管理器
class EventDelegator {
    constructor(container) {
        this.container = container;
        this.handlers = new Map();
    }

    on(selector, event, handler) {
        const key = `${selector}:${event}`;
        if (!this.handlers.has(key)) {
            this.handlers.set(key, []);
            
            this.container.addEventListener(event, (e) => {
                if (e.target.matches(selector) || e.target.closest(selector)) {
                    const handlers = this.handlers.get(key) || [];
                    handlers.forEach(h => h.call(this, e));
                }
            });
        }
        
        this.handlers.get(key).push(handler);
    }

    off(selector, event, handler) {
        const key = `${selector}:${event}`;
        if (this.handlers.has(key)) {
            const handlers = this.handlers.get(key);
            const index = handlers.indexOf(handler);
            if (index > -1) {
                handlers.splice(index, 1);
            }
        }
    }
}

// 全局API客户端实例
const apiClient = new ApiClient();

// 全局工具函数（为了兼容现有代码）
function showToast(message, type = 'success') {
    NotificationManager.show(message, type);
}

function toggleTheme() {
    ThemeManager.toggle();
}

function logout() {
    AuthManager.logout();
}

function checkAuth() {
    return AuthManager.checkAuth();
}

// 导出给其他模块使用
window.DesignAI = {
    ApiClient,
    AuthManager,
    ThemeManager,
    NotificationManager,
    ModalManager,
    FormValidator,
    Utils,
    LoadingManager,
    EventDelegator,
    apiClient
};

// DOM加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    // 初始化主题
    ThemeManager.init();
    
    // 添加全局样式
    if (!document.getElementById('common-animations')) {
        const style = document.createElement('style');
        style.id = 'common-animations';
        style.textContent = `
            @keyframes slideInRight {
                from { transform: translateX(100%); opacity: 0; }
                to { transform: translateX(0); opacity: 1; }
            }
            @keyframes slideOutRight {
                from { transform: translateX(0); opacity: 1; }
                to { transform: translateX(100%); opacity: 0; }
            }
            .loading-overlay {
                backdrop-filter: blur(5px);
            }
        `;
        document.head.appendChild(style);
    }
    
    console.log('DesignAI Common initialized');
});