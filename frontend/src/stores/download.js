import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useDownloadStore = defineStore('download', () => {
  const progress = ref({
    isDownloading: false,
    currentFile: '',
    progress: 0,
    currentIndex: 0,
    totalCount: 0,
    thumbnail: '',
    speed: '',
    eta: '',
    status: '',
    lastActivity: '',
    currentTitle: '',
    completedFiles: []
  })

  const isDownloading = computed(() => progress.value.isDownloading)

  const setDownloading = (status) => {
    console.log('setDownloading called with:', status)
    progress.value.isDownloading = status
    if (!status) {
      // 重置进度
      progress.value = {
        isDownloading: false,
        currentFile: '',
        progress: 0,
        currentIndex: 0,
        totalCount: 0,
        thumbnail: '',
        speed: '',
        eta: '',
        status: '',
        lastActivity: '',
        currentTitle: '',
        completedFiles: []
      }
    }
    console.log('setDownloading result:', progress.value.isDownloading)
  }

  const updateProgress = (newProgress) => {
    console.log('updateProgress called with:', newProgress)
    progress.value = { ...progress.value, ...newProgress }
    console.log('updateProgress result:', progress.value)
  }

  return {
    isDownloading,
    progress,
    setDownloading,
    updateProgress
  }
})
