<script setup lang="ts">
import {createInstance} from "@/views/index/core/master";
import {computed, h, onBeforeUnmount, onMounted, ref, watch} from "vue";
import {ElMessageBox, ElOption, ElSelect} from "element-plus";
import {
    ExportExcel,
    FetchActConfig,
    GetTempAct,
    ImportExcel,
    KeepActionConfig,
    TempActKeep
} from "../../../wailsjs/go/main/App";
import {dataparser} from "../../../wailsjs/go/models";
import {Message} from "@arco-design/web-vue";
import {runApi} from "@/config/apis/api";
import {throttle} from "lodash-es";

// 常量定义
const CONSTANTS = {
    ACTIVITY_PREFIX: 'Activity_',
    AUTO_SAVE_INTERVAL: 30_000, // 30秒
    THROTTLE_DELAY: 3000, // 3秒
    WATERMARK_FONT_SIZE: 18,
    MENU_TYPE: {
        IMPORT: 1,
        EXPORT: 2,
        RENEW: 3,
        HOT_UPDATE: 4,
        SWITCH_AB: 5,
        NET_DATA: 6,
    },
    DEFAULT_SHEET_NAME: '配置',
    DEFAULT_HEADERS: [
        {v: "中文字段名称"},
        {v: "英文字段名"},
        {v: "字段类型"},
        {v: "值"}
    ],
} as const;

type ABType = '' | 'A' | 'B';

const emitter = defineEmits(['onNewAct', 'update:actId'])

const props = defineProps({
    actId: {
        type: Number,
        default: 0
    }
})

const actIdModel = computed({
    get() {
        return Number(props.actId)
    },
    set(val) {
        emitter('update:actId', val)
    }
})

const changeAct = (id: number) => {
    actIdModel.value = id
}

const univerRef = ref<any>(null)
const univerAPIRef = ref<any>(null)
interface BtnGroupState {
    btnGroup: boolean;
    importBtn: boolean;
    exportBtn: boolean;
    newBtn: boolean;
    updateBtn: boolean;
    loadCache: boolean;
    abCfgBtn: boolean;
    forNetBtn: boolean;
}

const genBtnGroupShow = (): BtnGroupState => {
    return {
        btnGroup: false,
        importBtn: false,
        exportBtn: false,
        newBtn: false,
        updateBtn: false,
        loadCache: true,
        abCfgBtn: false,
        forNetBtn: false,
    }
}
const btnGroupShow = ref<BtnGroupState>(genBtnGroupShow())
/**
 * 获取工作簿数据快照
 * @param includeStyles 是否包含样式
 * @returns 工作簿数据
 */
const getWorkbookSnapshot = (includeStyles = false): Partial<dataparser.Workbook> | null => {
    if (!univerAPIRef.value) return null
    
    const fworkbook: any = univerAPIRef.value.getActiveWorkbook()
    if (!fworkbook) return null
    
    const fworksheets = fworkbook?.getSheets()
    const fworkbookdata: Partial<dataparser.Workbook> = {
        id: fworkbook?.id,
        name: fworkbook?.getName(),
        sheetOrder: [],
        sheets: {},
    }
    
    if (includeStyles) {
        fworkbookdata.styles = fworkbook.getSnapshot().styles
    }
    
    for (let i = 0; i < fworksheets.length; i++) {
        const sheet = fworksheets[i]?.getSheet()
        if (!sheet) continue
        
        const sheetId = String(sheet.getSheetId())
        fworkbookdata.sheets![sheetId] = sheet.getSnapshot()
        fworkbookdata.sheetOrder!.push(sheetId)
    }
    
    return fworkbookdata
}

/**
 * 创建并初始化工作簿
 * @param workbookData 工作簿数据
 */
const createAndInitWorkbook = (workbookData: any) => {
    const workbook = univerAPIRef.value.createWorkbook(workbookData)
    univerAPIRef.value.addWatermark('text', {
        content: `${workbook.getId()}`,
        fontSize: CONSTANTS.WATERMARK_FONT_SIZE,
        repeat: true
    })
    return workbook
}

const selectedAct = async (priorityNet = false, ab: ABType = "", silent = false) => {
    let actConfigInfo: any
    if (priorityNet) {
        actConfigInfo = await loadNetActInfo(ab, true, false)
        if (!actConfigInfo) {
            Message.warning('内网数据未拉取成功，尝试加载本地数据')
            actConfigInfo = await loadTempActInfo(String(props.actId), ab, silent)
        }
    } else {
        actConfigInfo = await loadTempActInfo(String(props.actId), ab, true, false)
        if (!actConfigInfo) {
            Message.warning('本地数据未拉取成功，尝试加载内网数据')
            actConfigInfo = await loadNetActInfo(ab, silent)
        }
    }

    if (!actConfigInfo || !actConfigInfo.id) {
        disposeUniver()
        btnGroupShow.value.btnGroup = true
        btnGroupShow.value.importBtn = true
        btnGroupShow.value.newBtn = true
        btnGroupShow.value.loadCache = false
        Message.warning("新活动配置，需要进行新建")
        return
    }
    await toTempKeepAct()
    disposeUniver()
    initUniver(actConfigInfo.id)
    createAndInitWorkbook(actConfigInfo)
    startTempKeepAct()
}

const disposeUniver = () => {
    btnGroupShow.value = genBtnGroupShow()
    univerAPIRef.value?.dispose()
    univerRef.value = null
    univerAPIRef.value = null
    stopTempKeepAct()
}

const initUniver = (workbookKey="") => {
    btnGroupShow.value = {
        btnGroup: true,
        importBtn: true,
        exportBtn: true,
        newBtn: true,
        updateBtn: true,
        loadCache: false,
        abCfgBtn: false,
        forNetBtn: true,
    }
    if (utilGetActAB(workbookKey) !== "") btnGroupShow.value.abCfgBtn = true
    const {univer, univerAPI} = createInstance("mainContainer")
    univerRef.value = univer
    univerAPIRef.value = univerAPI
    const el = document.getElementById("mainContainer")
    el?.classList.add("has-content")
}

type MenuType = typeof CONSTANTS.MENU_TYPE[keyof typeof CONSTANTS.MENU_TYPE]

const menuPopoverParams = ref<{
    visible: boolean
    type: MenuType
}>({
    visible: false,
    type: CONSTANTS.MENU_TYPE.IMPORT,
})

/**
 * 创建配置选择对话框
 */
const createConfigSelectDialog = (selectValue: { value: ABType }) => {
    return h('div', [
        h('div', {style: 'margin-bottom: 12px;'}, `确认为活动【${props.actId}】初始化配置吗`),
        h(ElSelect, {
            modelValue: selectValue.value,
            'onUpdate:modelValue': (val: ABType) => (selectValue.value = val),
            placeholder: '请选择配置类型',
            style: 'width: 100%',
        }, () => [
            h(ElOption, {label: '普通配置', value: ''}),
            h(ElOption, {label: 'A 配置', value: 'A'}),
            h(ElOption, {label: 'B 配置', value: 'B'}),
        ])
    ])
}

const toRenew = async (jumpWarning = false, ab: ABType = '') => {
    const selectValue = ref<ABType>(ab)
    try {
        if (!jumpWarning) {
            menuPopoverParams.value = {type: CONSTANTS.MENU_TYPE.RENEW, visible: false}
            await ElMessageBox({
                title: '提示',
                message: () => createConfigSelectDialog(selectValue),
                confirmButtonText: '确定',
                cancelButtonText: '取消',
            })
        }
        await toTempKeepAct()
        const workbookKey = utilGenWorkbookKey(selectValue.value)
        disposeUniver()
        initUniver(workbookKey)
        const fworkbook = univerAPIRef.value.createWorkbook({id: workbookKey, name: workbookKey})
        univerAPIRef.value.addWatermark('text', {
            content: `${fworkbook.getId()}`,
            fontSize: CONSTANTS.WATERMARK_FONT_SIZE,
            repeat: true
        })
        const curSheet = fworkbook.getActiveSheet()
        curSheet.setName(CONSTANTS.DEFAULT_SHEET_NAME)
        const frange = curSheet.getRange("A1:A4")
        frange.setValues(CONSTANTS.DEFAULT_HEADERS.map(header => [header]))
        curSheet.activate()
        startTempKeepAct()
        Message.success('已确认操作')
    } catch (err) {
        Message.info(`已取消操作:${err}`)
    }
}

const toSwitchAB = async () => {
    menuPopoverParams.value = {type: CONSTANTS.MENU_TYPE.SWITCH_AB, visible: false}
    const workbookKey = utilGetWorkbookKey()
    if (!workbookKey) return
    
    const curAB = utilGetActAB(workbookKey)
    if (curAB === '') return
    
    const targetAB: ABType = curAB === 'A' ? 'B' : 'A'
    await selectedAct(false, targetAB, true)
    if (!univerAPIRef.value) {
        await toRenew(true, targetAB)
    }
}

const toHotUpdate = async () => {
    menuPopoverParams.value = {type: CONSTANTS.MENU_TYPE.HOT_UPDATE, visible: false}
    const fworkbookdata = getWorkbookSnapshot()
    if (!fworkbookdata) {
        Message.error('获取工作簿数据失败')
        return
    }
    
    try {
        const token = window.localStorage.getItem("token")
        if (!token) {
            Message.error('未找到token，请先登录')
            return
        }
        await runApi(() => KeepActionConfig(fworkbookdata as dataparser.Workbook, token))
        Message.success('实时更新成功')
    } catch (e: any) {
        Message.error(`实时更新失败:${e.message}`)
    }
}

const toImport = async () => {
    menuPopoverParams.value = {type: CONSTANTS.MENU_TYPE.IMPORT, visible: false}
    const selectValue = ref<ABType>('')
    
    try {
        await ElMessageBox({
            title: '提示',
            message: () => createConfigSelectDialog(selectValue),
            confirmButtonText: '确定',
            cancelButtonText: '取消',
        })

        const res = await runApi(() => ImportExcel(String(props.actId), selectValue.value)) as any
        Message.success("导入成功")
        disposeUniver()
        initUniver(res.id)
        createAndInitWorkbook(res)
        startTempKeepAct()
    } catch (e: any) {
        Message.error(e.message)
    }
}


const loadTempActInfo = async (
    actId = "",
    ab: ABType = "",
    silent = false,
    throwOnError = true
) => {
    try {
        return await runApi(() => GetTempAct(actId, ab), {silent, throwOnError})
    } catch (e) {
        return undefined
    }
}

const loadNetActInfo = async (
    ab: ABType = "",
    silent = false,
    throwOnError = true
) => {
    try {
        return await runApi(() => FetchActConfig(String(props.actId), ab), {silent, throwOnError})
    } catch (e) {
        return undefined
    }
}

const judgeLoadTempActInfo = async () => {
    try {
        await ElMessageBox.confirm(
            '是否加载上次修改的活动信息',
            '提示',
            {
                confirmButtonText: '确定',
                cancelButtonText: '取消',
            }
        )
        const actInfo = await loadTempActInfo() as any
        if (actInfo && actInfo.id) {
            changeAct(Number(utilGetActId(actInfo.id)))
        }
    } catch (e: any) {
        Message.error(e.message)
    }
}

const toTempKeepAct = async () => {
    const fworkbookdata = getWorkbookSnapshot(true)
    if (!fworkbookdata || !fworkbookdata.id) return
    
    try {
        await runApi(() => TempActKeep(fworkbookdata as dataparser.Workbook))
    } catch (e: any) {
        Message.error(`临时存储失败:${e.message}`)
    }
}

const toExport = async () => {
    menuPopoverParams.value = {type: CONSTANTS.MENU_TYPE.EXPORT, visible: false}
    const fworkbookdata = getWorkbookSnapshot()
    if (!fworkbookdata || !fworkbookdata.id) {
        Message.error('获取工作簿数据失败')
        return
    }
    
    try {
        const data = await runApi(() => ExportExcel(fworkbookdata as dataparser.Workbook))
        Message.success(`导出目录:${data}`)
    } catch (e: any) {
        Message.error(`导出失败:${e.message}`)
    }
}

const toNetData = () => {
    menuPopoverParams.value = {type: CONSTANTS.MENU_TYPE.NET_DATA, visible: false}
    selectedAct(true)
}

let intervaler: number | null = null

const startTempKeepAct = () => {
    if (intervaler !== null) {
        return
    }
    intervaler = window.setInterval(() => {
        saveOnce()
    }, CONSTANTS.AUTO_SAVE_INTERVAL)
}

const stopTempKeepAct = () => {
    if (intervaler === null) return
    clearInterval(intervaler)
    intervaler = null
}

/**
 * 生成工作簿键名
 * @param ab AB配置类型
 * @returns 工作簿键名
 */
const utilGenWorkbookKey = (ab: ABType = ""): string => {
    let workbookKey = `${CONSTANTS.ACTIVITY_PREFIX}${props.actId}`
    if (ab) workbookKey += `_${ab}`
    return workbookKey
}

const utilGetWorkbookKey = (): string | undefined => {
    return univerAPIRef.value?.getActiveWorkbook?.()?.id
}

const utilGetActId = (workbookId: string): string => {
    if (workbookId.length <= 0) return ""
    workbookId = workbookId.slice(CONSTANTS.ACTIVITY_PREFIX.length)
    const workbookInfos = workbookId.split("_")
    return workbookInfos[0]
}

const utilGetActAB = (workbookId: string): ABType => {
    if (workbookId.length <= 0) return ""
    workbookId = workbookId.slice(CONSTANTS.ACTIVITY_PREFIX.length)
    const workbookInfos = workbookId.split("_")
    return (workbookInfos[1] ?? "") as ABType
}

const saveOnce = throttle(
    () => {
        toTempKeepAct()
    },
    CONSTANTS.THROTTLE_DELAY,
    {
        trailing: false,
    }
)

const keyBoardHandler = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
        e.preventDefault()
        saveOnce()
    }
}

watch(() => actIdModel.value, (val) => {
    if (!val) return
    selectedAct()
})

onMounted(() => {
    window.addEventListener('keydown', keyBoardHandler)
})


onBeforeUnmount(async () => {
    await toTempKeepAct()
    disposeUniver()
    window.removeEventListener('keydown', keyBoardHandler)
})
</script>

<template>
    <div class="main-container" id="mainContainer">
    </div>
    <div class="btn-load-cache" v-if="btnGroupShow.loadCache">
        <div class="tooltip-trigger" data-tooltip="点击重新加载本地缓存">
            <div class="warning-symbol" @click="judgeLoadTempActInfo"></div>
        </div>
    </div>
    <div ref="menuRef" class="menu draggable" v-if="btnGroupShow.btnGroup">
        <a-trigger
            :trigger="['click']"
            clickToClose
            position="top"
            v-model:popupVisible="menuPopoverParams.visible"
        >
            <div :class="`button-trigger ${menuPopoverParams.visible ? 'button-trigger-active' : ''}`">
                <IconClose size="22" v-if="menuPopoverParams.visible" />
                <IconMessage size="22" v-else />
            </div>
            <template #content>
                <a-menu
                    :style="{ marginBottom: '-4px' }"
                    mode="popButton"
                    :tooltipProps="{ position: 'left' }"
                    showCollapseButton
                >
                    <a-menu-item key="1" v-if="btnGroupShow.importBtn" @click="toImport">
                        <template #icon><icon-upload/></template>
                        导入配置
                    </a-menu-item>
                    <a-menu-item key="2" v-if="btnGroupShow.exportBtn" @click="toExport">
                        <template #icon><icon-download/></template>
                        导出配置
                    </a-menu-item>
                    <a-menu-item key="3" v-if="btnGroupShow.newBtn" @click="toRenew()">
                        <template #icon><icon-home/></template>
                        新建配置
                    </a-menu-item>
                    <a-menu-item key="4" v-if="btnGroupShow.updateBtn" @click="toHotUpdate">
                        <template #icon><icon-cloud/></template>
                        更新内网
                    </a-menu-item>
                    <a-menu-item key="5" v-if="btnGroupShow.abCfgBtn" @click="toSwitchAB">
                        <template #icon><icon-bold /></template>
                        AB配置
                    </a-menu-item>
                    <a-menu-item key="6" v-if="btnGroupShow.forNetBtn" @click="toNetData">
                        <template #icon><icon-drag-arrow/></template>
                        同步实时数据
                    </a-menu-item>
                </a-menu>
            </template>
        </a-trigger>
    </div>
</template>

<style scoped>
.main-container {
    height: 100%;
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    position: relative;

    > div {
        width: 100% !important;
    }
}

#mainContainer :deep([data-u-comp="workbench-layout"]) {
    /* 你想覆盖的样式 */
    width: 100%;
}

.main-container::before {
    content: "请选择活动";
    font-size: 24px;
    color: #999;
    letter-spacing: 2px;
    opacity: 1;
    transition: opacity 0.8s ease;
    position: absolute;
}

/* main-container 有内容时淡出 */
.main-container.has-content::before {
    opacity: 0;
}

.menu {
    position: fixed;        /* 🔑 关键 */
    z-index: 999;
    left: 90%;              /* 用 left/top，不要 right */
    bottom: 100px;
    user-select: none;
}

.draggable {
    cursor: grab;
}

.draggable:active {
    cursor: grabbing;
}

.dropdown {
    border: 1px solid #c1c2c5;
    border-radius: 12px;
    transition: all 300ms;
    display: flex;
    flex-direction: column;
    min-height: 58px;
    background-color: white;
    overflow: hidden;
    position: relative;
    inset-inline: auto;
    max-width: 298px;
    min-width: 298px;
}

.dropdown input:where(:checked) ~ .list {
    opacity: 1;
    transform: translateY(-3rem) scale(1);
    transition: all 500ms ease;
    margin-top: 32px;
    padding-top: 4px;
    margin-bottom: -32px;
    width: 340px;
}

.dropdown input:where(:not(:checked)) ~ .list {
    opacity: 0;
    transform: translateY(3rem);
    margin-top: -100%;
    user-select: none;
    height: 0px;
    max-height: 0px;
    min-height: 0px;
    pointer-events: none;
    transition: all 500ms ease-out;
    width: 340px;
}

.trigger {
    cursor: pointer;
    list-style: none;
    -webkit-user-select: none;
    -moz-user-select: none;
    user-select: none;
    font-weight: 600;
    color: inherit;
    width: 100%;
    display: flex;
    align-items: center;
    flex-flow: row;
    gap: 1rem;
    padding: 1rem;
    height: max-content;
    position: relative;
    z-index: 99;
    border-radius: inherit;
    background-color: white;
}

.sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border-width: 0;
}

.dropdown input:where(:checked) + .trigger {
    margin-bottom: 1rem;
}

.dropdown input:where(:checked) + .trigger:before {
    rotate: 90deg;
    transition-delay: 0ms;
}

.dropdown input:where(:checked) + .trigger::after {
    content: "关闭面板";
}

.trigger:before,
.trigger::after {
    position: relative;
    display: flex;
    justify-content: center;
    align-items: center;
}

.trigger:before {
    content: "›";
    rotate: -90deg;
    width: 17px;
    height: 17px;
    color: #262626;
    border-radius: 2px;
    font-size: 26px;
    transition: all 350ms ease;
    transition-delay: 85ms;
}

.trigger::after {
    content: "打开面板";
}

.list {
    height: 100%;
    max-height: 20rem;
    display: grid;
    grid-auto-flow: row;
    overflow: hidden auto;
    gap: 1rem;
    padding: 0 1rem;
    margin-right: -8px;
    --w-scrollbar: 8px;
}

.listitem {
    height: 100%;
    list-style: none;
}

.detail-info {
    border-radius: 5px;
    transition: all 0.2s ease;
}

.detail-info.selected {
    border: 2px solid #4a90e2; /* 蓝色边框 */
    background: rgba(74, 144, 226, 0.1); /* 浅蓝背景 */
}

.detail-info:hover {
    background: rgba(74, 144, 226, 0.05); /* 悬停时非常浅的蓝色 */
    cursor: pointer;
}


.detail-info {
    width: 100%;
    padding: 4px;
    display: flex;
    justify-content: space-between;
}

.webkit-scrollbar::-webkit-scrollbar {
    width: var(--w-scrollbar);
    height: var(--w-scrollbar);
    border-radius: 9999px;
}

.webkit-scrollbar::-webkit-scrollbar-track {
    background: #0000;
}

.webkit-scrollbar::-webkit-scrollbar-thumb {
    background: #0000;
    border-radius: 9999px;
}

.webkit-scrollbar:hover::-webkit-scrollbar-thumb {
    background: #c1c2c5;
}

:deep(.el-divider--horizontal) {
    margin: 0 5px 0 0;
}


#SocailIcons {
    min-width: 350px;
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    justify-content: space-around;
    align-items: center;
}

.icons {
    position: relative; /* 关键 */
    width: 50px;
    height: 50px;
    background: #fff;
    border-radius: 50%;
    cursor: pointer;
    border: none;
    text-align: center;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
}

/* 提示文字 */
.iconName {
    position: absolute;
    top: -38px;
    left: 50%;
    transform: translateX(-50%) scale(0);
    font-size: 12px;
    color: #fff;
    border-radius: 4px;
    padding: 4px 8px;
    white-space: nowrap;
    transition: transform 0.25s ease;
    z-index: 10;
}

/* hover 统一生效 */
.icons:hover .iconName {
    transform: translateX(-50%) scale(1);
}

/* 不同类型只管颜色，不管位移 */
.icons.instaIcon .iconName {
    background: linear-gradient(30deg, #0000ff, #f56040);
}

.icons.linkedin .iconName {
    background: #0274b3;
}

.icons.whatsapp .iconName {
    background: #25d366;
}

.icons.youtube .iconName {
    background: #ff0000;
}

.icons.hotupdate .iconName {
    background: #ff0000;
}

.icons.abcfg .iconName {
    background: #6cb400;
}

.icons.calibration .iconName {
    background: #eac221;
}


.icons:hover .icon {
    opacity: 1;
    color: #fff;
}


.icon {
    width: 100%;
    height: 100%;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;

    background: transparent;
    color: #333;
}

.icon::before {
    content: "";
    position: absolute;
    inset: 0;
    border-radius: 50%;
    height: 0;
    z-index: 0;
}

.icon:hover::before {
    height: 100%;
}

.icon.tube::before {
    background: red;
}

.icon.abcfg::before {
    background: #6cb400;
}

.icon.calibration::before {
    background: #eac221;
}

.icon.insta::before {
    background: linear-gradient(40deg, #0000ff, #f56040);
}

.icon.link::before {
    background: #0274b3;
}

.icon.whats::before {
    background: #25d366;
}

.iconName {
    position: absolute;
    top: -34px;
    left: 50%;
    transform: translateX(-50%) translateY(6px);
    opacity: 0;
    font-size: 12px;
    color: #fff;
    background: #333;
    padding: 4px 6px;
    border-radius: 4px;
    white-space: nowrap;
    transition: opacity 0.2s ease,
    transform 0.2s ease;
    pointer-events: none;
}


.icons:hover .iconName {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
}

.btn-load-cache {
    position: absolute;
    right: 140px;
    bottom: 60px;
}

.plate-tooltip-container {
    display: flex;
    justify-content: center;
    padding: 3rem;
}

.tooltip-trigger {
    --primary: #ffb200;

    width: 50px;
    height: 50px;

    background: linear-gradient(to bottom, #3a3d44 0%, #212329 100%);
    border: 1px solid #444;

    box-shadow: inset 0 2px 2px rgba(255, 255, 255, 0.18),
    inset 0 -4px 6px rgba(0, 0, 0, 0.7),
    0 8px 16px rgba(0, 0, 0, 0.45);

    border-radius: 6px;

    display: flex;
    justify-content: center;
    align-items: center;

    position: relative;
    cursor: help;

    transition: filter 0.15s ease-out;
}

/* ===============================
   Warning Triangle（按 50px 缩放）
   =============================== */

.warning-symbol {
    width: 0;
    height: 0;

    border-left: 12px solid transparent;
    border-right: 12px solid transparent;
    border-bottom: 20px solid var(--primary);

    position: relative;
}

.warning-symbol::after {
    content: "!";
    position: absolute;

    left: 50%;
    top: 13px;
    transform: translate(-50%, -50%);

    color: #111;
    font-size: 16px;
    font-weight: 900;
    font-family: sans-serif;
}

/* ===============================
   Tooltip Bubble（不显夸张）
   =============================== */

.tooltip-trigger::before {
    content: attr(data-tooltip);

    position: absolute;
    bottom: calc(100% + 10px);
    left: 50%;
    transform: translateX(-50%) translateY(6px);

    opacity: 0;
    pointer-events: none;

    background: var(--primary);
    color: #111;

    padding: 6px 12px;
    border-radius: 6px;

    font-family: "Share Tech Mono", monospace;
    font-size: 14px;
    font-weight: bold;
    white-space: nowrap;

    box-shadow: 0 10px 18px rgba(0, 0, 0, 0.4);

    transition: transform 0.25s cubic-bezier(0.2, 1.3, 0.4, 1),
    opacity 0.25s ease;
}

/* Tooltip Arrow */
.tooltip-trigger::after {
    content: "";

    position: absolute;
    bottom: 100%;
    left: 50%;

    width: 0;
    height: 0;

    border-left: 7px solid transparent;
    border-right: 7px solid transparent;
    border-top: 7px solid var(--primary);

    transform: translateX(-50%) translateY(6px);

    opacity: 0;
    pointer-events: none;

    transition: transform 0.25s cubic-bezier(0.2, 1.3, 0.4, 1),
    opacity 0.25s ease;
}

/* ===============================
   Hover Effects
   =============================== */

.tooltip-trigger:hover {
    filter: brightness(1.18);
}

.tooltip-trigger:hover .warning-symbol {
    animation: warning-pulse 0.9s ease-in-out infinite;
}

.tooltip-trigger:hover::before,
.tooltip-trigger:hover::after {
    transform: translateX(-50%) translateY(0);
    opacity: 1;
}

/* ===============================
   Animations
   =============================== */

@keyframes warning-pulse {
    0% {
        opacity: 1;
    }
    50% {
        opacity: 0.6;
    }
    100% {
        opacity: 1;
    }
}


@keyframes electric-shock {
    0% {
        transform: translate(0, 0);
        box-shadow: inset 0 2px 2px -1px rgba(255, 255, 255, 0.2),
        inset 0 -5px 5px -2px rgba(0, 0, 0, 0.8),
        0 10px 20px -3px rgba(0, 0, 0, 0.5);
    }
    20% {
        transform: translate(-1px, 1px);
        box-shadow: inset 0 2px 2px -1px rgba(255, 255, 255, 0.2),
        inset 0 -5px 5px -2px rgba(0, 0, 0, 0.8),
        0 10px 20px -3px rgba(0, 0, 0, 0.5),
        0 0 8px 1px var(--primary);
    }
    40% {
        transform: translate(-1px, -1px);
        box-shadow: inset 0 2px 2px -1px rgba(255, 255, 255, 0.2),
        inset 0 -5px 5px -2px rgba(0, 0, 0, 0.8),
        0 10px 20px -3px rgba(0, 0, 0, 0.5);
    }
    60% {
        transform: translate(1px, 1px);
        box-shadow: inset 0 2px 2px -1px rgba(255, 255, 255, 0.2),
        inset 0 -5px 5px -2px rgba(0, 0, 0, 0.8),
        0 10px 20px -3px rgba(0, 0, 0, 0.5),
        0 0 8px 1px var(--primary);
    }
    80% {
        transform: translate(1px, -1px);
        box-shadow: inset 0 2px 2px -1px rgba(255, 255, 255, 0.2),
        inset 0 -5px 5px -2px rgba(0, 0, 0, 0.8),
        0 10px 20px -3px rgba(0, 0, 0, 0.5);
    }
    100% {
        transform: translate(0, 0);
        box-shadow: inset 0 2px 2px -1px rgba(255, 255, 255, 0.2),
        inset 0 -5px 5px -2px rgba(0, 0, 0, 0.8),
        0 10px 20px -3px rgba(0, 0, 0, 0.5);
    }
}

@keyframes warning-pulse {
    0% {
        transform: scale(1);
        filter: drop-shadow(0 0 3px var(--primary));
    }
    50% {
        transform: scale(1.1);
        filter: drop-shadow(0 0 8px var(--primary));
    }
    100% {
        transform: scale(1);
        filter: drop-shadow(0 0 3px var(--primary));
    }
}
</style>
