/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { sanitizeAuthRedirect } from '@/features/auth/lib/auth-redirect'
import { api, applyAuthBundle, isAuthBundle } from '@/lib/api'

import { clearCentralReauthenticationAttempt } from './central-reauth'

const SSO_STORAGE_KEY = 'robocoding:account-sso'

type PendingSSO = {
  verifier: string
  state: string
  returnTo: string
}

export type CentralSSOOptions = {
  prompt?: 'none'
  reauthenticate?: boolean
}

export type CentralSSOMessageResult =
  | { status: 'ignored' }
  | { status: 'authenticated'; returnTo: string }
  | { status: 'interaction_required'; error: string }

export function resolveCentralSSOReturnTo(value: unknown): string {
  return sanitizeAuthRedirect(value, window.location.origin) ?? '/dashboard'
}

function base64Url(bytes: Uint8Array): string {
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary)
    .replaceAll('+', '-')
    .replaceAll('/', '_')
    .replace(/=+$/, '')
}

function randomValue(size: number): string {
  const bytes = new Uint8Array(size)
  crypto.getRandomValues(bytes)
  return base64Url(bytes)
}

export async function startCentralSSO(
  returnTo: string,
  options: CentralSSOOptions = {}
): Promise<string> {
  const verifier = randomValue(48)
  const state = randomValue(32)
  const digest = await crypto.subtle.digest(
    'SHA-256',
    new TextEncoder().encode(verifier)
  )
  const codeChallenge = base64Url(new Uint8Array(digest))
  const safeReturnTo = resolveCentralSSOReturnTo(returnTo)
  const redirectUri = `${window.location.origin}/account/callback`
  sessionStorage.setItem(
    SSO_STORAGE_KEY,
    JSON.stringify({
      verifier,
      state,
      returnTo: safeReturnTo,
    } satisfies PendingSSO)
  )
  const response = await api.get('/api/account/sso/start', {
    params: {
      codeChallenge,
      state,
      redirectUri,
      responseMode: 'web_message',
      ...(options.prompt ? { prompt: options.prompt } : {}),
      ...(options.reauthenticate ? { reauth: true } : {}),
    },
    skipAuthRefresh: true,
  })
  return response.data.data.authorizationUrl as string
}

export async function finishCentralSSOMessage(
  event: MessageEvent,
  iframeWindow: Window,
  authorizationUrl: string
): Promise<CentralSSOMessageResult> {
  const expectedOrigin = new URL(authorizationUrl, window.location.origin)
    .origin
  if (event.origin !== expectedOrigin || event.source !== iframeWindow) {
    return { status: 'ignored' }
  }
  const data = event.data
  if (
    typeof data !== 'object' ||
    data === null ||
    data.type !== 'robocoding.oauth.result'
  ) {
    return { status: 'ignored' }
  }
  if (typeof data.state !== 'string') {
    throw new Error('The account sign-in message was invalid.')
  }
  const serialized = sessionStorage.getItem(SSO_STORAGE_KEY)
  if (!serialized) {
    throw new Error('The sign-in request is missing or expired.')
  }
  const pending = JSON.parse(serialized) as PendingSSO
  if (pending.state !== data.state) {
    throw new Error('The sign-in state did not match.')
  }
  if (typeof data.error === 'string') {
    sessionStorage.removeItem(SSO_STORAGE_KEY)
    if (
      data.error === 'login_required' ||
      data.error === 'interaction_required'
    ) {
      return { status: 'interaction_required', error: data.error }
    }
    throw new Error(`Account sign in failed: ${data.error}`)
  }
  if (typeof data.code !== 'string' || data.code.length === 0) {
    throw new Error('The account sign-in message did not contain a code.')
  }
  return {
    status: 'authenticated',
    returnTo: await finishCentralSSO(data.code, data.state),
  }
}

export function cancelCentralSSO(): void {
  sessionStorage.removeItem(SSO_STORAGE_KEY)
}

export async function finishCentralSSO(
  code: string,
  state: string
): Promise<string> {
  const serialized = sessionStorage.getItem(SSO_STORAGE_KEY)
  if (!serialized) {
    throw new Error('The sign-in request is missing or expired.')
  }
  const pending = JSON.parse(serialized) as PendingSSO
  if (pending.state !== state) {
    throw new Error('The sign-in state did not match.')
  }
  sessionStorage.removeItem(SSO_STORAGE_KEY)
  const redirectUri = `${window.location.origin}/account/callback`
  const response = await api.post(
    '/api/account/sso/exchange',
    {
      code,
      codeVerifier: pending.verifier,
      redirectUri,
    },
    { skipAuthRefresh: true, singleUseAuthorization: true }
  )
  const bundle: unknown = response.data.data
  if (!isAuthBundle(bundle)) {
    throw new Error('The account service returned an invalid sign-in response.')
  }
  applyAuthBundle(bundle)
  clearCentralReauthenticationAttempt()
  return resolveCentralSSOReturnTo(pending.returnTo)
}
