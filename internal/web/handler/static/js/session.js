const TOKEN_KEY = 'accessToken';
const EMAIL_KEY = 'email';

const session = {
	token: () => localStorage.getItem(TOKEN_KEY),
	email: () => localStorage.getItem(EMAIL_KEY),
	save(token, email) {
		localStorage.setItem(TOKEN_KEY, token);
		localStorage.setItem(EMAIL_KEY, email);
	},
	clear() {
		localStorage.removeItem(TOKEN_KEY);
		localStorage.removeItem(EMAIL_KEY);
	},
};

function authHeaders() {
	const token = session.token();
	return token ? { Authorization: `Bearer ${token}` } : {};
}

function requireAuth() {
	if (session.token()) return true;
	location.replace('/login');
	return false;
}

function renderSession() {
	const holder = document.querySelector('#current-user');
	if (holder) holder.textContent = session.email() ?? '';

	const button = document.querySelector('#logout');
	if (!button) return;

	button.addEventListener('click', async () => {
		try {
			await fetch('/v1/sessions/current', { method: 'DELETE', headers: authHeaders() });
		} catch {}
		session.clear();
		location.href = '/';
	});
}
