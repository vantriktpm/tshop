import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import { env } from '@/core/config/env'

const TOKEN_STORAGE_KEY = 'tshop_token'

const config: AxiosRequestConfig = {
  baseURL: env.apiBaseUrl,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
}

export const httpClient: AxiosInstance = axios.create(config)

// Request interceptor: gắn token nếu có (cho GET /api/auth sau OAuth redirect)
httpClient.interceptors.request.use(
  (config) => {
    const token = typeof localStorage !== 'undefined' ? localStorage.getItem(TOKEN_STORAGE_KEY) : null
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
httpClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    // global error handling
    return Promise.reject(error)
  }
)
