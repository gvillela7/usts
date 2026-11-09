// Table and Form Management
// Alert function is defined globally in alerts.js
let rowCounter = 0;

window.resetOSRows = function () {
    rowCounter = 0;
    const tbody = document.getElementById('os-items-tbody');
    if (tbody) tbody.innerHTML = '';
    const table = document.getElementById('os-items-table');
    if (table) table.classList.add('hidden');
    calculateTotalUST();
};

/** Resolve UST/description from the project catalog rendered on the page (source of truth). */
window.resolveCatalogPricing = function (catalogId, catalogCod, fallback) {
    const id = String(catalogId || '');
    const cod = String(catalogCod || '').toUpperCase();
    let el = null;
    if (id) {
        el = document.querySelector(`#dropdown-menu-catalog [data-catalog-id="${id}"]`);
    }
    if (!el && cod) {
        el = document.querySelector(`#dropdown-menu-catalog [data-catalog-cod="${cod}"]`);
    }

    const fb = fallback || {};
    if (!el) {
        const ustNormal = Number(fb.UstNormal ?? fb.ustNormal ?? fb.ust_normal ?? 0) || 0;
        const ustNewTech = Number(fb.UstNewTech ?? fb.ustNewTech ?? fb.ust_new_tech ?? ustNormal) || ustNormal;
        return {
            catalogId: Number(fb.CatalogID ?? fb.catalogId ?? catalogId) || 0,
            catalogCod: (fb.CatalogCod ?? fb.catalogCod ?? catalogCod) || '',
            catalogDesc: (fb.CatalogDesc ?? fb.catalogDesc ?? fb.Description) || '',
            ustNormal,
            ustNewTech,
        };
    }

    const ustNormal = parseInt(el.getAttribute('data-ust-normal'), 10) || 0;
    const ustNewTech = parseInt(el.getAttribute('data-ust-new-tech'), 10) || ustNormal;
    return {
        catalogId: parseInt(el.getAttribute('data-catalog-id'), 10) || Number(catalogId) || 0,
        catalogCod: el.getAttribute('data-catalog-cod') || catalogCod || '',
        catalogDesc: el.getAttribute('data-catalog-desc') || '',
        ustNormal,
        ustNewTech,
    };
};

window.addOSRow = function (catalogId, catalogCod, catalogDesc, taskId, taskName, sprintId, ustNormal, ustNewTech, observation) {
    // Always prefer live project catalog values over template/API leftovers (e.g. ust_item=1).
    const priced = window.resolveCatalogPricing(catalogId, catalogCod, {
        CatalogID: catalogId,
        CatalogCod: catalogCod,
        CatalogDesc: catalogDesc,
        UstNormal: ustNormal,
        UstNewTech: ustNewTech,
    });
    const normal = priced.ustNormal;
    const newTech = priced.ustNewTech || normal;

    window.addOSConfiguredRow({
        catalogId: priced.catalogId || catalogId,
        catalogCod: priced.catalogCod || catalogCod,
        catalogDesc: priced.catalogDesc || catalogDesc,
        taskId,
        taskName,
        sprintId,
        ustNormal: normal,
        ustNewTech: newTech,
        ustItem: normal,
        ustRecurrence: 1,
        complexity: 1,
        ustEnd: normal * 1 * 1,
        kind: 'normal',
        description: priced.catalogDesc || catalogDesc || '',
        observation: observation || '',
        tarefaCandidta: 0,
        skipTaskColumn: true,
        clearCatalogSelection: true,
    });
};

window.addOSConfiguredRow = function (opts) {
    const {
        catalogId,
        catalogCod,
        catalogDesc,
        taskId = '',
        taskName = '',
        sprintId = '',
        ustNormal = 0,
        ustNewTech = 0,
        ustItem = ustNormal,
        ustRecurrence = 1,
        complexity = 1,
        ustEnd = null,
        kind = 'normal',
        description = '',
        observation = '',
        tarefaCandidta = 0,
        skipTaskColumn = false,
        clearCatalogSelection = false,
    } = opts;

    rowCounter++;
    const tbody = document.getElementById('os-items-tbody');
    const table = document.getElementById('os-items-table');

    if (!tbody) return;
    if (table) table.classList.remove('hidden');

    const isNewTech = kind === 'nova_tecnologia' || kind === 'ajuste';
    const catalogUst = isNewTech ? (ustNewTech || ustNormal) : ustNormal;
    const resolvedUstItem = catalogUst || ustItem || ustNormal || 1;
    const initialUstFim = ustEnd != null ? ustEnd : (resolvedUstItem * ustRecurrence * complexity);
    const kindNormalSelected = isNewTech ? '' : 'selected';
    const kindNewTechSelected = isNewTech ? 'selected' : '';
    const candidataSim = Number(tarefaCandidta) === 1;
    const safeDesc = escapeHtml(description || '');
    const safeObs = escapeHtml(observation || '');
    const safeCod = escapeHtml(catalogCod || '');
    const safeTaskName = escapeHtml(taskName || '');

    const hiddenTaskFields = skipTaskColumn
        ? `<input type="hidden" name="task_id_${rowCounter}" value="${taskId}" />
            <input type="hidden" name="sprint_id_${rowCounter}" value="${sprintId}" />`
        : `<input type="hidden" name="task_id_${rowCounter}" value="${taskId}" />
            <input type="hidden" name="sprint_id_${rowCounter}" value="${sprintId}" />`;

    const row = document.createElement('tr');
    row.id = `os-row-${rowCounter}`;
    row.className = 'border-b border-slate-100 hover:bg-slate-50/80';
    row.setAttribute('data-ust-normal', ustNormal);
    row.setAttribute('data-ust-new-tech', ustNewTech || ustNormal);
    row.setAttribute('data-catalog-id', catalogId);
    row.setAttribute('data-catalog-cod', (catalogCod || '').toUpperCase());

    row.innerHTML = `
        <td class="px-2 py-3 align-middle">
            <button type="button" onclick="window.moveOSRow(${rowCounter}, event.shiftKey ? -1 : 1)" class="os-drag-handle flex h-8 w-8 cursor-grab items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600" title="Reordenar (Shift+clique para subir)">
                <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                    <path d="M7 5a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM7 10a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM7 15a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM10 5a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM10 10a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM10 15a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0z"/>
                </svg>
            </button>
        </td>
        <td class="px-3 py-3 align-middle whitespace-nowrap">
            <input type="hidden" name="catalog_id_${rowCounter}" value="${catalogId}" />
            ${hiddenTaskFields}
            <div class="text-sm font-semibold text-indigo-700">${safeCod}</div>
            ${skipTaskColumn ? '' : `<div class="mt-0.5 text-xs text-slate-500 os-row-task-name">${safeTaskName || '—'}</div>`}
        </td>
        <td class="px-3 py-3 align-middle">
            <textarea name="description_${rowCounter}"
                class="w-full min-w-[260px] resize-y rounded-lg border border-slate-200 bg-white px-2.5 py-2 text-sm leading-5 text-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                rows="2"
                required>${safeDesc}</textarea>
        </td>
        <td class="px-3 py-3 align-middle">
            <select name="kind_${rowCounter}"
                class="w-full min-w-[140px] appearance-auto rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500"
                onchange="handleTypeChange(${rowCounter})"
                required>
                <option value="normal" ${kindNormalSelected}>Normal</option>
                <option value="nova_tecnologia" ${kindNewTechSelected}>Nova tecnologia</option>
            </select>
        </td>
        <td class="px-2 py-3 align-middle text-center">
            <input type="number"
                id="ust_item_${rowCounter}"
                name="ust_item_${rowCounter}"
                value="${resolvedUstItem}"
                class="mx-auto w-16 rounded-lg border border-slate-200 bg-slate-50 px-2 py-1.5 text-center text-sm text-slate-700 focus:outline-none"
                title="UST do catálogo (Normal ou Nova tecnologia, conforme o Tipo)"
                required min="1"
                readonly />
        </td>
        <td class="px-2 py-3 align-middle text-center">
            <input type="number"
                id="ust_recurrence_${rowCounter}"
                name="ust_recurrence_${rowCounter}"
                value="${ustRecurrence}"
                onchange="calculateUstFim(${rowCounter})"
                class="mx-auto w-14 rounded-lg border border-slate-200 px-2 py-1.5 text-center text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500"
                required min="1" />
        </td>
        <td class="px-2 py-3 align-middle text-center">
            <input type="number"
                id="complexity_${rowCounter}"
                name="complexity_${rowCounter}"
                value="${complexity}"
                onchange="calculateUstFim(${rowCounter})"
                class="mx-auto w-14 rounded-lg border border-slate-200 px-2 py-1.5 text-center text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500"
                required min="1" max="10" />
        </td>
        <td class="px-2 py-3 align-middle text-center">
            <input type="number"
                id="ust_end_${rowCounter}"
                name="ust_end_${rowCounter}"
                value="${initialUstFim}"
                class="mx-auto w-16 rounded-lg border border-slate-200 bg-slate-50 px-2 py-1.5 text-center text-sm text-slate-700 focus:outline-none"
                title="UST × Rec. × Comp."
                min="1"
                readonly />
        </td>
        <td class="px-3 py-3 align-middle">
            <input type="text"
                name="observation_${rowCounter}"
                value="${safeObs}"
                class="w-full min-w-[200px] rounded-lg border border-slate-200 px-2.5 py-1.5 text-sm text-slate-800 focus:outline-none focus:ring-1 focus:ring-indigo-500"
                maxlength="1000"
                placeholder="Observação"
                title="Observação (até 1000 caracteres)" />
        </td>
        <td class="px-3 py-3 align-middle text-center">
            <div class="inline-flex rounded-lg border border-slate-200 bg-white p-0.5" data-candidata-group="${rowCounter}">
                <label class="cursor-pointer">
                    <input type="radio" name="tarefa_candidata_${rowCounter}" value="0" class="peer sr-only" ${candidataSim ? '' : 'checked'} onchange="syncCandidataToggle(${rowCounter})" />
                    <span class="block rounded-md px-2.5 py-1 text-xs font-medium text-slate-600 peer-checked:bg-indigo-600 peer-checked:text-white">Não</span>
                </label>
                <label class="cursor-pointer">
                    <input type="radio" name="tarefa_candidata_${rowCounter}" value="1" class="peer sr-only" ${candidataSim ? 'checked' : ''} onchange="syncCandidataToggle(${rowCounter})" />
                    <span class="block rounded-md px-2.5 py-1 text-xs font-medium text-slate-600 peer-checked:bg-indigo-600 peer-checked:text-white">Sim</span>
                </label>
            </div>
        </td>
        <td class="px-2 py-3 align-middle text-center">
            <button type="button" onclick="window.removeOSRow(${rowCounter})"
                class="inline-flex h-8 w-8 items-center justify-center rounded-lg text-red-500 hover:bg-red-50 hover:text-red-700 focus:outline-none"
                title="Remover">
                <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
            </button>
        </td>
    `;
    tbody.appendChild(row);

    if (clearCatalogSelection) {
        const catalogInput = document.getElementById('catalog_id_input');
        if (catalogInput) catalogInput.value = '';
        if (typeof window.resetCatalogSearch === 'function') {
            window.resetCatalogSearch();
        }
    }

    calculateTotalUST();
    if (typeof window.syncENC04UstFromENC02 === 'function') {
        window.syncENC04UstFromENC02(tbody);
    }
};

function escapeHtml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

// Calculate UST Final = UST × Recorrência × Complexidade
window.calculateUstFim = function (rowId) {
    const ustItemInput = document.getElementById(`ust_item_${rowId}`);
    const recurrenceInput = document.getElementById(`ust_recurrence_${rowId}`);
    const complexityInput = document.getElementById(`complexity_${rowId}`);
    const ustEndInput = document.getElementById(`ust_end_${rowId}`);

    if (ustItemInput && complexityInput && ustEndInput) {
        const ustItem = parseInt(ustItemInput.value) || 0;
        const recurrence = parseInt(recurrenceInput?.value) || 1;
        const complexity = parseInt(complexityInput.value) || 1;
        ustEndInput.value = ustItem * recurrence * complexity;
        calculateTotalUST();
    }
};

// Calculate total UST from all rows
window.calculateTotalUST = function () {
    const tbody = document.getElementById('os-items-tbody');
    if (!tbody) return;
    const rows = tbody.getElementsByTagName('tr');
    let total = 0;

    for (let row of rows) {
        const rowId = row.id.split('-')[2];
        const ustEndInput = document.getElementById(`ust_end_${rowId}`);
        if (ustEndInput) {
            total += parseInt(ustEndInput.value) || 0;
        }
    }

    const totalElement = document.getElementById('ust-total');
    if (totalElement) {
        totalElement.textContent = total;
    }
};

// Sync UST column from catalog: Normal → ust_normal, Nova tecnologia → ust_new_tech
window.handleTypeChange = function (rowId) {
    const row = document.getElementById(`os-row-${rowId}`);
    if (!row) return;

    const catalogCod = (row.getAttribute('data-catalog-cod') || row.querySelector('.text-indigo-700')?.textContent || '').trim().toUpperCase();
    if (catalogCod === 'ENC04') {
        if (typeof window.syncENC04UstFromENC02 === 'function') {
            window.syncENC04UstFromENC02(document.getElementById('os-items-tbody'));
        }
        return;
    }

    const typeSelect = row.querySelector(`select[name="kind_${rowId}"]`);
    const ustItemInput = document.getElementById(`ust_item_${rowId}`);
    if (!typeSelect || !ustItemInput) return;

    const catalogId = row.getAttribute('data-catalog-id');
    const priced = window.resolveCatalogPricing(catalogId, catalogCod, {
        CatalogID: catalogId,
        CatalogCod: catalogCod,
        UstNormal: row.getAttribute('data-ust-normal'),
        UstNewTech: row.getAttribute('data-ust-new-tech'),
    });
    const ustNormal = priced.ustNormal || parseInt(row.getAttribute('data-ust-normal'), 10) || 0;
    const ustNewTech = priced.ustNewTech || ustNormal;
    row.setAttribute('data-ust-normal', String(ustNormal));
    row.setAttribute('data-ust-new-tech', String(ustNewTech));

    if (typeSelect.value === 'nova_tecnologia' || typeSelect.value === 'ajuste') {
        ustItemInput.value = String(ustNewTech || ustNormal || 1);
    } else {
        ustItemInput.value = String(ustNormal || 1);
    }

    calculateUstFim(rowId);
    if (typeof window.syncENC04UstFromENC02 === 'function') {
        window.syncENC04UstFromENC02(document.getElementById('os-items-tbody'));
    }
};

window.moveOSRow = function (rowId, direction) {
    const row = document.getElementById(`os-row-${rowId}`);
    const tbody = document.getElementById('os-items-tbody');
    if (!row || !tbody) return;
    if (direction < 0 && row.previousElementSibling) {
        tbody.insertBefore(row, row.previousElementSibling);
    } else if (direction > 0 && row.nextElementSibling) {
        tbody.insertBefore(row.nextElementSibling, row);
    }
};

window.syncCandidataToggle = function () {
    // peer-checked styling handled via CSS
};

window.removeOSRow = function (rowId) {
    const row = document.getElementById(`os-row-${rowId}`);
    if (row) {
        row.remove();
        const tbody = document.getElementById('os-items-tbody');
        if (tbody.children.length === 0) {
            document.getElementById('os-items-table').classList.add('hidden');
        }
        calculateTotalUST();
    }
};

function collectOSItems() {
    const tbody = document.getElementById('os-items-tbody');
    const rows = tbody.getElementsByTagName('tr');
    const items = [];

    // Get project ID from hidden input
    const projectIdInput = document.getElementById('page-project-id');
    const projectId = projectIdInput ? projectIdInput.value : null;

    if (!projectId) {
        console.error("Project ID not found");
        return [];
    }

    for (let row of rows) {
        const rowId = row.id.split('-')[2];
        const catalogId = parseInt(row.querySelector(`input[name="catalog_id_${rowId}"]`).value);
        const taskId = parseInt(row.querySelector(`input[name="task_id_${rowId}"]`).value);
        const sprintIdInput = row.querySelector(`input[name="sprint_id_${rowId}"]`);
        const sprintId = sprintIdInput ? parseInt(sprintIdInput.value) : null;
        const ustItem = parseInt(row.querySelector(`input[name="ust_item_${rowId}"]`).value);
        const ustRecurrence = parseInt(row.querySelector(`input[name="ust_recurrence_${rowId}"]`).value);
        const complexity = parseInt(row.querySelector(`input[name="complexity_${rowId}"]`).value);
        const ustEndInput = row.querySelector(`input[name="ust_end_${rowId}"]`).value;
        const computedEnd = ustItem * (ustRecurrence || 1) * (complexity || 1);
        const ustEnd = ustEndInput ? parseInt(ustEndInput) : computedEnd;
        const kind = row.querySelector(`select[name="kind_${rowId}"]`).value;
        const description = row.querySelector(`textarea[name="description_${rowId}"]`).value;
        const observation = (row.querySelector(`input[name="observation_${rowId}"], textarea[name="observation_${rowId}"]`)?.value || '').trim();

        // Get selected radio button value for tarefa_candidata
        const tarefaCandidtaRadio = row.querySelector(`input[name="tarefa_candidata_${rowId}"]:checked`);
        const tarefaCandidta = tarefaCandidtaRadio ? parseInt(tarefaCandidtaRadio.value) : 0;

        items.push({
            ProjectID: parseInt(projectId),
            CatalogID: catalogId,
            TaskID: taskId,
            SprintID: sprintId,
            USTItem: ustItem,
            USTRecurrence: ustRecurrence,
            Complexity: complexity,
            USTEnd: ustEnd,
            Kind: kind,
            Description: description,
            Observation: observation,
            TarefaCandidta: tarefaCandidta
        });
    }

    return items;
}

window.submitForm = function (event) {
    event.preventDefault();
    const form = event.target;
    const messageArea = document.getElementById('message-area');
    messageArea.innerHTML = '';

    const sprintId = document.getElementById('sprint_id_input')?.value;
    const activityCode = (document.getElementById('activity_code_input')?.value || '').trim();
    const projectId = document.getElementById('page-project-id')?.value;

    if (!sprintId) {
        messageArea.innerHTML = window.createAlert('warning', 'Selecione a sprint antes de criar a OS.');
        return false;
    }
    if (!activityCode || activityCode === 'G2MDS-') {
        messageArea.innerHTML = window.createAlert('warning', 'Informe o código da atividade Jira antes de criar a OS.');
        return false;
    }

    const responsavel = typeof window.getResponsavelValue === 'function'
        ? window.getResponsavelValue()
        : '';
    if (!responsavel) {
        messageArea.innerHTML = window.createAlert('warning', 'Informe o responsável da OS.');
        return false;
    }

    if (typeof window.syncENC04UstFromENC02 === 'function') {
        window.syncENC04UstFromENC02(document.getElementById('os-items-tbody'));
    }

    const osItems = collectOSItems();

    if (osItems.length === 0) {
        messageArea.innerHTML = window.createAlert('warning', 'Adicione pelo menos um item de OS.');
        return false;
    }

    for (const item of osItems) {
        if (Array.from(item.Observation || '').length > 1000) {
            messageArea.innerHTML = window.createAlert('warning', 'Observação deve ter no máximo 1000 caracteres.');
            return false;
        }
    }

    const payload = {
        ProjectID: parseInt(projectId, 10),
        SprintID: parseInt(sprintId, 10),
        ActivityCode: activityCode,
        Responsavel: responsavel,
        TemplateKind: (document.getElementById('template_kind_input')?.value || 'os').toLowerCase(),
        Items: osItems.map((item) => ({
            ...item,
            ProjectID: parseInt(projectId, 10),
            SprintID: parseInt(sprintId, 10),
        })),
    };

    fetch(form.action, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'HX-Request': 'true'
        },
        body: JSON.stringify(payload)
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    messageArea.innerHTML = window.createAlert('error', text);
                });
            }
            const redirect = response.headers.get('HX-Redirect');
            if (redirect) {
                window.location.href = redirect;
                return;
            }
            document.getElementById('os-items-tbody').innerHTML = '';
            document.getElementById('os-items-table').classList.add('hidden');
            messageArea.innerHTML = window.createAlert('success', 'OS criada com sucesso!');
            rowCounter = 0;
        })
        .catch(error => {
            messageArea.innerHTML = window.createAlert('error', `Erro ao criar OS: ${error.message}`);
        });

    return false;
};
