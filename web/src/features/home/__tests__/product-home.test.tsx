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
import { render, screen, cleanup } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, expect, it, vi } from 'vitest'

import { Home } from '..'

const state = vi.hoisted(() => ({ content: '', isLoaded: true, isUrl: false }))
vi.mock('../hooks', () => ({ useHomePageContent: () => state }))
vi.mock('@/context/theme-provider', () => ({
  useTheme: () => ({ resolvedTheme: 'light' }),
}))
vi.mock('@/stores/auth-store', () => ({
  useAuthStore: () => ({ auth: { user: null } }),
}))
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}))
vi.mock('@/components/layout', () => ({
  PublicLayout: (props: { children: ReactNode }) => <div>{props.children}</div>,
}))
vi.mock('@/components/layout/components/footer', () => ({
  Footer: () => <footer />,
}))
vi.mock('@/components/rich-content', () => ({
  RichContent: (props: { content: string }) => (
    <article>{props.content}</article>
  ),
}))
vi.mock('../components/product-home', () => ({
  ProductHome: () => <h1>OpenFox product home</h1>,
}))
afterEach(() => {
  cleanup()
  state.content = ''
  state.isLoaded = true
  state.isUrl = false
})
it('renders the product home when no administrator content is set', () => {
  render(<Home />)
  expect(screen.getByRole('heading').textContent).toContain('OpenFox')
})
it('preserves custom Markdown content', () => {
  state.content = '# Custom welcome'
  render(<Home />)
  expect(screen.getByRole('article').textContent).toBe('# Custom welcome')
  expect(screen.queryByRole('heading')).toBeNull()
})
it('preserves sandboxed external custom home content', () => {
  state.content = 'https://example.com/welcome'
  state.isUrl = true
  render(<Home />)
  const frame = screen.getByTitle('Custom Home Page')
  expect(frame.getAttribute('src')).toBe(state.content)
  expect(frame.getAttribute('sandbox')).not.toContain('allow-same-origin')
})
it('keeps loading state before home settings are resolved', () => {
  state.isLoaded = false
  render(<Home />)
  expect(screen.getByText('Loading...')).toBeTruthy()
  expect(screen.queryByRole('heading')).toBeNull()
})
