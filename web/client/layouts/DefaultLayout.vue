<template>
  <v-layout class="h-screen app-background">
    <!-- 渐变模糊导航栏 -->
    <v-app-bar :elevation="0" class="app-nav-bar border-neon">
      <!-- Logo 和标题区域 -->
      <div class="header-logo-section">
        <div class="logo-container">
          <img src="/logo.jpg" alt="Linker Logo" class="app-logo" />
        </div>
        <span class="app-title text-primary-glow">Linker</span>
      </div>
      
      <v-spacer></v-spacer>
    
      <!-- 中间：主题切换 -->
      <div class="theme-switcher">
        <v-btn
          variant="text"
          icon
          class="nav-btn theme-btn"
          @click="themeStore.toggleTheme"
        >
          <v-icon :icon="themeStore.currentTheme === 'dark' ? 'mdi-weather-night' : 'mdi-weather-sunny'" size="20"></v-icon>
          <v-tooltip activator="parent" location="bottom">切换主题</v-tooltip>
        </v-btn>
      </div>

      <v-spacer></v-spacer>

      <div class="nav-items text-slate-300">
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
          <v-list class="glass-menu" elevation="0">
            <v-list-item href="https://github.com/likun7981/hlink" target="_blank" prepend-icon="mdi-github" class="menu-item-hover">
              <v-list-item-title class="font-mono">Github</v-list-item-title>
            </v-list-item>
            <v-list-item href="https://hlink.likun.me" target="_blank" prepend-icon="mdi-book-open-variant" class="menu-item-hover">
              <v-list-item-title class="font-mono">文档</v-list-item-title>
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
import { useThemeStore } from '../stores/theme'

const $route = useRoute()
const themeStore = useThemeStore()
</script>

<style scoped>
/* Glassmorphism Header */
.app-nav-bar {
  /* Using global variable in style.css, but adding specific override here if needed */
}

.app-nav-bar :deep(.v-toolbar__content) {
  padding: 0 32px;
  height: 72px;
}

.header-logo-section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo-container {
  width: 40px;
  height: 40px;
  background: rgba(15, 23, 42, 0.8);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  box-shadow: 0 0 15px rgba(0, 240, 255, 0.2);
  border: 1px solid rgba(0, 240, 255, 0.3);
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.logo-container:hover {
  transform: scale(1.05);
  box-shadow: 0 0 20px rgba(0, 240, 255, 0.4);
  border-color: #00F0FF;
}

.app-logo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.app-title {
  font-family: 'Orbitron', sans-serif;
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: 1px;
}

/* Nav Buttons */
.nav-items {
  display: flex;
  align-items: center;
  gap: 4px;
}

.nav-btn {
  color: var(--color-text-muted) !important;
  font-family: 'Space Mono', monospace;
  font-weight: 600;
  height: 36px !important;
  letter-spacing: 0.5px;
  transition: all 0.3s ease !important;
  border: 1px solid transparent !important;
}

.nav-btn:hover {
  background: rgba(var(--color-primary-rgb), 0.05) !important;
  color: var(--color-primary) !important;
}

.nav-btn-active {
  background: rgba(var(--color-primary-rgb), 0.1) !important;
  color: var(--color-primary) !important;
  border: 1px solid rgba(var(--color-primary-rgb), 0.3) !important;
  box-shadow: 0 0 10px rgba(var(--color-primary-rgb), 0.1);
}

/* Menu Items */
.menu-item-hover {
  color: var(--color-text-muted) !important;
  transition: all 0.2s ease;
}

.menu-item-hover:hover {
  background: rgba(var(--color-primary-rgb), 0.1) !important;
  color: var(--color-primary) !important;
}

.main-content {
  position: relative;
}

/* Page Transition */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
  filter: blur(4px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
  filter: blur(4px);
}
</style>
