/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { UserAuthForm } from '../user-auth-form'

const { onPasswordSubmit } = vi.hoisted(() => ({
  onPasswordSubmit: vi.fn(),
}))

vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({
    status: {
      password_login_enabled: false,
      passkey_login: true,
      github_oauth: true,
      user_agreement_enabled: true,
    },
  }),
}))

vi.mock('@/features/auth/hooks/use-turnstile', () => ({
  useTurnstile: () => ({
    isTurnstileEnabled: true,
    turnstileSiteKey: 'site-key',
    turnstileToken: '',
    setTurnstileToken: vi.fn(),
    validateTurnstile: vi.fn(() => false),
  }),
}))

vi.mock('@/features/auth/hooks/use-auth-redirect', () => ({
  useAuthRedirect: () => ({ handleLoginResult: vi.fn() }),
}))

vi.mock('@/features/auth/components/oauth-providers', () => ({
  OAuthProviders: () => <div>unsupported-oauth</div>,
}))

vi.mock('@/features/auth/components/legal-consent', () => ({
  LegalConsent: (props: {
    checked: boolean
    onCheckedChange: (checked: boolean) => void
  }) => (
    <label>
      legal-consent
      <input
        type='checkbox'
        checked={props.checked}
        onChange={(event) => props.onCheckedChange(event.target.checked)}
      />
    </label>
  ),
}))

vi.mock('@/lib/passkey', () => ({
  isPasskeySupported: vi.fn(async () => true),
}))

vi.mock('@/lib/handle-server-error', () => ({
  handleServerError: vi.fn(),
}))

afterEach(() => {
  cleanup()
  vi.clearAllMocks()
})

it('re-enables the native account form after a rejected attempt', async () => {
  onPasswordSubmit
    .mockRejectedValueOnce(new Error('Invalid identifier or password.'))
    .mockResolvedValueOnce(undefined)
  render(
    <UserAuthForm
      passwordOnly
      forgotPasswordUrl='https://account.openzrob.com/account/recovery'
      onPasswordSubmit={onPasswordSubmit}
    />
  )

  fireEvent.change(screen.getByLabelText('Username or Email'), {
    target: { value: 'alice@example.com' },
  })
  fireEvent.change(screen.getByLabelText('Password'), {
    target: { value: 'wrong-password' },
  })
  fireEvent.click(screen.getByRole('checkbox', { name: 'legal-consent' }))
  const submit = screen.getByRole('button', { name: 'Sign in' })
  fireEvent.click(submit)
  await waitFor(() => expect(onPasswordSubmit).toHaveBeenCalledTimes(1))
  await waitFor(() => expect(submit).toBeEnabled())

  fireEvent.click(submit)
  await waitFor(() => expect(onPasswordSubmit).toHaveBeenCalledTimes(2))
})

it('keeps legal consent and submits the native password fields through the account handler', async () => {
  onPasswordSubmit.mockResolvedValue(undefined)
  render(
    <UserAuthForm
      passwordOnly
      forgotPasswordUrl='https://account.openzrob.com/account/recovery'
      onPasswordSubmit={onPasswordSubmit}
    />
  )

  expect(screen.queryByText('unsupported-oauth')).toBeNull()
  expect(screen.queryByText('Sign in with Passkey')).toBeNull()
  expect(screen.queryByTestId('turnstile-container')).toBeNull()
  expect(
    screen.getByRole('link', { name: 'Forgot password?' })
  ).toHaveAttribute('href', 'https://account.openzrob.com/account/recovery')
  fireEvent.change(screen.getByLabelText('Username or Email'), {
    target: { value: 'alice@example.com' },
  })
  fireEvent.change(screen.getByLabelText('Password'), {
    target: { value: 'secret-password' },
  })
  fireEvent.click(screen.getByRole('checkbox', { name: 'legal-consent' }))
  fireEvent.click(screen.getByRole('button', { name: 'Sign in' }))

  await waitFor(() =>
    expect(onPasswordSubmit).toHaveBeenCalledWith(
      { username: 'alice@example.com', password: 'secret-password' },
      expect.any(AbortSignal)
    )
  )
})
