// ENC04 UST mirrors ENC02 Recurrence (dynamic sprint rule).

function rowCatalogCod(row) {
    if (!row) return '';
    const fromAttr = row.getAttribute('data-catalog-cod');
    if (fromAttr) return fromAttr.trim().toUpperCase();
    const el = row.querySelector('.text-indigo-700');
    return (el?.textContent || '').trim().toUpperCase();
}

window.syncENC04UstFromENC02 = function (tbodyOrSelector) {
    let tbody = tbodyOrSelector;
    if (typeof tbodyOrSelector === 'string') {
        tbody = document.querySelector(tbodyOrSelector);
    }
    if (!tbody) {
        tbody = document.getElementById('os-items-tbody') || document.getElementById('task-os-edit-tbody');
    }
    if (!tbody) return;

    let enc02Rec = null;
    let enc04Row = null;

    tbody.querySelectorAll('tr').forEach((row) => {
        const cod = rowCatalogCod(row);
        if (cod === 'ENC02') {
            const recInput = row.querySelector('.os-ust-rec, input[name^="ust_recurrence_"]');
            const parsed = parseInt(recInput?.value, 10);
            enc02Rec = parsed > 0 ? parsed : 1;
        } else if (cod === 'ENC04') {
            enc04Row = row;
        }
    });

    if (!enc04Row || enc02Rec === null) return;

    const ustInput = enc04Row.querySelector('.os-ust-item, input[name^="ust_item_"]');
    if (!ustInput) return;

    ustInput.value = String(enc02Rec);
    enc04Row.setAttribute('data-ust-linked-enc02', '1');
    ustInput.title = 'UST dinâmico: igual à Rec. do código ENC02';

    if (tbody.id === 'os-items-tbody') {
        const rowId = enc04Row.id?.split('-')[2];
        if (rowId && typeof window.calculateUstFim === 'function') {
            window.calculateUstFim(rowId);
        } else if (typeof window.calculateTotalUST === 'function') {
            window.calculateTotalUST();
        }
    } else if (typeof window.syncTaskOSUstEnd === 'function') {
        window.syncTaskOSUstEnd(enc04Row);
    }
};

function enc02RecInput(target) {
    if (!(target instanceof HTMLInputElement)) return null;
    if (!target.matches('.os-ust-rec, input[name^="ust_recurrence_"]')) return null;
    const row = target.closest('#os-items-tbody tr, #task-os-edit-tbody tr.task-os-row');
    if (!row || rowCatalogCod(row) !== 'ENC02') return null;
    return row;
}

function onEnc02RecChanged(target) {
    const row = enc02RecInput(target);
    if (!row) return;
    const tbody = row.closest('tbody');
    window.syncENC04UstFromENC02(tbody);
}

document.addEventListener('input', (event) => onEnc02RecChanged(event.target));
document.addEventListener('change', (event) => onEnc02RecChanged(event.target));

function syncENC04OnPage() {
    if (typeof window.syncENC04UstFromENC02 !== 'function') return;
    window.syncENC04UstFromENC02(document.getElementById('os-items-tbody'));
    window.syncENC04UstFromENC02(document.getElementById('task-os-edit-tbody'));
}

document.addEventListener('DOMContentLoaded', syncENC04OnPage);
if (document.readyState !== 'loading') {
    syncENC04OnPage();
}
document.body?.addEventListener('htmx:afterSettle', syncENC04OnPage);
