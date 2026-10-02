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
import {
  cleanup,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { PublicHeader } from '../public-header'

beforeEach(() => {
  // jsdom has no layout; model a viewport without inset scrollbars.
  vi.spyOn(document.documentElement, 'clientWidth', 'get').mockReturnValue(
    window.innerWidth
  )
  vi.spyOn(document.body, 'clientWidth', 'get').mockReturnValue(
    window.innerWidth
  )
  vi.spyOn(window, 'scrollTo').mockImplementation(() => undefined)
  vi.stubGlobal('localStorage', {
    getItem: () => null,
    setItem: () => undefined,
    removeItem: () => undefined,
  })
  vi.spyOn(api, 'get').mockImplementation(
    async () => ({ data: { success: true, data: {} } }) as never
  )
})

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
  useAuthStore.getState().auth.reset()
})

async function renderHeader(status: Record<string, unknown> = {}) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  client.setQueryData(['status'], status)
  client.setQueryData(['notice'], { success: true, data: '' })

  function Probe() {
    return <PublicHeader />
  }
  const rootRoute = createRootRoute()
  const indexRoute = createRoute({
    getParentRoute: () => rootRoute,
    path: '/',
    component: Probe,
  })
  const router = createRouter({
    routeTree: rootRoute.addChildren([indexRoute]),
    history: createMemoryHistory({ initialEntries: ['/'] }),
  })
  await router.load()

  return render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}

describe('public header layout', () => {
  it('identifies the current page in desktop and mobile navigation', async () => {
    await renderHeader()
    const homeLinks = screen.getAllByRole('link', {
      name: 'Home',
      hidden: true,
    })
    expect(homeLinks).toHaveLength(2)
    for (const link of homeLinks) {
      expect(link).toHaveAttribute('aria-current', 'page')
    }
  })

  it('opens the mobile navigation and restores scrolling when a link closes it', async () => {
    const user = userEvent.setup()
    await renderHeader()
    const trigger = screen.getByRole('button', {
      name: 'Toggle navigation menu',
    })
    expect(trigger).toHaveAttribute('aria-expanded', 'false')
    await user.click(trigger)
    expect(trigger).toHaveAttribute('aria-expanded', 'true')
    expect(document.body.style.overflow).toBe('hidden')
    const mobile = document.querySelector<HTMLElement>(
      '#public-mobile-navigation'
    )
    if (!mobile) {
      throw new Error('Mobile navigation missing')
    }
    expect(mobile).toHaveAttribute('aria-hidden', 'false')
    await user.click(within(mobile).getByRole('link', { name: 'Home' }))
    expect(trigger).toHaveAttribute('aria-expanded', 'false')
    expect(document.body.style.overflow).toBe('')
  })

  it('keeps a protected mobile destination behind the existing sign-in prompt', async () => {
    const user = userEvent.setup()
    await renderHeader({
      HeaderNavModules: JSON.stringify({
        pricing: { enabled: true, requireAuth: true },
      }),
    })
    const trigger = screen.getByRole('button', {
      name: 'Toggle navigation menu',
    })
    await user.click(trigger)
    const mobile = document.querySelector<HTMLElement>(
      '#public-mobile-navigation'
    )
    if (!mobile) {
      throw new Error('Mobile navigation missing')
    }
    await user.click(within(mobile).getByRole('link', { name: 'Model Square' }))
    expect(
      await screen.findByRole('dialog', { name: 'Sign in required' })
    ).toBeVisible()
    expect(trigger).toHaveAttribute('aria-expanded', 'false')
    await user.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(
      screen.queryByRole('dialog', { name: 'Sign in required' })
    ).toBeNull()
    await waitFor(() => expect(document.body.style.overflow).toBe(''))
  })
  it('renders a static full-width top bar without the floating capsule scroll state', async () => {
    const { container } = await renderHeader()

    const header = container.querySelector('.robo-public-header')
    const bar = container.querySelector('.robo-public-header-inner')
    const nav = container.querySelector('.robo-public-nav')

    expect(header).not.toBeNull()
    expect(bar).not.toBeNull()
    expect(nav).not.toBeNull()
    expect(bar?.className).toContain('robo-public-header-inner')
    expect(bar?.className).not.toContain('--scrolled')
    expect(nav?.className).not.toContain('--scrolled')
  })

  it('keeps the header geometry unchanged while scrolling', async () => {
    const { container } = await renderHeader()

    Object.defineProperty(window, 'scrollY', {
      value: 200,
      configurable: true,
    })
    window.dispatchEvent(new Event('scroll'))

    const bar = container.querySelector('.robo-public-header-inner')
    const nav = container.querySelector('.robo-public-nav')
    expect(bar?.className).not.toContain('--scrolled')
    expect(nav?.className).not.toContain('--scrolled')
  })
})
