<script setup lang="ts">
import {createInstance, disposeInstance} from "@/views/core/master";
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from "vue";
import {ElMessage, ElMessageBox} from "element-plus";
import {ImportExcel, ExportExcel, FetchActConfig, KeepActionConfig, FetchActList} from "../../../wailsjs/go/main/App";
import {dataparser, Sheet} from "../../../wailsjs/go/models";
import debounce from "lodash/debounce";

const props = defineProps({
  actId: {
    type: Number,
    default: 0
  }
})

const univerRef = ref(null as any)
const univerAPIRef = ref(null as any)
const curAct = ref({})

const selectedAct = (actInfo) => {
  // 选择活动
  ElMessageBox.confirm(
      `你确定要选择 【${actInfo.ActName}】 活动吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
  )
      .then(async() => {
        try {
          const res = await FetchActConfig(String(actInfo.ActId))
          if (res.status !== 200) {
            ElMessage.error(`获取失败:${res.msg}`)
            return
          }
          disposeUniver()
          initUniver()
          const workbook = univerAPIRef.value.createWorkbook(res.data)
          univerAPIRef.value.addWatermark('text', {content:`${workbook.getId()}`, fontSize: 18, repeat: true})
          curAct.value = actInfo
        }catch (err) {
          ElMessage.error(`获取异常:${err}`)
        }
      })
      .catch((err) => {
        console.log(err)
      })
}

const colorHandler = (idx:number)=>{
  switch (idx%5){
    case 0:
      return "primary"
    case 1:
      return "success"
    case 2:
      return "info"
    case 3:
      return "warning"
    case 4:
      return "danger"
  }
}

const selectedHandler = (item) => {
  if (item.ActId === curAct.value.ActId){
    return true
  }
  return false
}

const disposeUniver = () => {
  univerAPIRef.value?.dispose()
  univerRef.value = null
  univerAPIRef.value = null
}

const initUniver = () => {
  const {univer, univerAPI} = createInstance("mainContainer")
  univerRef.value = univer
  univerAPIRef.value = univerAPI
  const el = document.getElementById("mainContainer")
  el?.classList.add("has-content")
}

const menuPopoverRef = ref()
// 拿原始数据进行渲染
const menuPopoverParams = ref({
  title: '',
  trigger:'click',
  visiable: false,
  width: 500,
  type: 1, // 1:ShowAct, 2:Renew, 3:Import, 4: Export, 5: hotUpdate
})

const toShowAct = () => {
  menuPopoverParams.value.type = 1
}
const toRenew = async() => {
  menuPopoverParams.value.type = 2
  const { value: actId } = await ElMessageBox.prompt('请输入活动ID', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^\d+$/, // 可选，只允许数字
    inputErrorMessage: '活动ID必须是数字',
  });
  disposeUniver()
  initUniver()
  let workbookKey = "Activity_"+actId
  const fworkbook = univerAPIRef.value.createWorkbook({ id: workbookKey, name: workbookKey });
  univerAPIRef.value.addWatermark('text', {content:`${fworkbook.getId()}`, fontSize: 18, repeat: true})
  const curSheet = fworkbook.getActiveSheet()
  curSheet.setName("配置");
  const frange = curSheet.getRange("A1:A4")
  frange.setValues([
    [{ v: "中文字段名称" }],
    [{ v: "英文字段名" }],
    [{ v: "字段类型" }],
    [{ v: "值" }]
  ])
  curSheet.activate()
}

const toHotUpdate = async() => {
  menuPopoverParams.value.type = 5
  const fworkbook: any = univerAPIRef.value.getActiveWorkbook()
  const fworksheets = fworkbook?.getSheets()
  const fworkbookdata: dataparser.Workbook = {
    id: fworkbook?.id,
    name: fworkbook?.getName(),
    sheetOrder: [],
    sheets: {},
  };
  for (let i = 0; i < fworksheets.length; i++) {
    const sheet = fworksheets[i]?.getSheet();
    if (!sheet) continue;

    const sheetId = String(sheet.getSheetId()); // 转成字符串
    fworkbookdata.sheets![sheetId] = sheet.getSnapshot(); // sheets 已初始化，!安全断言
    fworkbookdata.sheetOrder!.push(sheetId);          // sheetOrder 已初始化
  }
  try {
    const res = await KeepActionConfig(fworkbookdata)
    if (res.status === 200) {
      ElMessage.success(`实时更新成功`)
    } else {
      ElMessage.error(`实时更新失败:${res.msg}`)
    }
  }catch (e) {
    ElMessage.error(`实时更新失败:${e}`)
  }
}
const toImport = async() => {
  menuPopoverParams.value.type = 3
  // 把路径传给后端
  try {
    const { value: actId } = await ElMessageBox.prompt('请输入活动ID', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /^\d+$/, // 可选，只允许数字
      inputErrorMessage: '活动ID必须是数字',
    });
    const res = await ImportExcel(actId)
    if (res.status == 200) {
      ElMessage.success("导入成功")
      disposeUniver()
      initUniver()
      const workbook = univerAPIRef.value.createWorkbook(res.data)
      univerAPIRef.value.addWatermark('text', {content:`${workbook.getId()}`, fontSize: 18, repeat: true})
    } else {
      ElMessage.error(res.msg)
    }
  }catch (e) {
    ElMessage.error(e)
  }
}
const toGetActList = async() => {
  const genEmptyListData = () => {
    screenActData.value = [];
  }
  try {
    const res = await FetchActList(queryActInfo.value)
    if (res.status === 200) {
      screenActData.value = res.data || []
    } else {
      ElMessage.error(`获取活动列表失败:${res.msg}`)
      genEmptyListData()
    }
  }catch (e){
    ElMessage.error(`获取活动列表异常:${e}`)
    genEmptyListData()
  }
}
const toExport = async() => {
  menuPopoverParams.value.type = 4
  const fworkbook: any = univerAPIRef.value.getActiveWorkbook()
  const fworksheets = fworkbook?.getSheets()
  const fworkbookdata: dataparser.Workbook = {
    id: fworkbook?.id,
    name: fworkbook?.getName(),
    sheetOrder: [],
    sheets: {},
  };
  for (let i = 0; i < fworksheets.length; i++) {
    const sheet = fworksheets[i]?.getSheet();
    if (!sheet) continue;

    const sheetId = String(sheet.getSheetId()); // 转成字符串
    fworkbookdata.sheets![sheetId] = sheet.getSnapshot(); // sheets 已初始化，!安全断言
    fworkbookdata.sheetOrder!.push(sheetId);          // sheetOrder 已初始化
  }
  try {
    const res = await ExportExcel(fworkbookdata)
    if (res.status === 200) {
      ElMessage.success(`导出目录:${res.data}`)
    } else {
      ElMessage.error(`导出失败:${res.msg}`)
    }
  }catch (e) {
    ElMessage.error(`导出失败:${e}`)
  }
}
const beforeEnterMenuInfo = () => {
  if (menuPopoverParams.value.type !== 1) {
    menuPopoverRef.value.hide()
  }
}

// 活动查询过滤
const screenActData = ref([])
const queryActInfo = ref('')

const debouncedGetList = debounce(() => {
  toGetActList();
}, 600);

watch(queryActInfo, () => {
  debouncedGetList()
});
// 获取excel实例
onMounted(()=>{
  toGetActList()
})
onBeforeUnmount(()=>{
  disposeUniver()
})
</script>

<template>
  <div class="main-container" id="mainContainer">
  </div>
  <div ref="menuRef" class="menu draggable">
    <el-popover
        ref="menuPopoverRef"
        v-model="menuPopoverParams.visiable"
        placement="bottom"
        :title="menuPopoverParams.title"
        :width="menuPopoverParams.width"
        trigger="click"
        @before-enter="beforeEnterMenuInfo"
    >
      <template #reference>
        <el-card shadow="hover" style="width: 340px;height: 57px">
          <div id="SocailIcons">
            <div class="icons instaIcon" @click="toShowAct">
              <p class="iconName">活动列表</p>
              <div class="icon insta">
                <el-icon :size="30"><Grid /></el-icon>
              </div>
            </div>
            <div class="icons linkedin" @click="toImport">
              <p class="iconName">导入配置</p>
              <div class="icon link">
                <el-icon :size="30"><Upload /></el-icon>
              </div>
            </div >
            <div v-if="univerRef" class="icons whatsapp" @click="toExport">
              <p class="iconName">导出配置</p>
              <div class="icon whats">
                <el-icon :size="30" style="font-weight: 800"><Download /></el-icon>
              </div>
            </div>
            <div class="icons youtube" :style="univerRef?{}:{}" @click="toRenew">
              <p class="iconName">新建配置</p>
              <div class="icon tube">
                <el-icon :size="30"><HomeFilled /></el-icon>
              </div>
            </div>
            <div v-if="univerRef" class="icons hotupdate" @click="toHotUpdate">
              <p class="iconName">热更新</p>
              <div class="icon tube">
                <el-icon :size="30"><MagicStick /></el-icon>
              </div>
            </div>
          </div>
        </el-card>
      </template>
      <template #default>
        <el-input placeholder="输入活动关键词" v-model="queryActInfo"/>
        <el-divider/>
        <ul class="list webkit-scrollbar" role="list" dir="auto">
          <li class="listitem" role="listitem" v-for="(item,index) in screenActData" :key="index" @click="selectedAct(item)">
            <div class="detail-info" :class="selectedHandler(item)?'selected':''">
              <el-text class="mx-1" size="large" :type="colorHandler(index)">{{item.ActName}}</el-text>
              <el-tag size="small" :type="colorHandler(index)">{{item.ActId}}</el-tag>
            </div>
            <el-divider/>
          </li>
        </ul>
      </template>
    </el-popover>
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
  >div{
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

.menu{
  position: absolute;
  z-index: 999;
  right: 4%;
  top: 80%;
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
  border: 2px solid #4a90e2;            /* 蓝色边框 */
  background: rgba(74, 144, 226, 0.1);  /* 浅蓝背景 */
}

.detail-info:hover {
  background: rgba(74, 144, 226, 0.05); /* 悬停时非常浅的蓝色 */
  cursor: pointer;
}


.detail-info{
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
  position: relative;           /* 关键 */
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
  transition:
      opacity 0.2s ease,
      transform 0.2s ease;
  pointer-events: none;
}


.icons:hover .iconName {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}
</style>