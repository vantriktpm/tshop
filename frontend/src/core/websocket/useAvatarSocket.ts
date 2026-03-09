import { ref, watch, type Ref } from 'vue'
import { env } from '@/core/config/env'
import { httpClient } from '@/core/http'

/** URL WebSocket gateway (cùng host với API, path /ws) */
function getWsBaseUrl(): string {
  const base = env.oauthBaseUrl || (typeof window !== 'undefined' ? window.location.origin : '')
  return base.replace(/^http/, 'ws')
}

export interface AvatarSavedMessage {
  user_id: string
  image_id: string
}

/**
 * Kết nối WebSocket /ws?user_id=xxx, khi nhận avatar.saved (user_id, image_id)
 * thì gọi GET /api/images/:image_id/download-url và cập nhật avatarDownloadUrl.
 */
export function useAvatarSocket(userId: Ref<string | undefined>) {
  const avatarDownloadUrl = ref<string | null>(null)
  let ws: WebSocket | null = null

  function connect(uid: string) {
    const wsUrl = `${getWsBaseUrl()}/ws?user_id=${encodeURIComponent(uid)}`
    try {
      ws = new WebSocket(wsUrl)
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data as string) as AvatarSavedMessage
          if (data.image_id && data.user_id === uid) {
            fetchDownloadUrl(data.image_id)
          }
        } catch {
          // ignore parse error
        }
      }
      ws.onclose = () => {
        ws = null
      }
      ws.onerror = () => {
        ws = null
      }
    } catch {
      ws = null
    }
  }

  async function fetchDownloadUrl(imageId: string) {
    try {
      const res = await httpClient.get<{ download_url: string }>(
        `/images/${imageId}/download-url`
      ) as unknown as { download_url: string }
      if (res?.download_url) {
        avatarDownloadUrl.value = res.download_url
      }
    } catch {
      avatarDownloadUrl.value = null
    }
  }

  function disconnect() {
    if (ws) {
      ws.close()
      ws = null
    }
    avatarDownloadUrl.value = null
  }

  watch(
    userId,
    (uid) => {
      disconnect()
      if (uid) {
        connect(uid)
      }
    },
    { immediate: true }
  )

  function clearAvatar() {
    avatarDownloadUrl.value = null
  }

  return { avatarDownloadUrl, disconnect, clearAvatar }
}
