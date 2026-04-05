const { create, get } = webauthnJSON;

const elMessage = document.getElementById('message');
const elUsername = document.getElementById('username');
const btnRegister = document.getElementById('btn-register');
const btnLogin = document.getElementById('btn-login');
const btnLogout = document.getElementById('btn-logout');

const authSection = document.getElementById('auth-section');
const userSection = document.getElementById('user-section');
const loggedInUser = document.getElementById('logged-in-user');

function showMessage(msg, isError = false) {
    elMessage.textContent = msg;
    elMessage.style.color = isError ? '#ef4444' : '#fbbf24';
}

async function checkAuth() {
    try {
        const res = await fetch('/api/auth/me');
        if (res.ok) {
            const data = await res.json();
            authSection.style.display = 'none';
            userSection.style.display = 'block';
            loggedInUser.textContent = data.user_id;
        } else {
            authSection.style.display = 'block';
            userSection.style.display = 'none';
        }
    } catch (e) {
        console.error(e);
    }
}

btnRegister.addEventListener('click', async () => {
    const username = elUsername.value.trim();
    if (!username) return showMessage('Username required', true);

    try {
        showMessage('Starting registration...');
        
        // 1. Get challenge from server
        const optionsRes = await fetch('/api/auth/register/begin?username=' + encodeURIComponent(username));
        if (!optionsRes.ok) throw new Error((await optionsRes.json()).error || 'Failed to begin registration');
        const options = await optionsRes.json();

        // 2. Browser creates credential
        showMessage('Waiting for authenticator...');
        const credential = await create(options);

        // 3. Send back to server
        const finishRes = await fetch('/api/auth/register/finish', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credential),
        });

        if (!finishRes.ok) throw new Error((await finishRes.json()).error || 'Registration failed');
        
        showMessage('Registration successful! Logging in...');
        await checkAuth();
    } catch (e) {
        showMessage(e.message, true);
    }
});

btnLogin.addEventListener('click', async () => {
    const username = elUsername.value.trim();
    if (!username) return showMessage('Username required', true);

    try {
        showMessage('Starting login...');
        
        // 1. Get challenge
        const optionsRes = await fetch('/api/auth/login/begin?username=' + encodeURIComponent(username));
        if (!optionsRes.ok) throw new Error((await optionsRes.json()).error || 'Failed to begin login');
        const options = await optionsRes.json();

        // 2. Browser signs challenge
        showMessage('Waiting for authenticator...');
        const assertion = await get(options);

        // 3. Send to server
        const finishRes = await fetch('/api/auth/login/finish', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(assertion),
        });

        if (!finishRes.ok) throw new Error((await finishRes.json()).error || 'Login failed');
        
        showMessage('Login successful!');
        await checkAuth();
    } catch (e) {
        showMessage(e.message, true);
    }
});

btnLogout.addEventListener('click', async () => {
    await fetch('/api/auth/logout', { method: 'POST' });
    authSection.style.display = 'block';
    userSection.style.display = 'none';
    elUsername.value = '';
    showMessage('Logged out successfully');
});

// Initial check
checkAuth();
