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
import { renderHook, cleanup } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useTopNavLinks } from '../use-top-nav-links'

const state = vi.hoisted(() => ({
  status: {} as Record<string, unknown>,
  user: null as object | null,
}))
vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({ status: state.status }),
}))
vi.mock('@/stores/auth-store', () => ({
  useAuthStore: () => ({ auth: { user: state.user } }),
}))
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}))
afterEach(() => {
  cleanup()
  state.status = {}
  state.user = null
})
describe('product navigation', () => {
  it.each([true, false])(
    'keeps the public workbench entry separate when console is %s',
    (console) => {
      state.status = { HeaderNavModules: JSON.stringify({ console }) }
      const links = renderHook(useTopNavLinks).result.current
      expect(
        links.filter((link) => link.href === 'https://dash.openfox.work')
      ).toEqual([
        {
          title: 'Workbench',
          href: 'https://dash.openfox.work',
          external: true,
        },
      ])
      expect(links.find((link) => link.title === 'Console')?.href).toBe(
        console ? '/dashboard' : undefined
      )
    }
  )

  it('always provides a public download entry with old stored navigation settings', () => {
    state.status = {
      HeaderNavModules: JSON.stringify({
        home: false,
        console: true,
        about: false,
      }),
    }
    const { result } = renderHook(useTopNavLinks)
    expect(result.current.filter((link) => link.href === '/download')).toEqual([
      { title: 'Download', href: '/download' },
    ])
  })
  it.each([
    'https://docs.newapi.pro',
    'https://docs.newapi.ai/guide',
    undefined,
  ])('omits upstream or unconfigured documentation %s', (url) => {
    state.status = { docs_link: url }
    expect(
      renderHook(useTopNavLinks).result.current.some(
        (link) => link.title === 'Docs'
      )
    ).toBe(false)
  })
  it('retains configured business documentation', () => {
    state.status = { docs_link: 'https://help.example.com' }
    expect(renderHook(useTopNavLinks).result.current).toContainEqual({
      title: 'Docs',
      href: 'https://help.example.com',
      external: true,
    })
  })
  it('preserves pricing authentication gating while keeping downloads public', () => {
    state.status = {
      HeaderNavModules: JSON.stringify({
        pricing: { enabled: true, requireAuth: true },
      }),
    }
    const links = renderHook(useTopNavLinks).result.current
    expect(links.find((link) => link.href === '/pricing')?.requiresAuth).toBe(
      true
    )
    expect(
      links.find((link) => link.href === '/download')?.requiresAuth
    ).toBeUndefined()
  })
})
