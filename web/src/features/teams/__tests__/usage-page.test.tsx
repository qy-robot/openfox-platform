/* Copyright (C) 2023-2026 QuantumNous; licensed under GNU AGPLv3 or later. */
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import {
  createMemoryHistory,
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
} from '@tanstack/react-router'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { GeneralError } from '@/features/errors/general-error'
import { api } from '@/lib/api'

import { TeamWorkspace } from '..'
import type { TeamMember } from '../types'

const clients: QueryClient[] = []

afterEach(() => {
  cleanup()
  clients.splice(0).forEach((client) => client.clear())
  vi.restoreAllMocks()
})

async function renderUsage(members: TeamMember[] | null) {
  vi.spyOn(api, 'get').mockImplementation(async (url) => {
    if (url === '/api/teams/1') {
      return {
        data: {
          success: true,
          data: { id: 1, name: 'Robotics', role: 'owner' },
        },
      }
    }
    if (url === '/api/teams/1/usage') {
      return {
        data: {
          success: true,
          data: {
            month: '2026-09',
            total_usage_quota: members === null ? 0 : 50000,
            members,
          },
        },
      }
    }
    throw new Error(`Unexpected request: ${url}`)
  })
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  clients.push(client)
  const root = createRootRoute({ errorComponent: GeneralError })
  const usage = createRoute({
    getParentRoute: () => root,
    path: '/teams/$teamId/$section',
    component: () => <TeamWorkspace teamId={1} section='usage' />,
  })
  const router = createRouter({
    routeTree: root.addChildren([usage]),
    history: createMemoryHistory({ initialEntries: ['/teams/1/usage'] }),
  })
  await router.load()
  render(
    <QueryClientProvider client={client}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
  await waitFor(() => {
    expect(
      client.getQueryState([
        'teams',
        1,
        'usage',
        new Date().toISOString().slice(0, 7),
      ])?.status
    ).toBe('success')
  })
}

describe('team usage page', () => {
  it('keeps the usage page visible when the current month has no members', async () => {
    await renderUsage(null)

    expect(await screen.findByText('Team usage this month')).toBeVisible()
    expect(screen.getByText('0 points')).toBeVisible()
    expect(screen.queryByText('500')).not.toBeInTheDocument()
  })

  it('shows member consumption when the current month has usage', async () => {
    await renderUsage([
      {
        user_id: 7,
        username: 'operator',
        display_name: 'Robot operator',
        role: 'member',
        department: 'Lab',
        monthly_limit_quota: 100000,
        monthly_used_quota: 50000,
      },
    ])

    expect(await screen.findByText('Robot operator')).toBeVisible()
    expect(screen.getByText('Lab')).toBeVisible()
    expect(screen.getAllByText('1 points')).toHaveLength(2)
    expect(screen.queryByText('500')).not.toBeInTheDocument()
  })
})
