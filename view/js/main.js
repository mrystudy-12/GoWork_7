/**
 * main.js
 * 职责：存放全站通用的交互逻辑（如：登出、导航、全局图标初始化、标题设置等）。
 */

// 1. 全局初始化
document.addEventListener('DOMContentLoaded', () => {
    // 自动运行：为页面上所有带有 data-feather 属性的元素渲染图标
    initGlobalIcons();
    
    // 自动更新顶部导航的用户头像和用户名
    updateHeaderUserInfo();
});

/**
 * 2. 初始化全局图标
 * 封装此函数是为了在动态加载内容后也能手动触发
 */
function initGlobalIcons() {
    if (typeof feather !== 'undefined') {
        feather.replace();
    } else {
        console.warn('Feather 图标库尚未加载，请检查 HTML 中的脚本引入。');
    }
}

/**
 * 3. 更新顶部导航用户信息
 */
function updateHeaderUserInfo() {
    const userName = localStorage.getItem('user_name');
    const userAvatar = localStorage.getItem('user_avatar');

    // 添加调试日志，方便在控制台查看
    console.log("【调试】顶部导航栏更新 (来自 main.js)：", {
        userName: userName,
        userAvatar: userAvatar
    });

    if (userName) {
        const userNameElements = document.querySelectorAll('.user-name-display');
        userNameElements.forEach(el => el.innerText = userName);
    }

    // 无论有没有头像，都进行处理
    const userAvatarElements = document.querySelectorAll('.user-avatar-display');
    userAvatarElements.forEach(el => {
        if (userAvatar && userAvatar !== "") {
            // 如果有真实头像，直接使用
            el.src = userAvatar;
        } else if (userName) {
            // 如果没有头像但有用户名，生成一个字母头像
            el.src = `https://ui-avatars.com/api/?name=${encodeURIComponent(userName)}&background=random&size=128`;
        }
    });
}



/**
 * 4. 设置页面标题
 * @param {string} title - 页面显示的标题名称
 */
function setPageTitle(title) {
    const titleElement = document.getElementById('pageTitle');
    if (titleElement) {
        titleElement.textContent = title;
    }
    // 同时更新浏览器标签页标题
    document.title = title + ' - 后台管理系统';
}

/**
 * 5. 退出登录逻辑
 * 清除本地缓存并跳转至登录页
 */
function logout() {
    if (confirm('确定要退出系统吗？')) {
        // 清除存储的 Token 或用户信息
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user_id');
        localStorage.removeItem('user_role');
        sessionStorage.clear();

        // 跳转到登录页面 (根据你的路由结构调整)
        window.location.href = '/login.html';
    }
}

/**
 * 6. 通用工具函数：格式化日期
 * @param {string|Date} dateSource - 后端传来的日期字符串或对象
 * @returns {string} 格式化后的时间字符串
 */
function formatDateTime(dateSource) {
    if (!dateSource) return '无';
    const date = new Date(dateSource);
    if (isNaN(date.getTime())) return '无效日期';

    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });
}

/**
 * 7. 通用 Fetch 请求封装
 * 自动处理：Token 携带、New-Token 更新、401 自动跳转
 */
async function request(url, options = {}) {
    // 从本地获取 Token
    const token = localStorage.getItem('auth_token');

    // 默认 Headers
    const defaultHeaders = {
        'Content-Type': 'application/json',
    };

    // 如果有 Token，则按后端 auth.go 的逻辑加上 Bearer 前缀
    if (token) {
        defaultHeaders['Authorization'] = `Bearer ${token}`;
    }

    // 合并配置
    const config = {
        ...options,
        headers: {
            ...defaultHeaders,
            ...options.headers
        }
    };

    try {
        const response = await fetch(url, config);

        // A. 检查后端是否有新 Token 下发（角色变更逻辑）
        const newToken = response.headers.get('New-Token');
        if (newToken) {
            localStorage.setItem('auth_token', newToken);
            console.log("Token 已根据权限变更自动更新");
        }

        // B. 处理身份失效 (401)
        if (response.status === 401) {
            alert("登录已过期，请重新登录");
            localStorage.clear();
            window.location.href = '/html/login.html';
            return null;
        }

        // C. 处理禁止访问 (403 - 对应你 auth.go 中的账号禁用)
        if (response.status === 403) {
            const result = await response.json();
            alert(result.message || "您的账号已被禁用");
            return null;
        }

        return response;
    } catch (error) {
        console.error("请求失败:", error);
        throw error;
    }
}