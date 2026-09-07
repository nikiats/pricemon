const tableBody = document.querySelector('#summary');
const status = document.querySelector('#status');

const text = value => value ?? '—';

function price(value) {
	if (value == null) return '—';
	return new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 2 }).format(Number(value));
}

function difference(buy, sell) {
	if (buy == null || sell == null) return { value: '—', className: '' };
	const amount = Number(sell) - Number(buy);
	return {
		value: `${amount > 0 ? '+' : ''}${price(amount)}`,
		className: amount > 0 ? 'positive' : amount < 0 ? 'negative' : '',
	};
}

function cell(className, content) {
	const element = document.createElement('td');
	element.className = className;
	element.textContent = content;
	return element;
}

function platformCell(deal) {
	const element = document.createElement('td');
	element.className = 'platforms';
	element.append(text(deal.buyPlatform?.name));

	const arrow = document.createElement('span');
	arrow.className = 'arrow';
	arrow.textContent = '→';
	element.append(arrow, text(deal.sellPlatform?.name));

	return element;
}

function addRows(deals) {
	for (const deal of deals) {
		const row = document.createElement('tr');
		const delta = difference(deal.buyPrice, deal.sellPrice);

		row.append(
			cell('', text(deal.item?.name)),
			platformCell(deal),
			cell('prices', `${price(deal.buyPrice)} → ${price(deal.sellPrice)}`),
			cell(`difference ${delta.className}`, delta.value),
		);
		tableBody.append(row);
	}
}

async function load() {
	status.textContent = 'Загрузка…';

	try {
		const response = await fetch('/v1/summary', { headers: authHeaders() });
		if (!response.ok) throw new Error('request failed');

		const page = await response.json();
		tableBody.replaceChildren();
		addRows(page.deals ?? []);
		status.textContent = tableBody.rows.length ? '' : 'Сделки не найдены';
	} catch {
		status.textContent = 'Не удалось загрузить сделки';
	}
}

if (requireAuth()) {
	renderSession();
	load();
}
