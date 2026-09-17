/* Copyright (C) 2023-2026 QuantumNous; licensed under GNU AGPLv3 or later. */
import { api } from '@/lib/api'
import { requireServerSuccess } from '@/lib/server-error-message'

export interface DesktopAuthorization {
  user_code: string
  device_name: string
  expires_at: number
}

export async function getDesktopAuthorization(userCode: string) {
  const response = await api.get('/api/desktop/device/authorization', {
    params: { user_code: userCode },
    skipErrorHandler: true,
  })
  return requireServerSuccess(response.data).data as DesktopAuthorization
}

export async function decideDesktopAuthorization(
  userCode: string,
  approve: boolean
) {
  const response = await api.post(
    '/api/desktop/device/authorization',
    { user_code: userCode, approve },
    { singleUseAuthorization: true }
  )
  return requireServerSuccess(response.data).data
}
