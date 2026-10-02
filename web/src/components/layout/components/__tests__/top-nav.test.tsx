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
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, expect, it, vi } from 'vitest'

import { SidebarProvider } from '@/components/ui/sidebar'
import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { AppHeader } from '../app-header'
import { TopNav } from '../top-nav'

beforeEach(() => {
  vi.stubGlobal('localStorage', {
    getItem: () => null,
    setItem: () => undefined,
    removeItem: () => undefined,
  })
  vi.spyOn(window, 'scrollTo').mockImplementation(() => undefined)
  vi.spyOn(api, 'get').mockResolvedValue({
    data: { success: true, data: {} },
  } as never)
})
afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
  useAuthStore.getState().auth.reset()
})

async function renderNavigation(app = false) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  client.setQueryData(['status'], {})
  client.setQueryData(['notice'], { success: true, data: '' })
  function Probe() {
    if (app) {
      return (
        <SidebarProvider>
          <AppHeader
            showSearch={false}
            showNotifications={false}
            showConfigDrawer={false}
            showProfileDropdown={false}
          />
        </SidebarProvider>
      )
    }
    return (
      <TopNav
        links={[
          { title: 'Home', href: '/' },
          {
            title: 'Workbench',
            href: 'https://dash.example.com',
            external: true,
          },
          { title: 'Unavailable', href: '/unavailable', disabled: true },
        ]}
      />
    )
  }
  const root = createRootRoute()
  const index = createRoute({
    getParentRoute: () => root,
    path: '/',
    component: Probe,
  })
  const router = createRouter({
    routeTree: root.addChildren([index]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  await router.load()
  return render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}

it('exposes the app navigation menu outside desktop-only containers', async () => {
  await renderNavigation(true)
  const trigger = screen.getByRole('button', { name: 'Toggle navigation menu' })
  expect(trigger.closest('.hidden')).toBeNull()
  await userEvent.setup().click(trigger)
  expect(await screen.findByRole('menu')).toBeVisible()
  expect(screen.getByRole('menuitem', { name: 'Workbench' })).toHaveAttribute(
    'target',
    '_blank'
  )
})

it('keeps current, external and disabled destinations accessible in the mobile menu', async () => {
  const user = userEvent.setup()
  await renderNavigation()
  expect(screen.getByRole('link', { name: 'Home' })).toHaveAttribute(
    'aria-current',
    'page'
  )
  const trigger = screen.getByRole('button', { name: 'Toggle navigation menu' })
  await user.click(trigger)
  const menu = await screen.findByRole('menu')
  expect(within(menu).getByRole('menuitem', { name: 'Home' })).toHaveAttribute(
    'aria-current',
    'page'
  )
  expect(
    within(menu).getByRole('menuitem', { name: 'Workbench' })
  ).toHaveAttribute('rel', 'noopener noreferrer')
  expect(
    within(menu).getByRole('menuitem', { name: 'Unavailable' })
  ).toHaveAttribute('aria-disabled', 'true')
  await user.keyboard('{Escape}')
  expect(screen.queryByRole('menu')).toBeNull()
  expect(trigger).toHaveFocus()
})
