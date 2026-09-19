/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { t } from 'i18next'

import { getStatus } from '@/lib/api'

/** End the current browser's identity session before revoking the app session. */
export async function logoutCentralBrowserSession(): Promise<void> {
  const status = await getStatus()
  if (status.account_auth_enabled !== true) return
  const issuer = new URL(
    String(status.account_center_url || 'https://account.openfox.work')
  )
  const localHTTP =
    issuer.protocol === 'http:' &&
    (issuer.hostname === 'localhost' || issuer.hostname === '127.0.0.1')
  if (
    (issuer.protocol !== 'https:' && !localHTTP) ||
    issuer.username ||
    issuer.password
  ) {
    throw new Error(t('Failed to sign out session'))
  }
  // A separate fetch deliberately sends only the account host's cookie,
  // never the application's bearer token or its refresh interceptor.
  const response = await fetch(
    new URL('/v1/auth/browser-logout', issuer).toString(),
    {
      method: 'POST',
      credentials: 'include',
      redirect: 'error',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
      signal: AbortSignal.timeout(15_000),
    }
  )
  const body = (await response.json()) as { success?: boolean; code?: string }
  if (response.ok && body.success === true) return
  if (
    response.status === 401 &&
    (body.code === 'AUTH_REQUIRED' || body.code === 'AUTH_SESSION_INVALID')
  ) {
    return
  }
  throw new Error(t('Failed to sign out session'))
}
