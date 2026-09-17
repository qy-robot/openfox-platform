/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { AxiosError, type AxiosAdapter } from 'axios'
import { afterEach, expect, it, vi } from 'vitest'
import { toast } from 'sonner'
import { handleServerError } from '@/lib/handle-server-error'
import { createAppQueryClient } from '@/lib/query-client'

import { consumeFreshCentralLoginRequest } from '@/features/auth/sign-in/central-reauth'
import { beginExplicitSignOut, finishExplicitSignOut } from '@/lib/auth-session'
import { api } from '@/lib/http-client'
import { useAuthStore } from '@/stores/auth-store'

const originalAdapter = api.defaults.adapter

afterEach(() => {
  finishExplicitSignOut()
  api.defaults.adapter = originalAdapter
  localStorage.removeItem('status')
  sessionStorage.clear()
  window.history.replaceState({}, '', '/')
  useAuthStore.getState().auth.reset()
  vi.clearAllMocks()
})

it.each(['during sign out', 'on the sign-in page'])(
  'does not restart SSO for a protected response received %s',
  async (stage) => {
    localStorage.setItem(
      'status',
      JSON.stringify({ account_auth_enabled: true })
    )
    api.defaults.adapter = async (config) => {
      throw new AxiosError(
        'Unauthorized',
        'ERR_BAD_REQUEST',
        config,
        undefined,
        {
          data: { code: 'AUTH_SESSION_REVOKED' },
          status: 401,
          statusText: 'Unauthorized',
          headers: {},
          config,
        }
      )
    }
    if (stage === 'during sign out') beginExplicitSignOut()
    else window.history.replaceState({}, '', '/sign-in')

    await expect(api.get('/api/workbench/access')).rejects.toBeInstanceOf(
      AxiosError
    )

    expect(
      sessionStorage.getItem('robocoding:central-reauth-attempt')
    ).toBeNull()
  }
)

it('routes to local sign-in and never replays an unauthorized mutation', async () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))
  useAuthStore.getState().auth.setBundle({
    access_token: 'expired-central-access',
    token_type: 'Bearer',
    access_expires_at: 1,
    user: { id: 1, username: 'central-user', role: 1 },
    session: {
      sid: 'central-session',
      current: true,
      login_method: 'central',
      ip: '',
      user_agent: '',
      created_at: 1,
      last_active_at: 1,
      expires_at: 2_000_000_000,
    },
  })
  const adapter = vi.fn<AxiosAdapter>(async (config) => {
    throw new AxiosError('Unauthorized', 'ERR_BAD_REQUEST', config, undefined, {
      data: { code: 'AUTH_SESSION_REVOKED' },
      status: 401,
      statusText: 'Unauthorized',
      headers: {},
      config,
    })
  })
  api.defaults.adapter = adapter
  window.history.replaceState({}, '', '/workbench/skills?draft=1')

  await expect(
    api.post('/api/workbench/skills/example/publish', { revision: 2 })
  ).rejects.toBeInstanceOf(AxiosError)

  expect(adapter).toHaveBeenCalledTimes(1)
  expect(
    sessionStorage.getItem('robocoding:central-reauth-attempt')
  ).not.toBeNull()
  expect(consumeFreshCentralLoginRequest()).toBe(false)
  expect(useAuthStore.getState().auth.accessToken).toBeNull()
})

it('rate limits repeated local sign-in redirects', async () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))
  const adapter = vi.fn<AxiosAdapter>(async (config) => {
    throw new AxiosError('Unauthorized', 'ERR_BAD_REQUEST', config, undefined, {
      data: { code: 'AUTH_SESSION_REVOKED' },
      status: 401,
      statusText: 'Unauthorized',
      headers: {},
      config,
    })
  })
  api.defaults.adapter = adapter

  await expect(api.post('/api/first-sensitive-action')).rejects.toBeInstanceOf(
    AxiosError
  )
  await expect(api.post('/api/second-sensitive-action')).rejects.toBeInstanceOf(
    AxiosError
  )

  expect(adapter).toHaveBeenCalledTimes(2)
  expect(
    sessionStorage.getItem('robocoding:central-reauth-attempt')
  ).not.toBeNull()
  expect(consumeFreshCentralLoginRequest()).toBe(false)
})

it('forces a fresh account login only when the backend requires reauthentication', async () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))
  const adapter = vi.fn<AxiosAdapter>(async (config) => {
    throw new AxiosError('Unauthorized', 'ERR_BAD_REQUEST', config, undefined, {
      data: { code: 'AUTH_REAUTH_REQUIRED' },
      status: 401,
      statusText: 'Unauthorized',
      headers: {},
      config,
    })
  })
  api.defaults.adapter = adapter

  await expect(
    api.post('/api/channel/1/key', { proof: 'single-use-proof' })
  ).rejects.toBeInstanceOf(AxiosError)

  expect(adapter).toHaveBeenCalledTimes(1)
  expect(consumeFreshCentralLoginRequest()).toBe(true)
  expect(consumeFreshCentralLoginRequest()).toBe(false)
})


it.each(['during logout', 'after logout'])(
  'does not toast a revoked background request %s',
  async (stage) => {
    const notify = vi.spyOn(toast, 'error').mockReturnValue('error')
    localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))
    api.defaults.adapter = async (config) => {
      if (stage === 'during logout') beginExplicitSignOut()
      else {
        finishExplicitSignOut()
        window.history.replaceState({}, '', '/sign-in')
      }
      throw new AxiosError('Unauthorized', 'ERR_BAD_REQUEST', config, undefined, {
        data: { message: 'access token invalid' },
        status: 401, statusText: 'Unauthorized', headers: {}, config,
      })
    }
    const client = createAppQueryClient()
    await expect(client.fetchQuery({
      queryKey: ['logout-background', stage],
      queryFn: () => api.get('/api/user/self', {
        headers: { Authorization: 'Bearer revoked-session' },
      }),
      retry: false,
    })).rejects.toBeInstanceOf(AxiosError)
    expect(notify).not.toHaveBeenCalled()
    client.clear()
    notify.mockRestore()
  }
)

it.each(['/api/user/auth/logout', '/api/user/login'])(
  'still reports an explicit authentication failure from %s',
  async (url) => {
    const notify = vi.spyOn(toast, 'error').mockReturnValue('error')
    beginExplicitSignOut()
    api.defaults.adapter = async (config) => {
      throw new AxiosError('Unauthorized', 'ERR_BAD_REQUEST', config, undefined, {
        data: { message: 'Authentication failed' },
        status: 401, statusText: 'Unauthorized', headers: {}, config,
      })
    }
    await api.post(url, {}, { skipAuthRefresh: true,
      headers: { Authorization: 'Bearer session' },
    }).catch((error) => handleServerError(error))
    expect(notify).toHaveBeenCalledWith('Authentication failed')
    notify.mockRestore()
  }
)
