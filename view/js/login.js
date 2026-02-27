/**
 * 登录页面逻辑处理 - 增强版
 */
document.addEventListener('DOMContentLoaded', () => {
    const loginForm = document.getElementById('loginForm');
    const submitBtn = document.getElementById('submitBtn');

    // 监听表单提交事件
    loginForm.addEventListener('submit', async (e) => {

        // 1. 阻止表单默认提交（防止页面刷新）
        e.preventDefault();

        // 2. 获取表单数据
        const formData = new FormData(loginForm);
        const loginData = Object.fromEntries(formData);

        /**
         * 额外设计：记录前端访问/提交页面的时间
         * 对应你之前询问的“获取访问页面时间”
         */
        loginData.access_time = new Date().toLocaleString();

        // 3. UI 交互：禁用按钮，显示加载状态
        setLoading(true);

        try {
            // 4. 发起异步请求到后端服务器
            // 确保后端路由匹配 'POST /api/auth/login'
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(loginData)
            });
            // 解析 JSON 响应
            const result = await response.json();
            console.log("服务器响应状态:", response.status);
            // 5. 根据服务器返回的结果进行逻辑判断
            if (response.ok) {
                // 【核心修改】：从 result.data 中提取 token 和用户信息
                const userData = result.data;

                if (userData && userData.token) {
                    console.log("登录成功，正在保存 Token");

                    // 存储 Token 和相关信息
                    localStorage.setItem('auth_token', userData.token);
                    localStorage.setItem('user_id'  , userData.id || '');
                    localStorage.setItem('user_role', userData.role || '');
                    localStorage.setItem('user_name', userData.username || '');

                    // 【关键跳转】：确保路径正确
                    console.log("即将跳转至项目首页...");
                    window.location.href = '/html/index.html';
                } else {
                    console.error("跳转失败：在 result.data 中未找到 token 字段", result);
                    alert("服务器响应数据异常，请检查后端结构");
                }
            } else {
                // 情况 B: 校验失败 (HTTP 401, 403, 500 等)
                // 对应你在 Go 中处理的各种错误返回
                handleErrorResponse(response.status, result.message);
            }
        } catch (error) {
            // 情况 C: 网络异常处理（如服务器没启动、跨域问题）
            console.error('Fetch Error:', error);
            alert('无法连接到服务器，请检查后端程序是否运行');
        } finally {
            // 6. 恢复 UI 状态
            setLoading(false);
        }
    });

    /**
     * 根据不同的 HTTP 状态码提供精确反馈
     */
    function handleErrorResponse(status, message) {
        switch (status) {
            case 401:
                alert('登录失败：' + (message || '用户名或密码错误'));
                break;
            case 403:
                alert('访问受限：' + (message || '您的账号已被禁用'));
                break;
            case 500:
                alert('服务器内部错误，请检查后端数据库连接');
                break;
            default:
                alert('错误码: ' + status + ' - ' + (message || '未知错误'));
        }
    }

    /**
     * 切换按钮的加载状态
     */
    function setLoading(isLoading) {
        if (isLoading) {
            submitBtn.disabled = true;
            submitBtn.innerText = '正在核验身份...';
            submitBtn.classList.add('opacity-50', 'cursor-not-allowed');
        } else {
            submitBtn.disabled = false;
            submitBtn.innerText = '登录';
            submitBtn.classList.remove('opacity-50', 'cursor-not-allowed');
        }
    }
});