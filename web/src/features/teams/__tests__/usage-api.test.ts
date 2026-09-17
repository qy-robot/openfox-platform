/* Copyright (C) 2023-2026 QuantumNous; licensed under GNU AGPLv3 or later. */
import { afterEach, describe, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'

import { getTeamUsage } from '../api'
import type { TeamMember } from '../types'

afterEach(() => {
  vi.restoreAllMocks()
})

describe('getTeamUsage', () => {
  it('returns an empty member list when an unused team returns null', async () => {
    vi.spyOn(api, 'get').mockResolvedValue({
      data: {
        success: true,
        data: { month: '2026-09', total_usage_quota: 0, members: null },
      },
    })

    await expect(getTeamUsage(1, '2026-09')).resolves.toEqual({
      month: '2026-09',
      total_usage_quota: 0,
      members: [],
    })
  })

  it('preserves member usage and the requested team and month', async () => {
    const members: TeamMember[] = [
      {
        user_id: 7,
        username: 'operator',
        role: 'member',
        department: 'Robotics',
        monthly_limit_quota: 100,
        monthly_used_quota: 25,
      },
    ]
    const usage = { month: '2026-09', total_usage_quota: 25, members }
    const get = vi.spyOn(api, 'get').mockResolvedValue({
      data: { success: true, data: usage },
    })

    await expect(getTeamUsage(1, '2026-09')).resolves.toEqual(usage)
    expect(get).toHaveBeenCalledWith('/api/teams/1/usage', {
      params: { month: '2026-09' },
    })
  })

  it('rejects a failed response instead of reporting zero usage', async () => {
    vi.spyOn(api, 'get').mockResolvedValue({
      data: { success: false, message: 'Team access denied' },
    })

    await expect(getTeamUsage(1, '2026-09')).rejects.toThrow(
      'Team access denied'
    )
  })
})
