/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { expect, it, vi } from 'vitest'

import { AuthLayout } from '../auth-layout'

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, ...props }: { children: ReactNode; to: string }) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
}))

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))

vi.mock('@/hooks/use-system-config', () => ({
  useSystemConfig: () => ({
    systemName: 'OpenFox',
    logo: '/openfox-logo.png',
    loading: false,
  }),
}))

it('keeps the account form in a named main region with the OpenFox home link', () => {
  render(
    <AuthLayout>
      <form aria-label='Sign in form' />
    </AuthLayout>
  )

  expect(screen.getByRole('main')).toContainElement(
    screen.getByRole('form', { name: 'Sign in form' })
  )
  expect(screen.getByRole('link', { name: 'Logo OpenFox' })).toHaveAttribute(
    'href',
    '/'
  )
})

it('keeps the desktop app entry available from the decorative scene', () => {
  render(
    <AuthLayout>
      <div />
    </AuthLayout>
  )

  expect(
    screen.getByRole('link', { name: 'Download desktop app' })
  ).toHaveAttribute('href', '/download')
})
