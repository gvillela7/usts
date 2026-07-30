// Table and Form Management
// Alert function is defined globally in alerts.js
let rowCounter = 0;

window.addOSRow = function (catalogId, catalogCod, catalogDesc, taskId, taskName, sprintId, ustNormal, ustNewTech) {
    rowCounter++;
    const tbody = document.getElementById('os-items-tbody');
    const table = document.getElementById('os-items-table');

    if (table) table.classList.remove('hidden');

    const row = document.createElement('tr');
    row.id = `os-row-${rowCounter}`;
    row.className = 'border-b hover:bg-gray-50';
    row.setAttribute('data-ust-normal', ustNormal);
    row.setAttribute('data-ust-new-tech', ustNewTech || ustNormal);
    row.setAttribute('data-catalog-id', catalogId);

    // Calculate initial UST Fim (ustNormal * 1)
    const initialUstFim = ustNormal * 1;

    row.innerHTML = `
        <td class="px-3 py-2 text-sm">
            <input type="hidden" name="catalog_id_${rowCounter}" value="${catalogId}" />
            <div class="font-medium text-gray-900">${catalogCod}</div>
            <div class="text-xs text-gray-500">${catalogDesc}</div>
        </td>
        <td class="px-3 py-2 text-sm">
            <input type="hidden" name="task_id_${rowCounter}" value="${taskId}" />
            <input type="hidden" name="sprint_id_${rowCounter}" value="${sprintId}" />
            ${taskName}
        </td>
        <td class="px-1 py-2" style="max-width: 60px;">
            <input type="number"
                id="ust_item_${rowCounter}"
                name="ust_item_${rowCounter}"
                value="${ustNormal}"
                onchange="calculateUstFim(${rowCounter})"
                class="w-full border border-gray-300 rounded px-1 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                style="max-width: 60px;"
                required min="1" />
        </td>
        <td class="px-1 py-2" style="max-width: 60px;">
            <input type="number"
                id="ust_recurrence_${rowCounter}"
                name="ust_recurrence_${rowCounter}"
                value="1"
                class="w-full border border-gray-300 rounded px-1 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                style="max-width: 60px;"
                required min="1" />
        </td>
        <td class="px-1 py-2" style="max-width: 60px;">
            <input type="number"
                id="complexity_${rowCounter}"
                name="complexity_${rowCounter}"
                value="1"
                onchange="calculateUstFim(${rowCounter})"
                class="w-full border border-gray-300 rounded px-1 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                style="max-width: 60px;"
                required min="1" max="10" />
        </td>
        <td class="px-1 py-2" style="max-width: 60px;">
            <input type="number"
                id="ust_end_${rowCounter}"
                name="ust_end_${rowCounter}"
                value="${initialUstFim}"
                class="w-full border border-gray-300 rounded px-1 py-1 text-sm bg-gray-100 focus:outline-none focus:ring-1 focus:ring-blue-500"
                style="max-width: 60px;"
                min="1"
                readonly />
        </td>
        <td class="px-2 py-2" style="max-width: 100px;">
            <select name="kind_${rowCounter}"
                class="w-full border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                onchange="handleTypeChange(${rowCounter})"
                required>
                <option value="normal" selected>Normal</option>
                <option value="ajuste">Ajuste</option>
            </select>
        </td>
        <td class="px-3 py-2" style="min-width: 320px;">
            <textarea name="description_${rowCounter}"
                class="w-full border border-gray-300 rounded px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
                rows="2"
                required></textarea>
        </td>
        <td class="px-2 py-2" style="max-width: 60px;">
            <div class="flex flex-col gap-3">
                <label class="flex items-center cursor-pointer">
                    <input type="radio"
                        name="tarefa_candidata_${rowCounter}"
                        value="1"
                        class="mr-1 focus:ring-blue-500" />
                    <span class="text-xs">Sim</span>
                </label>
                <label class="flex items-center cursor-pointer">
                    <input type="radio"
                        name="tarefa_candidata_${rowCounter}"
                        value="0"
                        class="mr-1 focus:ring-blue-500"
                        checked />
                    <span class="text-xs">Não</span>
                </label>
            </div>
        </td>
        <td class="px-3 py-2 text-center">
            <button type="button" onclick="window.removeOSRow(${rowCounter})"
                class="text-red-600 hover:text-red-800 focus:outline-none"
                title="Remover">
                <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
            </button>
        </td>
    `;
    tbody.appendChild(row);

    document.getElementById('catalog_id_input').value = '';
    const catalogText = document.getElementById('catalog-selected-text');
    catalogText.textContent = 'Código';
    catalogText.classList.remove('text-gray-900');
    catalogText.classList.add('text-gray-500');

    calculateTotalUST();
};

// Calculate UST Fim based on UST Item and Complexity
window.calculateUstFim = function (rowId) {
    const ustItemInput = document.getElementById(`ust_item_${rowId}`);
    const complexityInput = document.getElementById(`complexity_${rowId}`);
    const ustEndInput = document.getElementById(`ust_end_${rowId}`);

    if (ustItemInput && complexityInput && ustEndInput) {
        const ustItem = parseInt(ustItemInput.value) || 0;
        const complexity = parseInt(complexityInput.value) || 1;
        const ustFim = ustItem * complexity;

        ustEndInput.value = ustFim;
        calculateTotalUST();
    }
};

// Calculate total UST from all rows
window.calculateTotalUST = function() {
    const tbody = document.getElementById('os-items-tbody');
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

// Handle type change between Normal and Ajuste
window.handleTypeChange = function(rowId) {
    const row = document.getElementById(`os-row-${rowId}`);
    const typeSelect = row.querySelector(`select[name="kind_${rowId}"]`);
    const ustItemInput = document.getElementById(`ust_item_${rowId}`);

    const ustNormal = parseInt(row.getAttribute('data-ust-normal')) || 0;
    const ustNewTech = parseInt(row.getAttribute('data-ust-new-tech')) || 0;

    if (typeSelect.value === 'ajuste') {
        ustItemInput.value = ustNewTech;
    } else {
        ustItemInput.value = ustNormal;
    }

    calculateUstFim(rowId);
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
        const ustEnd = ustEndInput ? parseInt(ustEndInput) : null;
        const kind = row.querySelector(`select[name="kind_${rowId}"]`).value;
        const description = row.querySelector(`textarea[name="description_${rowId}"]`).value;

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

    const osItems = collectOSItems();

    if (osItems.length === 0) {
        messageArea.innerHTML = window.createAlert('warning', 'Adicione pelo menos um item de OS.');
        return false;
    }

    fetch(form.action, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'HX-Request': 'true'
        },
        body: JSON.stringify(osItems)
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => {
                    messageArea.innerHTML = window.createAlert('error', text);
                });
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
