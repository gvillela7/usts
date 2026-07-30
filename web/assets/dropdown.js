// Dropdown Management
// Alert function is defined globally in alerts.js
window.toggleDropdown = function (id) {
    const dropdowns = ['dropdown-menu-catalog', 'dropdown-menu-tasks', 'dropdown-menu-sprints'];
    dropdowns.forEach(d => {
        if (d !== id) {
            const el = document.getElementById(d);
            if (el) el.classList.add('hidden');
        }
    });
    const el = document.getElementById(id);
    if (el) el.classList.toggle('hidden');
};

window.selectSprint = function (id, name) {
    document.getElementById('sprint_id_input').value = id;
    const sprintText = document.getElementById('sprint-selected-text');
    sprintText.textContent = name;
    sprintText.classList.remove('text-gray-500');
    sprintText.classList.add('text-gray-900');
    document.getElementById('dropdown-menu-sprints').classList.add('hidden');

    // Reset tasks dropdown and clear selection
    const taskInput = document.getElementById('task_id_input');
    const taskText = document.getElementById('task-selected-text');
    taskInput.value = '';
    taskText.textContent = 'Tarefa';
    taskText.classList.remove('text-gray-900');
    taskText.classList.add('text-gray-500');

    // Fetch tasks for this sprint
    const projectId = document.getElementById('page-project-id').value;
    fetch(`/api/tasks?sprint_id=${id}`, {
        headers: { 'HX-Request': 'true' }
    })
        .then(response => response.json())
        .then(tasks => {
            const dropdownMenu = document.getElementById('dropdown-menu-tasks');
            dropdownMenu.innerHTML = '';

            if (tasks && tasks.length > 0) {
                tasks.forEach(task => {
                    const div = document.createElement('div');
                    div.className = 'cursor-pointer select-none relative py-3 pl-4 pr-4 hover:bg-blue-50 border-b border-gray-100 last:border-0 transition-colors duration-150';
                    div.onclick = function() { selectTask(task.ID, task.Name); };
                    div.innerHTML = `
                        <div class="flex flex-col">
                            <span class="font-bold text-gray-900 block text-base">${task.Name}</span>
                        </div>
                    `;
                    dropdownMenu.appendChild(div);
                });
            } else {
                const div = document.createElement('div');
                div.className = 'py-3 pl-4 pr-4 text-gray-500 text-sm';
                div.textContent = 'Nenhuma tarefa nesta sprint';
                dropdownMenu.appendChild(div);
            }
        })
        .catch(error => {
            console.error('Failed to fetch tasks:', error);
        });

    // Update task form sprint_id for creating new tasks in this sprint
    const taskFormSprintId = document.getElementById('task-form-sprint-id');
    if (taskFormSprintId) {
        taskFormSprintId.value = id;
    }
};

window.selectCatalog = function (id, cod, description, ustNormal, ustNewTech) {
    const sprintId = document.getElementById('sprint_id_input').value;
    const sprintName = document.getElementById('sprint-selected-text').textContent;
    const taskId = document.getElementById('task_id_input').value;
    const taskName = document.getElementById('task-selected-text').textContent;

    if (!sprintId || sprintName === 'Selecione a Sprint') {
        const messageArea = document.getElementById('message-area');
        messageArea.innerHTML = window.createAlert('warning', 'Selecione uma sprint antes de escolher um código.');
        document.getElementById('dropdown-menu-catalog').classList.add('hidden');
        return;
    }

    if (!taskId || taskName === 'Tarefa') {
        const messageArea = document.getElementById('message-area');
        messageArea.innerHTML = window.createAlert('warning', 'Selecione uma tarefa antes de escolher um código.');
        document.getElementById('dropdown-menu-catalog').classList.add('hidden');
        return;
    }

    document.getElementById('message-area').innerHTML = '';
    // addOSRow is defined in table.js
    if (typeof window.addOSRow === 'function') {
        window.addOSRow(id, cod, description, taskId, taskName, sprintId, ustNormal, ustNewTech);
    } else {
        console.error('addOSRow function not found');
    }
    document.getElementById('dropdown-menu-catalog').classList.add('hidden');
};

window.selectTask = function (id, name) {
    document.getElementById('task_id_input').value = id;
    const taskText = document.getElementById('task-selected-text');
    taskText.textContent = name;
    taskText.classList.remove('text-gray-500');
    taskText.classList.add('text-gray-900');
    document.getElementById('dropdown-menu-tasks').classList.add('hidden');
};

// Sprint Modal Functions
window.toggleSprintForm = function () {
    const modal = document.getElementById('sprint-modal');
    modal.showModal();
    const sprintNameInput = document.getElementById('sprint_name');
    if (sprintNameInput) sprintNameInput.focus();
};

// Task Modal Functions
window.toggleTaskForm = function () {
    const sprintId = document.getElementById('sprint_id_input').value;
    const sprintName = document.getElementById('sprint-selected-text').textContent;

    if (!sprintId || sprintName === 'Sprint') {
        const messageArea = document.getElementById('message-area');
        messageArea.innerHTML = window.createAlert('warning', 'Selecione uma sprint antes de criar uma tarefa.');
        return;
    }

    // Atualizar sprint_id no hidden field
    document.getElementById('task-form-sprint-id').value = sprintId;

    const modal = document.getElementById('task-modal');
    modal.showModal();
    const taskNameInput = document.getElementById('task_name');
    if (taskNameInput) taskNameInput.focus();
};

// Close dropdown when clicking outside
document.addEventListener('click', function (event) {
    const catalogDropdown = document.getElementById('custom-dropdown-catalog');
    const tasksDropdown = document.getElementById('custom-dropdown-tasks');
    const sprintsDropdown = document.getElementById('custom-dropdown-sprints');

    if (catalogDropdown && !catalogDropdown.contains(event.target)) {
        const menu = document.getElementById('dropdown-menu-catalog');
        if (menu) menu.classList.add('hidden');
    }
    if (tasksDropdown && !tasksDropdown.contains(event.target)) {
        const menu = document.getElementById('dropdown-menu-tasks');
        if (menu) menu.classList.add('hidden');
    }
    if (sprintsDropdown && !sprintsDropdown.contains(event.target)) {
        const menu = document.getElementById('dropdown-menu-sprints');
        if (menu) menu.classList.add('hidden');
    }
});
