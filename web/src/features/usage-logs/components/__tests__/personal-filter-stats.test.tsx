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
import { getCoreRowModel, useReactTable } from '@tanstack/react-table'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, test, vi } from 'vitest'

import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { CommonLogsFilterBar } from '../common-logs-filter-bar'
import { UsageLogsProvider } from '../usage-logs-provider'

function FilterFixture() {
  // eslint-disable-next-line react/incompatible-library -- TanStack table fixture is not compiled.
  const table = useReactTable({
    data: [],
    columns: [],
    getCoreRowModel: getCoreRowModel(),
  })

  return (
    <UsageLogsProvider>
      <CommonLogsFilterBar table={table} />
    </UsageLogsProvider>
  )
}

async function renderFilter(initialEntry: string) {
  vi.spyOn(api, 'get').mockImplementation(async (url) => {
    if (url === '/api/group/') {
      return { data: { success: true, data: ['default'] } }
    }
    return {
      data: { success: true, data: { quota: 1, rpm: 12, tpm: 345 } },
    }
  })

  const root = createRootRoute()
  const authenticated = createRoute({
    getParentRoute: () => root,
    id: '_authenticated',
  })
  const logs = createRoute({
    getParentRoute: () => authenticated,
    path: '/usage-logs/$section',
    component: FilterFixture,
    validateSearch: (search: Record<string, unknown>) => search,
  })
  const router = createRouter({
    routeTree: root.addChildren([authenticated.addChildren([logs])]),
    history: createMemoryHistory({ initialEntries: [initialEntry] }),
  })
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  render(
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
  await screen.findByPlaceholderText('Model Name')

  return router
}

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  useAuthStore.getState().auth.setUser(null)
})

test('personal view shows only date and model filters with precise point usage', async () => {
  const router = await renderFilter(
    '/usage-logs/common?model=gpt-test&token=secret&group=private&type=%5B%222%22%5D&requestId=req-1&upstreamRequestId=up-1'
  )

  expect(await screen.findByText('Points used')).toBeVisible()
  expect(screen.getByPlaceholderText('Model Name')).toHaveValue('gpt-test')
  expect(screen.getByText('0.00002')).toBeVisible()
  for (const label of ['Group', 'Type', 'Hide', 'Show']) {
    expect(
      screen.queryByRole('button', { name: label })
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('combobox', { name: label })
    ).not.toBeInTheDocument()
  }
  for (const placeholder of [
    'Token Name',
    'Username',
    'Channel ID',
    'Request ID',
    'Upstream Request ID',
  ]) {
    expect(screen.queryByPlaceholderText(placeholder)).not.toBeInTheDocument()
  }
  expect(screen.queryByText('RPM')).not.toBeInTheDocument()
  expect(screen.queryByText('TPM')).not.toBeInTheDocument()
  expect(api.get).not.toHaveBeenCalledWith('/api/user/self/groups')
  const statsRequest = vi
    .mocked(api.get)
    .mock.calls.find(([url]) => url.startsWith('/api/log/self/stat?'))?.[0]
  expect(statsRequest).toContain('model_name=gpt-test')
  expect(statsRequest).not.toMatch(
    /token|group|type|request_id|upstream_request_id/
  )

  await userEvent.click(screen.getByRole('button', { name: 'Search' }))
  await waitFor(() => {
    expect(router.state.location.search).toMatchObject({
      model: 'gpt-test',
      page: 1,
    })
  })
  for (const key of [
    'token',
    'group',
    'type',
    'requestId',
    'upstreamRequestId',
  ]) {
    expect(router.state.location.search).not.toHaveProperty(key)
  }

  await userEvent.click(screen.getByRole('button', { name: 'Reset' }))
  await waitFor(() =>
    expect(router.state.location.search).not.toHaveProperty('model')
  )
})

test('administrator view keeps technical filters, visibility control, and rate stats', async () => {
  useAuthStore.getState().auth.setUser({ id: 1, username: 'admin', role: 10 })
  await renderFilter('/usage-logs/common')

  expect(await screen.findByText('Usage')).toBeVisible()
  expect(screen.getByRole('combobox', { name: 'Group' })).toBeVisible()
  expect(screen.getByRole('combobox', { name: 'Type' })).toBeVisible()
  expect(screen.getByRole('button', { name: 'Hide' })).toBeVisible()
  expect(screen.getByText('RPM')).toBeVisible()
  expect(screen.getByText('12')).toBeVisible()
  expect(screen.getByText('TPM')).toBeVisible()
  expect(screen.getByText('345')).toBeVisible()
})
