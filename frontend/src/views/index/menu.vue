<template>
  <div class="component-container">
   <div class="title-container" v-if="!collapseStatus">
<!--      <div class="title"><icon-apps></icon-apps>&nbsp;活动列表</div>-->
       <a-input-search :style="{width:'160px'}" placeholder="活动ID" v-model="keyword"/>
       <a-divider style="height: 8px;margin:0"/>
   </div>
      <div class="menu">
    <a-menu
        ref="menuRef"
        class="custom-menu"
        v-model:selected-keys="selectedKey"
        :style="{ width: '200px', height: '100%' }"
        show-collapse-button
        breakpoint="md"
        auto-open
        @collapse="collapseMenu"
        @menu-item-click="handleSelectedKey"
    >
<!--        <template #icon><icon-apps></icon-apps></template>-->
<!--        <template #title><div>活动列表</div></template>-->
        <a-sub-menu :key="0">
            <template #icon v-if="collapseStatus"><icon-apps></icon-apps></template>
        <a-menu-item
            v-for="(item) in screenActData"
            :key="item.ActId"
        >
          <span class="activity-tag" :class="`tag-${String(item.ActId).slice(-1) % 4}`">
            {{ item.ActId }}
          </span>
          <span class="activity-name" :title="item.ActName">{{item.ActName}}</span>
        </a-menu-item>
<!--          <a-menu-item-->
<!--              v-for="(item) in 100"-->
<!--              :key="item"-->
<!--          >-->
<!--          <span class="activity-tag">-->
<!--            {{ item }}-->
<!--          </span>-->
<!--              <span class="activity-name" :title="item">{{item}}</span>-->
<!--          </a-menu-item>-->
        </a-sub-menu>
    </a-menu>
      </div>
  </div>
</template>
<script lang="ts" setup>
import {computed, onMounted, ref, watch} from 'vue';
import {
  IconApps,
} from '@arco-design/web-vue/es/icon';
import {FetchActList} from "../../../wailsjs/go/main/App";
import debounce from "lodash/debounce";
import {Message} from "@arco-design/web-vue";
import {handleResp} from "@/config/apis/filter";

const props = defineProps({
  actId: {
    type: Number,
    default: 0
  }
})

const menuRef = ref()

const actIdModel = computed({
  get(){
    return Number(props.actId)
  },
  set(val){
    emitter('update:actId', val)
  }
})

const changeAct = (id) => {
  actIdModel.value = id
}

const keyword = ref('')
const doSearch = debounce((kw: string) => {
    const key = kw.trim()
    screenActData.value = !key
        ? allActData.value
        : allActData.value.filter(item => {
                return String(item.ActId).includes(key)
            }
        )
}, 500) // 停 500ms 才搜索

const emitter = defineEmits(['update:actId'])
const allActData = ref([])
const screenActData = ref([])

const toGetActList = async() => {
  const genEmptyListData = () => {
      allActData.value = [];
  }
  try {
    const resp = await FetchActList("")
    const res = handleResp(resp)
    allActData.value = res || []
    screenActData.value = allActData.value
  }catch (e){
    Message.error(`获取活动列表异常:${e.message}`)
    genEmptyListData()
  }
}

const selectedKey = ref([])

const handleSelectedKey = (actId) => {
  if (actIdModel.value === Number(actId)) return
  changeAct(actId)
}
const collapseStatus = ref(false)
const collapseMenu = (v) => {
    collapseStatus.value = v
}

const queryActInfo = ref('')

const debouncedGetList = debounce(() => {
  toGetActList();
}, 600);

watch(queryActInfo, () => {
  debouncedGetList()
});

watch(keyword, (kw) => {
    doSearch(kw)
})

watch(()=>actIdModel.value, ()=>{
  if (selectedKey.value.length <= 0) selectedKey.value = [actIdModel.value]
})
// 获取excel实例
onMounted( ()=>{
  toGetActList()
})
</script>

<style lang="scss" scoped>
.component-container{
  height: 100%;
    display: flex;
    flex-direction: column;
}
.title-container{
    height: 45px;
    padding-top: 5px;
    padding-left:8px;
    font-size: 20px;
    font-weight: 500;
    background: white;
}
.title{
    padding-bottom: 5px;
}
.menu{
    flex: 1;
    overflow: auto;
}
/* 整个 menu item 压扁 */
.activity-item {
  height: 32px;
  line-height: 32px;
  padding: 0 12px !important;
  display: flex;
  align-items: center;
  gap: 10px;
}

.activity-tag {
  font-size: 12px;
  line-height: 18px;
  padding: 0 6px;
  border-radius: 4px;
  background-color: rgba(255, 125, 0, 0.15);
  color: #ff7d00; margin-right: 8px;
  flex-shrink: 0;
}
.activity-name {
  font-size: 14px;
  color: var(--color-text-1);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 蓝 */
.tag-0 {
  color: #165dff;
  background-color: rgba(22, 93, 255, 0.12);
}

/* 橙 */
.tag-1 {
  color: #ff7d00;
  background-color: rgba(255, 125, 0, 0.12);
}

/* 绿 */
.tag-2 {
  color: #00b42a;
  background-color: rgba(0, 180, 42, 0.12);
}

/* 紫 */
.tag-3 {
  color: #722ed1;
  background-color: rgba(114, 46, 209, 0.12);
}

:deep(.custom-menu .arco-menu-inner::-webkit-scrollbar) {
  width: 8px;
}

:deep(.custom-menu .arco-menu-inner::-webkit-scrollbar-thumb) {
  background-color: rgba(0, 0, 0, 0.25);
  border-radius: 4px;
}

:deep(.custom-menu .arco-menu-inner::-webkit-scrollbar-thumb:hover) {
  background-color: rgba(0, 0, 0, 0.4);
}
:deep(.arco-menu-inline-header ){
  .arco-menu-icon-suffix{
    display: none;
  }
  pointer-events: none;
  user-select: none;
}
</style>
