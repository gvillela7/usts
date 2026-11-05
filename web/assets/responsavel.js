// Responsável da OS: lista fixa + opção Outros

const KNOWN_RESPONSAVEIS = ['Ane Carvalho', 'Gustavo Goulart', 'Murilo Sampaio'];

window.syncResponsavelOther = function (root) {
    const scope = root && root.querySelector ? root : document;
    const select = scope.querySelector('#responsavel-select');
    const otherWrap = scope.querySelector('#responsavel-other-wrap');
    const otherInput = scope.querySelector('#responsavel-other');
    if (!select) return;

    const isOther = select.value === '__other__';
    if (otherWrap) {
        otherWrap.classList.toggle('hidden', !isOther);
    }
    if (otherInput) {
        otherInput.required = isOther;
        if (!isOther) {
            otherInput.value = '';
        } else {
            otherInput.focus();
        }
    }
};

window.getResponsavelValue = function (root) {
    const scope = root && root.querySelector ? root : document;
    const select = scope.querySelector('#responsavel-select');
    if (!select) return '';
    if (select.value === '__other__') {
        return (scope.querySelector('#responsavel-other')?.value || '').trim();
    }
    return (select.value || '').trim();
};

window.initResponsavelField = function (root) {
    const scope = root && root.querySelector ? root : document;
    const select = scope.querySelector('#responsavel-select');
    if (!select) return;

    const current = (select.getAttribute('data-current') || '').trim();
    if (current) {
        if (KNOWN_RESPONSAVEIS.includes(current)) {
            select.value = current;
        } else {
            select.value = '__other__';
            const otherInput = scope.querySelector('#responsavel-other');
            if (otherInput) otherInput.value = current;
        }
    }
    window.syncResponsavelOther(scope);
};

document.addEventListener('DOMContentLoaded', () => window.initResponsavelField());
if (document.readyState !== 'loading') {
    window.initResponsavelField();
}
document.body?.addEventListener('htmx:afterSettle', () => window.initResponsavelField());
