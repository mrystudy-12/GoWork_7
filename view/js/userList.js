/**
 * userList.js
 * 职责：用户管理页面的核心逻辑，包括数据拉取、表格渲染、分页控制及弹窗交互。
 * 已整合 main.js 中的 request 封装，支持 JWT 自动校验与刷新。
 */

// 1. 全局状态变量
let currentPage = 1;      // 当前页码
const pageSize = 5;       // 每页显示条数
let totalItems = 0;       // 总记录数（从后端获取）
let cachedUsers = [];     // 用于存放当前页数据缓存，实现快速回显
let searchKeyword = '';   // 搜索关键词
let statusFilter = '';    // 状态筛选

// 2. 页面加载初始化
document.addEventListener('DOMContentLoaded', () => {
    const role = localStorage.getItem('user_role');

    if (role !== "admin"){
        const createBtn = document.querySelector('button[onclick="openUserModal()"]');
        if (createBtn){
            createBtn.remove();
            console.log("检测到非管理员身份，已移除");
        }
    }
    // 设置页面标题
    if (typeof setPageTitle === 'function') {
        setPageTitle('用户管理');
    }

    // 更新顶部导航的用户头像和用户名
    updateHeaderUserInfo();

    // 初始加载第一页数据
    loadUserList(currentPage);

    // 初始化模态框外部点击关闭监听
    const modal = document.getElementById('userModal');
    if (modal) {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) closeUserModal();
        });
    }

    // 初始化搜索框事件监听
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        // 使用防抖处理，避免频繁请求
        let searchTimeout;
        searchInput.addEventListener('input', () => {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                searchKeyword = searchInput.value.trim();
                currentPage = 1; // 重置到第一页
                loadUserList(currentPage);
            }, 300);
        });
    }

    // 初始化状态筛选事件监听
    const statusFilterSelect = document.getElementById('statusFilter');
    if (statusFilterSelect) {
        statusFilterSelect.addEventListener('change', () => {
            statusFilter = statusFilterSelect.value;
            currentPage = 1; // 重置到第一页
            loadUserList(currentPage);
        });
    }
});

/**
 * 3. 从后端获取用户数据
 * 使用 request 封装，自动处理 Authorization Header
 */
async function loadUserList(page) {
    try {
        currentPage = page;
        // 构建查询参数
        let queryParams = `page=${page}&limit=${pageSize}`;
        if (searchKeyword) {
            queryParams += `&keyword=${encodeURIComponent(searchKeyword)}`;
        }
        if (statusFilter) {
            queryParams += `&status=${statusFilter}`;
        }
        // 使用 main.js 封装的 request
        const response = await request(`/api/users?${queryParams}`);

        if (!response) return; // 如果返回空，说明 request 函数内部已处理了 401/403 跳转

        const result = await response.json();

        // 将后端返回的数据存入全局变量
        const usersData = result.data?.users || [];
        const totalData = result.data?.total || 0;
        cachedUsers = Array.isArray(usersData) ? usersData : [];
        totalItems = totalData || 0;

        // 按用户 ID 正序排序
        cachedUsers.sort((a, b) => a.id - b.id);

        // 渲染表格
        renderUserTable(cachedUsers);
        // 更新分页 UI
        updatePaginationUI();

    } catch (error) {
        console.error('加载用户列表失败:', error);
        document.getElementById('userTableBody').innerHTML =
            `<tr><td colspan="5" class="text-center py-4 text-red-500">数据加载失败，请检查身份验证或网络。</td></tr>`;
    }
}

/**
 * 4. 填充表单数据函数
 */
function fillUserDataToForm(id) {
    const user = cachedUsers.find(u => u.id === id);
    if (user) {
        const currentUserID = parseInt(localStorage.getItem('user_id'), 10);
        const roleSelect = document.getElementById('userRole');
        document.getElementById('userId').value = user.id;
        document.getElementById('userName').value = user.username;
        document.getElementById('userRole').value = user.role;
        document.getElementById('userStatus').value = user.enable ? "1" : "0";
        if (user.id === currentUserID) {
            roleSelect.disabled = true;
            roleSelect.classList.add('bg-gray-100', 'cursor-not-allowed');
            roleSelect.title = "为了安全，您不能修改自己的角色";
        } else {
            roleSelect.disabled = false;
            roleSelect.classList.remove('bg-gray-100', 'cursor-not-allowed');
            roleSelect.title = "";
        }

        document.getElementById('userPassword').value = "";
        document.getElementById('userPassword').placeholder = "若不修改请留空";
    }
}

/**
 * 5. 新建用户入口
 */
function openUserModal() {
    const modal = document.getElementById('userModal');
    const form = document.getElementById('userForm');
    const roleSelect = document.getElementById('userRole');

    form.reset();
    document.getElementById('userId').value = "";

    roleSelect.disabled = false;
    roleSelect.classList.remove('bg-gray-100', 'cursor-not-allowed');

    document.getElementById('modalTitle').innerText = "新建系统用户";
    document.getElementById('submitBtn').innerText = "确认提交";

    const userNameInput = document.getElementById('userName');
    userNameInput.readOnly = false;
    userNameInput.classList.remove('bg-gray-100');

    modal.classList.remove('hidden');
}

/**
 * 6. 修改用户入口
 */
async function editUser(id) {
    const modal = document.getElementById('userModal');
    const roleSelect = document.getElementById('userRole'); // 获取角色下拉框
    const statusSelect = document.getElementById('userStatus'); // 获取状态下拉框
    const currentUserID = parseInt(localStorage.getItem('user_id'), 10); // 获取当前登录者ID

    document.getElementById('modalTitle').innerText = "修改用户信息";
    document.getElementById('submitBtn').innerText = "保存修改";

    const userNameInput = document.getElementById('userName');
    userNameInput.readOnly = true;
    userNameInput.classList.add('bg-gray-100');

    if (id === currentUserID) {
        roleSelect.disabled = true;
        roleSelect.classList.add('bg-gray-100', 'cursor-not-allowed');
        roleSelect.title = "为了安全，您不能修改自己的角色";

        statusSelect.disabled = true;
        statusSelect.classList.add('bg-gray-100', 'cursor-not-allowed');
        statusSelect.title = "您不能修改自己的账号状态";
    } else {
        roleSelect.disabled = false;
        roleSelect.classList.remove('bg-gray-100', 'cursor-not-allowed');
        roleSelect.title = "";

        statusSelect.disabled = false;
        statusSelect.classList.remove('bg-gray-100', 'cursor-not-allowed');
        statusSelect.title = "";
    }


    fillUserDataToForm(id);
    modal.classList.remove('hidden');
}

/**
 * 7. 统一提交处理函数（新增与修改共用）
 * 同样使用 request 封装，支持角色变动后 Token 的实时刷新
 */

async function submitUserData() {
    const userIdInput = document.getElementById('userId').value;
    const isEdit = userIdInput !== "";
    const currentUserID = parseInt(localStorage.getItem('user_id'), 10);

    const inputRole = document.getElementById('userRole').value;
    const inputPassword = document.getElementById('userPassword').value;

    // --- 1. 权限变更拦截逻辑 ---
    if (isEdit && parseInt(userIdInput, 10) === currentUserID) {
        // 从缓存中获取当前用户修改前的数据
        const originalUser = cachedUsers.find(u => u.id === currentUserID);

        // 如果输入框的角色值与原始角色不符
        if (originalUser && inputRole !== originalUser.role) {
            // 在控制台（终端）显示明确的错误提示
            console.error(`[权限拦截] 用户 ID: ${currentUserID} 试图修改自己的角色从 "${originalUser.role}" 到 "${inputRole}"。操作被拒绝。`);

            // 弹出用户友好提示
            alert("提交失败：您不能修改自己的用户权限！");
            return; // 彻底拦截，不进入后续请求逻辑
        }
    }

    // --- 2. 构造提交数据 (Payload) ---
    const payload = {
        username: document.getElementById('userName').value,
        role: inputRole,
        enable: document.getElementById('userStatus').value === "1"
    };

    if (isEdit) {
        payload.id = parseInt(userIdInput, 10);

        // --- 核心逻辑：不输入新密码就不修改密码 ---
        // 只有当密码框不为空字符串（去除空格后）时，才加入 password 字段
        if (inputPassword.trim() !== "") {
            payload.password = inputPassword;
        }
    } else {
        // 新建用户：必须设置初始密码
        if (!inputPassword) {
            return alert("新建用户请设置初始密码");
        }
        payload.password = inputPassword;
    }

    // --- 3. 处理头像上传 ---
    const avatarFile = document.getElementById('userAvatar').files[0];
    if (avatarFile) {
        // 如果是编辑，传当前用户ID；如果是新建，使用通用上传接口
        const uploadUrl = isEdit ? `/api/users/${userIdInput}/avatar` : '/api/uploads/avatar';
        const uploadResult = await uploadAvatar(avatarFile, uploadUrl);
        if (uploadResult) {
            payload.avatar = uploadResult.path;
        }
    }

    // --- 4. 提交数据 ---
    const url = isEdit ? `/api/users/${userIdInput}` : '/api/users';
    const method = isEdit ? 'PUT' : 'POST';

    try {
        const response = await request(url, {
            method: method,
            body: JSON.stringify(payload)
        });

        if (!response) return;

        const result = await response.json();

        if (result.code === 200) {
            alert(isEdit ? "修改成功" : "添加成功");
            closeUserModal();
            // 刷新当前页面数据
            await loadUserList(currentPage);
        } else {
            alert("操作失败：" + (result.message || "权限不足或服务器错误"));
        }
    } catch (error) {
        console.error("提交异常:", error);
        alert("系统异常，请稍后再试");
    }
}

/**
 * 头像上传预览
 */
document.addEventListener('DOMContentLoaded', () => {
    const avatarInput = document.getElementById('userAvatar');
    if (avatarInput) {
        avatarInput.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                // 验证文件格式
                const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];
                if (!allowedTypes.includes(file.type)) {
                    alert('只能上传以下格式的图片: JPEG、PNG、GIF');
                    // 清空文件选择
                    avatarInput.value = '';
                    return;
                }
                
                const reader = new FileReader();
                reader.onload = (e) => {
                    const preview = document.getElementById('avatarPreview');
                    preview.src = e.target.result;
                    preview.classList.remove('hidden');
                };
                reader.readAsDataURL(file);
            }
        });
    }
});

/**
 * 上传头像 (RESTful: POST /api/users/{id}/avatar 或 /api/uploads/avatar)
 */
async function uploadAvatar(file, url) {
    // 验证文件格式
    const allowedTypes = ['image/jpeg', 'image/png', 'image/gif'];
    if (!allowedTypes.includes(file.type)) {
        alert('只能上传以下格式的图片: JPEG、PNG、GIF');
        return null;
    }
    
    const formData = new FormData();
    formData.append('avatar', file);
    
    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('auth_token')}`
            },
            body: formData
        });
        
        if (!response.ok) {
            throw new Error('上传失败');
        }
        
        const result = await response.json();
        if (result.success) {
            // 直接返回相对路径
            return result.data;
        } else {
            throw new Error(result.message || '上传失败');
        }
    } catch (error) {
        console.error('上传错误:', error);
        alert('头像上传失败: ' + error.message);
        return null;
    }
}

/**
 * 8. 删除用户
 */
async function deleteUser(id) {
    const numericId = parseInt(id, 10);
    const currentUserID = parseInt(localStorage.getItem('user_id'), 10);
    const currentRole = localStorage.getItem('user_role');
    
    // 管理员不能删除自己
    if (currentRole === 'admin' && numericId === currentUserID) {
        alert('管理员不能删除自己的账号');
        return;
    }
    
    if (!confirm(`确定要删除 ID 为 ${numericId} 的用户吗？`)) return;

    try {
        const response = await request(`/api/users/${numericId}`, {
            method: 'DELETE'
        });

        if (!response) return;

        const result = await response.json();

        if (result.code === 200) {
            alert('删除成功');
            await loadUserList(currentPage);
        } else {
            alert('删除失败：' + (result.message || '无法执行删除'));
        }
    } catch (error) {
        console.error('删除操作异常:', error);
    }
}

/**
 * 9. 关闭模态框
 */
function closeUserModal() {
    const modal = document.getElementById('userModal');
    if (modal) modal.classList.add('hidden');
}

/**
 * 10. 渲染表格 UI
 */
function renderUserTable(users) {
    const tbody = document.getElementById('userTableBody');
    if (!tbody) return;

    // --- 1. 获取当前登录者的权限信息 ---
    const currentRole = localStorage.getItem('user_role');
    const currentUserID = parseInt(localStorage.getItem('user_id'), 10);

    if (!users || users.length === 0) {
        tbody.innerHTML = `<tr><td colspan="5" class="text-center py-10 text-gray-400">暂无相关用户信息</td></tr>`;
        return;
    }

    const html = users.map(user => {
        // --- 2. 判定当前行用户的权限状态 ---
        const isAdmin = currentRole === 'admin';
        const isSelf = user.id === currentUserID;

        const lastLoginTime = user.last_login
            ? new Date(user.last_login).toLocaleString('zh-CN', { hour12: false })
            : '从未登录';

        const roleConfig = user.role === 'admin'
            ? { label: '管理员', class: 'bg-purple-100 text-purple-800' }
            : { label: '普通用户', class: 'bg-blue-100 text-blue-800' };

        const statusConfig = user.enable
            ? { label: '正常启用', class: 'bg-green-100 text-green-800', dot: 'bg-green-500' }
            : { label: '已禁用', class: 'bg-red-100 text-red-800', dot: 'bg-red-500' };

        // --- 3. 动态生成操作按钮的 HTML ---
        let actionButtons = '';
        if (isAdmin) {
            // 管理员：可以操作普通用户，也可以操作自己，但不能操作其他管理员
            if (user.role !== 'admin') {
                // 普通用户：显示操作按钮
                actionButtons = `
                    <button onclick="editUser(${user.id})" class="p-2 hover:bg-blue-50 text-blue-600 rounded-lg transition-colors">
                        <i data-feather="edit-3" class="w-4 h-4"></i>
                    </button>
                    <button onclick="deleteUser(${user.id})" class="p-2 hover:bg-red-50 text-red-600 rounded-lg transition-colors">
                        <i data-feather="trash-2" class="w-4 h-4"></i>
                    </button>`;
            } else if (isSelf) {
                // 当前登录的管理员自己：只显示编辑按钮，不显示删除按钮
                actionButtons = `
                    <button onclick="editUser(${user.id})" class="p-2 hover:bg-blue-50 text-blue-600 rounded-lg transition-colors">
                        <i data-feather="edit-3" class="w-4 h-4"></i>
                    </button>
                    <span class="text-xs text-gray-300 italic">本人</span>`;
            } else {
                // 其他管理员：隐藏操作按钮
                actionButtons = `<span class="text-xs text-gray-300 italic">同等级</span>`;
            }
        } else if (isSelf) {
            // 普通用户本人：只能编辑自己，不能删除自己
            actionButtons = `
                <button onclick="editUser(${user.id})" class="p-2 hover:bg-blue-50 text-blue-600 rounded-lg transition-colors">
                    <i data-feather="edit-3" class="w-4 h-4"></i>
                </button>
                <span class="text-xs text-blue-500 font-medium ml-1">本人</span>`;
        } else {
            // 其他人：无操作权限
            actionButtons = `<span class="text-xs text-gray-300 italic">只读</span>`;
        }

        return `
            <tr class="hover:bg-gray-50 border-b transition-colors" id="user-row-${user.id}">
                <td class="px-6 py-4">
                    <div class="flex items-center gap-3">
                            <img src="${user.avatar ? user.avatar : 'https://ui-avatars.com/api/?name=' + encodeURIComponent(user.username) + '&background=random&size=128'}" 
                                 class="w-10 h-10 rounded-full border border-gray-200" alt="头像">
                            <div>
                                <p class="font-semibold text-gray-900">
                                    ${user.username} 
                                    ${isSelf ? '<span class="ml-1 text-[10px] bg-blue-50 text-blue-500 px-1 rounded">YOU</span>' : ''}
                                </p>
                                <p class="text-xs text-gray-500">ID: ${user.id}</p>
                            </div>
                    </div>
                </td>
                <td class="px-6 py-4">
                    <span class="px-2.5 py-0.5 rounded-full text-xs font-medium ${roleConfig.class}">
                        ${roleConfig.label}
                    </span>
                </td>
                <td class="px-6 py-4 text-sm text-gray-600">
                    <div class="flex items-center gap-1.5">
                        <i data-feather="clock" class="w-3.5 h-3.5 text-gray-400"></i>
                        ${lastLoginTime}
                    </div>
                </td>
                <td class="px-6 py-4">
                    <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${statusConfig.class}">
                        <span class="w-1.5 h-1.5 rounded-full ${statusConfig.dot}"></span>
                        ${statusConfig.label}
                    </span>
                </td>
                <td class="px-6 py-4">
                    <div class="flex items-center gap-1">
                        ${actionButtons}
                    </div>
                </td>
            </tr>
        `;
    }).join('');

    tbody.innerHTML = html;
    if (window.feather) feather.replace();
}

/**
 * 11. 分页控制
 */
function updatePaginationUI() {
    const totalPages = Math.ceil(totalItems / pageSize) || 1;
    
    // 更新分页信息显示
    const startIndex = (currentPage - 1) * pageSize + 1;
    const endIndex = Math.min(currentPage * pageSize, totalItems);
    
    const startIndexElement = document.getElementById('start-index');
    const endIndexElement = document.getElementById('end-index');
    const totalCountElement = document.getElementById('total-count');
    
    if (startIndexElement) startIndexElement.innerText = startIndex;
    if (endIndexElement) endIndexElement.innerText = endIndex;
    if (totalCountElement) totalCountElement.innerText = totalItems;
}

function changePage(delta) {
    const totalPages = Math.ceil(totalItems / pageSize) || 1;
    const newPage = currentPage + delta;
    if (newPage >= 1 && newPage <= totalPages) {
        loadUserList(newPage);
        window.scrollTo({ top: 0, behavior: 'smooth' });
    }
}

/**
 * 12. 更新顶部导航用户信息
 */
async function updateHeaderUserInfo() {
    const userID = localStorage.getItem('user_id');
    if (!userID) return;

    try {
        // 获取用户信息
        const response = await request('/api/GetAllUsers?page=1&limit=100');
        if (!response) return;

        const result = await response.json();
        if (result.code === 200 && result.data && result.data.users) {
            const users = result.data.users;
            const currentUser = users.find(user => user.id === parseInt(userID, 10));

            if (currentUser) {
                // 更新头像
                const headerAvatar = document.querySelector('header img');
                if (headerAvatar) {
                    if (currentUser.avatar) {
                        headerAvatar.src = currentUser.avatar;
                    } else {
                        // 使用默认头像
                        headerAvatar.src = `https://ui-avatars.com/api/?name=${encodeURIComponent(currentUser.username)}&background=random&size=128`;
                    }
                    // 添加鼠标悬停效果
                    headerAvatar.title = currentUser.username;
                }

                // 更新用户信息容器
                const userInfoContainer = document.querySelector('header .flex.items-center.gap-2');
                if (userInfoContainer) {
                    // 检查是否已有用户名显示
                    let usernameElement = userInfoContainer.querySelector('.username-display');
                    if (!usernameElement) {
                        // 创建用户名显示元素
                        usernameElement = document.createElement('span');
                        usernameElement.className = 'username-display text-sm font-medium';
                        userInfoContainer.insertBefore(usernameElement, userInfoContainer.querySelector('button'));
                    }
                    // 设置用户名
                    usernameElement.textContent = currentUser.username;
                }
            }
        }
    } catch (error) {
        console.error('更新顶部导航用户信息失败:', error);
    }
}

/**
 * 13. 退出登录
 */
async function logout() {
    if (!confirm('确定要退出登录吗？')) return;

    // 清除本地存储的认证信息
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user_id');
    localStorage.removeItem('user_role');

    // 跳转到登录页面
    window.location.href = '/login.html';
}