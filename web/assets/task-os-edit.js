// Task OS inline editor (sprint task detail)

function getTaskOSCatalogs() {
    const el = document.getElementById('task-os-catalogs-json');
    if (!el) return [];
    try {
        return JSON.parse(el.textContent || '[]');
    } catch (e) {
        return [];
    }
}

function taskOSRowPricing(row) {
    const catalogId = parseInt(row.getAttribute('data-catalog-id'), 10);
    const ustNormal = parseInt(row.getAttribute('data-ust-normal'), 10);
    const ustNewTech = parseInt(row.getAttribute('data-ust-new-tech'), 10);
    if (ustNormal > 0) {
        return { ustNormal, ustNewTech: ustNewTech > 0 ? ustNewTech : ustNormal };
    }
    const catalogs = getTaskOSCatalogs();
    const item = catalogs.find((c) => Number(c.id ?? c.ID) === catalogId);
    if (!item) return { ustNormal: 0, ustNewTech: 0 };
    const normal = Number(item.ustNormal ?? item.UstNormal ?? 0);
    const newTech = Number(item.ustNewTech ?? item.UstNewTech ?? normal) || normal;
    return { ustNormal: normal, ustNewTech: newTech };
}

function refreshTaskOSRowCatalogPricing(row) {
    const catalogId = parseInt(row.getAttribute('data-catalog-id'), 10) || 0;
    if (!catalogId) return;
    const catalogs = getTaskOSCatalogs();
    const item = catalogs.find((c) => Number(c.id ?? c.ID) === catalogId);
    if (!item) return;
    const normal = Number(item.ustNormal ?? item.UstNormal ?? 0);
    const newTech = Number(item.ustNewTech ?? item.UstNewTech ?? normal) || normal;
    if (normal > 0) {
        row.setAttribute('data-ust-normal', String(normal));
        row.setAttribute('data-ust-new-tech', String(newTech));
    }
}

window.ustFromCatalogKind = function (ustNormal, ustNewTech, kind) {
    const k = (kind || 'normal').trim();
    if (k === 'nova_tecnologia' || k === 'ajuste') {
        return ustNewTech > 0 ? ustNewTech : ustNormal;
    }
    return ustNormal;
};

window.moveTaskOSRow = function (btn, direction) {
    const row = btn.closest('tr');
    const tbody = document.getElementById('task-os-edit-tbody');
    if (!row || !tbody) return;
    if (direction < 0 && row.previousElementSibling) {
        tbody.insertBefore(row, row.previousElementSibling);
    } else if (direction > 0 && row.nextElementSibling) {
        tbody.insertBefore(row.nextElementSibling, row);
    }
    refreshTaskOSMeta();
};

window.removeTaskOSRow = function (btn) {
    const row = btn.closest('tr');
    if (row) row.remove();
    refreshTaskOSMeta();
};

window.focusTaskOSRow = function (btn) {
    const row = btn?.closest('tr.task-os-row');
    const desc = row?.querySelector('.os-desc');
    if (desc) {
        desc.focus();
        if (typeof desc.select === 'function') desc.select();
    }
};

window.addCatalogFromTaskOSSelect = function () {
    const select = document.getElementById('task-os-add-catalog');
    const message = document.getElementById('task-os-message');
    if (!select) return;
    if (!select.value) {
        if (message) {
            message.innerHTML = window.createAlert
                ? window.createAlert('warning', 'Selecione um código do catálogo antes de adicionar.')
                : 'Selecione um código do catálogo antes de adicionar.';
        }
        return;
    }
    if (message) message.innerHTML = '';
    window.addCatalogToTaskOS(select);
};

function updateTaskOSCodeCount() {
    const n = document.querySelectorAll('#task-os-edit-tbody tr.task-os-row').length;
    const el = document.getElementById('task-os-code-count');
    if (!el) return;
    el.textContent = n === 1 ? '1 código adicionado' : `${n} códigos adicionados`;
}

function updateTaskOSUstFinalTotal() {
    let total = 0;
    document.querySelectorAll('#task-os-edit-tbody tr.task-os-row .os-ust-final').forEach((input) => {
        total += parseInt(input.value, 10) || 0;
    });
    const el = document.getElementById('task-os-ust-final-total');
    if (el) el.textContent = String(total);
}

function updateTaskOSRowNumbers() {
    document.querySelectorAll('#task-os-edit-tbody tr.task-os-row').forEach((row, i) => {
        const num = row.querySelector('.task-os-row-num');
        if (num) num.textContent = String(i + 1);
    });
}

function refreshTaskOSMeta() {
    updateTaskOSCodeCount();
    updateTaskOSRowNumbers();
    updateTaskOSUstFinalTotal();
}

function syncTaskOSDescriptions() {
    document.querySelectorAll('#task-os-edit-tbody tr.task-os-row').forEach((row) => {
        const catalogDesc = (row.getAttribute('data-catalog-desc') || '').trim();
        const descInput = row.querySelector('.os-desc');
        if (!descInput || !catalogDesc) return;
        const current = (descInput.value || '').trim();
        if (!current || (current.length < 30 && catalogDesc.length > current.length)) {
            descInput.value = catalogDesc;
        }
    });
}

window.syncTaskOSUstEnd = function (rowOrInput) {
    const row = rowOrInput?.closest?.('tr.task-os-row') || rowOrInput;
    if (!row || !row.querySelector) return;
    const ustItem = parseInt(row.querySelector('.os-ust-item')?.value, 10) || 0;
    const recurrence = parseInt(row.querySelector('.os-ust-rec')?.value, 10) || 1;
    const complexity = parseInt(row.querySelector('.os-complexity')?.value, 10) || 1;
    const finalInput = row.querySelector('.os-ust-final');
    if (finalInput) {
        finalInput.value = String(ustItem * recurrence * complexity);
    }
    updateTaskOSUstFinalTotal();
};

window.syncTaskOSKindUST = function (select) {
    const row = select?.closest('tr.task-os-row') || select?.closest('tr');
    if (!row) return;

    const cod = (row.getAttribute('data-catalog-cod') || row.querySelector('.text-indigo-700')?.textContent || '').trim().toUpperCase();
    if (cod === 'ENC04') {
        window.syncENC04UstFromENC02(row.closest('tbody'));
        return;
    }

    const ustInput = row.querySelector('.os-ust-item');
    if (!ustInput) return;

    const { ustNormal, ustNewTech } = taskOSRowPricing(row);
    const kind = (select.value || 'normal').trim();
    ustInput.value = String(window.ustFromCatalogKind(ustNormal, ustNewTech, kind));
    window.syncTaskOSUstEnd(row);
};

const taskOSDragHandleHtml = `
    <button type="button" class="task-os-drag-handle flex h-8 w-8 cursor-grab items-center justify-center rounded-lg text-slate-400 hover:bg-slate-100 hover:text-slate-600" title="Arrastar para reordenar" draggable="true">
        <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
            <path d="M7 5a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM7 10a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM7 15a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM10 5a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM10 10a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zM10 15a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0z"/>
        </svg>
    </button>
`;

const taskOSActionsHtml = `
    <div class="flex items-center justify-center gap-1">
        <button type="button" class="task-os-focus-row flex h-8 w-8 items-center justify-center rounded-lg text-indigo-600 hover:bg-indigo-50" title="Editar descrição" onclick="focusTaskOSRow(this)">
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z"></path>
            </svg>
        </button>
        <button type="button" class="flex h-8 w-8 items-center justify-center rounded-lg text-red-600 hover:bg-red-50" title="Excluir" onclick="removeTaskOSRow(this)">
            <svg class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm4 0a1 1 0 012 0v6a1 1 0 11-2 0V8z" clip-rule="evenodd"></path>
            </svg>
        </button>
    </div>
`;

function taskOSOrderCellHtml(rowNum) {
    return `
        <div class="flex items-center gap-0.5">
            ${taskOSDragHandleHtml}
            <span class="task-os-row-num w-4 text-center text-xs font-medium tabular-nums text-slate-400">${rowNum}</span>
        </div>
    `;
}

window.addCatalogToTaskOS = function (select) {
    const option = select.options[select.selectedIndex];
    if (!option || !option.value) return;

    const catalogId = option.value;
    const cod = option.getAttribute('data-cod') || '';
    const desc = option.getAttribute('data-desc') || '';
    const ustNormal = parseInt(option.getAttribute('data-ust-normal'), 10) || 0;
    const ustNewTech = parseInt(option.getAttribute('data-ust-new-tech'), 10) || ustNormal;

    const tbody = document.getElementById('task-os-edit-tbody');
    if (!tbody) return;

    const rowNum = tbody.querySelectorAll('tr.task-os-row').length + 1;

    const tr = document.createElement('tr');
    tr.className = 'task-os-row border-t border-slate-100 hover:bg-slate-50/60';
    tr.setAttribute('data-os-id', '0');
    tr.setAttribute('data-catalog-id', catalogId);
    tr.setAttribute('data-catalog-cod', cod.toUpperCase());
    tr.setAttribute('data-catalog-desc', desc);
    tr.setAttribute('data-ust-normal', String(ustNormal));
    tr.setAttribute('data-ust-new-tech', String(ustNewTech));
    tr.innerHTML = `
        <td class="px-2 py-2 align-middle">${taskOSOrderCellHtml(rowNum)}</td>
        <td class="px-2 py-2 align-middle text-sm font-semibold text-indigo-700 whitespace-nowrap">${escapeHtml(cod)}</td>
        <td class="px-2 py-2 align-middle">
            <textarea class="os-desc task-os-field-compact w-full rounded-lg border border-slate-200 px-2 py-1.5 text-sm text-slate-800" rows="2">${escapeHtml(desc)}</textarea>
        </td>
        <td class="px-2 py-2 align-middle">
            <select class="os-kind w-full max-w-[9rem] rounded-lg border border-slate-200 px-2 py-1.5 text-sm">
                <option value="normal" selected>Normal</option>
                <option value="nova_tecnologia">Nova tecnologia</option>
            </select>
        </td>
        <td class="px-2 py-2 align-middle text-center">
            <input type="number" min="1" class="os-ust-item os-num-input rounded-lg border border-slate-200 bg-slate-50 px-1 py-1.5 text-sm tabular-nums text-slate-700" value="${ustNormal}" readonly title="UST do catálogo conforme o Tipo"/>
        </td>
        <td class="px-2 py-2 align-middle text-center">
            <input type="number" min="1" class="os-ust-rec os-num-input rounded-lg border border-slate-200 px-1 py-1.5 text-sm tabular-nums" value="1"/>
        </td>
        <td class="px-2 py-2 align-middle text-center">
            <input type="number" min="1" max="10" class="os-complexity os-num-input rounded-lg border border-slate-200 px-1 py-1.5 text-sm tabular-nums" value="1"/>
        </td>
        <td class="px-2 py-2 align-middle text-center">
            <input type="number" min="1" class="os-ust-final os-num-input rounded-lg border border-slate-200 bg-slate-50 px-1 py-1.5 text-sm font-medium tabular-nums text-slate-700" value="${ustNormal}" readonly title="UST × Rec. × Comp."/>
        </td>
        <td class="px-2 py-2 align-middle">
            <textarea class="os-obs task-os-field-compact w-full rounded-lg border border-slate-200 px-2 py-1.5 text-sm" rows="1" maxlength="1000" placeholder="Observação (até 1000)"></textarea>
        </td>
        <td class="px-2 py-2 align-middle text-center">
            <select class="os-candidata rounded-lg border border-slate-200 px-2 py-1.5 text-sm">
                <option value="0" selected>Não</option>
                <option value="1">Sim</option>
            </select>
        </td>
        <td class="px-2 py-2 align-middle">${taskOSActionsHtml}</td>
    `;
    tbody.appendChild(tr);
    select.value = '';
    refreshTaskOSMeta();
    if (typeof window.syncENC04UstFromENC02 === 'function') {
        window.syncENC04UstFromENC02(tbody);
    }
};

window.saveTaskOSEditor = async function () {
    const editor = document.getElementById('task-os-editor');
    const message = document.getElementById('task-os-message');
    const saveBtn = document.querySelector('[data-task-os-save]');
    if (!editor) return;

    const taskId = editor.getAttribute('data-task-id');
    if (!taskId || taskId === '0') {
        if (message) {
            message.innerHTML = window.createAlert
                ? window.createAlert('error', 'ID da tarefa inválido. Recarregue a página.')
                : 'ID da tarefa inválido.';
        }
        return;
    }

    const tbody = document.getElementById('task-os-edit-tbody');
    if (typeof window.syncENC04UstFromENC02 === 'function') {
        window.syncENC04UstFromENC02(tbody);
    }

    const rows = Array.from(document.querySelectorAll('#task-os-edit-tbody tr.task-os-row'));
    const items = [];

    const responsavel = typeof window.getResponsavelValue === 'function'
        ? window.getResponsavelValue(editor)
        : '';
    if (!responsavel) {
        if (message) {
            message.innerHTML = window.createAlert
                ? window.createAlert('warning', 'Informe o responsável da OS.')
                : 'Informe o responsável da OS.';
        }
        return;
    }

    for (const row of rows) {
        const observation = (row.querySelector('.os-obs')?.value || '').trim();
        if (Array.from(observation).length > 1000) {
            if (message) {
                message.innerHTML = window.createAlert
                    ? window.createAlert('warning', 'Observação deve ter no máximo 1000 caracteres.')
                    : 'Observação deve ter no máximo 1000 caracteres.';
            }
            return;
        }
        const kindSelect = row.querySelector('.os-kind');
        if (kindSelect) window.syncTaskOSKindUST(kindSelect);

        const catalogId = parseInt(row.getAttribute('data-catalog-id'), 10) || 0;
        if (catalogId <= 0) {
            if (message) {
                message.innerHTML = window.createAlert
                    ? window.createAlert('error', 'Um ou mais itens estão sem código do catálogo. Remova e adicione novamente.')
                    : 'Item sem código do catálogo.';
            }
            return;
        }

        const ustItem = parseInt(row.querySelector('.os-ust-item')?.value, 10) || 0;
        const recurrence = parseInt(row.querySelector('.os-ust-rec')?.value, 10) || 1;
        const complexity = parseInt(row.querySelector('.os-complexity')?.value, 10) || 1;
        window.syncTaskOSUstEnd(row);
        const ustFinal = parseInt(row.querySelector('.os-ust-final')?.value, 10) || (ustItem * recurrence * complexity);
        items.push({
            ID: parseInt(row.getAttribute('data-os-id'), 10) || 0,
            CatalogID: catalogId,
            USTItem: ustItem,
            USTRecurrence: recurrence,
            Complexity: complexity,
            USTEnd: ustFinal,
            Kind: kindSelect?.value || 'normal',
            Description: row.querySelector('.os-desc')?.value || '',
            Observation: observation,
            TarefaCandidta: parseInt(row.querySelector('.os-candidata')?.value, 10) || 0,
        });
    }

    if (message) message.innerHTML = '';
    if (saveBtn) saveBtn.disabled = true;

    try {
        const response = await fetch(`/tasks/${taskId}/os`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'HX-Request': 'true',
            },
            body: JSON.stringify({ Responsavel: responsavel, Items: items }),
        });
        const text = await response.text();
        if (!response.ok) {
            throw new Error(text || 'Falha ao salvar');
        }
        let payload = {};
        try {
            payload = JSON.parse(text);
        } catch (e) {
            payload = {};
        }
        window.location.href = payload.redirect || `/sprints/detail?id=${taskId}`;
    } catch (error) {
        let errorMsg = error?.message || 'Erro ao salvar';
        if (errorMsg === 'Failed to fetch') {
            errorMsg = 'Não foi possível conectar ao servidor. Verifique se a aplicação está rodando em http://localhost:8000 e tente novamente.';
        }
        if (message) {
            message.innerHTML = window.createAlert
                ? window.createAlert('error', errorMsg)
                : errorMsg;
        }
    } finally {
        if (saveBtn) saveBtn.disabled = false;
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

function syncAllTaskOSKindUST() {
    document.querySelectorAll('#task-os-edit-tbody tr.task-os-row').forEach((row) => {
        refreshTaskOSRowCatalogPricing(row);
        const select = row.querySelector('select.os-kind');
        if (select) window.syncTaskOSKindUST(select);
        else window.syncTaskOSUstEnd(row);
    });
    if (typeof window.syncENC04UstFromENC02 === 'function') {
        window.syncENC04UstFromENC02(document.getElementById('task-os-edit-tbody'));
    }
    syncTaskOSDescriptions();
    refreshTaskOSMeta();
}

// Drag-and-drop reorder for task OS rows
let dragTaskOSRow = null;

function clearTaskOSDragStyles(tbody) {
    if (!tbody) return;
    tbody.querySelectorAll('tr.task-os-row').forEach((el) => {
        el.classList.remove('opacity-50', 'bg-indigo-50/60');
    });
}

document.addEventListener('dragstart', (event) => {
    const handle = event.target.closest('.task-os-drag-handle');
    const row = handle?.closest('tr.task-os-row');
    if (!row) return;
    event.stopPropagation();
    dragTaskOSRow = row;
    row.classList.add('opacity-50');
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', row.getAttribute('data-os-id') || 'row');
});

document.addEventListener('dragend', () => {
    const tbody = dragTaskOSRow?.closest('#task-os-edit-tbody');
    clearTaskOSDragStyles(tbody);
    dragTaskOSRow = null;
    refreshTaskOSMeta();
});

document.addEventListener('dragover', (event) => {
    if (!dragTaskOSRow) return;
    const tbody = document.getElementById('task-os-edit-tbody');
    const row = event.target.closest('tr.task-os-row');
    if (!tbody || !tbody.contains(dragTaskOSRow)) return;
    if (row && !tbody.contains(row)) return;

    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
    if (!row || row === dragTaskOSRow) return;

    const rect = row.getBoundingClientRect();
    const before = event.clientY < rect.top + rect.height / 2;
    const target = before ? row : row.nextElementSibling;
    if (dragTaskOSRow.nextElementSibling === target || dragTaskOSRow === target) return;
    tbody.insertBefore(dragTaskOSRow, target);
});

document.addEventListener('drop', (event) => {
    if (!dragTaskOSRow) return;
    const tbody = event.target.closest('#task-os-edit-tbody') || dragTaskOSRow.closest('#task-os-edit-tbody');
    if (!tbody || !tbody.contains(dragTaskOSRow)) return;
    event.preventDefault();
    event.stopPropagation();
    refreshTaskOSMeta();
});

document.addEventListener('change', (event) => {
    const target = event.target;
    if (!(target instanceof HTMLElement)) return;
    const row = target.closest('#task-os-edit-tbody tr.task-os-row');
    if (!row) return;

    if (target instanceof HTMLSelectElement && target.classList.contains('os-kind')) {
        window.syncTaskOSKindUST(target);
        return;
    }
    if (target.classList.contains('os-ust-rec') || target.classList.contains('os-complexity')) {
        window.syncTaskOSUstEnd(row);
    }
});

document.addEventListener('input', (event) => {
    const target = event.target;
    if (!(target instanceof HTMLElement)) return;
    const row = target.closest('#task-os-edit-tbody tr.task-os-row');
    if (!row) return;
    if (target.classList.contains('os-ust-rec') || target.classList.contains('os-complexity')) {
        window.syncTaskOSUstEnd(row);
    }
});

document.addEventListener('DOMContentLoaded', syncAllTaskOSKindUST);
if (document.readyState !== 'loading') {
    syncAllTaskOSKindUST();
}
