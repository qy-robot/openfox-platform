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
import { cleanup, render, renderHook, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { checkIsActive } from '@/components/layout/lib/url-utils'
import {
  parseSidebarModulesAdmin,
  serializeSidebarModulesAdmin,
} from '@/features/system-settings/maintenance/config'
import { ROLE } from '@/lib/roles'
import { useAuthStore } from '@/stores/auth-store'

import { useSidebarConfig } from '../use-sidebar-config'
import { useSidebarData } from '../use-sidebar-data'
import { useSidebarView } from '../use-sidebar-view'

beforeEach(() => {
  vi.stubGlobal('localStorage', {
    getItem: () => null,
    setItem: () => undefined,
    removeItem: () => undefined,
  })
})

afterEach(() => {
  cleanup()
  vi.unstubAllGlobals()
  useAuthStore.getState().auth.reset()
})

function sidebarFor(
  admin?: object,
  user?: object,
  canConfigure = true,
  role: number = ROLE.USER
) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  client.setQueryData(['status'], {
    SidebarModulesAdmin: admin ? JSON.stringify(admin) : '',
  })
  useAuthStore.getState().auth.setUser({
    id: 1,
    username: 'alice',
    role,
    permissions: { sidebar_settings: canConfigure },
    sidebar_modules: user ? JSON.stringify(user) : '',
  })
  function Wrapper(props: { children: ReactNode }) {
    return (
      <QueryClientProvider client={client}>
        {props.children}
      </QueryClientProvider>
    )
  }
  const result = renderHook(
    () => useSidebarConfig(useSidebarData().navGroups),
    { wrapper: Wrapper }
  )
  return result
}

/**
 * Resolve the sidebar exactly as the layout component does (`useSidebarView`),
 * so role gating (`requiredRole` + the admin group) is exercised for real.
 */
async function sidebarViewFor(
  role: number,
  admin?: object,
  user?: object,
  canConfigure = true
) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  client.setQueryData(['status'], {
    SidebarModulesAdmin: admin ? JSON.stringify(admin) : '',
  })
  useAuthStore.getState().auth.setUser({
    id: 1,
    username: 'alice',
    role,
    permissions: { sidebar_settings: canConfigure },
    sidebar_modules: user ? JSON.stringify(user) : '',
  })

  function Probe() {
    const view = useSidebarView()
    return (
      <ul>
        {view.navGroups.flatMap((group) =>
          group.items.map((item) => (
            <li
              key={item.title}
              data-url={'url' in item ? item.url : undefined}
            >
              {item.title}
            </li>
          ))
        )}
      </ul>
    )
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
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
  const entries = screen.getAllByRole('listitem')
  return {
    titles: entries.map((entry) => entry.textContent ?? ''),
    entryByTitle: (title: string) =>
      entries.find((entry) => entry.textContent === title),
  }
}

describe('role-based sidebar visibility', () => {
  it('removes the entire Chat group for every role', () => {
    const userSidebar = sidebarFor()
    const adminSidebar = sidebarFor(undefined, undefined, true, ROLE.ADMIN)

    for (const sidebar of [userSidebar, adminSidebar]) {
      expect(sidebar.result.current.some((group) => group.id === 'chat')).toBe(
        false
      )
      expect(
        sidebar.result.current
          .flatMap((group) => group.items)
          .some((item) => item.title === 'Playground')
      ).toBe(false)
    }
  })

  it('shows regular users only the slimmed console while admins keep the full view', async () => {
    const userView = await sidebarViewFor(ROLE.USER)
    expect(userView.titles).toEqual([
      'Available Models',
      'API Keys',
      'Usage Logs',
      'Wallet',
      'Teams',
    ])

    const adminView = await sidebarViewFor(ROLE.ADMIN)
    expect(adminView.titles).toEqual(
      expect.arrayContaining([
        'Available Models',
        'Overview',
        'Dashboard',
        'Usage Logs',
        'Audit Logs',
        'Task Logs',
        'Wallet',
        'Teams',
        'Profile',
      ])
    )
  })

  it('points the Available Models entry at the dedicated page for every role', async () => {
    for (const role of [ROLE.USER, ROLE.ADMIN]) {
      const view = await sidebarViewFor(role)
      expect(
        view.entryByTitle('Available Models')?.getAttribute('data-url')
      ).toBe('/available-models')
    }
  })
})

describe('security sidebar visibility', () => {
  it('replaces legacy identity pages with the central account entry', () => {
    vi.stubGlobal('localStorage', {
      getItem: (key: string) =>
        key === 'status'
          ? JSON.stringify({ account_auth_enabled: true })
          : null,
      setItem: () => undefined,
      removeItem: () => undefined,
    })

    const { result } = sidebarFor(undefined, undefined, true, ROLE.ADMIN)
    const items =
      result.current.find((group) => group.id === 'personal')?.items ?? []

    expect(items.map((item) => item.title)).toEqual([
      'Wallet',
      'Teams',
      'Account center',
    ])
    expect(items[2]).toMatchObject({
      url: 'https://account.openzrob.com/account/profile',
    })
  })

  it('old configurations show Security & Access immediately after Profile and keep API Keys', () => {
    const { result } = sidebarFor(
      { personal: { enabled: true, personal: true, topup: true } },
      { personal: { enabled: true, personal: true } },
      true,
      ROLE.ADMIN
    )
    expect(
      result.current
        .find((group) => group.id === 'personal')
        ?.items.map((item) => item.title)
    ).toEqual(['Wallet', 'Teams', 'Profile', 'Security & Access'])
    expect(
      result.current
        .flatMap((group) => group.items)
        .some((item) => item.title === 'API Keys')
    ).toBe(true)
  })
  it.each([
    [{ personal: { enabled: true, security: false } }, undefined],
    [{ personal: { enabled: false } }, { personal: { security: true } }],
    [undefined, { personal: { enabled: true, security: false } }],
    [undefined, { personal: { enabled: false } }],
  ])(
    'admin or user disablement hides Security & Access (%j, %j)',
    (admin, user) => {
      const { result } = sidebarFor(admin, user, true, ROLE.ADMIN)
      expect(
        result.current
          .flatMap((group) => group.items)
          .some((item) => item.title === 'Security & Access')
      ).toBe(false)
    }
  )
  it('users without sidebar configuration permission retain the admin view', () => {
    const { result } = sidebarFor(
      undefined,
      { personal: { security: false } },
      false,
      ROLE.ADMIN
    )
    expect(
      result.current
        .flatMap((group) => group.items)
        .some((item) => item.title === 'Security & Access')
    ).toBe(true)
  })
})

describe('audit log sidebar entry', () => {
  it('admin settings default Audit Logs to visible and preserve its independent toggle when saved', () => {
    const config = parseSidebarModulesAdmin(
      '{"console":{"enabled":true,"log":true}}'
    )
    expect(config.console.audit).toBe(true)
    config.console.audit = false
    const saved = parseSidebarModulesAdmin(serializeSidebarModulesAdmin(config))
    const { result } = sidebarFor(saved, undefined, true, ROLE.ADMIN)
    const titles = result.current
      .flatMap((group) => group.items)
      .map((item) => item.title)
    expect(titles).not.toContain('Audit Logs')
    expect(titles).toContain('Usage Logs')
  })

  it('legacy configurations show a separate Audit Logs link immediately after Usage Logs', () => {
    const { result } = sidebarFor(
      { console: { enabled: true, log: true } },
      { console: { enabled: true, log: true } },
      true,
      ROLE.ADMIN
    )
    const items =
      result.current.find((group) => group.id === 'general')?.items ?? []
    const usageIndex = items.findIndex((item) => item.title === 'Usage Logs')
    expect(items[usageIndex + 1]).toMatchObject({
      title: 'Audit Logs',
      url: '/usage-logs/audit',
    })
    const selected = items.filter((item) =>
      checkIsActive('/usage-logs/audit', item)
    )
    expect(selected.map((item) => item.title)).toEqual(['Audit Logs'])
  })

  it.each([
    [{ console: { enabled: true, audit: false } }, undefined],
    [{ console: { enabled: false } }, { console: { audit: true } }],
    [undefined, { console: { enabled: true, audit: false } }],
    [undefined, { console: { enabled: false } }],
  ])(
    'admin and personal visibility rules can hide Audit Logs (%j, %j)',
    (admin, user) => {
      const { result } = sidebarFor(admin, user, true, ROLE.ADMIN)
      expect(
        result.current
          .flatMap((group) => group.items)
          .some((item) => item.title === 'Audit Logs')
      ).toBe(false)
    }
  )

  it('hiding Usage Logs does not hide the independently configured Audit Logs entry', () => {
    const { result } = sidebarFor(
      { console: { enabled: true, log: false } },
      undefined,
      true,
      ROLE.ADMIN
    )
    const titles = result.current
      .flatMap((group) => group.items)
      .map((item) => item.title)
    expect(titles).not.toContain('Usage Logs')
    expect(titles).toContain('Audit Logs')
  })
})
