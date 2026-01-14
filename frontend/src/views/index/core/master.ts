import { UniverSheetsConditionalFormattingPreset } from '@univerjs/preset-sheets-conditional-formatting'
import sheetsConditionalFormattingZhCN from '@univerjs/preset-sheets-conditional-formatting/locales/zh-CN'
import {UniverRPCMainThreadPlugin, UniverSheetsCorePreset} from '@univerjs/preset-sheets-core'
import sheetsCoreZhCN from '@univerjs/preset-sheets-core/locales/zh-CN'
import { UniverSheetsDataValidationPreset } from '@univerjs/preset-sheets-data-validation'
import sheetsDataValidationZhCN from '@univerjs/preset-sheets-data-validation/locales/zh-CN'
import { UniverSheetsDrawingPreset } from '@univerjs/preset-sheets-drawing'
import sheetsDrawingZhCN from '@univerjs/preset-sheets-drawing/locales/zh-CN'
import { UniverSheetsFilterPreset } from '@univerjs/preset-sheets-filter'
import UniverPresetSheetsFilterZhCN from '@univerjs/preset-sheets-filter/locales/zh-CN'
import { UniverSheetsFindReplacePreset } from '@univerjs/preset-sheets-find-replace'
import UniverPresetSheetsFindReplaceZhCN from '@univerjs/preset-sheets-find-replace/locales/zh-CN'
import {
    UniverSheetsHyperLinkPlugin,
    UniverSheetsHyperLinkUIPlugin
} from '@univerjs/preset-sheets-hyper-link'
import sheetsHyperLinkZhCN from '@univerjs/preset-sheets-hyper-link/locales/zh-CN'
import { UniverSheetsSortPreset } from '@univerjs/preset-sheets-sort'
import SheetsSortZhCN from '@univerjs/preset-sheets-sort/locales/zh-CN'
import { UniverSheetsThreadCommentPreset } from '@univerjs/preset-sheets-thread-comment'
import UniverPresetSheetsThreadCommentZhCN from '@univerjs/preset-sheets-thread-comment/locales/zh-CN'
import { createUniver, LocaleType, mergeLocales } from '@univerjs/presets'
import { UniverSheetsCrosshairHighlightPlugin } from '@univerjs/sheets-crosshair-highlight'
import SheetsCrosshairHighlightZhCN from '@univerjs/sheets-crosshair-highlight/locale/zh-CN'
import { UniverSheetsZenEditorPlugin } from '@univerjs/sheets-zen-editor'
import SheetsZenEditorZhCN from '@univerjs/sheets-zen-editor/locale/zh-CN'
import { UniverWatermarkPlugin } from '@univerjs/watermark'
import '@univerjs/watermark/facade'

// import {Univer} from "@univerjs/presets";
// import {FUniver} from "@univerjs/docs-ui/lib/types/facade";
import './styles.css'

export const createInstance = (id: string) => {
    const {univer, univerAPI} = createUniver({
        locale: LocaleType.ZH_CN,
        locales: {
            [LocaleType.ZH_CN]: mergeLocales(
                sheetsCoreZhCN,
                SheetsSortZhCN,
                UniverPresetSheetsFilterZhCN,
                sheetsConditionalFormattingZhCN,
                sheetsDataValidationZhCN,
                UniverPresetSheetsFindReplaceZhCN,
                sheetsDrawingZhCN,
                sheetsHyperLinkZhCN,
                UniverPresetSheetsThreadCommentZhCN,
                SheetsCrosshairHighlightZhCN,
                SheetsZenEditorZhCN,
            ),
        },
        presets: [
            UniverSheetsCorePreset({
                container: id,
                // workerURL: new Worker(new URL('./worker.ts', import.meta.url), { type: 'module' }),
            }),
            UniverSheetsFindReplacePreset(),
            UniverSheetsSortPreset(),
            UniverSheetsConditionalFormattingPreset(),
            UniverSheetsDataValidationPreset({
                // 是否在下拉菜单中显示编辑按钮
                showEditOnDropdown: true,
            }),
            UniverSheetsDrawingPreset(),
            UniverSheetsFilterPreset(),
            UniverSheetsThreadCommentPreset(),
        ],
        plugins: [
            [UniverSheetsHyperLinkPlugin],
            [UniverSheetsHyperLinkUIPlugin],
            [UniverWatermarkPlugin],
            UniverSheetsCrosshairHighlightPlugin,
            UniverSheetsZenEditorPlugin,
        ],
    })
    return {univer, univerAPI}
}

export function disposeInstance(univer, univerAPI){
    univer?.dispose()
    univerAPI?.dispose()
}