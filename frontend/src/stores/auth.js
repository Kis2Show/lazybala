import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const loggedIn = ref(false)

  // 检查服务器端cookies状态
  const checkServerCookies = async () => {
    try {
      const response = await fetch('/api/config')
      if (response.ok) {
        const config = await response.json()
        if (config.has_cookies && config.cookies_valid) {
          loggedIn.value = true
          console.log('检测到有效的cookies，自动设置为已登录状态')
        }
      }
    } catch (error) {
      console.error('检查服务器cookies状态失败:', error)
    }
  }

  // 初始化时检查cookies状态
  checkServerCookies()

  const isLoggedIn = computed(() => loggedIn.value)

  const setLoggedIn = (status) => {
    loggedIn.value = status
    // 保存到 localStorage
    localStorage.setItem('lazybala_auth', JSON.stringify({
      loggedIn: status
    }))
  }

  const logout = () => {
    setLoggedIn(false)
  }

  return {
    isLoggedIn,
    setLoggedIn,
    logout
  }
})
