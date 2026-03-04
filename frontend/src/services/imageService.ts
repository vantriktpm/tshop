import { httpClient, env } from '@/core'

export interface DownloadUrlResponse {
  download_url: string
}

/** Lấy URL tải ảnh avatar user theo userId (qua gateway -> image-service). */
export async function getUserAvatarUrl(userId: string): Promise<DownloadUrlResponse> {
  return httpClient.get<DownloadUrlResponse, any, {}>(
    `/images/${userId}/download-url`,
    { baseURL: env.imageBaseUrl },
  )
}

/** Lấy URL tải ảnh sản phẩm theo imageId (tùy backend map, dùng chung endpoint). */
export async function getProductImageUrl(imageId: string): Promise<DownloadUrlResponse> {
  return httpClient.get<DownloadUrlResponse, any, {}>(
    `/images/${imageId}/download-url`,
    { baseURL: env.imageBaseUrl },
  )
}

