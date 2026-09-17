/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { sanitizeAuthRedirect } from '@/features/auth/lib/auth-redirect'

const REAUTH_ATTEMPT_KEY = 'robocoding:central-reauth-attempt'
const FRESH_LOGIN_KEY = 'robocoding:central-fresh-login'
const SIGNED_OUT_KEY = 'robocoding:central-signed-out'
const REAUTH_LOOP_WINDOW_MS = 60_000

export function isCentralAccountMode(): boolean {
  try {
    const status = JSON.parse(
      window.localStorage.getItem('status') ?? 'null'
    ) as Record<string, unknown> | null
    return status?.account_auth_enabled === true
  } catch {
    return false
  }
}

export function currentCentralReturnTo(): string {
  const target = `${window.location.pathname}${window.location.search}${window.location.hash}`
  return sanitizeAuthRedirect(target, window.location.origin) ?? '/dashboard'
}

export function reserveCentralReauthentication(now = Date.now()): boolean {
  try {
    const stored = window.sessionStorage.getItem(REAUTH_ATTEMPT_KEY)
    const previous = stored === null ? Number.NaN : Number(stored)
    if (Number.isFinite(previous) && now - previous < REAUTH_LOOP_WINDOW_MS) {
      return false
    }
    window.sessionStorage.setItem(REAUTH_ATTEMPT_KEY, String(now))
    return true
  } catch {
    return false
  }
}

export function clearCentralReauthenticationAttempt(): void {
  try {
    window.sessionStorage.removeItem(REAUTH_ATTEMPT_KEY)
  } catch {
    /* Storage may be unavailable in private mode. */
  }
}

export function requestFreshCentralLogin(): void {
  try {
    window.sessionStorage.setItem(FRESH_LOGIN_KEY, '1')
  } catch {
    /* Storage may be unavailable in private mode. */
  }
}

export function consumeFreshCentralLoginRequest(): boolean {
  try {
    const requested = window.sessionStorage.getItem(FRESH_LOGIN_KEY) === '1'
    window.sessionStorage.removeItem(FRESH_LOGIN_KEY)
    return requested
  } catch {
    return false
  }
}

export function markCentralSignedOut(): void {
  try {
    window.sessionStorage.setItem(SIGNED_OUT_KEY, '1')
  } catch {
    /* Storage may be unavailable in private mode. */
  }
}

export function markCentralSignedOutIfEnabled(): void {
  if (isCentralAccountMode()) markCentralSignedOut()
}

export function isCentralSignedOut(): boolean {
  try {
    return window.sessionStorage.getItem(SIGNED_OUT_KEY) === '1'
  } catch {
    return false
  }
}

export function clearCentralSignedOut(): void {
  try {
    window.sessionStorage.removeItem(SIGNED_OUT_KEY)
  } catch {
    /* Storage may be unavailable in private mode. */
  }
}
