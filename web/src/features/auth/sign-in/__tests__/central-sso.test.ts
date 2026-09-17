/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { beforeEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'

import {
  clearCentralReauthenticationAttempt,
  reserveCentralReauthentication,
} from '../central-reauth'
import {
  finishCentralSSO,
  finishCentralSSOMessage,
  resolveCentralSSOReturnTo,
  startCentralSSO,
} from '../central-sso'

vi.mock('@/lib/api', () => ({
  api: { get: vi.fn(), post: vi.fn() },
  applyAuthBundle: vi.fn(),
  isAuthBundle: vi.fn(),
}))

describe('central account SSO callback', () => {
  beforeEach(() => {
    sessionStorage.clear()
    vi.clearAllMocks()
  })

  it('rejects a callback when the returned state differs before exchanging the code', async () => {
    sessionStorage.setItem(
      'robocoding:account-sso',
      JSON.stringify({
        verifier: 'verifier',
        state: 'expected',
        returnTo: '/dashboard',
      })
    )

    await expect(finishCentralSSO('code', 'forged')).rejects.toThrow(
      'state did not match'
    )
    expect(api.post).not.toHaveBeenCalled()
  })

  it('rejects protocol-relative and cross-origin return locations', () => {
    expect(resolveCentralSSOReturnTo('//evil.example')).toBe('/dashboard')
    expect(resolveCentralSSOReturnTo('https://evil.example/account')).toBe(
      '/dashboard'
    )
    expect(resolveCentralSSOReturnTo('/dashboard?tab=usage')).toBe(
      '/dashboard?tab=usage'
    )
  })

  it('only requests forced account reauthentication for renewal flows', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: { data: { authorizationUrl: 'https://account.example/authorize' } },
    })

    await startCentralSSO('/dashboard', { prompt: 'none' })
    await startCentralSSO('/dashboard', { reauthenticate: true })

    expect(api.get).toHaveBeenNthCalledWith(
      1,
      '/api/account/sso/start',
      expect.objectContaining({
        params: expect.objectContaining({
          responseMode: 'web_message',
          prompt: 'none',
        }),
        skipAuthRefresh: true,
      })
    )
    expect(api.get).toHaveBeenNthCalledWith(
      2,
      '/api/account/sso/start',
      expect.objectContaining({
        params: expect.objectContaining({
          responseMode: 'web_message',
          reauth: true,
        }),
        skipAuthRefresh: true,
      })
    )
  })

  it('rejects an invalid authentication bundle before updating auth state', async () => {
    const { applyAuthBundle, isAuthBundle } = await import('@/lib/api')
    sessionStorage.setItem(
      'robocoding:account-sso',
      JSON.stringify({
        verifier: 'verifier',
        state: 'expected',
        returnTo: '/dashboard',
      })
    )
    vi.mocked(api.post).mockResolvedValue({ data: { data: { user: {} } } })
    vi.mocked(isAuthBundle).mockReturnValue(false)

    await expect(finishCentralSSO('code', 'expected')).rejects.toThrow(
      'invalid sign-in response'
    )
    expect(applyAuthBundle).not.toHaveBeenCalled()
  })

  it('validates web_message origin, source, type, and state before consuming it', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        data: {
          authorizationUrl: 'https://account.openzrob.com/v1/oauth/authorize',
        },
      },
    })
    await startCentralSSO('/dashboard', { prompt: 'none' })
    const pending = JSON.parse(
      sessionStorage.getItem('robocoding:account-sso') || '{}'
    ) as { state: string }
    const valid = {
      origin: 'https://account.openzrob.com',
      source: window,
      data: {
        type: 'robocoding.oauth.result',
        state: pending.state,
        error: 'login_required',
      },
    } as unknown as MessageEvent

    await expect(
      finishCentralSSOMessage(
        { ...valid, origin: 'https://evil.example' },
        window,
        'https://account.openzrob.com/v1/oauth/authorize'
      )
    ).resolves.toEqual({ status: 'ignored' })
    await expect(
      finishCentralSSOMessage(
        { ...valid, source: null },
        window,
        'https://account.openzrob.com/v1/oauth/authorize'
      )
    ).resolves.toEqual({ status: 'ignored' })
    await expect(
      finishCentralSSOMessage(
        { ...valid, data: { type: 'unrelated.message' } },
        window,
        'https://account.openzrob.com/v1/oauth/authorize'
      )
    ).resolves.toEqual({ status: 'ignored' })
    await expect(
      finishCentralSSOMessage(
        { ...valid, data: { ...valid.data, state: 'forged' } },
        window,
        'https://account.openzrob.com/v1/oauth/authorize'
      )
    ).rejects.toThrow('state did not match')
    await expect(
      finishCentralSSOMessage(
        valid,
        window,
        'https://account.openzrob.com/v1/oauth/authorize'
      )
    ).resolves.toEqual({
      status: 'interaction_required',
      error: 'login_required',
    })
    expect(sessionStorage.getItem('robocoding:account-sso')).toBeNull()
  })

  it('blocks repeated central authorization attempts inside the loop window', () => {
    expect(reserveCentralReauthentication(1_000)).toBe(true)
    expect(reserveCentralReauthentication(30_000)).toBe(false)
    clearCentralReauthenticationAttempt()
    expect(reserveCentralReauthentication(30_000)).toBe(true)
  })
})
