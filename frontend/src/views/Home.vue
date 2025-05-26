<template>
  <div class="home-container" style="background: red !important; color: white !important; padding: 20px !important; position: fixed !important; top: 0 !important; left: 0 !important; width: 100% !important; height: 100% !important; z-index: 9999 !important;">
    <h1 style="font-size: 24px !important; margin: 20px 0 !important;">🎉 测试 - Home组件已渲染</h1>
    <p style="font-size: 18px !important; margin: 10px 0 !important;">downloadProgress存在: {{ !!downloadProgress }}</p>
    <p style="font-size: 18px !important; margin: 10px 0 !important;">isDownloading: {{ isDownloading }}</p>
    <p style="font-size: 18px !important; margin: 10px 0 !important;">localIsDownloading: {{ localIsDownloading }}</p>
    <p style="font-size: 18px !important; margin: 10px 0 !important;">showSettings: {{ showSettings }}</p>

    <button @click="showSettings = !showSettings" style="background: blue !important; color: white !important; padding: 10px 20px !important; border: none !important; margin: 10px 0 !important; font-size: 16px !important;">
      切换设置: {{ showSettings ? '关闭' : '打开' }}
    </button>

    <!-- 原始导航栏（隐藏） -->
    <div class="navbar" style="display: none;">
      <div class="navbar-left">
        <div class="app-icon">
          <i class="icon-app"></i>
        </div>
      </div>
      <div class="navbar-center">
        <h1 class="app-title">LazyBala</h1>
      </div>
      <div class="navbar-right">
        <button @click="showSettings = !showSettings" class="settings-btn">
          <i class="icon-settings">⋯</i>
        </button>
      </div>
    </div>

    <!-- 设置面板 -->
    <div class="settings-panel" v-if="showSettings">
      <div class="settings-overlay" @click="showSettings = false"></div>
      <div class="settings-content">
        <div class="settings-header">
          <h3>设置</h3>
          <button @click="showSettings = false" class="close-btn">×</button>
        </div>
        <div class="settings-body">
          <!-- 登录设置 -->
          <div class="setting-item">
            <h4>账号登录</h4>
            <div v-if="!isLoggedIn">
              <button @click="showLoginDialog" class="setting-btn">扫码登录</button>
            </div>
            <div v-else>
              <span class="login-status">已登录</span>
              <button @click="logout" class="setting-btn secondary">退出登录</button>
            </div>
          </div>

          <!-- 主题设置 -->
          <div class="setting-item">
            <h4>主题模式</h4>
            <button @click="toggleTheme" class="setting-btn">
              {{ isDarkMode ? '切换到浅色模式' : '切换到深色模式' }}
            </button>
          </div>

          <!-- 标题格式设置 -->
          <div class="setting-item">
            <h4>文件名格式</h4>
            <input
              v-model="titleFormat"
              type="text"
              class="setting-input"
              placeholder="%(title)s.%(ext)s"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 登录弹窗 -->
    <div class="login-modal" v-if="showLoginOption">
      <div class="modal-overlay" @click="skipLogin"></div>
      <div class="modal-content">
        <div class="modal-header">
          <h3>扫码登录</h3>
          <button @click="skipLogin" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div class="qr-container">
            <canvas ref="qrCanvas" v-show="qrCode"></canvas>
            <div v-if="!qrCode" class="loading">
              <span>生成二维码中...</span>
            </div>
            <p class="login-status">{{ loginStatus }}</p>
          </div>
          <div class="modal-actions">
            <button @click="generateQR" class="btn primary" :disabled="isGeneratingQR">
              {{ isGeneratingQR ? '生成中...' : '重新生成' }}
            </button>
            <button @click="() => generateQR(true)" class="btn secondary" :disabled="isGeneratingQR">
              强制重新扫码
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- 输入区域 -->
      <div class="input-section">
        <div class="input-group">
          <label class="input-label">链接</label>
          <input
            v-model="downloadUrl"
            type="text"
            class="input-field"
            placeholder="https://www.bilibili.com/video/BV1YeXDYTELy"
            @keyup.enter="checkPlaylist"
          />
        </div>

        <div class="input-group">
          <label class="input-label">保存路径</label>
          <input
            v-model="savePath"
            type="text"
            class="input-field"
            placeholder="test"
          />
        </div>
      </div>

      <!-- 检查结果区域 - 位于保存路径下方，按钮上方 -->
      <div class="check-result" v-if="precheckResult" :class="{ 'downloading': downloadProgress?.isDownloading || localIsDownloading }">
        <div class="result-content">
          <div class="result-icon">
            <i class="icon-check">✓</i>
          </div>
          <div class="result-info">
            <div class="result-title">{{ precheckResult.title }} - 共{{ precheckResult.audio_count }}集</div>
            <div class="result-desc">准备就绪，可以开始下载</div>
          </div>
          <div class="result-actions">
            <button @click="stopDownload" class="btn-stop" v-if="downloadProgress?.isDownloading || localIsDownloading">
              终止下载
            </button>
          </div>
        </div>

        <!-- 橘色边框进度条 - 根据文件数/总文件数显示总进度 -->
        <div class="progress-border" v-if="downloadProgress?.isDownloading || localIsDownloading">
          <div class="progress-fill" :style="{ width: getTotalProgressPercent() + '%' }"></div>
        </div>
      </div>

      <!-- 按钮区域 - 两个按钮一排 -->
      <div class="button-section">
        <button
          @click="handleCheckListAction"
          class="btn primary"
          :class="{ 'disabled': getCheckListButtonDisabled() }"
          :disabled="getCheckListButtonDisabled()"
        >
          {{ getCheckListButtonText() }}
        </button>

        <button
          @click="handleBackgroundTasks"
          class="btn secondary"
          :class="{ 'disabled': !hasBackgroundTasks() }"
          :disabled="!hasBackgroundTasks()"
        >
          后台列表
        </button>
      </div>

      <!-- WebSocket状态 -->
      <div class="ws-status" v-if="downloadProgress?.isDownloading || localIsDownloading">
        <div class="status-item">
          <span class="status-label">状态:</span>
          <span class="status-value">{{ downloadProgress?.status || '连接中...' }}</span>
        </div>
        <div class="status-item">
          <span class="status-label">进度:</span>
          <span class="status-value">{{ downloadProgress?.currentIndex || 0 }}/{{ downloadProgress?.totalCount || 0 }}</span>
        </div>
        <div class="status-item" v-if="downloadProgress?.speed">
          <span class="status-label">速度:</span>
          <span class="status-value">{{ downloadProgress.speed }}</span>
        </div>
      </div>
    </div>

    <!-- 调试信息 (开发时使用) -->
    <div class="debug-info" v-if="showDebug">
      <p>调试信息:</p>
      <p>downloadProgress存在: {{ !!downloadProgress }}</p>
      <p>downloadProgress.isDownloading: {{ downloadProgress?.isDownloading || false }}</p>
      <p>isDownloading: {{ isDownloading || false }}</p>
      <p>localIsDownloading: {{ localIsDownloading || false }}</p>
      <p>progress: {{ downloadProgress?.progress || 0 }}</p>
      <p>status: {{ downloadProgress?.status || '无状态' }}</p>
    </div>

  </div>
</template>

<script>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useDownloadStore } from '../stores/download'
import { useAuthStore } from '../stores/auth'
import QRCode from 'qrcode'

export default {
  name: 'Home',
  setup() {
    const downloadStore = useDownloadStore()
    const authStore = useAuthStore()

    const qrCanvas = ref(null)
    const qrCode = ref('')
    const qrCodeKey = ref('')
    const isGeneratingQR = ref(false)
    const loginStatus = ref('请使用哔哩哔哩 APP 扫码登录')
    const loginCheckInterval = ref(null)

    const downloadUrl = ref('')
    const savePath = ref('')
    const showLoginOption = ref(false)
    const isChecking = ref(false)
    const precheckResult = ref(null)
    const localIsDownloading = ref(false)
    const showSettings = ref(false)
    const showDebug = ref(false)
    const titleFormat = ref('%(title)s.%(ext)s')
    const isDarkMode = ref(false)

    let ws = null

    // 生成二维码
    const generateQR = async (force = false) => {
      // 如果已经登录且不是强制扫码，直接返回
      if (authStore.isLoggedIn && !force) {
        loginStatus.value = '已成功登录，不需要扫码'
        showLoginOption.value = false
        return
      }

      isGeneratingQR.value = true
      loginStatus.value = '正在生成二维码...'

      try {
        const endpoint = force ? '/api/qrcode/force-generate' : '/api/qrcode/generate'
        const response = await fetch(endpoint, {
          method: 'POST'
        })
        const data = await response.json()

        if (data.already_logged_in && !force) {
          // 已经登录，显示提示
          loginStatus.value = data.message
          authStore.setLoggedIn(true)
          showLoginOption.value = false
        } else if (data.qrcode_key) {
          qrCodeKey.value = data.qrcode_key
          qrCode.value = data.qrcode

          // 在 canvas 上绘制二维码
          if (qrCanvas.value) {
            await QRCode.toCanvas(qrCanvas.value, data.url, {
              width: 160,
              height: 160,
              color: {
                dark: '#000000',
                light: '#FFFFFF'
              }
            })
          }

          // 开始检查登录状态
          startLoginCheck()
        }
      } catch (error) {
        console.error('生成二维码失败:', error)
        loginStatus.value = '生成二维码失败，请重试'
      } finally {
        isGeneratingQR.value = false
      }
    }

    // 开始检查登录状态
    const startLoginCheck = () => {
      if (loginCheckInterval.value) {
        clearInterval(loginCheckInterval.value)
      }

      loginCheckInterval.value = setInterval(async () => {
        try {
          const response = await fetch('/api/qrcode/scan', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({
              qrcode_key: qrCodeKey.value
            })
          })

          const data = await response.json()

          switch (data.code) {
            case 86101:
              loginStatus.value = '请使用哔哩哔哩 APP 扫码登录'
              break
            case 86090:
              loginStatus.value = '已扫码，请在手机上确认登录'
              break
            case 0:
              loginStatus.value = '登录成功！'
              authStore.setLoggedIn(true)
              showLoginOption.value = false
              clearInterval(loginCheckInterval.value)
              break
            default:
              loginStatus.value = '登录失败，请重新扫码'
              clearInterval(loginCheckInterval.value)
              break
          }
        } catch (error) {
          console.error('检查登录状态失败:', error)
        }
      }, 2000)
    }

    // 显示登录对话框
    const showLoginDialog = () => {
      showLoginOption.value = true
      generateQR()
    }

    // 跳过登录
    const skipLogin = () => {
      showLoginOption.value = false
    }

    // 主题切换
    const toggleTheme = () => {
      isDarkMode.value = !isDarkMode.value
      document.documentElement.setAttribute('data-theme', isDarkMode.value ? 'dark' : 'light')
    }

    // 退出登录
    const logout = () => {
      authStore.setLoggedIn(false)
      showSettings.value = false
    }

    // 获取检查列表按钮文字
    const getCheckListButtonText = () => {
      if (isChecking.value) {
        return '检查中...'
      }
      if (downloadProgress.value?.isDownloading || localIsDownloading.value) {
        return downloadProgress.value?.isPaused ? '继续下载' : '暂停下载'
      }
      if (precheckResult.value) {
        return '确认下载'
      }
      return '检查列表'
    }

    // 获取检查列表按钮禁用状态
    const getCheckListButtonDisabled = () => {
      if (isChecking.value) return true
      // 仅且仅在用户给定链接的时候启用
      if (!downloadUrl.value && !precheckResult.value && !(downloadProgress.value?.isDownloading || localIsDownloading.value)) {
        return true
      }
      return false
    }

    // 处理检查列表按钮操作
    const handleCheckListAction = () => {
      if (downloadProgress.value?.isDownloading || localIsDownloading.value) {
        // 如果正在下载，切换暂停/继续
        if (downloadProgress.value?.isPaused) {
          resumeDownload()
        } else {
          pauseDownload()
        }
      } else if (precheckResult.value) {
        // 如果有检查结果，开始下载
        confirmDownload()
      } else {
        // 否则检查列表
        checkPlaylist()
      }
    }

    // 任务历史状态
    const taskHistory = ref([])
    const hasActiveTask = ref(false)

    // 检查是否有后台任务
    const hasBackgroundTasks = () => {
      // 检查是否有正在进行的下载任务或历史任务
      return hasActiveTask.value || taskHistory.value.length > 0 || downloadProgress.value?.isDownloading
    }

    // 获取任务历史
    const fetchTaskHistory = async () => {
      try {
        const response = await fetch('/api/download/history')
        if (response.ok) {
          const data = await response.json()
          taskHistory.value = data.tasks || []
          hasActiveTask.value = data.hasActiveTask || false
        }
      } catch (error) {
        console.error('获取任务历史失败:', error)
      }
    }

    // 处理后台任务
    const handleBackgroundTasks = async () => {
      // 首先检查是否有正在进行的下载任务
      if (downloadProgress.value?.isDownloading) {
        // 恢复显示当前下载任务
        localIsDownloading.value = true
        console.log('恢复显示当前下载任务')
        return
      }

      // 获取任务历史
      await fetchTaskHistory()

      if (hasActiveTask.value) {
        // 有活跃任务，尝试恢复
        try {
          const response = await fetch('/api/download/resume-background', {
            method: 'POST'
          })
          if (response.ok) {
            localIsDownloading.value = true
            console.log('恢复后台下载任务')
          } else {
            alert('恢复下载任务失败')
          }
        } catch (error) {
          console.error('恢复下载任务失败:', error)
          alert('恢复下载任务失败: ' + error.message)
        }
      } else if (taskHistory.value.length > 0) {
        // 显示最近3个任务的信息
        const recentTasks = taskHistory.value.slice(0, 3)
        const taskInfo = recentTasks.map((task, index) =>
          `${index + 1}. ${task.title} (${task.status})`
        ).join('\n')
        alert('最近任务:\n' + taskInfo)
      } else {
        // 无任务
        alert('暂无后台任务')
      }
    }

    // 检查播放列表
    const checkPlaylist = async () => {
      if (!downloadUrl.value) return

      isChecking.value = true
      precheckResult.value = null

      try {
        const response = await fetch('/api/download/precheck', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            url: downloadUrl.value
          })
        })

        if (response.ok) {
          const result = await response.json()
          precheckResult.value = result
        } else {
          const error = await response.json()
          alert('检查失败: ' + (error.error || '未知错误'))
        }
      } catch (error) {
        console.error('检查失败:', error)
        alert('检查失败: ' + error.message)
      } finally {
        isChecking.value = false
      }
    }

    // 确认下载
    const confirmDownload = async () => {
      if (!downloadUrl.value) return

      try {
        const response = await fetch('/api/download', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            url: downloadUrl.value,
            save_path: savePath.value,
            title_regex: '%(title)s.%(ext)s'
          })
        })

        if (response.ok) {
          console.log('下载请求成功，立即设置状态...')

          // 方法1：直接设置store状态
          downloadStore.setDownloading(true)

          // 方法2：同时设置本地状态作为备用
          localIsDownloading.value = true

          // 方法3：更新详细进度信息
          downloadStore.updateProgress({
            isDownloading: true,
            progress: 0,
            status: '正在启动下载...',
            currentIndex: 0,
            totalCount: precheckResult.value ? precheckResult.value.audio_count : 0,
            currentFile: '',
            speed: '',
            eta: '',
            currentTitle: '',
            lastActivity: '下载任务已提交',
            completedFiles: []
          })

          console.log('立即状态设置完成:')
          console.log('- downloadStore.isDownloading:', downloadStore.isDownloading)
          console.log('- downloadStore.progress:', downloadStore.progress)
          console.log('- localIsDownloading:', localIsDownloading.value)

          // 延迟重置检查结果
          setTimeout(() => {
            precheckResult.value = null
          }, 100)
        } else {
          const error = await response.json()
          alert('下载失败: ' + (error.error || '未知错误'))
        }
      } catch (error) {
        console.error('开始下载失败:', error)
        alert('下载失败: ' + error.message)
      }
    }

    // 停止下载
    const stopDownload = async () => {
      try {
        await fetch('/api/download/stop', {
          method: 'POST'
        })
        downloadStore.setDownloading(false)
        localIsDownloading.value = false
      } catch (error) {
        console.error('停止下载失败:', error)
      }
    }

    // 暂停下载
    const pauseDownload = async () => {
      try {
        const response = await fetch('/api/download/pause', {
          method: 'POST'
        })
        if (response.ok) {
          console.log('下载已暂停')
        } else {
          const error = await response.json()
          alert('暂停失败: ' + (error.error || '未知错误'))
        }
      } catch (error) {
        console.error('暂停下载失败:', error)
        alert('暂停失败: ' + error.message)
      }
    }

    // 继续下载
    const resumeDownload = async () => {
      try {
        const response = await fetch('/api/download/resume', {
          method: 'POST'
        })
        if (response.ok) {
          console.log('下载已继续')
        } else {
          const error = await response.json()
          alert('继续失败: ' + (error.error || '未知错误'))
        }
      } catch (error) {
        console.error('继续下载失败:', error)
        alert('继续失败: ' + error.message)
      }
    }

    // 获取短文件名（用于显示）
    const getShortFileName = (fileName) => {
      if (!fileName) return ''
      // 提取文件名中的关键部分，去掉路径和扩展名
      const name = fileName.split('\\').pop().split('/').pop()
      return name.length > 30 ? name.substring(0, 30) + '...' : name
    }

    // 获取已完成的文件
    const getCompletedFile = () => {
      if (downloadProgress.value.completedFiles && downloadProgress.value.completedFiles.length > 0) {
        const lastCompleted = downloadProgress.value.completedFiles[downloadProgress.value.completedFiles.length - 1]
        return getShortFileName(lastCompleted)
      }
      return null
    }

    // 检查是否正在下载
    const isCurrentlyDownloading = () => {
      return downloadProgress.value.isDownloading &&
             downloadProgress.value.progress >= 0
    }

    // 获取当前下载文件名
    const getCurrentFileName = () => {
      if (downloadProgress.value.currentTitle) {
        return getShortFileName(downloadProgress.value.currentTitle)
      }
      if (downloadProgress.value.currentFile) {
        return getShortFileName(downloadProgress.value.currentFile)
      }
      // 从状态中提取文件名
      if (downloadProgress.value.currentIndex) {
        return `第${downloadProgress.value.currentIndex}集`
      }
      return '准备中...'
    }

    // 获取即将下载的文件名
    const getUpcomingFileName = () => {
      if (downloadProgress.value.currentIndex && downloadProgress.value.totalCount) {
        const nextIndex = downloadProgress.value.currentIndex + 1
        if (nextIndex <= downloadProgress.value.totalCount) {
          return `第${nextIndex}集`
        }
      }
      return null
    }

    // 获取总进度百分比
    const getTotalProgressPercent = () => {
      if (!downloadProgress.value.currentIndex || !downloadProgress.value.totalCount) {
        return 0
      }
      const percent = Math.round((downloadProgress.value.currentIndex / downloadProgress.value.totalCount) * 100)

      // 动态更新检查结果区域的进度样式
      if (typeof document !== 'undefined') {
        document.documentElement.style.setProperty('--progress-width', `${percent}%`)
      }

      return percent
    }

    // 获取队列位置
    const getQueuePosition = () => {
      if (downloadProgress.value.currentIndex && downloadProgress.value.totalCount) {
        const nextIndex = downloadProgress.value.currentIndex + 1
        if (nextIndex <= downloadProgress.value.totalCount) {
          return `${nextIndex}`
        }
      }
      return ''
    }

    // 格式化状态信息
    const getFormattedStatus = () => {
      if (!downloadProgress.value.status) return ''

      // 处理中文乱码问题
      let status = downloadProgress.value.status
      if (status.includes('Downloading item')) {
        const match = status.match(/Downloading item (\d+) of (\d+)/)
        if (match) {
          return `正在处理第${match[1]}个，共${match[2]}个`
        }
      }

      if (status.includes('Extracting')) {
        return '正在提取视频信息...'
      }

      if (status.includes('Playlist')) {
        return '正在解析播放列表...'
      }

      return status
    }

    // 格式化最后活动时间
    const formatLastActivity = () => {
      if (!downloadProgress.value.lastActivity) return ''

      // 简化显示，只显示关键信息
      const activity = downloadProgress.value.lastActivity
      if (activity.includes('[download]')) {
        return '下载中...'
      }
      if (activity.includes('[info]')) {
        return '获取信息中...'
      }
      if (activity.includes('[BiliBili]')) {
        return '连接B站中...'
      }

      return '处理中...'
    }

    // 测试状态设置
    const testStateSet = () => {
      console.log('=== 测试状态设置 ===')

      try {
        console.log('设置前状态:')
        console.log('- downloadStore:', downloadStore)
        console.log('- downloadStore.isDownloading:', downloadStore.isDownloading)
        console.log('- downloadStore.progress:', downloadStore.progress)

        // 使用updateProgress方法
        downloadStore.updateProgress({
          isDownloading: true,
          progress: 50,
          status: '测试状态',
          currentIndex: 1,
          totalCount: 10,
          currentFile: '测试文件',
          speed: '1MB/s',
          eta: '5分钟',
          currentTitle: '测试标题'
        })

        console.log('设置后状态:')
        console.log('- downloadStore.isDownloading:', downloadStore.isDownloading)
        console.log('- downloadStore.progress:', downloadStore.progress)

      } catch (error) {
        console.error('状态设置错误:', error)
        console.error('错误堆栈:', error.stack)
      }
    }

    // 强制显示进度条
    const forceShowProgress = () => {
      console.log('=== 强制显示进度条 ===')

      // 方法1：设置本地状态
      localIsDownloading.value = true
      console.log('localIsDownloading设置为:', localIsDownloading.value)

      // 方法2：设置store状态
      downloadStore.setDownloading(true)
      console.log('store.isDownloading设置为:', downloadStore.isDownloading)

      // 方法3：设置详细进度
      downloadStore.updateProgress({
        isDownloading: true,
        progress: 25,
        status: '强制测试状态',
        currentIndex: 1,
        totalCount: 4,
        currentFile: '测试文件.m4a',
        speed: '1.5MB/s',
        eta: '2分钟',
        currentTitle: '测试标题',
        lastActivity: '强制显示测试',
        completedFiles: []
      })

      console.log('强制显示完成，当前状态:')
      console.log('- localIsDownloading:', localIsDownloading.value)
      console.log('- downloadProgress.isDownloading:', downloadStore.progress.value.isDownloading)
      console.log('- store.isDownloading:', downloadStore.isDownloading)
    }

    // 简单测试方法
    const simpleTest = () => {
      console.log('Vue事件绑定正常工作！')
      alert('Vue事件绑定成功！')

      // 测试状态设置
      localIsDownloading.value = true
      console.log('localIsDownloading设置为:', localIsDownloading.value)
    }

    // 初始化 WebSocket
    const initWebSocket = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/ws`
      console.log('正在连接WebSocket:', wsUrl)

      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('WebSocket连接已建立')
      }

      ws.onmessage = (event) => {
        const message = JSON.parse(event.data)
        console.log('WebSocket收到消息:', message)
        if (message.type === 'progress') {
          console.log('收到进度更新:', message.data)

          // 更新store
          downloadStore.updateProgress(message.data)

          // 强制更新本地状态以触发响应式更新
          if (message.data.isDownloading !== undefined) {
            localIsDownloading.value = message.data.isDownloading
            console.log('强制更新localIsDownloading:', localIsDownloading.value)
          }

          console.log('更新后的状态:')
          console.log('- store.progress:', downloadStore.progress.value)
          console.log('- store.isDownloading:', downloadStore.isDownloading)
          console.log('- localIsDownloading:', localIsDownloading.value)
        }
      }

      ws.onclose = (event) => {
        console.log('WebSocket连接关闭:', event.code, event.reason)
        // 重连
        setTimeout(() => {
          console.log('尝试重新连接WebSocket...')
          initWebSocket()
        }, 3000)
      }

      ws.onerror = (error) => {
        console.error('WebSocket连接错误:', error)
      }
    }

    onMounted(async () => {
      console.log('=== Home组件已挂载 ===')
      console.log('downloadStore:', downloadStore)
      console.log('downloadStore.progress:', downloadStore.progress)
      console.log('downloadStore.isDownloading:', downloadStore.isDownloading)
      console.log('authStore:', authStore)

      // 将方法暴露到window对象用于调试
      window.testStateSet = testStateSet
      window.forceShowProgress = forceShowProgress
      window.downloadStore = downloadStore
      window.localIsDownloading = localIsDownloading

      // 检查cookies状态
      try {
        const response = await fetch('/api/config')
        if (response.ok) {
          const config = await response.json()
          if (config.has_cookies && config.cookies_valid) {
            authStore.setLoggedIn(true)
          }
        }
      } catch (error) {
        console.error('检查cookies状态失败:', error)
      }

      // 初始化 WebSocket
      initWebSocket()

      // 获取任务历史
      fetchTaskHistory()

      console.log('=== Home组件挂载完成 ===')
    })

    onUnmounted(() => {
      if (loginCheckInterval.value) {
        clearInterval(loginCheckInterval.value)
      }
      if (ws) {
        ws.close()
      }
    })

    return {
      // refs
      qrCanvas,
      qrCode,
      isGeneratingQR,
      loginStatus,
      downloadUrl,
      savePath,
      showLoginOption,
      isChecking,
      precheckResult,
      localIsDownloading,
      showSettings,
      showDebug,
      titleFormat,
      isDarkMode,

      // computed
      isLoggedIn: authStore.isLoggedIn,
      isDownloading: downloadStore.isDownloading,
      downloadProgress: downloadStore.progress,

      // methods
      generateQR,
      showLoginDialog,
      skipLogin,
      toggleTheme,
      logout,
      getCheckListButtonText,
      getCheckListButtonDisabled,
      handleCheckListAction,
      hasBackgroundTasks,
      handleBackgroundTasks,
      fetchTaskHistory,
      checkPlaylist,
      confirmDownload,
      stopDownload,
      pauseDownload,
      resumeDownload,
      getShortFileName,
      getCompletedFile,
      isCurrentlyDownloading,
      getCurrentFileName,
      getUpcomingFileName,
      getTotalProgressPercent,
      getQueuePosition,
      getFormattedStatus,
      formatLastActivity,
      testStateSet,
      forceShowProgress,
      simpleTest
    }
  }
}
</script>

<style scoped>
/* 基础样式 */
.container {
  background-color: var(--bg-primary);
  min-height: 100vh;
  color: var(--text-primary);
  transition: all 0.3s ease;
}

/* WeUI 组件覆盖 */
.weui-form {
  background-color: var(--bg-primary);
}

.weui-form__text-area {
  background-color: var(--bg-primary);
}

.weui-form__title {
  color: var(--text-primary) !important;
}

.weui-form__desc {
  color: var(--text-secondary) !important;
}

.weui-cells__group {
  background-color: var(--bg-primary);
}

.weui-cells__title {
  color: var(--text-secondary) !important;
  background-color: var(--bg-primary);
}

.weui-cells {
  background-color: var(--bg-secondary);
  border-radius: 10px;
  overflow: hidden;
  transition: background-color 0.3s ease;
}

.weui-cell {
  background-color: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
  transition: all 0.3s ease;
}

.weui-cell:last-child {
  border-bottom: none;
}

.weui-label {
  color: var(--text-primary) !important;
}

.weui-input {
  background-color: var(--bg-secondary) !important;
  color: var(--text-primary) !important;
  border: none !important;
}

.weui-input::placeholder {
  color: var(--text-secondary) !important;
}

.weui-btn {
  border-radius: 8px !important;
  font-weight: 500 !important;
  transition: all 0.3s ease !important;
}

.weui-btn_primary {
  background-color: var(--accent-color) !important;
  border-color: var(--accent-color) !important;
  color: #ffffff !important;
}

.weui-btn_primary:hover {
  background-color: var(--accent-hover) !important;
  border-color: var(--accent-hover) !important;
}

.weui-btn_default {
  background-color: var(--bg-tertiary) !important;
  border-color: var(--bg-tertiary) !important;
  color: var(--text-primary) !important;
}

.weui-btn_default:hover {
  background-color: var(--border-secondary) !important;
  border-color: var(--border-secondary) !important;
}

.weui-btn_warn {
  background-color: var(--warning-color) !important;
  border-color: var(--warning-color) !important;
  color: #ffffff !important;
}

.weui-btn_warn:hover {
  opacity: 0.8;
}

.weui-btn_disabled {
  background-color: var(--bg-tertiary) !important;
  border-color: var(--bg-tertiary) !important;
  color: var(--text-secondary) !important;
  opacity: 0.6;
  pointer-events: none;
}

.weui-msg {
  background-color: var(--bg-primary);
}

.weui-msg__title {
  color: var(--text-primary) !important;
}

.weui-msg__desc {
  color: var(--text-secondary) !important;
}

.weui-icon-success {
  color: var(--success-color) !important;
}

.qr-container {
  text-align: center;
  padding: 20px 0;
  background-color: var(--bg-secondary);
  border-radius: 10px;
  margin: 10px 0;
  transition: background-color 0.3s ease;
}

.qr-status {
  margin-top: 10px;
  color: var(--text-secondary);
  font-size: 14px;
}

.weui-loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 20px;
  background-color: var(--bg-secondary);
  border-radius: 10px;
  transition: background-color 0.3s ease;
}

.weui-loading-content__text {
  color: var(--text-secondary);
  font-size: 14px;
}

.weui-form__tips-area {
  color: var(--text-secondary) !important;
  background-color: var(--bg-primary);
}

/* WeUI 图标补充 */
.weui-icon-success-no-circle::before {
  content: '✓';
  font-weight: bold;
  color: var(--success-color);
}

.weui-icon-waiting::before {
  content: '○';
  color: var(--text-secondary);
}

.weui-loading {
  border-color: var(--info-color) transparent transparent transparent !important;
}

/* 进度条样式 */
.weui-progress {
  display: flex;
  align-items: center;
  margin-top: 8px;
}

.weui-progress__bar {
  flex: 1;
  height: 4px;
  background-color: var(--border-color);
  border-radius: 2px;
  overflow: hidden;
  margin-right: 10px;
  transition: background-color 0.3s ease;
}

.weui-progress__inner-bar {
  height: 100%;
  background-color: var(--success-color);
  transition: width 0.3s ease;
}

.weui-progress__opr {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
}

/* 表单间距 */
.weui-form + .weui-form {
  margin-top: 32px;
}

.weui-form + .weui-msg {
  margin-top: 32px;
}

.weui-msg + .weui-form {
  margin-top: 32px;
}

/* 检查结果样式 */
.check-result-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-primary) !important;
  margin-bottom: 4px;
  line-height: 1.4;
}

.check-result-desc {
  font-size: 13px;
  color: var(--text-secondary) !important;
  margin: 0;
}

/* 按钮禁用状态 */
.weui-btn_disabled {
  opacity: 0.6;
  pointer-events: none;
}

/* 单元格描述文字 */
.weui-cell__desc {
  font-size: 12px;
  color: var(--text-secondary) !important;
  margin-top: 4px;
}

/* 增强的下载进度样式 */
.progress-overview {
  padding: 8px 0;
  background-color: var(--bg-secondary);
  border-radius: 8px;
  transition: background-color 0.3s ease;
}

.progress-stats {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  font-size: 14px;
  color: var(--text-primary);
}

.current-index {
  font-size: 18px;
  font-weight: 600;
  color: var(--success-color);
}

.separator {
  margin: 0 4px;
  color: var(--text-secondary);
}

.total-count {
  color: var(--text-secondary);
}

.progress-percent {
  font-weight: 600;
  color: var(--success-color);
  margin-left: auto;
}

.weui-progress--enhanced .weui-progress__bar {
  height: 6px;
  background: var(--border-color);
  border-radius: 3px;
}

.weui-progress--enhanced .weui-progress__inner-bar {
  background: linear-gradient(90deg, var(--success-color) 0%, var(--success-color) 100%);
  border-radius: 3px;
  box-shadow: 0 1px 2px var(--shadow-color);
}

/* 下载项目样式 */
.download-item {
  position: relative;
  transition: all 0.3s ease;
  background-color: var(--bg-secondary) !important;
}

:root.dark-theme .download-item.completed {
  background-color: #0d2818 !important;
  border-left: 3px solid var(--success-color);
}

:root.light-theme .download-item.completed {
  background-color: #f6ffed !important;
  border-left: 3px solid var(--success-color);
}

:root.dark-theme .download-item.downloading {
  background-color: #0a1929 !important;
  border-left: 3px solid var(--info-color);
}

:root.light-theme .download-item.downloading {
  background-color: #e6f7ff !important;
  border-left: 3px solid var(--info-color);
}

.download-item.pending {
  background-color: var(--bg-secondary) !important;
  border-left: 3px solid var(--border-secondary);
}

.status-icon {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  transition: all 0.3s ease;
}

:root.dark-theme .status-icon.completed {
  background-color: #1e3a28;
  color: var(--success-color);
}

:root.light-theme .status-icon.completed {
  background-color: #f6ffed;
  color: var(--success-color);
}

:root.dark-theme .status-icon.downloading {
  background-color: #1e2a3a;
  color: var(--info-color);
}

:root.light-theme .status-icon.downloading {
  background-color: #e6f7ff;
  color: var(--info-color);
}

.status-icon.pending {
  background-color: var(--bg-tertiary);
  color: var(--text-secondary);
}

.file-info {
  flex: 1;
}

.file-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary) !important;
  margin-bottom: 4px;
  line-height: 1.4;
}

.file-status {
  font-size: 12px;
  color: var(--text-secondary) !important;
  margin: 0;
}

.download-details {
  margin-top: 8px;
}

.weui-progress--file {
  margin-bottom: 6px;
}

.weui-progress--file .weui-progress__bar {
  height: 3px;
  background-color: var(--border-color);
}

.downloading-bar {
  background: linear-gradient(90deg, var(--info-color) 0%, var(--info-color) 100%);
  transition: width 0.5s ease;
}

.progress-text {
  font-weight: 600;
  color: var(--info-color);
}

.speed-text {
  color: var(--text-secondary);
  margin-left: 8px;
}

.completion-badge {
  background-color: var(--success-color);
  color: #ffffff;
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
}

.queue-number {
  background-color: var(--bg-tertiary);
  color: var(--text-secondary);
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 500;
}

.download-status {
  background-color: var(--bg-tertiary) !important;
  border-radius: 8px;
  margin-top: 8px;
  transition: background-color 0.3s ease;
}

.status-info {
  padding: 4px 0;
}

.status-text {
  font-size: 13px;
  color: var(--text-primary) !important;
  margin-bottom: 2px;
}

.last-activity {
  font-size: 11px;
  color: var(--text-secondary) !important;
  margin: 0;
}

/* 新UI样式 */
/* 导航栏 */
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px 20px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
}

.navbar-left {
  width: 40px;
}

.app-icon {
  width: 24px;
  height: 24px;
  background: #ff6b35;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 14px;
}

.navbar-center {
  flex: 1;
  text-align: center;
}

.app-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.navbar-right {
  width: 40px;
  display: flex;
  justify-content: flex-end;
}

.settings-btn {
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 20px;
  cursor: pointer;
  padding: 5px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.settings-btn:hover {
  background-color: var(--bg-tertiary);
}

/* 设置面板 */
.settings-panel {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
}

.settings-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
}

.settings-content {
  position: absolute;
  top: 0;
  right: 0;
  width: 300px;
  height: 100%;
  background: var(--bg-primary);
  box-shadow: -2px 0 10px rgba(0, 0, 0, 0.1);
  animation: slideInRight 0.3s ease;
}

@keyframes slideInRight {
  from { transform: translateX(100%); }
  to { transform: translateX(0); }
}

.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid var(--border-color);
}

.settings-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.close-btn:hover {
  background-color: var(--bg-tertiary);
}

.settings-body {
  padding: 20px;
}

.setting-item {
  margin-bottom: 30px;
}

.setting-item h4 {
  margin: 0 0 10px 0;
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 600;
}

.setting-btn {
  background: #ff6b35;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.setting-btn:hover {
  background: #e55a2b;
}

.setting-btn.secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.setting-btn.secondary:hover {
  background: var(--border-secondary);
}

.setting-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 14px;
  box-sizing: border-box;
}

.login-status {
  color: var(--success-color);
  font-size: 14px;
  margin-right: 10px;
}

/* 登录弹窗 */
.login-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
}

.modal-content {
  position: relative;
  background: var(--bg-primary);
  border-radius: 12px;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  animation: fadeInScale 0.3s ease;
}

@keyframes fadeInScale {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
}

.modal-body {
  padding: 20px;
}

.loading {
  padding: 60px 0;
  color: var(--text-secondary);
}

.modal-actions {
  display: flex;
  gap: 10px;
}

.btn {
  flex: 1;
  padding: 10px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s;
}

.btn.primary {
  background: #ff6b35;
  color: white;
}

.btn.primary:hover {
  background: #e55a2b;
}

.btn.secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.btn.secondary:hover {
  background: var(--border-secondary);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 主要内容区域 */
.main-content {
  padding: 20px;
  max-width: 600px;
  margin: 0 auto;
}

/* 输入区域 */
.input-section {
  margin-bottom: 20px;
}

.input-group {
  margin-bottom: 15px;
}

.input-label {
  display: block;
  margin-bottom: 5px;
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}

.input-field {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 16px;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.input-field:focus {
  outline: none;
  border-color: #ff6b35;
}

/* 检查结果区域 */
.check-result {
  background: var(--bg-secondary);
  border: 2px solid #ff6b35;
  border-radius: 12px;
  margin-bottom: 20px;
  overflow: hidden;
  position: relative;
}

.result-content {
  display: flex;
  align-items: center;
  padding: 16px;
}

.result-icon {
  width: 40px;
  height: 40px;
  background: #34c759;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
  margin-right: 12px;
}

.result-info {
  flex: 1;
}

.result-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.result-desc {
  font-size: 14px;
  color: var(--text-secondary);
}

.result-actions {
  margin-left: 12px;
}

.btn-stop {
  background: #ff3b30;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-stop:hover {
  background: #d70015;
}

/* 橘色边框进度条 */
.progress-border {
  height: 4px;
  background: rgba(255, 107, 53, 0.2);
  position: relative;
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: #ff6b35;
  transition: width 0.3s ease;
  border-radius: 2px;
}

/* 检查结果区域进度指示 */
.check-result.downloading {
  border: 2px solid #ff6b35;
  position: relative;
  overflow: hidden;
}

.check-result.downloading::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  background: linear-gradient(90deg, rgba(255, 107, 53, 0.1) 0%, transparent 100%);
  width: var(--progress-width, 0%);
  transition: width 0.3s ease;
  pointer-events: none;
}

/* 按钮区域 */
.button-section {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.button-section .btn {
  flex: 1;
  padding: 14px;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.button-section .btn.primary {
  background: #ff6b35;
  color: white;
}

.button-section .btn.primary:hover:not(.disabled) {
  background: #e55a2b;
}

.button-section .btn.secondary {
  background: var(--bg-tertiary);
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.button-section .btn.secondary:hover:not(.disabled) {
  background: var(--border-secondary);
}

.button-section .btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* WebSocket状态 */
.ws-status {
  background: var(--bg-secondary);
  border-radius: 8px;
  padding: 12px;
  border: 1px solid var(--border-color);
}

.status-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 6px;
}

.status-item:last-child {
  margin-bottom: 0;
}

.status-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.status-value {
  font-size: 14px;
  color: var(--text-primary);
  font-weight: 500;
}

/* 调试信息 */
.debug-info {
  background: var(--bg-tertiary);
  padding: 10px;
  margin: 10px 0;
  font-size: 12px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

/* 响应式设计 */
@media (max-width: 480px) {
  .main-content {
    padding: 15px;
  }

  .navbar {
    padding: 12px 15px;
  }

  .settings-content {
    width: 280px;
  }

  .button-section {
    flex-direction: column;
  }

  .button-section .btn {
    margin-bottom: 8px;
  }
}

</style>
