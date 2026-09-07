const form = document.querySelector('.auth-form');
const error = document.querySelector('.form-error');

function show(message) {
	error.textContent = message;
}

async function send(path, email, password) {
	const response = await fetch(path, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password }),
	});

	if (response.ok) return response;

	let code = '';
	try { code = (await response.json()).code; } catch {}

	const messages = {
		email_taken: 'Такой email уже зарегистрирован',
		invalid_credentials: 'Неверный email или пароль',
		invalid_json: 'Заполните все поля',
	};
	throw new Error(messages[code] ?? 'Что-то пошло не так, попробуйте позже');
}

form.addEventListener('submit', async event => {
	event.preventDefault();
	show('');

	const data = new FormData(form);
	const email = data.get('email').trim();
	const password = data.get('password');
	const confirmation = data.get('password-confirmation');

	if (confirmation !== null && password !== confirmation) {
		show('Пароли не совпадают');
		return;
	}

	const button = form.querySelector('button');
	button.disabled = true;

	try {
		if (confirmation !== null) {
			await send('/v1/users', email, password);
			location.href = '/login';
			return;
		}

		const response = await send('/v1/sessions', email, password);
		const created = await response.json();
		session.save(created.accessToken, email);
		location.href = '/menu';
	} catch (failure) {
		show(failure.message);
	} finally {
		button.disabled = false;
	}
});
