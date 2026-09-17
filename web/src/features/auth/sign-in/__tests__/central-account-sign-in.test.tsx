/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { act, cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { isCentralSignedOut, markCentralSignedOut } from '../central-reauth'
import { CentralAccountSignIn } from '../components/central-account-sign-in'

const {
  capturedFormProps,
  finishCentralSSOResult,
  startCentralSSO,
  toastError,
} = vi.hoisted(() => ({
  capturedFormProps: { current: undefined as unknown },
  finishCentralSSOResult: vi.fn(),
  startCentralSSO: vi.fn(),
  toastError: vi.fn(),
}))

vi.mock('../central-sso', () => ({
  cancelCentralSSO: vi.fn(),
  finishCentralSSOResult,
  startCentralSSO,
}))

vi.mock('../components/user-auth-form', () => ({
  UserAuthForm: (props: unknown) => {
    capturedFormProps.current = props
    return <div>native-account-form</div>
  },
}))

vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({
    status: { account_center_url: 'https://account.openzrob.com' },
  }),
}))

vi.mock('sonner', () => ({ toast: { error: toastError } }))

afterEach(() => {
  cleanup()
  sessionStorage.clear()
  vi.unstubAllGlobals()
  vi.clearAllMocks()
  capturedFormProps.current = undefined
})

it('renders the native form without embedding the account site', () => {
  markCentralSignedOut()
  render(<CentralAccountSignIn redirectTo='/dashboard' />)

  expect(screen.getByText('native-account-form')).toBeVisible()
  expect(document.querySelector('iframe')).toBeNull()
  expect(capturedFormProps.current).toEqual(
    expect.objectContaining({
      passwordOnly: true,
      forgotPasswordUrl: 'https://account.openzrob.com/account/recovery',
    })
  )
})

it('sends credentials only to account then exchanges JSON authorization', async () => {
  markCentralSignedOut()
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce({
      ok: true,
      json: async () => ({ success: true }),
    })
    .mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        success: true,
        data: {
          type: 'robocoding.oauth.result',
          state: 'state',
          code: 'code',
        },
      }),
    })
  vi.stubGlobal('fetch', fetchMock)
  startCentralSSO.mockResolvedValue(
    'https://account.openzrob.com/v1/oauth/authorize?response_mode=json'
  )
  finishCentralSSOResult.mockResolvedValue({
    status: 'authenticated',
    returnTo: '/dashboard',
  })

  render(<CentralAccountSignIn redirectTo='/dashboard' />)
  const props = capturedFormProps.current as {
    onPasswordSubmit: (
      credentials: { username: string; password: string },
      signal: AbortSignal
    ) => Promise<void>
  }
  await act(() =>
    props.onPasswordSubmit(
      { username: 'alice@example.com', password: 'secret' },
      new AbortController().signal
    )
  )

  expect(fetchMock).toHaveBeenNthCalledWith(
    1,
    'https://account.openzrob.com/v1/auth/login',
    expect.objectContaining({
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        identifier: 'alice@example.com',
        password: 'secret',
      }),
    })
  )
  expect(fetchMock.mock.calls[0]?.[1]?.headers).not.toHaveProperty(
    'Authorization'
  )
  expect(fetchMock).toHaveBeenNthCalledWith(
    2,
    expect.stringContaining('response_mode=json'),
    expect.objectContaining({ credentials: 'include', method: 'GET' })
  )
  expect(finishCentralSSOResult).toHaveBeenCalledWith(
    expect.objectContaining({ code: 'code', state: 'state' }),
    expect.any(AbortSignal)
  )
  expect(isCentralSignedOut()).toBe(false)
})

it('rejects an authorization URL outside the configured account origin', async () => {
  markCentralSignedOut()
  const fetchMock = vi.fn().mockResolvedValue({
    ok: true,
    json: async () => ({ success: true }),
  })
  vi.stubGlobal('fetch', fetchMock)
  startCentralSSO.mockResolvedValue(
    'https://evil.example/v1/oauth/authorize?response_mode=json'
  )

  render(<CentralAccountSignIn />)
  const props = capturedFormProps.current as {
    onPasswordSubmit: (
      credentials: { username: string; password: string },
      signal: AbortSignal
    ) => Promise<void>
  }
  await expect(
    props.onPasswordSubmit(
      { username: 'alice@example.com', password: 'secret' },
      new AbortController().signal
    )
  ).rejects.toThrow('authorization URL is invalid')

  expect(fetchMock).toHaveBeenCalledTimes(1)
  expect(fetchMock).toHaveBeenCalledWith(
    'https://account.openzrob.com/v1/auth/login',
    expect.anything()
  )
})

it('preserves explicit sign-out when account credentials are rejected', async () => {
  markCentralSignedOut()
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({
        success: false,
        code: 'AUTH_INVALID_CREDENTIALS',
        message: 'Invalid identifier or password.',
      }),
    })
  )

  render(<CentralAccountSignIn />)
  const props = capturedFormProps.current as {
    onPasswordSubmit: (
      credentials: { username: string; password: string },
      signal: AbortSignal
    ) => Promise<void>
  }
  await expect(
    props.onPasswordSubmit(
      { username: 'alice@example.com', password: 'wrong-password' },
      new AbortController().signal
    )
  ).rejects.toThrow('Invalid identifier or password.')

  expect(isCentralSignedOut()).toBe(true)
  expect(startCentralSSO).not.toHaveBeenCalled()
})

it('silently restores an account cookie unless the user explicitly signed out', async () => {
  const fetchMock = vi.fn().mockResolvedValue({
    ok: false,
    json: async () => ({
      success: false,
      data: {
        type: 'robocoding.oauth.result',
        state: 'state',
        error: 'login_required',
      },
    }),
  })
  vi.stubGlobal('fetch', fetchMock)
  startCentralSSO.mockResolvedValue(
    'https://account.openzrob.com/v1/oauth/authorize?response_mode=json'
  )
  finishCentralSSOResult.mockResolvedValue({
    status: 'interaction_required',
    error: 'login_required',
  })

  render(<CentralAccountSignIn />)
  await waitFor(() => expect(startCentralSSO).toHaveBeenCalledTimes(1))
  cleanup()
  markCentralSignedOut()
  vi.clearAllMocks()
  render(<CentralAccountSignIn />)
  await act(async () => {})
  expect(startCentralSSO).not.toHaveBeenCalled()
})
