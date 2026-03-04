import { jwtDecode } from 'jwt-decode'

export interface DecodedTokenPayload {
  userID?: string
  email?: string
  secret?: string
  expire?: string
  [key: string]: unknown
}

export function decodeToken(token: string | null | undefined): DecodedTokenPayload | null {
  if (!token) return null
  try {
    return jwtDecode<DecodedTokenPayload>(token)
  } catch {
    return null
  }
}

export function isTokenExpired(token: string | null | undefined, skewSeconds = 30): boolean {
  const payload = decodeToken(token)
  if (!payload || !payload.expire) return true
  // Backend gửi expire dạng chuỗi thời gian; cố gắng parse thành Date.
  const expireTime = new Date(payload.expire).getTime()
  if (Number.isNaN(expireTime)) return true
  const nowMs = Date.now()
  return expireTime <= nowMs + skewSeconds * 1000
}

export function getValidTokenFromStorage(storageKey = 'tshop_token'): string | null {
  const token = typeof window !== 'undefined' ? localStorage.getItem(storageKey) : null
  if (!token) return null
  return isTokenExpired(token) ? null : token
}

