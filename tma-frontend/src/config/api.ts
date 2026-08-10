/// <reference types="vite/client" />

function resolveApiBaseUrl() {
  const configured = (import.meta as any).env.VITE_API_URL as string | undefined
  if (configured && !configured.includes('localhost') && !configured.includes('127.0.0.1')) {
    return configured
  }
  return '/api/v1'
}

export const API_BASE_URL = resolveApiBaseUrl()
export const API_URL = API_BASE_URL
export const BOT_USERNAME = (import.meta as any).env.VITE_BOT_USERNAME || 'coinmintshop'
