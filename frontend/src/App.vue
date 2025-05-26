<template>
  <div id="app">
    <!-- WeUI 风格的顶部导航 -->
    <div class="weui-navigation-bar">
      <div class="weui-navigation-bar__inner">
        <div class="weui-navigation-bar__left">
          <a href="javascript:" class="weui-navigation-bar__btn theme-toggle" @click="toggleTheme" :title="themeStore ? themeStore.getThemeDescription() : '主题切换'">
            <i class="weui-icon-theme">{{ themeStore ? themeStore.getThemeIcon() : '🔄' }}</i>
          </a>
        </div>
        <div class="weui-navigation-bar__center">
          <strong class="weui-navigation-bar__title">{{ pageTitle }}</strong>
        </div>
        <div class="weui-navigation-bar__right">
          <a href="javascript:" class="weui-navigation-bar__btn" @click="router.push('/settings')" v-if="route.path === '/'">
            <i class="weui-icon-more"></i>
          </a>
          <a href="javascript:" class="weui-navigation-bar__btn" @click="router.go(-1)" v-if="route.path === '/settings'">
            <i class="weui-icon-back"></i>
          </a>
        </div>
      </div>
    </div>

    <!-- 主要内容区域 -->
    <div class="weui-tab">
      <div class="weui-tab__panel">
        <router-view />
      </div>
    </div>
  </div>
</template>

<script>
import { useThemeStore } from './stores/theme'
import { onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export default {
  name: 'App',
  setup() {
    const themeStore = useThemeStore()
    const route = useRoute()
    const router = useRouter()

    const toggleTheme = () => {
      if (themeStore && themeStore.toggleTheme) {
        themeStore.toggleTheme()
      } else {
        console.error('主题store未正确初始化')
      }
    }

    const pageTitle = computed(() => {
      return route.path === '/settings' ? '设置' : 'LazyBala'
    })

    onMounted(() => {
      if (themeStore && themeStore.initTheme) {
        themeStore.initTheme()
      } else {
        console.error('主题store未正确初始化')
      }
    })

    return {
      themeStore,
      route,
      router,
      toggleTheme,
      pageTitle
    }
  }
}
</script>

<style>
@import 'weui/dist/style/weui.css';

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* CSS 变量定义 */
:root {
  /* 浅色模式 */
  --bg-primary: #f7f7f7;
  --bg-secondary: #ffffff;
  --bg-tertiary: #f8f8f8;
  --text-primary: #000000;
  --text-secondary: #666666;
  --text-tertiary: #999999;
  --border-color: #e5e5e5;
  --border-secondary: #f0f0f0;
  --accent-color: #ff6b35;
  --accent-hover: #e55a2b;
  --success-color: #52c41a;
  --warning-color: #fa541c;
  --info-color: #1890ff;
  --shadow-color: rgba(0, 0, 0, 0.05);
}

:root.dark-theme {
  /* 暗黑模式 */
  --bg-primary: #000000;
  --bg-secondary: #1c1c1e;
  --bg-tertiary: #2c2c2e;
  --text-primary: #ffffff;
  --text-secondary: #8e8e93;
  --text-tertiary: #8e8e93;
  --border-color: #2c2c2e;
  --border-secondary: #3a3a3c;
  --accent-color: #ff6b35;
  --accent-hover: #e55a2b;
  --success-color: #30d158;
  --warning-color: #ff453a;
  --info-color: #007aff;
  --shadow-color: rgba(255, 255, 255, 0.1);
}

/* 全局样式 */
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica, 'Segoe UI', Arial, Roboto, 'Miui', 'Hiragino Sans GB', 'Microsoft Yahei', sans-serif;
  background-color: var(--bg-primary);
  color: var(--text-primary);
  line-height: 1.4;
  transition: background-color 0.3s ease, color 0.3s ease;
}

#app {
  min-height: 100vh;
  background-color: var(--bg-primary);
}

/* WeUI 导航栏样式 */
.weui-navigation-bar {
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 1000;
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.weui-navigation-bar__inner {
  display: flex;
  align-items: center;
  height: 44px;
  padding: 0 16px;
  position: relative;
}

.weui-navigation-bar__left,
.weui-navigation-bar__right {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 44px;
}

.weui-navigation-bar__center {
  flex: 1;
  text-align: center;
}

.weui-navigation-bar__title {
  font-size: 17px;
  font-weight: 600;
  color: var(--text-primary);
  transition: color 0.3s ease;
}

.weui-navigation-bar__btn {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent-color);
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.weui-navigation-bar__btn:hover {
  background-color: var(--shadow-color);
}

.theme-toggle {
  margin-left: 0;
}

.weui-icon-back::before {
  content: '‹';
  font-size: 24px;
  font-weight: bold;
}

.weui-icon-more::before {
  content: '⋯';
  font-size: 20px;
  font-weight: bold;
}

/* 主题切换图标 */
.weui-icon-theme {
  font-size: 16px;
  transition: transform 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.theme-toggle:hover .weui-icon-theme {
  transform: scale(1.1);
}

.theme-toggle:active .weui-icon-theme {
  transform: scale(0.95);
}

/* 主要内容区域 */
.weui-tab__panel {
  padding: 0;
  background-color: var(--bg-primary);
  min-height: calc(100vh - 44px);
  transition: background-color 0.3s ease;
}


</style>
