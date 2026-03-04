import { httpClient, env } from '@/core'
import type { User, LoginPayload, RegisterPayload, AuthTokens } from '@/modules/auth/domain'

export interface AuthResponse {
  user: User
  tokens: AuthTokens
}

export async function login(payload: LoginPayload): Promise<AuthResponse> {
  return httpClient.post<AuthResponse, AuthResponse>('/auth/login', payload, {
    baseURL: env.userBaseUrl,
  })
}

export async function register(payload: RegisterPayload): Promise<AuthResponse> {
  return httpClient.post<AuthResponse, AuthResponse>('/auth/register', payload, {
    baseURL: env.userBaseUrl,
  })
}

