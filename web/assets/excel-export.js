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
    window.exportSprintToExcel = async function(sprintId, sprintName, projectName, osData, deliveryDate) {
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
            buildHeaderSection(worksheet, projectName, sprintName, deliveryDate);
            buildTableHeader(worksheet);
            buildDataRows(worksheet, osData);
            formatAsExcelTable(worksheet, osData.length);
            applyColumnWidths(worksheet);

            // Generate and download file
            const buffer = await workbook.xlsx.writeBuffer();
            downloadExcelFile(buffer, sprintId, projectName);

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

    /**
     * Build header section (rows 1-4) with project info
     */
    function buildHeaderSection(worksheet, projectName, sprintName, deliveryDate) {
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

        // Row 3: Total UST (with formula)
        worksheet.getCell('A3').value = 'Total de USTs da Ordem de Serviço:';
        worksheet.getCell('C3').value = { formula: 'SUM(Tabela1[UST Final])', result: 0 };
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
            row.getCell(6).value = {
                formula: `C${rowIndex}*D${rowIndex}*E${rowIndex}`,
                result: os.ustEnd || 0
            };
            row.getCell(6).numFmt = '0';

            // G: Tipo
            row.getCell(7).value = mapKindToPortuguese(os.kind);

            // H: Descrição do trabalho realizado
            row.getCell(8).value = os.description || '';
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

        // Create blob and download
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
            'ajuste': 'Ajuste',
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
