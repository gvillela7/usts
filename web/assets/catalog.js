function bindCatalogPage() {
	const search = document.getElementById('catalog-search');
	if (search) {
		search.addEventListener('input', window.applyCatalogFilters);
	}

	document.querySelectorAll('[data-catalog-filter]').forEach((button) => {
		button.addEventListener('click', () => {
			window.filterCatalogByPrefix(button.getAttribute('data-catalog-filter'));
		});
	});

	bindCatalogItemForm();
}

function bindCatalogItemForm() {
	const form = document.getElementById('catalog-item-form');
	const cancel = document.getElementById('catalog-item-cancel');
	if (!form) {
		return;
	}

	document.querySelectorAll('.catalog-item-edit').forEach((button) => {
		button.addEventListener('click', () => {
			const row = button.closest('tr');
			if (!row) {
				return;
			}
			window.startCatalogItemEdit(row);
		});
	});

	if (cancel) {
		cancel.addEventListener('click', window.cancelCatalogItemEdit);
	}
}

window.startCatalogItemEdit = function (row) {
	const form = document.getElementById('catalog-item-form');
	const title = document.getElementById('catalog-item-form-title');
	const hint = document.getElementById('catalog-item-form-hint');
	const submit = document.getElementById('catalog-item-submit');
	const cancel = document.getElementById('catalog-item-cancel');
	const card = document.getElementById('catalog-item-form-card');

	document.getElementById('item_id').value = row.getAttribute('data-item-id') || '';
	document.getElementById('cod').value = row.getAttribute('data-item-cod') || '';
	document.getElementById('description').value = row.getAttribute('data-item-description') || '';
	document.getElementById('ust_normal').value = row.getAttribute('data-item-ust-normal') || '0';
	document.getElementById('ust_new_tech').value = row.getAttribute('data-item-ust-new-tech') || '';

	form.action = '/catalogs/items/update';
	if (title) title.textContent = 'Editar item';
	if (hint) hint.textContent = 'Altere os campos e salve para atualizar esta linha do catálogo.';
	if (submit) submit.textContent = 'Salvar';
	if (cancel) cancel.classList.remove('hidden');
	if (card) card.scrollIntoView({ behavior: 'smooth', block: 'start' });
	document.getElementById('cod').focus();
};

window.cancelCatalogItemEdit = function () {
	const form = document.getElementById('catalog-item-form');
	const title = document.getElementById('catalog-item-form-title');
	const hint = document.getElementById('catalog-item-form-hint');
	const submit = document.getElementById('catalog-item-submit');
	const cancel = document.getElementById('catalog-item-cancel');

	form.reset();
	document.getElementById('item_id').value = '';
	form.action = '/catalogs/items/create';
	if (title) title.textContent = 'Novo item';
	if (hint) hint.textContent = 'Inclua um serviço com código, descrição e valores em UST. Para alterar um existente, use Editar na linha.';
	if (submit) submit.textContent = 'Incluir';
	if (cancel) cancel.classList.add('hidden');
};

window.applyCatalogFilters = function () {
	const query = (document.getElementById('catalog-search')?.value || '').trim().toLowerCase();
	const active = document.querySelector('[data-catalog-filter].is-active');
	const prefix = active ? active.getAttribute('data-catalog-filter') : 'all';
	let visible = 0;

	document.querySelectorAll('[data-catalog-group]').forEach((group) => {
		const groupPrefix = group.getAttribute('data-catalog-group');
		const prefixMatch = prefix === 'all' || prefix === groupPrefix;
		let groupVisible = 0;

		group.querySelectorAll('[data-catalog-row]').forEach((row) => {
			const haystack = row.getAttribute('data-catalog-row') || '';
			const match = prefixMatch && (!query || haystack.includes(query));
			row.classList.toggle('hidden', !match);
			if (match) {
				groupVisible += 1;
			}
		});

		group.classList.toggle('hidden', groupVisible === 0);
		visible += groupVisible;
	});

	const empty = document.getElementById('catalog-empty');
	if (empty) {
		empty.classList.toggle('hidden', visible > 0);
	}

	const count = document.getElementById('catalog-count');
	if (count) {
		count.textContent = visible === 1 ? '1 serviço' : `${visible} serviços`;
	}
};

window.filterCatalogByPrefix = function (prefix) {
	document.querySelectorAll('[data-catalog-filter]').forEach((button) => {
		const isActive = button.getAttribute('data-catalog-filter') === prefix;
		button.classList.toggle('is-active', isActive);
		button.classList.toggle('border-slate-900', isActive);
		button.classList.toggle('bg-slate-900', isActive);
		button.classList.toggle('text-white', isActive);
		button.classList.toggle('border-slate-200', !isActive);
		button.classList.toggle('bg-white', !isActive);
		button.classList.toggle('text-slate-600', !isActive);
	});
	window.applyCatalogFilters();
};

if (document.readyState === 'loading') {
	document.addEventListener('DOMContentLoaded', bindCatalogPage);
} else {
	bindCatalogPage();
}
