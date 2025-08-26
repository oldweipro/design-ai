/**
 * DesignAI MinIO Manager - MinIO配置管理模块
 */

class MinIOManager {
    constructor() {
        this.minioConfigs = [];
        this.editingMinioConfig = null;
        this.files = [];
        this.selectedFiles = [];
        
        this.init();
    }

    init() {
        this.bindEvents();
    }

    bindEvents() {
        // MinIO配置表单提交
        const minioConfigForm = document.getElementById('minioConfigForm');
        if (minioConfigForm) {
            minioConfigForm.addEventListener('submit', (e) => {
                e.preventDefault();
                const formData = new FormData(minioConfigForm);
                this.handleMinioConfigSave(formData);
            });
        }

        // 文件上传相关事件
        this.bindFileUploadEvents();
    }

    bindFileUploadEvents() {
        // 文件上传表单提交
        const uploadForm = document.getElementById('fileUploadForm');
        if (uploadForm) {
            uploadForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.handleFileUpload();
            });
        }

        // 文件选择区域点击事件
        const uploadZone = document.getElementById('uploadZone');
        const fileInput = document.getElementById('fileInput');
        
        if (uploadZone && fileInput) {
            uploadZone.addEventListener('click', () => {
                fileInput.click();
            });

            // 拖拽上传事件
            uploadZone.addEventListener('dragover', (e) => {
                e.preventDefault();
                uploadZone.classList.add('drag-over');
            });

            uploadZone.addEventListener('dragleave', (e) => {
                e.preventDefault();
                uploadZone.classList.remove('drag-over');
            });

            uploadZone.addEventListener('drop', (e) => {
                e.preventDefault();
                uploadZone.classList.remove('drag-over');
                const files = Array.from(e.dataTransfer.files);
                this.handleFileSelection(files);
            });

            // 文件选择事件
            fileInput.addEventListener('change', (e) => {
                const files = Array.from(e.target.files);
                this.handleFileSelection(files);
            });
        }
    }

    // 加载MinIO配置列表
    async loadMinioConfigs() {
        try {
            const result = await apiClient.request('/admin/minio');
            this.minioConfigs = result.configs || [];
            this.renderMinioConfigsList();
            
            // 同时加载文件列表
            await this.loadFileList();
        } catch (error) {
            console.error('Failed to load minio configs:', error);
            NotificationManager.error('加载MinIO配置失败');
        }
    }

    // 渲染MinIO配置列表
    renderMinioConfigsList() {
        const container = document.getElementById('minioConfigsList');
        if (!container) return;
        
        if (!this.minioConfigs || this.minioConfigs.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">📁</div>
                    <h3 class="empty-title">暂无MinIO配置</h3>
                    <p class="empty-description">您还没有添加任何MinIO存储配置</p>
                    <button class="btn btn-primary" onclick="minioManager.showAddMinioConfig()">
                        <span>➕</span>
                        <span>添加第一个配置</span>
                    </button>
                </div>
            `;
            return;
        }

        container.innerHTML = this.minioConfigs.map(config => {
            const statusClass = config.is_active ? 'status-success' : 'status-warning';
            const statusText = config.is_active ? '激活' : '未激活';
            
            return `
                <div class="user-item" style="margin-bottom: 1rem;">
                    <div class="user-info">
                        <div class="user-avatar-large" style="background: ${config.is_active ? 'var(--success-color)' : 'var(--text-secondary)'}">
                            📁
                        </div>
                        <div class="user-details">
                            <h4>${this.escapeHtml(config.name)}</h4>
                            <p>🌐 ${this.escapeHtml(config.endpoint)}</p>
                            <p>🗂️ 存储桶：${this.escapeHtml(config.bucket_name)}</p>
                            <p>📝 ${this.escapeHtml(config.description || '无描述')}</p>
                        </div>
                    </div>
                    <div style="display: flex; align-items: center; gap: 1rem;">
                        <div class="status-badge ${statusClass}">${statusText}</div>
                        <div class="user-actions">
                            ${!config.is_active ? `
                                <button class="btn btn-sm btn-success" onclick="minioManager.activateMinioConfig(${config.id})">
                                    <span>✅</span>
                                    <span>激活</span>
                                </button>
                            ` : ''}
                            <button class="btn btn-sm btn-secondary" onclick="minioManager.editMinioConfig(${config.id})">
                                <span>✏️</span>
                                <span>编辑</span>
                            </button>
                            <button class="btn btn-sm" style="background: var(--info-color); color: white;" onclick="minioManager.testMinioConfigConnection(${config.id})">
                                <span>🔧</span>
                                <span>测试</span>
                            </button>
                            <button class="btn btn-sm" style="background: var(--error-color); color: white;" onclick="minioManager.deleteMinioConfig(${config.id})">
                                <span>🗑️</span>
                                <span>删除</span>
                            </button>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
    }

    // 显示添加配置对话框
    showAddMinioConfig() {
        this.editingMinioConfig = null;
        
        const modalTitle = document.getElementById('minioModalTitle');
        const secretKeyInput = document.getElementById('minioSecretKey');
        const secretKeyHint = document.getElementById('secretKeyHint');
        const form = document.getElementById('minioConfigForm');
        
        if (modalTitle) modalTitle.textContent = '添加MinIO配置';
        if (form) form.reset();
        if (document.getElementById('minioConfigId')) document.getElementById('minioConfigId').value = '';
        if (secretKeyInput) secretKeyInput.required = true;
        if (secretKeyHint) secretKeyHint.style.display = 'none';
        
        ModalManager.show('minioConfigModal');
    }

    // 编辑MinIO配置
    async editMinioConfig(configId) {
        try {
            const result = await apiClient.request(`/admin/minio/${configId}`);
            const config = result.config;
            
            this.editingMinioConfig = config;
            
            const modalTitle = document.getElementById('minioModalTitle');
            const secretKeyInput = document.getElementById('minioSecretKey');
            const secretKeyHint = document.getElementById('secretKeyHint');
            
            if (modalTitle) modalTitle.textContent = '编辑MinIO配置';
            if (secretKeyInput) {
                secretKeyInput.required = false;
                secretKeyInput.value = '';
            }
            if (secretKeyHint) secretKeyHint.style.display = 'block';
            
            // 填充表单数据
            this.fillMinioConfigForm(config);
            
            ModalManager.show('minioConfigModal');
            
        } catch (error) {
            console.error('Failed to load minio config:', error);
            NotificationManager.error('加载配置失败');
        }
    }

    // 填充MinIO配置表单
    fillMinioConfigForm(config) {
        const fields = {
            'minioConfigId': config.id,
            'minioName': config.name,
            'minioEndpoint': config.endpoint,
            'minioAccessKey': config.access_key,
            'minioBucket': config.bucket_name,
            'minioRegion': config.region || 'us-east-1',
            'minioUrlExpiry': config.url_expiry || 3600,
            'minioDescription': config.description || ''
        };

        for (const [fieldId, value] of Object.entries(fields)) {
            const element = document.getElementById(fieldId);
            if (element && value !== undefined) {
                element.value = value;
            }
        }

        // 处理复选框
        const checkboxes = {
            'minioUseSSL': config.use_ssl,
            'minioIsPrivate': config.is_private,
            'minioIsActive': config.is_active
        };

        for (const [checkboxId, checked] of Object.entries(checkboxes)) {
            const element = document.getElementById(checkboxId);
            if (element) {
                element.checked = Boolean(checked);
            }
        }
    }

    // 关闭配置对话框
    closeMinioModal() {
        ModalManager.hide('minioConfigModal');
        this.editingMinioConfig = null;
    }

    // 测试MinIO连接
    async testMinioConnection() {
        const form = document.getElementById('minioConfigForm');
        if (!form) return;
        
        const formData = new FormData(form);
        const secretKey = formData.get('secret_key');
        
        // 如果是编辑模式且密钥为空，提示用户
        if (this.editingMinioConfig && !secretKey) {
            if (!confirm('当前密钥为空，将使用现有配置的密钥进行测试。继续？')) {
                return;
            }
        }
        
        const configData = {
            endpoint: formData.get('endpoint'),
            access_key: formData.get('access_key'),
            secret_key: secretKey || (this.editingMinioConfig ? this.editingMinioConfig.secret_key : ''),
            bucket_name: formData.get('bucket_name'),
            use_ssl: formData.has('use_ssl'),
            region: formData.get('region') || 'us-east-1'
        };

        // 验证必填字段
        if (!configData.endpoint || !configData.access_key || !configData.secret_key || !configData.bucket_name) {
            NotificationManager.error('请填写所有必填字段');
            return;
        }

        try {
            LoadingManager.show('minioConfigForm', '测试连接中...');
            
            await apiClient.request('/admin/minio/test', {
                method: 'POST',
                body: JSON.stringify(configData)
            });
            
            NotificationManager.success('连接测试成功！');
            
        } catch (error) {
            console.error('Connection test failed:', error);
            NotificationManager.error('连接测试失败：' + error.message);
        } finally {
            LoadingManager.hide('minioConfigForm');
        }
    }

    // 测试指定配置的连接
    async testMinioConfigConnection(configId) {
        try {
            const config = this.minioConfigs.find(c => c.id === configId);
            if (!config) {
                NotificationManager.error('配置未找到');
                return;
            }

            LoadingManager.show('minioConfigsList', '测试连接中...');

            // 使用专门的测试端点
            await apiClient.request(`/admin/minio/${configId}/test`, {
                method: 'POST'
            });
            
            NotificationManager.success(`配置"${config.name}"连接测试成功！`);
            
        } catch (error) {
            console.error('Connection test failed:', error);
            NotificationManager.error('连接测试失败：' + error.message);
        } finally {
            LoadingManager.hide('minioConfigsList');
        }
    }

    // 激活MinIO配置
    async activateMinioConfig(configId) {
        try {
            const config = this.minioConfigs.find(c => c.id === configId);
            if (!config) {
                NotificationManager.error('配置未找到');
                return;
            }

            if (!confirm(`确定要激活配置"${config.name}"吗？这将取消其他配置的激活状态。`)) {
                return;
            }

            await apiClient.request(`/admin/minio/${configId}/activate`, {
                method: 'POST'
            });
            
            NotificationManager.success('配置激活成功！');
            await this.loadMinioConfigs();
            
        } catch (error) {
            console.error('Failed to activate config:', error);
            NotificationManager.error('激活配置失败：' + error.message);
        }
    }

    // 删除MinIO配置
    async deleteMinioConfig(configId) {
        const config = this.minioConfigs.find(c => c.id === configId);
        if (!confirm(`确定要删除配置"${config?.name || configId}"吗？此操作不可撤销。`)) {
            return;
        }

        try {
            await apiClient.request(`/admin/minio/${configId}`, {
                method: 'DELETE'
            });
            
            NotificationManager.success('配置删除成功！');
            await this.loadMinioConfigs();
            
        } catch (error) {
            console.error('Failed to delete config:', error);
            NotificationManager.error('删除配置失败：' + error.message);
        }
    }

    // 保存MinIO配置
    async handleMinioConfigSave(formData) {
        try {
            const secretKey = formData.get('secret_key');
            const configId = formData.get('id');
            
            // 验证必填字段
            const requiredFields = ['name', 'endpoint', 'access_key', 'bucket_name'];
            const missingFields = requiredFields.filter(field => !formData.get(field));
            
            if (missingFields.length > 0) {
                NotificationManager.error('请填写所有必填字段');
                return;
            }

            // 新建配置时必须提供密钥
            if (!configId && !secretKey) {
                NotificationManager.error('新建配置时必须提供秘密密钥');
                return;
            }

            const configData = {
                name: formData.get('name'),
                endpoint: formData.get('endpoint'),
                access_key: formData.get('access_key'),
                bucket_name: formData.get('bucket_name'),
                region: formData.get('region') || 'us-east-1',
                url_expiry: parseInt(formData.get('url_expiry')) || 3600,
                use_ssl: formData.has('use_ssl'),
                is_private: formData.has('is_private'),
                is_active: formData.has('is_active'),
                description: formData.get('description')
            };

            // 只有在提供了密钥时才包含它
            if (secretKey) {
                configData.secret_key = secretKey;
            }

            let result;

            if (configId && this.editingMinioConfig) {
                // 更新配置
                result = await apiClient.request(`/admin/minio/${configId}`, {
                    method: 'PUT',
                    body: JSON.stringify(configData)
                });
                NotificationManager.success('配置更新成功！');
            } else {
                // 创建新配置
                result = await apiClient.request('/admin/minio', {
                    method: 'POST',
                    body: JSON.stringify(configData)
                });
                NotificationManager.success('配置创建成功！');
            }

            this.closeMinioModal();
            await this.loadMinioConfigs();
            
        } catch (error) {
            console.error('Failed to save minio config:', error);
            NotificationManager.error('保存配置失败：' + error.message);
        }
    }

    // 获取所有配置
    getConfigs() {
        return this.minioConfigs;
    }

    // 获取激活的配置
    getActiveConfig() {
        return this.minioConfigs.find(config => config.is_active);
    }

    // 工具方法：HTML转义
    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // ============ 文件管理功能 ============

    // 处理文件选择
    handleFileSelection(files) {
        this.selectedFiles = files;
        this.renderSelectedFiles();
        
        const filePreview = document.getElementById('filePreview');
        if (filePreview) {
            filePreview.style.display = files.length > 0 ? 'block' : 'none';
        }
    }

    // 渲染选择的文件预览
    renderSelectedFiles() {
        const container = document.getElementById('selectedFiles');
        if (!container) return;

        container.innerHTML = this.selectedFiles.map((file, index) => `
            <div class="selected-file-item">
                <div class="file-info">
                    <div class="file-icon">${this.getFileIcon(file.type)}</div>
                    <div class="file-details">
                        <div class="file-name">${file.name}</div>
                        <div class="file-size">${this.formatFileSize(file.size)}</div>
                    </div>
                </div>
                <button type="button" class="remove-file-btn" onclick="minioManager.removeSelectedFile(${index})">
                    &times;
                </button>
            </div>
        `).join('');
    }

    // 移除选择的文件
    removeSelectedFile(index) {
        this.selectedFiles.splice(index, 1);
        this.renderSelectedFiles();
        
        if (this.selectedFiles.length === 0) {
            const filePreview = document.getElementById('filePreview');
            if (filePreview) filePreview.style.display = 'none';
        }
    }

    // 处理文件上传
    async handleFileUpload() {
        if (this.selectedFiles.length === 0) {
            NotificationManager.warning('请先选择要上传的文件');
            return;
        }

        const uploadBtn = document.getElementById('uploadBtn');
        const progressContainer = document.getElementById('uploadProgress');
        const progressFill = document.getElementById('progressFill');
        const progressText = document.getElementById('progressText');

        try {
            // 禁用上传按钮
            if (uploadBtn) {
                uploadBtn.disabled = true;
                uploadBtn.textContent = '上传中...';
            }

            // 显示进度条
            if (progressContainer) progressContainer.style.display = 'block';

            const tags = document.getElementById('fileTags')?.value || '';
            const isPublic = document.getElementById('isPublic')?.checked || false;

            let successCount = 0;
            let failCount = 0;

            for (let i = 0; i < this.selectedFiles.length; i++) {
                const file = this.selectedFiles[i];
                
                // 更新进度
                const progress = Math.round(((i + 1) / this.selectedFiles.length) * 100);
                if (progressFill) progressFill.style.width = `${progress}%`;
                if (progressText) progressText.textContent = `${progress}% (${i + 1}/${this.selectedFiles.length})`;

                try {
                    await this.uploadSingleFile(file, tags, isPublic);
                    successCount++;
                } catch (error) {
                    console.error(`Failed to upload file ${file.name}:`, error);
                    failCount++;
                }
            }

            // 显示结果
            if (successCount > 0) {
                NotificationManager.success(`成功上传 ${successCount} 个文件${failCount > 0 ? `，失败 ${failCount} 个` : ''}`);
            } else {
                NotificationManager.error('文件上传失败');
            }

            // 刷新文件列表
            await this.loadFileList();
            
            // 关闭上传模态框
            this.closeUploadModal();

        } catch (error) {
            console.error('Upload error:', error);
            NotificationManager.error('上传失败：' + error.message);
        } finally {
            // 重置UI状态
            if (uploadBtn) {
                uploadBtn.disabled = false;
                uploadBtn.innerHTML = '<span>📤</span><span>开始上传</span>';
            }
            if (progressContainer) progressContainer.style.display = 'none';
        }
    }

    // 上传单个文件
    async uploadSingleFile(file, tags, isPublic) {
        console.log('Uploading file:', file.name, 'Size:', file.size);
        console.log('Tags:', tags, 'IsPublic:', isPublic);
        
        const formData = new FormData();
        formData.append('file', file);
        
        if (tags) {
            formData.append('tags', tags);
        }
        
        if (isPublic) {
            formData.append('is_public', 'true');
        }

        console.log('FormData entries:');
        for (let [key, value] of formData.entries()) {
            console.log(key, value);
        }

        const result = await apiClient.request('/files/upload', {
            method: 'POST',
            body: formData,
        });

        return result;
    }

    // 加载文件列表
    async loadFileList() {
        try {
            const result = await apiClient.request('/files');
            this.files = result.files || [];
            this.renderFileList();
            this.updateFileStats();
        } catch (error) {
            console.error('Failed to load file list:', error);
            NotificationManager.error('加载文件列表失败');
        }
    }

    // 渲染文件列表
    renderFileList() {
        const container = document.getElementById('filesList');
        if (!container) return;

        if (!this.files || this.files.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <div class="empty-icon">📁</div>
                    <h3 class="empty-title">暂无文件</h3>
                    <p class="empty-description">点击上传按钮开始上传文件</p>
                </div>
            `;
            return;
        }

        container.innerHTML = `
            <div class="file-grid">
                ${this.files.map(file => `
                    <div class="file-item">
                        <div class="file-header">
                            <div class="file-icon-large">${this.getFileIcon(file.content_type)}</div>
                            <div class="file-actions">
                                <button class="btn btn-small btn-secondary" onclick="minioManager.downloadFile('${file.id}')" title="下载">
                                    <span>📥</span>
                                </button>
                                <button class="btn btn-small btn-danger" onclick="minioManager.deleteFile('${file.id}')" title="删除">
                                    <span>🗑️</span>
                                </button>
                            </div>
                        </div>
                        <div class="file-info">
                            <div class="file-name" title="${file.original_name}">${file.original_name}</div>
                            <div class="file-meta">
                                <span class="file-size">${this.formatFileSize(file.file_size)}</span>
                                <span class="file-date">${Utils.formatDate(file.created_at)}</span>
                            </div>
                            ${file.tags ? `<div class="file-tags">${file.tags}</div>` : ''}
                            <div class="file-status">
                                ${file.is_public ? '<span class="status-badge status-success">公开</span>' : '<span class="status-badge status-warning">私有</span>'}
                            </div>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }

    // 更新文件统计
    updateFileStats() {
        const totalCount = this.files.length;
        const totalSize = this.files.reduce((sum, file) => sum + (file.file_size || 0), 0);

        const countElement = document.getElementById('totalFilesCount');
        const sizeElement = document.getElementById('totalFilesSize');

        if (countElement) countElement.textContent = totalCount;
        if (sizeElement) sizeElement.textContent = this.formatFileSize(totalSize);
    }

    // 下载文件
    async downloadFile(fileId) {
        try {
            const result = await apiClient.request(`/files/${fileId}/url`);
            if (result.url) {
                const a = document.createElement('a');
                a.href = result.url;
                a.download = '';
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
            }
        } catch (error) {
            console.error('Failed to download file:', error);
            NotificationManager.error('下载文件失败');
        }
    }

    // 删除文件
    async deleteFile(fileId) {
        if (!confirm('确定要删除这个文件吗？此操作无法撤销。')) {
            return;
        }

        try {
            await apiClient.request(`/files/${fileId}`, {
                method: 'DELETE'
            });
            
            NotificationManager.success('文件删除成功');
            await this.loadFileList(); // 刷新列表
            
        } catch (error) {
            console.error('Failed to delete file:', error);
            NotificationManager.error('删除文件失败：' + error.message);
        }
    }

    // 显示上传模态框
    showUploadModal() {
        // 重置表单
        const form = document.getElementById('fileUploadForm');
        if (form) form.reset();
        
        this.selectedFiles = [];
        const filePreview = document.getElementById('filePreview');
        if (filePreview) filePreview.style.display = 'none';
        
        ModalManager.show('uploadModal');
    }

    // 关闭上传模态框
    closeUploadModal() {
        ModalManager.hide('uploadModal');
    }

    // 刷新文件列表
    async refreshFileList() {
        await this.loadFileList();
        NotificationManager.info('文件列表已刷新');
    }

    // 获取文件图标
    getFileIcon(mimeType) {
        if (!mimeType) return '📄';
        
        if (mimeType.startsWith('image/')) return '🖼️';
        if (mimeType.startsWith('video/')) return '🎥';
        if (mimeType.startsWith('audio/')) return '🎵';
        if (mimeType.includes('pdf')) return '📕';
        if (mimeType.includes('word') || mimeType.includes('document')) return '📝';
        if (mimeType.includes('sheet') || mimeType.includes('excel')) return '📊';
        if (mimeType.includes('presentation') || mimeType.includes('powerpoint')) return '📽️';
        if (mimeType.includes('zip') || mimeType.includes('rar') || mimeType.includes('tar')) return '📦';
        
        return '📄';
    }

    // 格式化文件大小
    formatFileSize(bytes) {
        if (!bytes) return '0 B';
        
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(1024));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }
}

// 全局MinIO管理器实例
let minioManager;

// 全局函数（为了兼容现有的HTML onclick事件）
function showAddMinioConfig() {
    if (minioManager) {
        minioManager.showAddMinioConfig();
    }
}

function closeMinioModal() {
    if (minioManager) {
        minioManager.closeMinioModal();
    }
}

function testMinioConnection() {
    if (minioManager) {
        minioManager.testMinioConnection();
    }
}

function showUploadModal() {
    if (minioManager) {
        minioManager.showUploadModal();
    }
}

function closeUploadModal() {
    if (minioManager) {
        minioManager.closeUploadModal();
    }
}

function refreshFileList() {
    if (minioManager) {
        minioManager.refreshFileList();
    }
}

// DOM加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    // 延迟初始化，确保其他脚本已加载
    setTimeout(() => {
        minioManager = new MinIOManager();
        window.minioManager = minioManager;
        
        console.log('MinIO Manager initialized');
    }, 100);
});