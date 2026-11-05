// Template-aware OS row helpers (extends table.js)

let templateRowCounter = 0;

window.selectOSTemplate = function (id, name) {
    const input = document.getElementById('os_template_id_input');
    const text = document.getElementById('os-template-selected-text');
    if (input) input.value = id || '';
    if (text) {
        text.textContent = name || 'Selecione o template';
        text.classList.toggle('text-slate-900', Boolean(id));
        text.classList.toggle('text-gray-900', Boolean(id));
        text.classList.toggle('text-slate-400', !id);
        text.classList.toggle('text-gray-500', !id);
    }
    const menu = document.getElementById('dropdown-menu-os-template');
    if (menu) menu.classList.add('hidden');
};

window.syncJiraActivityPreview = function () {
    const suffixEl = document.getElementById('jira-activity-suffix');
    const previewEl = document.getElementById('jira-activity-preview');
    const hiddenEl = document.getElementById('activity_code_input');
    const suffix = (suffixEl?.value || '').trim().replace(/^G2MDS-?/i, '');
    const code = suffix ? `G2MDS-${suffix}` : '';
    if (previewEl) previewEl.textContent = code || 'G2MDS-0000';
    if (hiddenEl) hiddenEl.value = code;
};

window.applySelectedOSTemplate = async function () {
    const messageArea = document.getElementById('message-area');
    const templateId = document.getElementById('os_template_id_input')?.value;
    const taskId = document.getElementById('task_id_input')?.value;
    const sprintId = document.getElementById('sprint_id_input')?.value;
    const taskName = document.getElementById('task-selected-text')?.textContent || '';
    const pageKind = (document.getElementById('template_kind_input')?.value || '').toLowerCase();

    if (messageArea) messageArea.innerHTML = '';

    if (!templateId) {
        if (messageArea && window.createAlert) {
            messageArea.innerHTML = window.createAlert('warning', 'Selecione um template para aplicar.');
        }
        return;
    }
    if (!sprintId) {
        if (messageArea && window.createAlert) {
            messageArea.innerHTML = window.createAlert('warning', 'Selecione a sprint antes de aplicar o template.');
        }
        return;
    }

    try {
        const response = await fetch(`/api/os-templates/${templateId}/items`);
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || 'Falha ao carregar template');
        }
        const payload = await response.json();
        const items = Array.isArray(payload) ? payload : (payload.items || []);
        const templateKind = (Array.isArray(payload) ? pageKind : (payload.kind || pageKind || '')).toLowerCase();
        const isEpico = templateKind === 'epico';

        if (!Array.isArray(items) || items.length === 0) {
            if (messageArea && window.createAlert) {
                messageArea.innerHTML = window.createAlert('warning', 'Este template não possui itens.');
            }
            return;
        }

        if (typeof window.resetOSRows === 'function') {
            window.resetOSRows();
        } else {
            const tbody = document.getElementById('os-items-tbody');
            if (tbody) tbody.innerHTML = '';
        }

        const resolvedTaskId = taskId || '';
        const resolvedTaskName = resolvedTaskId ? taskName : '';
        const hisCodes = ['HIS01', 'HIS02'];
        const hisByCod = {};
        const otherItems = [];

        items.forEach((item) => {
            const cod = String(item.CatalogCod || '').toUpperCase();
            if (hisCodes.includes(cod)) {
                hisByCod[cod] = item;
            } else {
                otherItems.push(item);
            }
        });

        if (isEpico) {
            hisCodes.forEach((cod) => {
                if (hisByCod[cod]) return;
                const el = document.querySelector(`#dropdown-menu-catalog [data-catalog-cod="${cod}"]`);
                if (!el) return;
                hisByCod[cod] = {
                    CatalogID: parseInt(el.getAttribute('data-catalog-id'), 10),
                    CatalogCod: cod,
                    CatalogDesc: el.getAttribute('data-catalog-desc') || '',
                    UstNormal: parseInt(el.getAttribute('data-ust-normal'), 10) || 0,
                    UstNewTech: parseInt(el.getAttribute('data-ust-new-tech'), 10) || 0,
                };
            });
        }

        otherItems.forEach((item) => {
            const priced = window.resolveCatalogPricing(item.CatalogID, item.CatalogCod, item);
            window.addOSRow(
                priced.catalogId,
                priced.catalogCod,
                priced.catalogDesc,
                resolvedTaskId,
                resolvedTaskName,
                sprintId,
                priced.ustNormal,
                priced.ustNewTech
            );
        });

        if (isEpico) {
            const sprintTasks = await window.getSprintTasksForOS(sprintId, { force: true });
            if (!sprintTasks.length) {
                if (messageArea && window.createAlert) {
                    messageArea.innerHTML = window.createAlert(
                        'warning',
                        'Template OS Épico aplicado sem HIS01/HIS02: não há tarefas cadastradas nesta sprint.'
                    );
                }
            } else {
                // HIS01 para todas as tarefas (na ordem da listagem), depois HIS02 na mesma ordem
                hisCodes.forEach((cod) => {
                    const item = hisByCod[cod];
                    if (!item) return;
                    const priced = window.resolveCatalogPricing(item.CatalogID, item.CatalogCod || cod, item);
                    sprintTasks.forEach((task) => {
                        const obs = task.Name || '';
                        window.addOSRow(
                            priced.catalogId,
                            priced.catalogCod,
                            priced.catalogDesc,
                            task.ID || resolvedTaskId,
                            obs,
                            sprintId,
                            priced.ustNormal,
                            priced.ustNewTech,
                            obs
                        );
                    });
                });
                if (messageArea && window.createAlert) {
                    messageArea.innerHTML = window.createAlert(
                        'success',
                        `Template OS Épico aplicado: HIS01 e HIS02 na ordem das ${sprintTasks.length} tarefa(s) da sprint.`
                    );
                }
            }
        } else {
            hisCodes.forEach((cod) => {
                const item = hisByCod[cod];
                if (!item) return;
                const priced = window.resolveCatalogPricing(item.CatalogID, item.CatalogCod || cod, item);
                window.addOSRow(
                    priced.catalogId,
                    priced.catalogCod,
                    priced.catalogDesc,
                    resolvedTaskId,
                    resolvedTaskName,
                    sprintId,
                    priced.ustNormal,
                    priced.ustNewTech
                );
            });
            if (messageArea && window.createAlert) {
                messageArea.innerHTML = window.createAlert(
                    'success',
                    'Códigos do template copiados com sucesso. Revise os itens e defina a tarefa candidata, se necessário.'
                );
            }
        }
        if (typeof window.syncENC04UstFromENC02 === 'function') {
            window.syncENC04UstFromENC02(document.getElementById('os-items-tbody'));
        }
    } catch (error) {
        if (messageArea && window.createAlert) {
            messageArea.innerHTML = window.createAlert('error', `Erro ao aplicar template: ${error.message}`);
        }
    }
};

window.normalizeSprintTasksOrder = function (tasks) {
    return (Array.isArray(tasks) ? tasks : [])
        .slice()
        .map((task) => ({
            ID: Number(task.ID ?? task.id ?? 0),
            Name: String(task.Name ?? task.name ?? ''),
            SortOrder: Number(task.SortOrder ?? task.sort_order ?? 0),
        }))
        .filter((task) => task.ID > 0 || task.Name)
        .sort((a, b) => {
            if (a.SortOrder !== b.SortOrder) return a.SortOrder - b.SortOrder;
            return a.ID - b.ID;
        });
};

window.getSprintTasksForOS = async function (sprintId, opts = {}) {
    const force = Boolean(opts && opts.force);
    const cached = window.normalizeSprintTasksOrder(window.readSprintTasksJSON());
    if (!force && cached.length > 0 && String(window._sprintTasksSprintId || '') === String(sprintId)) {
        return cached;
    }
    try {
        const response = await fetch(`/api/tasks?sprint_id=${sprintId}`, {
            headers: { 'HX-Request': 'true' },
        });
        if (!response.ok) return cached;
        const tasks = await response.json();
        const list = window.normalizeSprintTasksOrder(tasks);
        window.writeSprintTasksJSON(list, sprintId);
        return list;
    } catch (e) {
        return cached;
    }
};

window.readSprintTasksJSON = function () {
    const el = document.getElementById('sprint-tasks-json');
    if (!el) return [];
    try {
        const parsed = JSON.parse(el.textContent || '[]');
        return Array.isArray(parsed) ? parsed : [];
    } catch (e) {
        return [];
    }
};

window.writeSprintTasksJSON = function (tasks, sprintId) {
    let el = document.getElementById('sprint-tasks-json');
    if (!el) {
        el = document.createElement('div');
        el.id = 'sprint-tasks-json';
        el.hidden = true;
        document.body.appendChild(el);
    }
    el.textContent = JSON.stringify(window.normalizeSprintTasksOrder(tasks));
    window._sprintTasksSprintId = sprintId || '';
};

window.selectTemplateCatalogFromEl = function (el) {
    if (!el) return;
    window.selectTemplateCatalog(
        el.getAttribute('data-catalog-id'),
        el.getAttribute('data-catalog-cod'),
        el.getAttribute('data-catalog-desc')
    );
};

window.selectTemplateCatalog = function (id, cod, description) {
    const menu = document.getElementById('dropdown-menu-catalog');
    if (menu) menu.classList.add('hidden');
    window.addTemplateRow({
        catalogId: id,
        catalogCod: cod,
        catalogDesc: description,
    });
};

window.addTemplateRow = function (opts) {
    const tbody = document.getElementById('os-template-items-tbody');
    const table = document.getElementById('os-items-table');
    if (!tbody) return;

    templateRowCounter += 1;
    if (table) table.classList.remove('hidden');

    const row = document.createElement('tr');
    row.id = `os-template-row-${templateRowCounter}`;
    row.className = 'border-b hover:bg-gray-50';
    row.innerHTML = `
        <td class="px-3 py-2 text-sm">
            <input type="hidden" name="catalog_id_${templateRowCounter}" value="${opts.catalogId}" />
            <div class="font-medium text-gray-900">${escapeTemplateHtml(opts.catalogCod || '')}</div>
            <div class="text-xs text-gray-500">${escapeTemplateHtml(opts.catalogDesc || '')}</div>
        </td>
        <td class="px-3 py-2 text-center">
            <button type="button" onclick="window.removeTemplateRow(${templateRowCounter})"
                class="text-red-600 hover:text-red-800 focus:outline-none"
                title="Remover">
                <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
            </button>
        </td>
    `;
    tbody.appendChild(row);
    if (table) table.scrollIntoView({ behavior: 'smooth', block: 'nearest' });

    const catalogInput = document.getElementById('catalog_id_input');
    if (catalogInput) catalogInput.value = '';
    if (opts.resetSearch !== false && typeof window.resetCatalogSearch === 'function') {
        window.resetCatalogSearch();
    }
};

window.removeTemplateRow = function (rowId) {
    const row = document.getElementById(`os-template-row-${rowId}`);
    if (row) row.remove();
    const tbody = document.getElementById('os-template-items-tbody');
    const table = document.getElementById('os-items-table');
    if (tbody && table && tbody.children.length === 0) {
        table.classList.add('hidden');
    }
};

function escapeTemplateHtml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

window.submitOSTemplateForm = function (event) {
    event.preventDefault();
    const messageArea = document.getElementById('message-area');
    if (messageArea) messageArea.innerHTML = '';

    const projectId = document.getElementById('page-project-id')?.value;
    const name = document.getElementById('template_name')?.value?.trim();
    const templateId = document.getElementById('template_id_input')?.value;
    const items = collectTemplateItems();

    if (!name) {
        if (messageArea && window.createAlert) {
            messageArea.innerHTML = window.createAlert('warning', 'Informe o nome do template.');
        }
        return false;
    }
    if (items.length === 0) {
        if (messageArea && window.createAlert) {
            messageArea.innerHTML = window.createAlert('warning', 'Adicione pelo menos um código ao template.');
        }
        return false;
    }

    const kind = document.getElementById('template_kind')?.value || 'os';
    const payload = {
        ProjectID: parseInt(projectId, 10),
        Name: name,
        Kind: kind,
        Items: items,
    };
    if (templateId) {
        payload.ID = parseInt(templateId, 10);
    }

    const url = templateId ? '/os/templates/update' : '/os/templates/create';

    fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'HX-Request': 'true',
        },
        body: JSON.stringify(payload),
    })
        .then(async (response) => {
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || 'Falha ao salvar template');
            }
            window.location.href = `/os/templates?project_id=${projectId}`;
        })
        .catch((error) => {
            if (messageArea && window.createAlert) {
                messageArea.innerHTML = window.createAlert('error', error.message);
            }
        });

    return false;
};

function collectTemplateItems() {
    const tbody = document.getElementById('os-template-items-tbody');
    if (!tbody) return [];
    const items = [];

    for (const row of tbody.getElementsByTagName('tr')) {
        const input = row.querySelector('input[name^="catalog_id_"]');
        const catalogId = parseInt(input?.value, 10);
        if (catalogId) {
            items.push({ CatalogID: catalogId });
        }
    }
    return items;
}

document.addEventListener('click', (event) => {
    const option = event.target.closest('#os-template-form #dropdown-menu-catalog [data-catalog-id]');
    if (!option) return;
    event.preventDefault();
    event.stopPropagation();
    window.selectTemplateCatalogFromEl(option);
});

document.addEventListener('DOMContentLoaded', () => {
    const initialEl = document.getElementById('os-template-initial-items');
    if (initialEl && typeof window.addTemplateRow === 'function') {
        try {
            const items = JSON.parse(initialEl.textContent || '[]');
            items.forEach((item) => {
                window.addTemplateRow({
                    catalogId: item.CatalogID,
                    catalogCod: item.CatalogCod,
                    catalogDesc: item.CatalogDesc,
                    resetSearch: false,
                });
            });
        } catch (e) {
            console.error('Failed to load template items', e);
        }
    }

    // Pre-selected template on Cadastro de OS: apply automatically when sprint is ready
    const sprintIdInput = document.getElementById('sprint_id_input');
    window._sprintTasksSprintId = sprintIdInput?.value || '';
    const templateInput = document.getElementById('os_template_id_input');
    const autoApplyID = templateInput?.getAttribute('data-auto-apply') || templateInput?.value;
    const sprintId = sprintIdInput?.value;
    const itemsTbody = document.getElementById('os-items-tbody');
    const alreadyFilled = itemsTbody && itemsTbody.children.length > 0;
    if (autoApplyID && sprintId && !alreadyFilled && typeof window.applySelectedOSTemplate === 'function') {
        window.applySelectedOSTemplate();
    }
});
