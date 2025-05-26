import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  // 主题状态：'light', 'dark', 'auto'
  const theme = ref(localStorage.getItem('theme') || 'auto')

  // 系统主题检测
  const systemTheme = ref('dark')

  // 实际应用的主题
  const actualTheme = computed(() => {
    if (theme.value === 'auto') {
      return systemTheme.value
    }
    return theme.value
  })

  // 检测系统主题
  const detectSystemTheme = () => {
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      systemTheme.value = 'dark'
    } else {
      systemTheme.value = 'light'
    }
  }

  // 监听系统主题变化
  const watchSystemTheme = () => {
    if (window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      mediaQuery.addEventListener('change', (e) => {
        systemTheme.value = e.matches ? 'dark' : 'light'
        if (theme.value === 'auto') {
          applyTheme()
        }
      })
    }
  }

  // 切换主题（循环：auto -> light -> dark -> auto）
  const toggleTheme = () => {
    const themes = ['auto', 'light', 'dark']
    const currentIndex = themes.indexOf(theme.value)
    const nextIndex = (currentIndex + 1) % themes.length
    theme.value = themes[nextIndex]
    localStorage.setItem('theme', theme.value)
    applyTheme()
  }

  // 设置主题
  const setTheme = (newTheme) => {
    theme.value = newTheme
    localStorage.setItem('theme', newTheme)
    applyTheme()
  }

  // 应用主题到DOM
  const applyTheme = () => {
    const root = document.documentElement
    const currentTheme = actualTheme.value

    if (currentTheme === 'dark') {
      root.classList.add('dark-theme')
      root.classList.remove('light-theme')
    } else {
      root.classList.add('light-theme')
      root.classList.remove('dark-theme')
    }

    // 更新meta标签（移动端状态栏）
    updateMetaTheme(currentTheme)
  }

  // 更新meta主题色
  const updateMetaTheme = (currentTheme) => {
    const metaThemeColor = document.querySelector('meta[name="theme-color"]')
    const metaStatusBar = document.querySelector('meta[name="apple-mobile-web-app-status-bar-style"]')

    if (currentTheme === 'dark') {
      if (metaThemeColor) metaThemeColor.content = '#1c1c1e'
      if (metaStatusBar) metaStatusBar.content = 'black-translucent'
    } else {
      if (metaThemeColor) metaThemeColor.content = '#ffffff'
      if (metaStatusBar) metaStatusBar.content = 'default'
    }
  }

  // 初始化主题
  const initTheme = () => {
    detectSystemTheme()
    watchSystemTheme()
    applyTheme()
  }

  // 获取主题图标
  const getThemeIcon = () => {
    switch (theme.value) {
      case 'light':
        return '☀️'
      case 'dark':
        return '🌙'
      case 'auto':
      default:
        return '🔄'
    }
  }

  // 获取主题描述
  const getThemeDescription = () => {
    switch (theme.value) {
      case 'light':
        return '浅色模式'
      case 'dark':
        return '暗黑模式'
      case 'auto':
      default:
        return '跟随系统'
    }
  }

  return {
    theme,
    systemTheme,
    actualTheme,
    toggleTheme,
    setTheme,
    initTheme,
    getThemeIcon,
    getThemeDescription
  }
})
