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
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { render, screen, cleanup } from '@testing-library/react'
import { createInstance } from 'i18next'
import { I18nextProvider } from 'react-i18next'
import { afterEach, beforeEach, expect, test, vi } from 'vitest'

import zh from '@/i18n/locales/zh.json'
import {
  DEFAULT_CURRENCY_CONFIG,
  useSystemConfigStore,
} from '@/stores/system-config-store'

import { usageLogSchema } from '../../data/schema'
import { useColumnsByCategory } from '../../lib/columns'
import { UsageLogsMobileList } from '../usage-logs-mobile-card'

vi.mock('@lobehub/icons', () => ({}))
vi.hoisted(() =>
  vi.stubGlobal('localStorage', {
    getItem: () => null,
    setItem: () => undefined,
    removeItem: () => undefined,
  })
)
const i18n = createInstance()
beforeEach(async () => {
  await i18n.init({ lng: 'zh', resources: { zh } })
  useSystemConfigStore
    .getState()
    .setConfig({ currency: DEFAULT_CURRENCY_CONFIG })
})
afterEach(cleanup)
function Preview(props: {
  mobile?: boolean
  name?: string
  quota?: number
  type?: number
  empty?: boolean
}) {
  const log = usageLogSchema.parse({
    id: 1,
    user_id: 1,
    created_at: 1700000000,
    type: props.type ?? 2,
    content: 'private billing details',
    model_name: 'gpt-test',
    token_name: props.name ?? 'robocoding desktop',
    quota: props.quota ?? 50000,
    group: 'private-group',
    prompt_tokens: 555,
    other: '{}',
  })
  // eslint-disable-next-line react/incompatible-library -- TanStack table fixture is not compiled.
  const table = useReactTable({
    data: props.empty ? [] : [log],
    columns: useColumnsByCategory('common', false, false),
    getCoreRowModel: getCoreRowModel(),
  })
  if (props.mobile) {
    return <UsageLogsMobileList table={table} logCategory='common' />
  }
  return (
    <table>
      <thead>
        {table.getHeaderGroups().map((group) => (
          <tr key={group.id}>
            {group.headers.map((header) => (
              <th key={header.id}>
                {flexRender(
                  header.column.columnDef.header,
                  header.getContext()
                )}
              </th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody>
        {table.getRowModel().rows.map((row) => (
          <tr key={row.id}>
            {row.getVisibleCells().map((cell) => (
              <td key={cell.id}>
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}
function show(props: Parameters<typeof Preview>[0]) {
  return render(
    <I18nextProvider i18n={i18n}>
      <Preview {...props} />
    </I18nextProvider>
  )
}
test.each([false, true])(
  'personal usage hides technical fields and currency (mobile %s)',
  (mobile) => {
    const { container } = show({ mobile })
    expect(screen.getByText('Robo Coding 客户端消耗')).toBeVisible()
    expect(screen.getByText('gpt-test')).toBeVisible()
    expect(screen.getByText('消耗点数')).toBeVisible()
    expect(screen.getByText('1')).toBeVisible()
    expect(container.textContent).not.toMatch(
      /¥|人民币|Token|private|555|robocoding desktop/
    )
    expect(screen.queryByRole('button')).toBeNull()
  }
)
test('API consumption is not mislabeled as desktop consumption', () => {
  show({ name: 'personal-api-key' })
  expect(screen.getByText('API 调用消耗')).toBeVisible()
  expect(screen.queryByText('Robo Coding 客户端消耗')).toBeNull()
  expect(screen.queryByText('personal-api-key')).toBeNull()
})
test('small nonzero charges retain point precision', () => {
  show({ quota: 1 })
  expect(screen.getByText('0.00002')).toBeVisible()
})
test('refunds preserve their event description and sign', () => {
  show({ type: 6, quota: -50000 })
  expect(screen.getByText('-1')).toBeVisible()
  expect(screen.queryByText('Robo Coding 客户端消耗')).toBeNull()
})
