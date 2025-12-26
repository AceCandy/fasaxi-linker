<template>
  <v-layout class="rounded rounded-md h-screen">
    <!-- 渐变导航栏 -->
    <v-app-bar :elevation="0" class="app-nav-bar">
      <v-app-bar-title class="font-weight-bold d-flex align-center">
        <div class="logo-container mr-3">
          <v-icon icon="mdi-link-variant" size="28" color="white"></v-icon>
        </div>
        <span class="app-title">Fasaxi Linker</span>
      </v-app-bar-title>
      
      <v-spacer></v-spacer>

      <v-btn 
        to="/" 
        prepend-icon="mdi-format-list-bulleted"
        variant="text"
        class="nav-btn mx-1"
        :class="{ 'nav-btn-active': $route.path === '/' }"
      >任务列表</v-btn>
      <v-btn 
        to="/config" 
        prepend-icon="mdi-cog"
        variant="text"
        class="nav-btn mx-1"
        :class="{ 'nav-btn-active': $route.path === '/config' }"
      >配置管理</v-btn>
      
      <v-menu>
        <template v-slot:activator="{ props }">
          <v-btn icon="mdi-dots-vertical" v-bind="props" variant="text" class="nav-btn"></v-btn>
        </template>
        <v-list class="rounded-lg" elevation="8">
          <v-list-item href="https://github.com/likun7981/hlink" target="_blank" prepend-icon="mdi-github">
            <v-list-item-title>Github</v-list-item-title>
          </v-list-item>
          <v-list-item href="https://hlink.likun.me" target="_blank" prepend-icon="mdi-book-open-variant">
            <v-list-item-title>文档</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </v-app-bar>

   <v-main class="main-content">
      <v-container fluid class="pa-2">
        <div class="w-full max-w-7xl mx-auto">
          <router-view v-slot="{ Component, route }">
            <component :is="Component" :key="route.path" />
          </router-view>
        </div>
      </v-container>
    </v-main>
  </v-layout>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

const $route = useRoute()
</script>

<style scoped>
.app-nav-bar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%) !important;
  box-shadow: 0 4px 20px rgba(102, 126, 234, 0.3) !important;
  height: 72px !important; /* 增加导航栏高度 */
}

.app-nav-bar :deep(.v-toolbar__content) {
  padding: 0 24px;
  height: 72px; /* 增加内容高度 */
}

.logo-container {
  width: 42px;
  height: 42px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(10px);
}

.app-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: white;
  letter-spacing: 0.5px;
}

.nav-btn {
  color: rgba(255, 255, 255, 0.9) !important;
  font-weight: 500;
  border-radius: 8px !important;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1) !important;
  position: relative;
}

.nav-btn:hover {
  background: rgba(255, 255, 255, 0.15) !important;
  color: white !important;
  transform: translateY(-1px);
}

.nav-btn-active {
  background: rgba(255, 255, 255, 0.2) !important;
  color: white !important;
  font-weight: 600;
}

.nav-btn-active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 10%;
  right: 10%;
  height: 3px;
  background: white;
  border-radius: 3px 3px 0 0;
}

.main-content {
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e8f0 100%);
}

/* 确保页面可以正常滚动 - 移除之前的限制 */
:deep(.v-layout) {
  height: auto !important;
  min-height: 100vh;
}

:deep(.v-main) {
  position: relative;
  flex: 1 1 auto;
  max-width: 100%;
  height: auto !important;
  overflow: auto !important;
}

/* 修复容器间距 */
:deep(.v-container) {
  padding-top: 0px !important; /* 完全消除顶部间距 */
  padding-left: 16px !important;
  padding-right: 16px !important;
  padding-bottom: 16px !important;
}

/* 页面切换动画 */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.15s cubic-bezier(0.4, 0, 0.2, 1),
              transform 0.15s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
