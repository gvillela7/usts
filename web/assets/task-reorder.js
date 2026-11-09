// Sprint task list reordering (drag and drop) — order is persisted until the user changes it again.

window.saveSprintTaskOrder = async function (list) {
    if (!list) return false;
    const sprintID = parseInt(list.getAttribute('data-sprint-id'), 10);
    const taskIDs = Array.from(list.querySelectorAll('li[data-task-id]'))
        .map((el) => parseInt(el.getAttribute('data-task-id'), 10))
        .filter((id) => id > 0);

    if (!sprintID || taskIDs.length === 0) return false;

    try {
        const response = await fetch('/tasks/reorder', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ SprintID: sprintID, TaskIDs: taskIDs }),
        });
        if (!response.ok) {
            const text = await response.text();
            throw new Error(text || 'Falha ao salvar ordem');
        }
        return true;
    } catch (error) {
        console.error(error);
        if (window.showAlert) {
            window.showAlert('error', error.message || 'Erro ao reordenar tarefas');
        } else {
            alert(error.message || 'Erro ao reordenar tarefas');
        }
        window.location.reload();
        return false;
    }
};

window.moveSprintTask = async function (btn, direction) {
    if (typeof event !== 'undefined' && event) {
        event.preventDefault();
        event.stopPropagation();
    }
    const row = btn?.closest('li[data-task-id]');
    const list = row?.closest('ul[data-sprint-id]');
    if (!row || !list) return;

    if (direction < 0 && row.previousElementSibling) {
        list.insertBefore(row, row.previousElementSibling);
    } else if (direction > 0 && row.nextElementSibling) {
        list.insertBefore(row.nextElementSibling, row);
    } else {
        return;
    }

    await window.saveSprintTaskOrder(list);
};

let dragTaskRow = null;
let dragTaskList = null;
let dragOrderDirty = false;

function clearTaskDragStyles(list) {
    if (!list) return;
    list.querySelectorAll('li[data-task-id]').forEach((el) => {
        el.classList.remove('opacity-50', 'bg-indigo-50/60', 'border-t-2', 'border-indigo-400');
    });
}

document.addEventListener('dragstart', (event) => {
    const handle = event.target.closest('.task-drag-handle');
    const row = handle?.closest('li[data-task-id]');
    if (!row) return;
    event.stopPropagation();
    dragTaskRow = row;
    dragTaskList = row.closest('ul[data-sprint-id]');
    dragOrderDirty = false;
    row.classList.add('opacity-50');
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', row.getAttribute('data-task-id') || '');
});

document.addEventListener('dragend', async () => {
    const list = dragTaskList;
    clearTaskDragStyles(list);
    if (list && dragOrderDirty) {
        await window.saveSprintTaskOrder(list);
    }
    dragTaskRow = null;
    dragTaskList = null;
    dragOrderDirty = false;
});

document.addEventListener('dragover', (event) => {
    const row = event.target.closest('li[data-task-id]');
    const list = row?.closest('ul[data-sprint-id]') || dragTaskList;
    if (!dragTaskRow || !list || !list.contains(dragTaskRow)) return;
    if (row && !list.contains(row)) return;

    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
    if (!row || row === dragTaskRow) return;

    const rect = row.getBoundingClientRect();
    const before = event.clientY < rect.top + rect.height / 2;
    const target = before ? row : row.nextElementSibling;
    if (dragTaskRow.nextElementSibling === target || dragTaskRow === target) return;
    list.insertBefore(dragTaskRow, target);
    dragOrderDirty = true;
});

document.addEventListener('drop', (event) => {
    const list = event.target.closest('ul[data-sprint-id]') || dragTaskList;
    if (!list || !dragTaskRow || !list.contains(dragTaskRow)) return;
    event.preventDefault();
    event.stopPropagation();
    dragOrderDirty = true;
});
