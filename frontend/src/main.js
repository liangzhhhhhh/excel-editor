import * as XLSX from 'xlsx';

// Wails绑定（Wails v2暴露的后端方法）
const OpenFile = window.wails.bindings['.bindings'].OpenFile;
const SaveFile = window.wails.bindings['.bindings'].SaveFile;

let workbook; // 当前工作簿
let currentSheetName; // 当前工作表名
let currentCell; // 当前选中单元格（简化单选）

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    // 打开文件
    document.getElementById('open-btn').addEventListener('click', async () => {
        const base64 = await OpenFile();
        if (!base64) return;
        const binary = atob(base64);
        const array = new Uint8Array(binary.length);
        for (let i = 0; i < binary.length; i++) {
            array[i] = binary.charCodeAt(i);
        }
        workbook = XLSX.read(array, { type: 'array' });
        currentSheetName = workbook.SheetNames[0];
        renderTabs();
        renderSheet();
    });

    // 保存文件
    document.getElementById('save-btn').addEventListener('click', async () => {
        if (!workbook) return;
        const wbout = XLSX.write(workbook, { bookType: 'xlsx', type: 'base64' });
        const result = await SaveFile(wbout);
        console.log(result);
    });

    // 加粗
    document.getElementById('bold-btn').addEventListener('click', () => {
        if (!currentCell || !workbook) return;
        const sheet = workbook.Sheets[currentSheetName];
        const ref = currentCell.dataset.ref;
        const cell = sheet[ref] || { t: 's', v: '' };
        cell.s = cell.s || {};
        cell.s.font = cell.s.font || {};
        cell.s.font.bold = !cell.s.font.bold;
        currentCell.style.fontWeight = cell.s.font.bold ? 'bold' : 'normal';
    });

    // 颜色选择
    document.getElementById('color-picker').addEventListener('change', (e) => {
        if (!currentCell || !workbook) return;
        const sheet = workbook.Sheets[currentSheetName];
        const ref = currentCell.dataset.ref;
        const cell = sheet[ref] || { t: 's', v: '' };
        cell.s = cell.s || {};
        cell.s.fgColor = { rgb: e.target.value.slice(1) }; // RGB格式（如'FF0000'）
        currentCell.style.color = e.target.value;
    });

    // 排序（假设选中列为第一列，简化）
    document.getElementById('sort-btn').addEventListener('click', () => {
        if (!workbook) return;
        const sheet = workbook.Sheets[currentSheetName];
        const data = XLSX.utils.sheet_to_json(sheet, { header: 1 });
        data.sort((a, b) => (a[0] > b[0] ? 1 : -1)); // 按第一列排序
        workbook.Sheets[currentSheetName] = XLSX.utils.aoa_to_sheet(data);
        renderSheet();
    });
});

// 渲染工作表Tab
function renderTabs() {
    let tabsHtml = '';
    workbook.SheetNames.forEach(name => {
        tabsHtml += `<button onclick="switchSheet('${name}')">${name}</button>`;
    });
    document.getElementById('tabs').innerHTML = tabsHtml;
}

// 切换工作表
window.switchSheet = function(name) {
    currentSheetName = name;
    renderSheet();
};

// 渲染当前工作表
function renderSheet() {
    if (!workbook) return;
    const sheet = workbook.Sheets[currentSheetName];
    if (!sheet || !sheet['!ref']) return;
    const range = XLSX.utils.decode_range(sheet['!ref']);
    let table = '<table>';
    for (let row = range.s.r; row <= range.e.r; row++) {
        table += '<tr>';
        for (let col = range.s.c; col <= range.e.c; col++) {
            const cellRef = XLSX.utils.encode_cell({ r: row, c: col });
            const cell = sheet[cellRef] || { v: '' };
            let style = '';
            if (cell.s && cell.s.font && cell.s.font.bold) style += 'font-weight: bold;';
            if (cell.s && cell.s.fgColor && cell.s.fgColor.rgb) style += `color: #${cell.s.fgColor.rgb};`;
            table += `<td contenteditable="true" data-ref="${cellRef}" style="${style}">${cell.v}</td>`;
        }
        table += '</tr>';
    }
    table += '</table>';
    document.getElementById('spreadsheet').innerHTML = table;

    // 添加编辑和选中监听
    document.querySelectorAll('#spreadsheet td').forEach(td => {
        td.addEventListener('click', () => { currentCell = td; });
        td.addEventListener('input', (e) => {
            const ref = e.target.dataset.ref;
            const value = e.target.innerText;
            sheet[ref] = { t: 's', v: value };
            if (value.startsWith('=')) {
                sheet[ref].f = value.substring(1);
                XLSX.utils.book_calc(workbook); // 计算公式（注意：SheetJS的calc方法可能需调整为book_calc或自定义）
                e.target.innerText = sheet[ref].v || 'ERROR';
            }
        });
    });
}