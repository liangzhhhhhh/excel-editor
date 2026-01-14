<script lang="ts" setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import Canva from "@/views/index/canva.vue";
import Menu from "@/views/index/menu.vue";

const selectedAct = ref()

onMounted(async () => {
})

onBeforeUnmount(() => {})
</script>

<template>
  <div ref="container" class="container">
    <div class="container-menu"><Menu v-model:act-id="selectedAct"/></div>
    <div class="container-core"><Canva v-model:act-id="selectedAct"/></div>
  </div>
</template>

<style lang="scss">
.container{
  width: 100%;
  height: 100%;
  display: flex;
  background: rgb(240,240,240);
}
.container-core{
  flex: 1;
  height: 100%;
}
.container-menu{
  height: 100%;
}
</style>
