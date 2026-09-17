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

For commercial licensing, please contact support@quantumnous.com
*/
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ReactNode } from 'react'
import { beforeEach, describe, expect, test, vi } from 'vitest'

import { DesktopAuthorizationPage } from '..'
import * as authorizationApi from '../api'

vi.mock('../api', () => ({
  getDesktopAuthorization: vi.fn(),
  decideDesktopAuthorization: vi.fn(),
}))

function renderPage() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  function Wrapper(props: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>
        {props.children}
      </QueryClientProvider>
    )
  }
  return render(<DesktopAuthorizationPage userCode='ABCD-1234' />, {
    wrapper: Wrapper,
  })
}

describe('desktop authorization approval', () => {
  beforeEach(() => {
    vi.mocked(authorizationApi.getDesktopAuthorization).mockResolvedValue({
      user_code: 'ABCD-1234',
      device_name: 'Office Mac',
      expires_at: 1_800_000_000,
    })
    vi.mocked(authorizationApi.decideDesktopAuthorization).mockResolvedValue({})
  })

  test('shows matching device and code and waits for explicit approval', async () => {
    const user = userEvent.setup()
    renderPage()
    expect(await screen.findByText('Office Mac')).toBeVisible()
    expect(screen.getByText('ABCD-1234')).toBeVisible()
    expect(authorizationApi.decideDesktopAuthorization).not.toHaveBeenCalled()

    await user.click(screen.getByRole('button', { name: 'Approve' }))
    await waitFor(() =>
      expect(authorizationApi.decideDesktopAuthorization).toHaveBeenCalledWith(
        'ABCD-1234',
        true
      )
    )
    expect(await screen.findByText('Desktop access approved')).toBeVisible()
  })
})
