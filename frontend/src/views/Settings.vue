<template>
  <div class="container">
    <!-- 主题设置 -->
    <div class="weui-form">
      <div class="weui-form__text-area">
        <h2 class="weui-form__title">外观设置</h2>
        <p class="weui-form__desc">选择您喜欢的界面主题</p>
      </div>

      <div class="weui-form__control-area">
        <div class="weui-cells__group">
          <div class="weui-cells">
            <!-- 跟随系统 -->
            <div class="weui-cell weui-cell_active theme-option" @click="setTheme('auto')" :class="{ 'selected': themeStore.theme === 'auto' }">
              <div class="weui-cell__hd">
                <div class="theme-icon">🔄</div>
              </div>
              <div class="weui-cell__bd">
                <p class="theme-name">跟随系统</p>
                <p class="theme-desc">根据系统设置自动切换主题</p>
                <p class="theme-current" v-if="themeStore.theme === 'auto'">
                  当前: {{ themeStore.systemTheme === 'dark' ? '暗黑模式' : '浅色模式' }}
                </p>
              </div>
              <div class="weui-cell__ft">
                <i class="weui-icon-checked" v-if="themeStore.theme === 'auto'"></i>
              </div>
            </div>

            <!-- 浅色模式 -->
            <div class="weui-cell weui-cell_active theme-option" @click="setTheme('light')" :class="{ 'selected': themeStore.theme === 'light' }">
              <div class="weui-cell__hd">
                <div class="theme-icon">☀️</div>
              </div>
              <div class="weui-cell__bd">
                <p class="theme-name">浅色模式</p>
                <p class="theme-desc">明亮清爽的界面风格</p>
              </div>
              <div class="weui-cell__ft">
                <i class="weui-icon-checked" v-if="themeStore.theme === 'light'"></i>
              </div>
            </div>

            <!-- 暗黑模式 -->
            <div class="weui-cell weui-cell_active theme-option" @click="setTheme('dark')" :class="{ 'selected': themeStore.theme === 'dark' }">
              <div class="weui-cell__hd">
                <div class="theme-icon">🌙</div>
              </div>
              <div class="weui-cell__bd">
                <p class="theme-name">暗黑模式</p>
                <p class="theme-desc">护眼舒适的深色界面</p>
              </div>
              <div class="weui-cell__ft">
                <i class="weui-icon-checked" v-if="themeStore.theme === 'dark'"></i>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 基本设置 -->
    <div class="weui-form">
      <div class="weui-form__text-area">
        <h2 class="weui-form__title">下载设置</h2>
      </div>

      <div class="weui-form__control-area">
        <div class="weui-cells__group">
          <div class="weui-cells">

            <div class="weui-cell">
              <div class="weui-cell__hd">
                <label class="weui-label">默认保存路径</label>
              </div>
              <div class="weui-cell__bd">
                <input
                  v-model="config.savePath"
                  type="text"
                  class="weui-input"
                  placeholder="留空则保存到 audiobooks 根目录"
                />
              </div>
            </div>

            <div class="weui-cell">
              <div class="weui-cell__hd">
                <label class="weui-label">音频质量</label>
              </div>
              <div class="weui-cell__bd">
                <select v-model="config.quality" class="weui-select">
                  <option value="bestaudio/best">最佳音频质量</option>
                  <option value="bestaudio[ext=m4a]/best[ext=m4a]">M4A 格式</option>
                  <option value="bestaudio[ext=mp3]/best[ext=mp3]">MP3 格式</option>
                  <option value="worst">最低质量（节省空间）</option>
                </select>
              </div>
            </div>

            <div class="weui-cell">
              <div class="weui-cell__hd">
                <label class="weui-label">文件名格式</label>
              </div>
              <div class="weui-cell__bd">
                <input
                  v-model="config.titleRegex"
                  type="text"
                  class="weui-input"
                  placeholder="%(title)s.%(ext)s"
                />
              </div>
            </div>

            <div class="weui-cell">
              <div class="weui-cell__hd">
                <label class="weui-label">重试次数</label>
              </div>
              <div class="weui-cell__bd">
                <input
                  v-model.number="config.retryCount"
                  type="number"
                  class="weui-input"
                  min="1"
                  max="10"
                />
              </div>
            </div>

            <div class="weui-cell weui-cell_switch">
              <div class="weui-cell__bd">下载缩略图</div>
              <div class="weui-cell__ft">
                <input
                  v-model="config.writeThumbnail"
                  type="checkbox"
                  class="weui-switch"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 版本管理 -->
    <div class="weui-form">
      <div class="weui-form__text-area">
        <h2 class="weui-form__title">版本管理</h2>
      </div>

      <div class="weui-form__control-area">
        <div class="weui-cells__group">
          <div class="weui-cells">
            <div class="weui-cell">
              <div class="weui-cell__bd">
                <p>当前 yt-dlp 版本</p>
              </div>
              <div class="weui-cell__ft">
                <span class="version-text">{{ versionInfo.currentVersion || '检测中...' }}</span>
              </div>
            </div>

            <div class="weui-cell" v-if="versionInfo.latestVersion">
              <div class="weui-cell__bd">
                <p>最新版本</p>
              </div>
              <div class="weui-cell__ft">
                <span class="version-text">{{ versionInfo.latestVersion }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="weui-form__opr-area">
        <a @click="checkVersion" class="weui-btn weui-btn_default" :class="{ 'weui-btn_disabled': isCheckingVersion }">
          {{ isCheckingVersion ? '检查中...' : '检查更新' }}
        </a>

        <a @click="updateVersion" class="weui-btn weui-btn_primary" v-if="versionInfo.hasUpdate" :class="{ 'weui-btn_disabled': isUpdating }">
          {{ isUpdating ? '更新中...' : '更新到最新版本' }}
        </a>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="weui-form">
      <div class="weui-form__opr-area">
        <a @click="saveSettings" class="weui-btn weui-btn_primary" :class="{ 'weui-btn_disabled': isSaving }">
          {{ isSaving ? '保存中...' : '保存设置' }}
        </a>

        <a @click="resetSettings" class="weui-btn weui-btn_default">
          重置为默认设置
        </a>
      </div>
    </div>

    <!-- 状态消息 -->
    <div v-if="message" class="weui-msg" :class="messageType">
      <div class="weui-msg__icon-area">
        <i class="weui-icon-success weui-icon_msg" v-if="messageType === 'success'"></i>
        <i class="weui-icon-warn weui-icon_msg" v-if="messageType === 'error'"></i>
      </div>
      <div class="weui-msg__text-area">
        <p class="weui-msg__desc">{{ message }}</p>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useThemeStore } from '../stores/theme'

export default {
  name: 'Settings',
  setup() {
    const themeStore = useThemeStore()
    const config = ref({
      savePath: '',
      titleRegex: '%(title)s.%(ext)s',
      quality: 'bestaudio/best',
      retryCount: 5,
      writeThumbnail: true
    })

    const versionInfo = ref({
      currentVersion: '',
      latestVersion: '',
      hasUpdate: false
    })

    const isSaving = ref(false)
    const isCheckingVersion = ref(false)
    const isUpdating = ref(false)
    const message = ref('')
    const messageType = ref('success')

    // 加载配置
    const loadConfig = async () => {
      try {
        const response = await fetch('/api/config')
        if (response.ok) {
          const data = await response.json()
          config.value = { ...config.value, ...data }
        }
      } catch (error) {
        console.error('加载配置失败:', error)
        showMessage('加载配置失败', 'error')
      }
    }

    // 保存设置
    const saveSettings = async () => {
      isSaving.value = true
      try {
        const response = await fetch('/api/config', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(config.value)
        })

        if (response.ok) {
          showMessage('设置已保存', 'success')
        } else {
          showMessage('保存设置失败', 'error')
        }
      } catch (error) {
        console.error('保存设置失败:', error)
        showMessage('保存设置失败', 'error')
      } finally {
        isSaving.value = false
      }
    }

    // 重置设置
    const resetSettings = () => {
      config.value = {
        savePath: '',
        titleRegex: '%(title)s.%(ext)s',
        quality: 'bestaudio/best',
        retryCount: 5,
        writeThumbnail: true
      }
      showMessage('设置已重置', 'success')
    }

    // 检查版本
    const checkVersion = async () => {
      isCheckingVersion.value = true
      try {
        const response = await fetch('/api/version/check')
        if (response.ok) {
          const data = await response.json()
          versionInfo.value = data

          if (data.hasUpdate) {
            showMessage('发现新版本可用', 'success')
          } else {
            showMessage('已是最新版本', 'success')
          }
        }
      } catch (error) {
        console.error('检查版本失败:', error)
        showMessage('检查版本失败', 'error')
      } finally {
        isCheckingVersion.value = false
      }
    }

    // 更新版本
    const updateVersion = async () => {
      isUpdating.value = true
      try {
        const response = await fetch('/api/version/update', {
          method: 'POST'
        })

        if (response.ok) {
          showMessage('更新任务已启动，请稍候...', 'success')
          // 延迟检查更新结果
          setTimeout(checkVersion, 10000)
        } else {
          showMessage('启动更新失败', 'error')
        }
      } catch (error) {
        console.error('更新版本失败:', error)
        showMessage('更新版本失败', 'error')
      } finally {
        isUpdating.value = false
      }
    }

    // 设置主题
    const setTheme = (theme) => {
      themeStore.setTheme(theme)
    }

    // 显示消息
    const showMessage = (text, type = 'success') => {
      message.value = text
      messageType.value = type
      setTimeout(() => {
        message.value = ''
      }, 3000)
    }

    onMounted(() => {
      loadConfig()
      checkVersion()
    })

    return {
      themeStore,
      config,
      versionInfo,
      isSaving,
      isCheckingVersion,
      isUpdating,
      message,
      messageType,
      setTheme,
      saveSettings,
      resetSettings,
      checkVersion,
      updateVersion
    }
  }
}
</script>

<style scoped>
.container {
  background-color: var(--bg-primary);
  min-height: 100vh;
  color: var(--text-primary);
  transition: all 0.3s ease;
}

/* 主题选项样式 */
.theme-option {
  transition: all 0.3s ease;
  cursor: pointer;
  background-color: var(--bg-secondary) !important;
  border-bottom: 1px solid var(--border-color) !important;
}

.theme-option:hover {
  background-color: var(--bg-tertiary) !important;
}

.theme-option.selected {
  background-color: var(--bg-tertiary) !important;
  border-left: 3px solid var(--accent-color) !important;
}

.theme-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: var(--bg-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  transition: all 0.3s ease;
  border: 2px solid var(--border-color);
}

.theme-option.selected .theme-icon {
  background-color: var(--accent-color);
  border-color: var(--accent-color);
  color: #ffffff;
  transform: scale(1.05);
}

.theme-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.theme-desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
}

.theme-current {
  font-size: 12px;
  color: var(--accent-color);
  margin: 4px 0 0 0;
  font-weight: 500;
}

.weui-icon-checked::before {
  content: '✓';
  color: var(--accent-color);
  font-weight: bold;
  font-size: 16px;
}

/* WeUI 组件样式覆盖 */
.weui-form {
  background-color: var(--bg-primary) !important;
}

.weui-form__text-area {
  background-color: var(--bg-primary) !important;
}

.weui-form__title {
  color: var(--text-primary) !important;
}

.weui-form__desc {
  color: var(--text-secondary) !important;
}

.weui-cells__group {
  background-color: var(--bg-primary) !important;
}

.weui-cells {
  background-color: var(--bg-secondary) !important;
  border-radius: 10px !important;
}

.weui-cell {
  background-color: var(--bg-secondary) !important;
  border-bottom: 1px solid var(--border-color) !important;
  color: var(--text-primary) !important;
}

.weui-cell:last-child {
  border-bottom: none !important;
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

.weui-select {
  background-color: var(--bg-secondary) !important;
  color: var(--text-primary) !important;
  border: none !important;
  width: 100%;
}

.weui-select option {
  background-color: var(--bg-secondary);
  color: var(--text-primary);
}

.weui-switch {
  background-color: var(--border-color) !important;
}

.weui-switch:checked {
  background-color: var(--accent-color) !important;
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

.version-text {
  color: var(--text-secondary);
  font-size: 14px;
}

/* 版本信息样式 */
.version-info {
  background: var(--bg-secondary);
  padding: 1.5rem;
  border-radius: 8px;
  transition: background-color 0.3s ease;
}

.version-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.version-label {
  font-weight: 500;
  color: var(--text-primary);
}

.version-value {
  color: var(--text-secondary);
}

.version-actions {
  margin-top: 1rem;
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.settings-actions {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  margin-top: 2rem;
}

.message {
  margin-top: 1rem;
  padding: 0.75rem;
  border-radius: 8px;
  font-weight: 500;
}

.message.success {
  background: var(--success-color);
  color: #ffffff;
  opacity: 0.9;
}

.message.error {
  background: var(--warning-color);
  color: #ffffff;
  opacity: 0.9;
}

/* 表单间距和布局 */
.weui-form {
  margin-bottom: 24px;
}

.weui-form + .weui-form {
  margin-top: 32px;
}

.weui-form__text-area {
  padding: 20px 16px 16px 16px !important;
}

.weui-form__control-area {
  padding: 0 16px !important;
}

.weui-form__opr-area {
  padding: 16px !important;
  gap: 12px;
  display: flex;
  flex-direction: column;
}

.weui-form__opr-area .weui-btn {
  margin-bottom: 8px;
}

.weui-form__opr-area .weui-btn:last-child {
  margin-bottom: 0;
}

@media (max-width: 768px) {
  .version-item {
    flex-direction: column;
    gap: 0.25rem;
  }

  .version-actions,
  .settings-actions {
    flex-direction: column;
  }

  .btn {
    width: 100%;
  }
}
</style>
