/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { cleanup, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, expect, it, vi } from 'vitest'

import { SignIn } from '../sign-in'
import { SignUp } from '../sign-up'

const { navigateToAccountRegistration, state } = vi.hoisted(() => ({
  navigateToAccountRegistration: vi.fn(),
  state: {
    loading: false,
    search: {} as { continue?: string; redirect?: string },
    status: null as Record<string, unknown> | null,
    theme: 'light' as 'dark' | 'light',
  },
}))

vi.mock('@tanstack/react-router', () => ({
  Link: (props: { children: ReactNode; to: string }) => (
    <a href={props.to}>{props.children}</a>
  ),
  useSearch: () => state.search,
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))

vi.mock('@/context/theme-provider', () => ({
  useTheme: () => ({ resolvedTheme: state.theme }),
}))

vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({ loading: state.loading, status: state.status }),
}))

vi.mock('../central-account-navigation', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../central-account-navigation')>()),
  navigateToAccountRegistration,
}))

vi.mock('../auth-layout', () => ({
  AuthLayout: (props: { children: ReactNode }) => <main>{props.children}</main>,
}))

vi.mock('../components/terms-footer', () => ({
  TermsFooter: () => <footer />,
}))

vi.mock('../sign-in/components/central-account-sign-in', () => ({
  CentralAccountSignIn: () => <div>central-sign-in</div>,
}))

vi.mock('../sign-in/components/user-auth-form', () => ({
  UserAuthForm: () => <div>legacy-sign-in</div>,
}))

vi.mock('../sign-up/components/sign-up-form', () => ({
  SignUpForm: () => <div>legacy-sign-up</div>,
}))

afterEach(() => {
  cleanup()
  state.loading = false
  state.search = {}
  state.status = null
  state.theme = 'light'
  vi.clearAllMocks()
})

it('links directly from central sign in to the themed account registration page', () => {
  state.status = {
    account_auth_enabled: true,
    account_center_url: 'https://account.example.com',
    register_enabled: true,
  }
  state.theme = 'dark'

  render(<SignIn />)

  expect(screen.getByRole('link', { name: 'Sign up' })).toHaveAttribute(
    'href',
    'https://account.example.com/account/register?theme=dark'
  )
  expect(screen.getByText('central-sign-in')).toBeVisible()
})

it('keeps the platform registration route in legacy account mode', () => {
  state.status = { account_auth_enabled: false, register_enabled: true }

  render(<SignIn />)

  expect(screen.getByRole('link', { name: 'Sign up' })).toHaveAttribute(
    'href',
    '/sign-up'
  )
  expect(screen.getByText('legacy-sign-in')).toBeVisible()
})

it('redirects a direct central sign-up visit without rendering an intermediary page', () => {
  state.status = {
    account_auth_enabled: true,
    account_center_url: 'https://account.example.com',
  }
  state.search = {
    continue: '/v1/oauth/authorize?client_id=ai&state=one',
  }
  state.theme = 'dark'

  const view = render(<SignUp />)

  expect(navigateToAccountRegistration).toHaveBeenCalledWith(
    'https://account.example.com/account/register?theme=dark&continue=%2Fv1%2Foauth%2Fauthorize%3Fclient_id%3Dai%26state%3Done'
  )
  expect(view.container).toBeEmptyDOMElement()
})
