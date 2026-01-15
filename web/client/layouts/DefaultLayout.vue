<template>
  <v-layout class="rounded rounded-md h-screen app-layout">
    <!-- 渐变模糊导航栏 -->
    <v-app-bar :elevation="0" class="app-nav-bar">
      <!-- Logo 和标题区域 -->
      <div class="header-logo-section">
        <div class="logo-container">
          <img src="/logo.jpg" alt="Linker Logo" class="app-logo" />
        </div>
        <span class="app-title">Linker</span>
      </div>
      
      <v-spacer></v-spacer>

      <div class="nav-items">
        <v-btn 
          to="/" 
          prepend-icon="mdi-format-list-bulleted"
          variant="text"
          class="nav-btn mx-1"
          :class="{ 'nav-btn-active': $route.path === '/' }"
        >任务列表</v-btn>
        <v-btn 
          to="/config" 
          prepend-icon="mdi-cog-outline"
          variant="text"
          class="nav-btn mx-1"
          :class="{ 'nav-btn-active': $route.path === '/config' }"
        >配置管理</v-btn>
        
        <v-menu location="bottom end" transition="scale-transition">
          <template v-slot:activator="{ props }">
            <v-btn icon="mdi-dots-vertical" v-bind="props" variant="text" class="nav-btn ml-1"></v-btn>
          </template>
          <v-list class="glass-menu rounded-xl" elevation="0">
            <v-list-item href="https://github.com/likun7981/hlink" target="_blank" prepend-icon="mdi-github" class="menu-item-hover">
              <v-list-item-title>Github</v-list-item-title>
            </v-list-item>
            <v-list-item href="https://hlink.likun.me" target="_blank" prepend-icon="mdi-book-open-variant" class="menu-item-hover">
              <v-list-item-title>文档</v-list-item-title>
            </v-list-item>
          </v-list>
        </v-menu>
      </div>
    </v-app-bar>

   <v-main class="main-content">
      <v-container fluid class="pa-0 h-100">
        <div class="w-full max-w-7xl mx-auto px-4 py-6 h-100">
          <router-view v-slot="{ Component, route }">
            <transition name="page-fade" mode="out-in">
              <component :is="Component" :key="route.path" />
            </transition>
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
.app-layout {
  background: #f0f2f5; /* Fallback */
  background: linear-gradient(135deg, #f3f5f9 0%, #eef2f7 100%);
}

/* Glassmorphism Header */
.app-nav-bar {
  background: rgba(255, 255, 255, 0.8) !important;
  backdrop-filter: blur(20px) !important;
  -webkit-backdrop-filter: blur(20px) !important;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05) !important;
  height: 80px !important;
}

.app-nav-bar :deep(.v-toolbar__content) {
  padding: 0 32px;
  height: 80px;
}

.header-logo-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo-container {
  width: 44px;
  height: 44px;
  background: white;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-shadow: 0 8px 24px -6px rgba(102, 126, 234, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.5);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-container:hover {
  transform: rotate(-5deg) scale(1.05);
}

.app-logo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.app-title {
  font-size: 1.5rem;
  font-weight: 800;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -0.5px;
}

/* Nav Buttons */
.nav-items {
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(0, 0, 0, 0.03);
  padding: 4px;
  border-radius: 12px;
}

.nav-btn {
  color: #64748b !important;
  font-weight: 600;
  border-radius: 10px !important;
  height: 40px !important;
  letter-spacing: 0.3px;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1) !important;
}

.nav-btn:hover {
  background: rgba(0, 0, 0, 0.04) !important;
  color: #1e293b !important;
}

.nav-btn-active {
  background: white !important;
  color: #667eea !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

/* Glass Menu */
.glass-menu {
  background: rgba(255, 255, 255, 0.9) !important;
  backdrop-filter: blur(24px) !important;
  border: 1px solid rgba(255, 255, 255, 0.5);
  box-shadow: 0 20px 40px -4px rgba(0, 0, 0, 0.1) !important;
  padding: 8px;
  overflow: visible !important;
}

.menu-item-hover {
  border-radius: 8px;
  transition: all 0.2s ease;
  margin-bottom: 2px;
}

.menu-item-hover:hover {
  background: rgba(102, 126, 234, 0.08) !important;
  color: #667eea !important;
}

.main-content {
  position: relative;
}

/* Page Transition */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(10px) scale(0.99);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px) scale(0.99);
}

/* Global Scrollbar */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.2);
}
</style>
