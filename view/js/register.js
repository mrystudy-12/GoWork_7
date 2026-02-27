document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('registerForm');
    const usernameInput = document.getElementById('username');
    const passwordInput = document.getElementById('password');
    const confirmInput = document.getElementById('confirm_password');

    // 提示元素
    const userHint = document.getElementById('usernameHint');
    const passHint = document.getElementById('passwordHint');
    const passError = document.getElementById('passwordError');

    // 正则表达式：4-16位用户名，6位数字密码
    const userRegex = /^[a-zA-Z0-9_]{4,16}$/;
    const passRegex = /^\d{6}$/;

    // --- 校验函数 ---
    const validateUsername = () => {
        if (userRegex.test(usernameInput.value)) {
            userHint.textContent = "用户名格式正确 ✅";
            userHint.className = "mt-1 text-xs text-green-600";
            usernameInput.classList.replace('border-gray-300', 'border-green-500');
            return true;
        } else {
            userHint.textContent = "格式错误：需4-16位字母/数字/下划线";
            userHint.className = "mt-1 text-xs text-red-500";
            usernameInput.classList.add('border-red-500');
            return false;
        }
    };

    const validatePassword = () => {
        if (passRegex.test(passwordInput.value)) {
            passHint.textContent = "密码格式正确 ✅";
            passHint.className = "mt-1 text-xs text-green-600";
            passwordInput.classList.replace('border-gray-300', 'border-green-500');
            return true;
        } else {
            passHint.textContent = "错误：必须是6位数字";
            passHint.className = "mt-1 text-xs text-red-500";
            passwordInput.classList.add('border-red-500');
            return false;
        }
    };

    const validateMatch = () => {
        if (confirmInput.value === passwordInput.value && confirmInput.value !== "") {
            passError.classList.add('hidden');
            confirmInput.classList.replace('border-gray-300', 'border-green-500');
            confirmInput.classList.remove('border-red-500');
            return true;
        } else {
            passError.classList.remove('hidden');
            confirmInput.classList.add('border-red-500');
            return false;
        }
    };

    // --- 事件监听 ---
    usernameInput.addEventListener('input', validateUsername);
    passwordInput.addEventListener('input', validatePassword);
    confirmInput.addEventListener('input', validateMatch);

    // --- 表单提交 ---
    form.addEventListener('submit', (e) => {
        e.preventDefault(); // 阻止表单默认提交行为，改为 AJAX 提交

        const isUserOk = validateUsername(); //
        const isPassOk = validatePassword(); //
        const isMatch = validateMatch();     //

        if (!isUserOk || !isPassOk || !isMatch) {
            alert("请完善信息后再提交！");
            return;
        }

        // 构建 JSON 数据
        const registerData = {
            username: usernameInput.value,
            password: passwordInput.value
        };

        // 发送请求到后端
        fetch('/api/auth/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(registerData)
        })
            .then(response => response.json()) // 将后端返回的 JSON 字符串转为对象
            .then(res => {
                // 根据你后端的输出格式：{"code": 200, "message": "注册成功", "data": {...}}
                if (res.code === 200) {
                    alert(res.message + "！点击确定跳转登录。");

                    // 成功跳转：这里路径需对应你的 main.go 静态资源挂载路径
                    window.location.href = "/html/login.html";
                } else {
                    // 注册失败提示（如用户名已存在）
                    alert("注册失败: " + res.message);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert("服务器响应异常，请稍后再试");
            });
    });
});