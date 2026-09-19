import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  Outlet,
  RouterProvider,
} from '@tanstack/react-router'
import { cleanup, render, screen } from '@testing-library/react'
import type { AxiosAdapter } from 'axios'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { z } from 'zod'

import { getIdentityProfileURL } from '@/features/auth/central-account-navigation'
import { api } from '@/lib/api'
import * as authSession from '@/lib/auth-session'
import {
  beginExplicitSignOut,
  clearAuthentication,
  finishExplicitSignOut,
} from '@/lib/auth-session'
import { resolveSignInAuthentication } from '@/routes/(auth)/sign-in'
import { resolveProtectedAuthentication } from '@/routes/_authenticated/route'
import { useAuthStore, type AuthBundle } from '@/stores/auth-store'

import { CentralSSOCallback } from '../central-sso-callback'

const originalAdapter = api.defaults.adapter

const bundle: AuthBundle = {
  access_token: 'central-access',
  token_type: 'Bearer',
  access_expires_at: 9_999_999_999,
  user: { id: 42, username: 'central-user', role: 1 },
  session: {
    sid: 'central-session',
    current: true,
    login_method: 'central',
    ip: '',
    user_agent: '',
    created_at: 1,
    last_active_at: 1,
    expires_at: 9_999_999_999,
  },
}

beforeEach(() => {
  vi.spyOn(window, 'scrollTo').mockImplementation(() => {})
})

afterEach(() => {
  cleanup()
  api.defaults.adapter = originalAdapter
  localStorage.removeItem('status')
  sessionStorage.clear()
  useAuthStore.getState().auth.reset('idle')
  vi.restoreAllMocks()
})

it('keeps the exchanged token while routing from the callback to the workbench', async () => {
  sessionStorage.setItem(
    'robocoding:account-sso',
    JSON.stringify({
      verifier: 'pkce-verifier',
      state: 'expected-state',
      returnTo: '/workbench/skills?draft=1',
    })
  )
  const adapter = vi.fn<AxiosAdapter>(async (config) => ({
    data: { success: true, data: bundle },
    status: 200,
    statusText: 'OK',
    headers: {},
    config,
  }))
  api.defaults.adapter = adapter

  const root = createRootRoute({ component: Outlet })
  const searchSchema = z.object({
    code: z.string().optional(),
    state: z.string().optional(),
  })
  const callbackRoute = createRoute({
    getParentRoute: () => root,
    path: '/account/callback',
    validateSearch: searchSchema,
    component: () => {
      const search = callbackRoute.useSearch()
      return <CentralSSOCallback code={search.code} state={search.state} />
    },
  })
  const destinationRoute = createRoute({
    getParentRoute: () => root,
    path: '/workbench/skills',
    component: () => <div>Workbench destination</div>,
  })
  const router = createRouter({
    routeTree: root.addChildren([callbackRoute, destinationRoute]),
    history: createMemoryHistory({
      initialEntries: [
        '/account/callback?code=one-time-code&state=expected-state',
      ],
    }),
  })

  render(<RouterProvider router={router} />)

  expect(await screen.findByText('Workbench destination')).toBeVisible()
  expect(router.state.location.href).toBe('/workbench/skills?draft=1')
  expect(useAuthStore.getState().auth.accessToken).toBe('central-access')
  expect(adapter).toHaveBeenCalledTimes(1)
  expect(adapter.mock.calls[0]?.[0].url).toBe('/api/account/sso/exchange')
})

it('restores a cold central session before entering a protected route', async () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))
  const resolveAuthenticationSpy = vi
    .spyOn(authSession, 'resolveAuthentication')
    .mockResolvedValue({ kind: 'authenticated', bundle })

  await expect(
    resolveProtectedAuthentication('/workbench/skills')
  ).resolves.toBeNull()

  expect(resolveAuthenticationSpy).toHaveBeenCalledOnce()
})

it('keeps explicit central sign-in public without silently restoring a session', async () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))
  const adapter = vi.fn<AxiosAdapter>()
  api.defaults.adapter = adapter

  await expect(resolveSignInAuthentication()).resolves.toBe(false)

  expect(adapter).not.toHaveBeenCalled()
})

it('routes identity management to the account subdomain in central mode', () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))

  expect(getIdentityProfileURL()).toBe(
    'https://account.openfox.work/account/profile'
  )
})

it('does not restart central SSO while an explicit sign-out leaves a protected route', async () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))
  useAuthStore.getState().auth.setBundle(bundle)
  const adapter = vi.fn<AxiosAdapter>()
  api.defaults.adapter = adapter

  beginExplicitSignOut()
  clearAuthentication(false)
  try {
    await expect(
      resolveProtectedAuthentication('/workbench/skills')
    ).resolves.toBeNull()
  } finally {
    finishExplicitSignOut()
  }

  expect(adapter).not.toHaveBeenCalled()
})
