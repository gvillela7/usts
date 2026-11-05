function closeAppMenu() {
	const sidebar = document.getElementById('app-sidebar');
	const backdrop = document.getElementById('app-sidebar-backdrop');
	const toggle = document.getElementById('menu-toggle');
	if (sidebar) sidebar.classList.remove('is-open');
	if (backdrop) backdrop.classList.remove('is-open');
	if (toggle) toggle.setAttribute('aria-expanded', 'false');
}

function openAppMenu() {
	const sidebar = document.getElementById('app-sidebar');
	const backdrop = document.getElementById('app-sidebar-backdrop');
	const toggle = document.getElementById('menu-toggle');
	if (sidebar) sidebar.classList.add('is-open');
	if (backdrop) backdrop.classList.add('is-open');
	if (toggle) toggle.setAttribute('aria-expanded', 'true');
}

function bindAppMenu() {
	const toggle = document.getElementById('menu-toggle');
	const backdrop = document.getElementById('app-sidebar-backdrop');
	if (toggle) {
		toggle.addEventListener('click', () => {
			const sidebar = document.getElementById('app-sidebar');
			if (sidebar && sidebar.classList.contains('is-open')) {
				closeAppMenu();
			} else {
				openAppMenu();
			}
		});
	}
	if (backdrop) {
		backdrop.addEventListener('click', closeAppMenu);
	}
	document.addEventListener('keydown', (event) => {
		if (event.key === 'Escape') {
			closeAppMenu();
		}
	});
}

if (document.readyState === 'loading') {
	document.addEventListener('DOMContentLoaded', bindAppMenu);
} else {
	bindAppMenu();
}
