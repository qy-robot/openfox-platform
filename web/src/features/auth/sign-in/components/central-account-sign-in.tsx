/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { useCallback, useEffect, useMemo, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { useStatus } from '@/hooks/use-status'
import { getServerErrorMessage } from '@/lib/server-error-message'

import {
  clearCentralSignedOut,
  consumeFreshCentralLoginRequest,
  isCentralSignedOut,
  requestFreshCentralLogin,
} from '../central-reauth'
import {
  cancelCentralSSO,
  finishCentralSSOResult,
  startCentralSSO,
} from '../central-sso'
import { UserAuthForm } from './user-auth-form'

const ACCOUNT_REQUEST_TIMEOUT_MS = 15_000

type AccountResponse = {
  success?: boolean
  code?: string
  message?: string
  data?: unknown
}

function accountCenterURL(value: unknown): URL {
  const url = new URL(
    typeof value === 'string' && value ? value : 'https://account.openfox.work'
  )
  const localHTTP =
    url.protocol === 'http:' &&
    (url.hostname === 'localhost' || url.hostname === '127.0.0.1')
  if (url.protocol !== 'https:' && !localHTTP) {
    throw new Error('The account service URL is invalid.')
  }
  return url
}

async function accountFetch(
  input: string,
  init: RequestInit,
  signal: AbortSignal
): Promise<AccountResponse> {
  signal.throwIfAborted()
  const controller = new AbortController()
  const abort = () => controller.abort(signal.reason)
  signal.addEventListener('abort', abort, { once: true })
  const timeout = window.setTimeout(
    () =>
      controller.abort(new DOMException('Request timed out', 'TimeoutError')),
    ACCOUNT_REQUEST_TIMEOUT_MS
  )
  try {
    const response = await fetch(input, {
      ...init,
      credentials: 'include',
      redirect: 'error',
      signal: controller.signal,
    })
    const body = (await response.json()) as AccountResponse
    signal.throwIfAborted()
    if (!response.ok && body.data === undefined) {
      const error = new Error(body.message || 'Account sign in failed.')
      if (body.code) error.name = body.code
      throw error
    }
    return body
  } finally {
    window.clearTimeout(timeout)
    signal.removeEventListener('abort', abort)
  }
}

export function CentralAccountSignIn(props: { redirectTo?: string }) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const silentOperation = useRef<AbortController | null>(null)
  const returnTo = props.redirectTo || '/dashboard'
  const accountURL = useMemo(
    () => accountCenterURL(status?.account_center_url),
    [status?.account_center_url]
  )

  const authorize = useCallback(
    async (
      signal: AbortSignal,
      options: { prompt?: 'none'; reauthenticate?: boolean } = {}
    ) => {
      const authorizationUrl = await startCentralSSO(returnTo, {
        ...options,
        signal,
      })
      signal.throwIfAborted()
      const parsedAuthorizationURL = new URL(authorizationUrl)
      if (
        parsedAuthorizationURL.origin !== accountURL.origin ||
        parsedAuthorizationURL.pathname !== '/v1/oauth/authorize' ||
        parsedAuthorizationURL.searchParams.get('response_mode') !== 'json'
      ) {
        cancelCentralSSO()
        throw new Error('The account authorization URL is invalid.')
      }
      const authorization = await accountFetch(
        parsedAuthorizationURL.toString(),
        { method: 'GET', headers: { Accept: 'application/json' } },
        signal
      )
      signal.throwIfAborted()
      const result = await finishCentralSSOResult(authorization.data, signal)
      signal.throwIfAborted()
      return result
    },
    [accountURL.origin, returnTo]
  )

  useEffect(() => {
    if (isCentralSignedOut()) return
    const controller = new AbortController()
    silentOperation.current = controller
    const reauthenticate = consumeFreshCentralLoginRequest()
    void authorize(
      controller.signal,
      reauthenticate ? { reauthenticate: true } : { prompt: 'none' }
    )
      .then((result) => {
        if (result.status === 'authenticated') {
          window.location.assign(result.returnTo)
        }
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted) return
        cancelCentralSSO()
        if (error instanceof Error && error.name === 'AUTH_SESSION_REVOKED' && !reauthenticate) {
          requestFreshCentralLogin()
          window.location.assign(`/sign-in?redirect=${encodeURIComponent(returnTo)}`)
          return
        }
        toast.error(
          getServerErrorMessage(error, t('Unable to start account sign in'))
        )
      })
    return () => {
      controller.abort()
      cancelCentralSSO()
    }
  }, [authorize, t])

  async function signIn(
    credentials: { username: string; password: string },
    signal: AbortSignal
  ): Promise<void> {
    silentOperation.current?.abort()
    silentOperation.current = null
    cancelCentralSSO()
    const loginURL = new URL('/v1/auth/login', accountURL)
    const login = await accountFetch(
      loginURL.toString(),
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          identifier: credentials.username,
          password: credentials.password,
        }),
      },
      signal
    )
    if (login.success !== true) {
      throw new Error(login.message || t('Login failed'))
    }
    const result = await authorize(signal)
    if (result.status !== 'authenticated') {
      throw new Error(t('Unable to complete account sign in'))
    }
    clearCentralSignedOut()
    window.location.assign(result.returnTo)
  }

  return (
    <UserAuthForm
      redirectTo={returnTo}
      passwordOnly
      forgotPasswordUrl={new URL('/account/recovery', accountURL).toString()}
      onPasswordSubmit={signIn}
    />
  )
}
