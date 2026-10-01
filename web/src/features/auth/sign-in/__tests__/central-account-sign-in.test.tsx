/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { userEvent } from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import {
  isCentralSignedOut,
  markCentralSignedOut,
  requestFreshCentralLogin,
} from '../central-reauth'
import { CentralAccountSignIn } from '../components/central-account-sign-in'

const { assign, startCentralSSO } = vi.hoisted(() => ({
  assign: vi.fn(),
  startCentralSSO: vi.fn(),
}))

vi.mock('../central-sso', () => ({ startCentralSSO }))

const VALID_AUTHORIZATION_URL =
  'https://account.openfox.work/v1/oauth/authorize?client_id=openfox-platform&redirect_uri=https%3A%2F%2Fai.openfox.work%2Faccount%2Fcallback&code_challenge=challenge&code_challenge_method=S256&state=state'

function stubLocation(): void {
  vi.stubGlobal('location', {
    origin: 'http://localhost:5173',
    assign,
  })
}

afterEach(() => {
  cleanup()
  sessionStorage.clear()
  vi.unstubAllGlobals()
  vi.clearAllMocks()
})

it('redirects the whole page to the backend-provided authorization URL', async () => {
  stubLocation()
  startCentralSSO.mockResolvedValue(VALID_AUTHORIZATION_URL)

  render(<CentralAccountSignIn redirectTo='/usage-logs/common' />)

  await waitFor(() => expect(assign).toHaveBeenCalledTimes(1))
  expect(startCentralSSO).toHaveBeenCalledWith('/usage-logs/common', {
    reauthenticate: false,
  })
  expect(assign).toHaveBeenCalledWith(VALID_AUTHORIZATION_URL)
})

it('requests forced account reauthentication when a fresh login was reserved', async () => {
  stubLocation()
  requestFreshCentralLogin()
  startCentralSSO.mockResolvedValue(VALID_AUTHORIZATION_URL)

  render(<CentralAccountSignIn />)

  await waitFor(() => expect(assign).toHaveBeenCalledTimes(1))
  expect(startCentralSSO).toHaveBeenCalledWith('/dashboard', {
    reauthenticate: true,
  })
})

it('does not redirect automatically after an explicit sign-out until the user continues', async () => {
  stubLocation()
  markCentralSignedOut()
  startCentralSSO.mockResolvedValue(VALID_AUTHORIZATION_URL)

  render(<CentralAccountSignIn />)
  await waitFor(() =>
    expect(
      screen.getByRole('button', { name: /account center/i })
    ).toBeVisible()
  )
  expect(startCentralSSO).not.toHaveBeenCalled()
  expect(isCentralSignedOut()).toBe(true)

  await userEvent
    .setup()
    .click(screen.getByRole('button', { name: /account center/i }))

  await waitFor(() =>
    expect(assign).toHaveBeenCalledWith(VALID_AUTHORIZATION_URL)
  )
  expect(isCentralSignedOut()).toBe(false)
})

it('rejects an authorization URL that is not a redirect-mode authorize request', async () => {
  stubLocation()
  startCentralSSO.mockResolvedValue(
    'https://evil.example/v1/oauth/authorize?response_mode=json&state=state'
  )

  render(<CentralAccountSignIn />)
  const button = await screen.findByRole('button', { name: /account center/i })

  expect(assign).not.toHaveBeenCalled()
  expect(
    screen.getByText('The account authorization URL is invalid.')
  ).toBeVisible()

  startCentralSSO.mockResolvedValue(VALID_AUTHORIZATION_URL)
  await userEvent.setup().click(button)
  await waitFor(() =>
    expect(assign).toHaveBeenCalledWith(VALID_AUTHORIZATION_URL)
  )
})
