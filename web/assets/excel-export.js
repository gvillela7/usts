// Excel Export Module
// Uses ExcelJS to export sprint data to Excel format matching the template

import ExcelJS from 'exceljs';

(function() {
    'use strict';

    /**
     * Main export function exposed globally
     * @param {number} sprintId - Sprint ID
     * @param {string} sprintName - Sprint name (competência)
     * @param {string} projectName - Project name
     * @param {Array} osData - Array of OS items with catalog data
     * @param {string} deliveryDate - Delivery date in ISO format
     */
    window.exportSprintToExcel = async function(sprintId, sprintName, projectName, osData, deliveryDate, options) {
        // Check browser support
        if (!window.Blob || !window.URL || !window.URL.createObjectURL) {
            if (window.showAlert) {
                window.showAlert('error', 'Seu navegador não suporta exportação de Excel. Atualize para a versão mais recente.');
            }
            return;
        }

        // Validate data
        if (!osData || osData.length === 0) {
            if (window.showAlert) {
                window.showAlert('info', 'Nenhuma OS para exportar');
            }
            return;
        }

        try {
            // Create workbook and worksheet
            const workbook = new ExcelJS.Workbook();
            const worksheet = workbook.addWorksheet('OS Data', {
                views: [{ showGridLines: false, zoomScale: 130, zoomScaleNormal: 130 }]
            });

            // Build the Excel structure
            const totalUST = calculateTotalUST(osData);
            buildHeaderSection(worksheet, projectName, sprintName, deliveryDate, totalUST, osData.length);
            buildTableHeader(worksheet);
            buildDataRows(worksheet, osData);
            formatAsExcelTable(worksheet, osData.length);
            applyColumnWidths(worksheet);

            // Generate and download file
            const buffer = await workbook.xlsx.writeBuffer();
            const filename = (options && options.filename)
                ? options.filename
                : buildTaskOSFilename(sprintName || sprintId, (options && options.templateKind) || 'os', projectName);
            downloadBuffer(buffer, filename);

            // Show success message
            if (window.showAlert) {
                window.showAlert('success', 'Excel exportado com sucesso!');
            }
        } catch (error) {
            console.error('Error exporting to Excel:', error);
            if (window.showAlert) {
                window.showAlert('error', 'Erro ao exportar Excel: ' + error.message);
            }
        }
    };

    /**
     * Alias for task-based exports
     * Uses the same function as sprint exports since the functionality is identical
     */
    window.exportTaskToExcel = window.exportSprintToExcel;

    function templateKindFileToken(kind) {
        return String(kind || '').toLowerCase().trim() === 'epico' ? 'EPICO' : 'US';
    }

    function buildTaskOSFilename(taskCode, templateKind, projectName) {
        const code = sanitizeFilename(taskCode || 'TAREFA');
        const kind = templateKindFileToken(templateKind);
        const project = sanitizeFilename(projectName || 'PROJETO');
        return `OS_${code}_${kind}_${project}.xlsx`;
    }

    window.exportTaskByID = async function (taskID, taskName, projectName, deliveryDate) {
        try {
            const response = await fetch(`/api/tasks/${taskID}/os`);
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || 'Falha ao carregar OS da tarefa');
            }
            const payload = await response.json();
            const osData = Array.isArray(payload.items) ? payload.items : [];
            if (osData.length === 0) {
                if (window.showAlert) {
                    window.showAlert('info', 'Nenhuma OS para exportar nesta tarefa.');
                } else if (window.createAlert) {
                    const area = document.getElementById('message-area');
                    if (area) area.innerHTML = window.createAlert('info', 'Nenhuma OS para exportar nesta tarefa.');
                }
                return;
            }
            const name = taskName || payload.task?.name || `Tarefa_${taskID}`;
            const project = projectName || payload.project?.name || 'PROJETO';
            const templateKind = payload.task?.templateKind || 'os';
            await window.exportTaskToExcel(taskID, name, project, osData, deliveryDate, {
                filename: buildTaskOSFilename(name, templateKind, project),
                templateKind,
            });
        } catch (error) {
            console.error(error);
            if (window.showAlert) {
                window.showAlert('error', error.message || 'Erro ao gerar Excel');
            } else {
                alert(error.message || 'Erro ao gerar Excel');
            }
        }
    };

    /**
     * Sum UST Final values (UST Item × Recorrência × Complexidade), matching Excel row formulas.
     */
    function calculateTotalUST(osData) {
        return (osData || []).reduce((sum, os) => {
            const ustItem = Number(os.ustItem) || 0;
            const recurrence = Number(os.ustRecurrence) || 0;
            const complexity = Number(os.complexity) || 0;
            const fromFormula = ustItem * recurrence * complexity;
            if (fromFormula > 0) return sum + fromFormula;
            return sum + (Number(os.ustEnd) || 0);
        }, 0);
    }

    /**
     * Build header section (rows 1-4) with project info
     */
    function buildHeaderSection(worksheet, projectName, sprintName, deliveryDate, totalUST, dataRowCount) {
        // Row 1: Nome do Projeto
        worksheet.getCell('A1').value = 'Nome do Projeto:';
        worksheet.getCell('C1').value = projectName;
        worksheet.mergeCells('A1:B1');
        worksheet.mergeCells('C1:I1');

        // Row 2: Competência
        worksheet.getCell('A2').value = 'Competência da Ordem de Serviço:';
        worksheet.getCell('C2').value = sprintName;
        worksheet.mergeCells('A2:B2');
        worksheet.mergeCells('C2:I2');

        // Row 3: Total UST (formula + cached result so the value shows without recalculation)
        const lastDataRow = 5 + (dataRowCount || 0);
        const totalFormula = dataRowCount > 0 ? `SUM(F6:F${lastDataRow})` : '0';
        worksheet.getCell('A3').value = 'Total de USTs da Ordem de Serviço:';
        worksheet.getCell('C3').value = { formula: totalFormula, result: totalUST || 0 };
        worksheet.getCell('C3').numFmt = '0';
        worksheet.mergeCells('A3:B3');
        worksheet.mergeCells('C3:I3');

        // Row 4: Data de entrega
        worksheet.getCell('A4').value = 'Data de entrega da O.S.:';
        const date = parseDeliveryDate(deliveryDate);
        worksheet.getCell('C4').value = date;
        worksheet.getCell('C4').numFmt = 'dd/mm/yyyy';
        worksheet.mergeCells('A4:B4');
        worksheet.mergeCells('C4:I4');

        // Apply bold to label cells
        ['A1', 'A2', 'A3', 'A4'].forEach(cell => {
            worksheet.getCell(cell).font = { bold: true };
            worksheet.getCell(cell).alignment = { horizontal: 'left', vertical: 'middle' };
        });

        // Apply styles to value cells
        ['C1', 'C2', 'C3', 'C4'].forEach(cell => {
            worksheet.getCell(cell).alignment = { horizontal: 'left', vertical: 'middle' };
        });
    }

    /**
     * Build table header (row 5)
     */
    function buildTableHeader(worksheet) {
        const headers = [
            'Código Catálogo',      // A5
            'Item catálogo',        // B5
            'UST Item',             // C5
            'Recorrência da UST',   // D5
            'Complexidade',         // E5
            'UST Final',            // F5
            'Tipo',                 // G5
            'Descrição do trabalho realizado', // H5
            'Tarefa Candidata'      // I5
        ];

        const headerRow = worksheet.getRow(5);
        headers.forEach((header, index) => {
            const cell = headerRow.getCell(index + 1);
            cell.value = header;
            cell.font = { bold: true };
            cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true };
            cell.fill = {
                type: 'pattern',
                pattern: 'solid',
                fgColor: { argb: 'FFD9E1F2' } // Light blue background
            };
            cell.border = {
                top: { style: 'thin' },
                left: { style: 'thin' },
                bottom: { style: 'thin' },
                right: { style: 'thin' }
            };
        });
    }

    /**
     * Build data rows (row 6+)
     */
    function buildDataRows(worksheet, osData) {
        let rowIndex = 6;

        osData.forEach((os) => {
            const row = worksheet.getRow(rowIndex);

            // A: Código Catálogo
            row.getCell(1).value = os.catalogCod || '';

            // B: Item catálogo
            row.getCell(2).value = os.catalogDescription || '';

            // C: UST Item
            row.getCell(3).value = os.ustItem || 0;
            row.getCell(3).numFmt = '0';

            // D: Recorrência da UST
            row.getCell(4).value = os.ustRecurrence || 0;
            row.getCell(4).numFmt = '0';

            // E: Complexidade
            row.getCell(5).value = os.complexity || 0;
            row.getCell(5).numFmt = '0';

            // F: UST Final (FORMULA)
            const ustItem = Number(os.ustItem) || 0;
            const recurrence = Number(os.ustRecurrence) || 0;
            const complexity = Number(os.complexity) || 0;
            const ustFinal = (ustItem * recurrence * complexity) || Number(os.ustEnd) || 0;
            row.getCell(6).value = {
                formula: `C${rowIndex}*D${rowIndex}*E${rowIndex}`,
                result: ustFinal
            };
            row.getCell(6).numFmt = '0';

            // G: Tipo
            row.getCell(7).value = mapKindToPortuguese(os.kind);

            // H: Descrição do trabalho realizado (observação da OS)
            row.getCell(8).value = (os.observation || '').trim();
            row.getCell(8).alignment = {
                horizontal: 'left',
                vertical: 'top',
                wrapText: true
            };

            // I: Tarefa Candidata
            row.getCell(9).value = os.tarefaCandidta === 1 ? 'Sim' : 'Não';

            // Apply borders to all cells
            for (let col = 1; col <= 9; col++) {
                const cell = row.getCell(col);
                cell.border = {
                    top: { style: 'thin' },
                    left: { style: 'thin' },
                    bottom: { style: 'thin' },
                    right: { style: 'thin' }
                };

                // Default alignment if not set
                if (!cell.alignment) {
                    cell.alignment = { horizontal: 'left', vertical: 'middle' };
                }
            }

            // Apply center alignment to numeric columns
            [1, 3, 4, 5, 6, 7, 9].forEach(col => {
                row.getCell(col).alignment = { horizontal: 'center', vertical: 'middle' };
            });

            rowIndex++;
        });
    }

    /**
     * Format as Excel Table
     */
    function formatAsExcelTable(worksheet, dataRowCount) {
        const tableRange = `A5:I${5 + dataRowCount}`;

        worksheet.addTable({
            name: 'Tabela1',
            ref: tableRange,
            headerRow: true,
            totalsRow: false,
            style: {
                theme: 'TableStyleMedium2',
                showRowStripes: true,
                showColumnStripes: false
            },
            columns: [
                { name: 'Código Catálogo', filterButton: false },
                { name: 'Item catálogo', filterButton: false },
                { name: 'UST Item', filterButton: false },
                { name: 'Recorrência da UST', filterButton: false },
                { name: 'Complexidade', filterButton: false },
                { name: 'UST Final', filterButton: false },
                { name: 'Tipo', filterButton: false },
                { name: 'Descrição do trabalho realizado', filterButton: false },
                { name: 'Tarefa Candidata', filterButton: false }
            ],
            rows: [] // Data already populated in buildDataRows
        });
    }

    /**
     * Apply column widths
     */
    function applyColumnWidths(worksheet) {
        worksheet.columns = [
            { key: 'A', width: 15 },  // Código Catálogo
            { key: 'B', width: 54 },  // Item catálogo
            { key: 'C', width: 10 },  // UST Item
            { key: 'D', width: 18 },  // Recorrência da UST
            { key: 'E', width: 12 },  // Complexidade
            { key: 'F', width: 10 },  // UST Final
            { key: 'G', width: 11 },  // Tipo
            { key: 'H', width: 33 },  // Descrição
            { key: 'I', width: 15 }   // Tarefa Candidata
        ];
    }

    /**
     * Download Excel file
     */
    function downloadExcelFile(buffer, sprintId, projectName) {
        // Sanitize project name for filename
        const sanitizedProjectName = sanitizeFilename(projectName);

        const filename = `OS_${sprintId}_${sanitizedProjectName}.xlsx`;
        downloadBuffer(buffer, filename);
    }

    function downloadBuffer(buffer, filename) {
        const blob = new Blob([buffer], {
            type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
        });

        const url = window.URL.createObjectURL(blob);
        const anchor = document.createElement('a');
        anchor.href = url;
        anchor.download = filename;
        document.body.appendChild(anchor);
        anchor.click();

        // Cleanup
        document.body.removeChild(anchor);
        window.URL.revokeObjectURL(url);
    }

    /**
     * Sprint summary Excel (tasks list with responsável + UST Final).
     */
    window.exportSprintSummaryByID = async function (sprintID) {
        try {
            const response = await fetch(`/api/sprints/${sprintID}/summary`);
            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || 'Falha ao carregar resumo da sprint');
            }
            const payload = await response.json();
            const tasks = Array.isArray(payload.tasks) ? payload.tasks : [];
            if (tasks.length === 0) {
                if (window.showAlert) {
                    window.showAlert('info', 'Nenhuma tarefa para exportar nesta sprint.');
                } else {
                    alert('Nenhuma tarefa para exportar nesta sprint.');
                }
                return;
            }

            const workbook = new ExcelJS.Workbook();
            const worksheet = workbook.addWorksheet('Sprint', {
                views: [{ showGridLines: true, zoomScale: 120, zoomScaleNormal: 120 }]
            });

            buildSprintSummarySheet(worksheet, payload);

            const buffer = await workbook.xlsx.writeBuffer();
            const sprintTitle = payload.sprint?.title || `Sprint_${sprintID}`;
            const projectName = payload.project?.name || 'PROJETO';
            downloadBuffer(
                buffer,
                `Sprint_${sanitizeFilename(sprintTitle)}_${sanitizeFilename(projectName)}.xlsx`
            );

            if (window.showAlert) {
                window.showAlert('success', 'Excel da sprint exportado com sucesso!');
            }
        } catch (error) {
            console.error(error);
            if (window.showAlert) {
                window.showAlert('error', error.message || 'Erro ao gerar Excel da sprint');
            } else {
                alert(error.message || 'Erro ao gerar Excel da sprint');
            }
        }
    };

    function responsavelSiglaFill(name) {
        const n = String(name || '').toLowerCase().trim();
        if (n === 'ane carvalho') return { bg: 'FFFCE7F3', fg: 'FF9D174D' };
        if (n === 'gustavo goulart') return { bg: 'FFE0F2FE', fg: 'FF0369A1' };
        if (n === 'murilo sampaio') return { bg: 'FFFEF3C7', fg: 'FF92400E' };
        let sum = 0;
        for (let i = 0; i < n.length; i++) sum += n.charCodeAt(i);
        const palette = [
            { bg: 'FFFCE7F3', fg: 'FF9D174D' },
            { bg: 'FFE0F2FE', fg: 'FF0369A1' },
            { bg: 'FFD1FAE5', fg: 'FF047857' },
            { bg: 'FFFEF3C7', fg: 'FF92400E' },
            { bg: 'FFEDE9FE', fg: 'FF6D28D9' },
        ];
        return palette[sum % palette.length];
    }

    function thinBorder() {
        return {
            top: { style: 'thin', color: { argb: 'FFCBD5E1' } },
            left: { style: 'thin', color: { argb: 'FFCBD5E1' } },
            bottom: { style: 'thin', color: { argb: 'FFCBD5E1' } },
            right: { style: 'thin', color: { argb: 'FFCBD5E1' } },
        };
    }

    function buildSprintSummarySheet(worksheet, payload) {
        const sprintTitle = payload.sprint?.title || 'Sprint';
        const status = payload.sprint?.status || '';
        const totalTasks = Number(payload.totalTasks) || (payload.tasks || []).length;
        const totalUST = Number(payload.totalUSTFinal) || 0;
        const tasks = payload.tasks || [];

        worksheet.mergeCells('A1:D1');
        worksheet.mergeCells('A2:D2');
        worksheet.mergeCells('A3:D3');

        const headerFill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFDBEAFE' } };
        for (const addr of ['A1', 'B1', 'C1', 'D1', 'A2', 'B2', 'C2', 'D2', 'A3', 'B3', 'C3', 'D3']) {
            worksheet.getCell(addr).fill = headerFill;
        }

        const titleCell = worksheet.getCell('A1');
        titleCell.value = sprintTitle;
        titleCell.font = { bold: true, size: 16, color: { argb: 'FF1E3A8A' } };
        titleCell.alignment = { vertical: 'middle', horizontal: 'left' };

        worksheet.getCell('A2').value = status;
        worksheet.getCell('A2').font = { bold: true, size: 11, color: { argb: 'FF475569' } };
        worksheet.getCell('A2').alignment = { vertical: 'middle', horizontal: 'left' };

        worksheet.getCell('A3').value = `Total de tarefas: ${totalTasks}    UST Final total: ${totalUST}`;
        worksheet.getCell('A3').font = { size: 11, color: { argb: 'FF1D4ED8' } };
        worksheet.getCell('A3').alignment = { vertical: 'middle', horizontal: 'left' };

        // Summary boxes E1:F2
        const summaryHeaderFill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF1E40AF' } };
        worksheet.getCell('E1').value = 'Total de tarefas';
        worksheet.getCell('F1').value = 'UST Final total';
        ['E1', 'F1'].forEach((addr) => {
            const cell = worksheet.getCell(addr);
            cell.fill = summaryHeaderFill;
            cell.font = { bold: true, color: { argb: 'FFFFFFFF' } };
            cell.alignment = { horizontal: 'center', vertical: 'middle' };
            cell.border = thinBorder();
        });

        worksheet.getCell('E2').value = totalTasks;
        worksheet.getCell('F2').value = totalUST;
        ['E2', 'F2'].forEach((addr) => {
            const cell = worksheet.getCell(addr);
            cell.font = { bold: true, size: 12 };
            cell.alignment = { horizontal: 'center', vertical: 'middle' };
            cell.border = thinBorder();
            cell.numFmt = '0';
        });

        worksheet.getRow(1).height = 28;
        worksheet.getRow(2).height = 22;
        worksheet.getRow(3).height = 22;

        const headers = ['#', 'Código da tarefa', 'Responsável (Sigla)', 'Responsável', 'UST Final', 'Observações'];
        const headerRow = worksheet.getRow(4);
        headers.forEach((header, index) => {
            const cell = headerRow.getCell(index + 1);
            cell.value = header;
            cell.fill = summaryHeaderFill;
            cell.font = { bold: true, color: { argb: 'FFFFFFFF' } };
            cell.alignment = { horizontal: 'center', vertical: 'middle', wrapText: true };
            cell.border = thinBorder();
        });
        headerRow.height = 24;

        tasks.forEach((task, i) => {
            const rowIndex = 5 + i;
            const row = worksheet.getRow(rowIndex);
            const responsavel = (task.responsavel || '').trim();
            const sigla = (task.sigla || '').trim() || (responsavel ? '?' : '');

            row.getCell(1).value = i + 1;
            row.getCell(2).value = task.code || '';
            row.getCell(3).value = sigla;
            row.getCell(4).value = responsavel;
            row.getCell(5).value = Number(task.ustFinal) || 0;
            row.getCell(5).numFmt = '0';
            row.getCell(6).value = task.observacoes || '';

            for (let col = 1; col <= 6; col++) {
                const cell = row.getCell(col);
                cell.border = thinBorder();
                cell.alignment = { vertical: 'middle', horizontal: col === 2 || col === 4 || col === 6 ? 'left' : 'center' };
            }

            if (sigla) {
                const colors = responsavelSiglaFill(responsavel);
                const siglaCell = row.getCell(3);
                siglaCell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: colors.bg } };
                siglaCell.font = { bold: true, color: { argb: colors.fg } };
            }

            if (i % 2 === 1) {
                [1, 2, 4, 5, 6].forEach((col) => {
                    const cell = row.getCell(col);
                    if (!cell.fill || !cell.fill.fgColor) {
                        cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF8FAFC' } };
                    }
                });
            }
        });

        worksheet.columns = [
            { key: 'A', width: 6 },
            { key: 'B', width: 18 },
            { key: 'C', width: 18 },
            { key: 'D', width: 22 },
            { key: 'E', width: 12 },
            { key: 'F', width: 28 },
        ];

        if (tasks.length > 0) {
            worksheet.autoFilter = {
                from: { row: 4, column: 1 },
                to: { row: 4 + tasks.length, column: 6 },
            };
        }
    }

    /**
     * Helper: Parse delivery date
     */
    function parseDeliveryDate(dateString) {
        const date = new Date(dateString);
        if (isNaN(date.getTime())) {
            console.warn('Invalid delivery date, using current date');
            return new Date();
        }
        return date;
    }

    /**
     * Helper: Map Kind to Portuguese
     */
    function mapKindToPortuguese(kind) {
        if (!kind) return 'Normal';

        const normalized = kind.toLowerCase().trim();
        const kindMap = {
            'normal': 'Normal',
            'nova_tecnologia': 'Nova tecnologia',
            'nova tecnologia': 'Nova tecnologia',
            'ajuste': 'Nova tecnologia',
            'nova funcionalidade': 'Normal',
            'melhoria': 'Normal',
            'correção': 'Ajuste',
            'correcao': 'Ajuste',
            'refatoração': 'Normal',
            'refatoracao': 'Normal'
        };
        return kindMap[normalized] || 'Normal';
    }

    /**
     * Helper: Sanitize filename
     */
    function sanitizeFilename(name) {
        if (!name) return 'PROJETO';

        return name
            .replace(/[^a-zA-Z0-9_\-]/g, '_')  // Replace special chars
            .replace(/_+/g, '_')                 // Collapse multiple underscores
            .replace(/^_|_$/g, '')               // Trim underscores
            .toUpperCase();                      // Uppercase
    }

})();
