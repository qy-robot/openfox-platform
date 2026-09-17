/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { act, render } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { SignOutDialog } from '../sign-out-dialog'

type ConfirmProps = { handleConfirm: () => Promise<void> }

const { captured, centralLogout, localLogout, navigate, toastSuccess } =
  vi.hoisted(() => ({
    captured: { current: undefined as ConfirmProps | undefined },
    centralLogout: vi.fn(),
    localLogout: vi.fn(),
    navigate: vi.fn(),
    toastSuccess: vi.fn(),
  }))

vi.mock('@/components/confirm-dialog', () => ({
  ConfirmDialog: (props: unknown) => {
    captured.current = props as ConfirmProps
    return null
  },
}))
vi.mock('@/features/auth/api', () => ({ logout: localLogout }))
vi.mock('@/features/auth/sign-in/central-logout', () => ({
  logoutCentralBrowserSession: centralLogout,
}))
vi.mock('@/lib/auth-session', () => ({
  beginExplicitSignOut: vi.fn(),
  clearAuthenticatedClientState: vi.fn(),
  finishExplicitSignOut: vi.fn(),
}))
vi.mock('@/features/auth/sign-in/central-reauth', () => ({
  markCentralSignedOutIfEnabled: vi.fn(),
}))
vi.mock('@/lib/handle-server-error', () => ({ handleServerError: vi.fn() }))
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => navigate }))
vi.mock('sonner', () => ({ toast: { success: toastSuccess } }))
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (value: string) => value }),
}))

afterEach(() => {
  vi.clearAllMocks()
  captured.current = undefined
})

function setup() {
  const queryClient = new QueryClient()
  render(
    <QueryClientProvider client={queryClient}>
      <SignOutDialog open onOpenChange={vi.fn()} />
    </QueryClientProvider>
  )
  if (!captured.current) {
    throw new Error('ConfirmDialog props were not captured')
  }
  return captured.current
}

it('cancels queries, signs out centrally, then revokes the local session', async () => {
  const order: string[] = []
  vi.spyOn(QueryClient.prototype, 'cancelQueries').mockImplementation(
    async () => {
      order.push('cancel')
    }
  )
  centralLogout.mockImplementation(async () => order.push('central'))
  localLogout.mockImplementation(async () => {
    order.push('local')
    return { success: true }
  })
  navigate.mockImplementation(async () => order.push('navigate'))
  const confirm = setup()
  await act(() => confirm.handleConfirm())
  expect(order).toEqual(['cancel', 'central', 'local', 'navigate'])
  expect(toastSuccess).toHaveBeenCalledWith('Signed out')
})

it('keeps the local session and suppresses success toast when central logout fails', async () => {
  centralLogout.mockRejectedValue(new Error('account logout failed'))
  const confirm = setup()
  await act(() => confirm.handleConfirm())
  expect(localLogout).not.toHaveBeenCalled()
  expect(navigate).not.toHaveBeenCalled()
  expect(toastSuccess).not.toHaveBeenCalled()
})
