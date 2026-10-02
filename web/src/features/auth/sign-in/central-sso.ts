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

// Effect re-entry may await the same exchange, but settled codes remain single-use.
let callbackFlight: {
  code: string
  state: string
  promise: Promise<string>
} | null = null

export type CentralSSOOptions = {
  reauthenticate?: boolean
  signal?: AbortSignal
}

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
  options.signal?.throwIfAborted()
  const verifier = randomValue(48)
  const state = randomValue(32)
  const digest = await crypto.subtle.digest(
    'SHA-256',
    new TextEncoder().encode(verifier)
  )
  const codeChallenge = base64Url(new Uint8Array(digest))
  options.signal?.throwIfAborted()
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
      ...(options.reauthenticate ? { reauth: true } : {}),
    },
    skipAuthRefresh: true,
    signal: options.signal,
  })
  options.signal?.throwIfAborted()
  return response.data.data.authorizationUrl as string
}

export async function finishCentralSSO(
  code: string,
  state: string,
  signal?: AbortSignal
): Promise<string> {
  signal?.throwIfAborted()
  if (callbackFlight?.code === code && callbackFlight.state === state) {
    return callbackFlight.promise
  }
  const promise = exchangeCentralCallback(code, state, signal)
  callbackFlight = { code, state, promise }
  try {
    return await promise
  } finally {
    if (callbackFlight?.promise === promise) callbackFlight = null
  }
}

async function exchangeCentralCallback(
  code: string,
  state: string,
  signal?: AbortSignal
): Promise<string> {
  signal?.throwIfAborted()
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
    { skipAuthRefresh: true, singleUseAuthorization: true, signal }
  )
  signal?.throwIfAborted()
  const bundle: unknown = response.data.data
  if (!isAuthBundle(bundle)) {
    throw new Error('The account service returned an invalid sign-in response.')
  }
  applyAuthBundle(bundle)
  clearCentralReauthenticationAttempt()
  return resolveCentralSSOReturnTo(pending.returnTo)
}
