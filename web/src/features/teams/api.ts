/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.
*/
import { api } from '@/lib/api'

import type {
  ApiEnvelope,
  TeamDetail,
  TeamInvite,
  TeamJoinRequest,
  TeamMember,
  TeamMembership,
  TeamUsage,
} from './types'

async function data<T>(request: Promise<{ data: ApiEnvelope<T> }>): Promise<T> {
  const response = (await request).data
  if (!response.success || response.data === undefined) {
    throw new Error(response.message || 'Request failed')
  }
  return response.data
}

export const getMyTeams = () =>
  data<TeamMembership[]>(api.get('/api/teams/self'))
export const createTeam = (name: string) =>
  data<TeamDetail>(api.post('/api/teams', { name }))
export const joinTeam = (code: string) =>
  data<unknown>(api.post('/api/teams/join', { code }))
export const joinTeamInvite = (token: string) =>
  data<{ team_id: number }>(api.post('/api/teams/join/invite', { token }))
export const getTeam = (id: number) =>
  data<TeamDetail>(api.get(`/api/teams/${id}`))
export const getTeamMembers = (id: number) =>
  data<TeamMember[]>(api.get(`/api/teams/${id}/members`))
export const updateTeamMember = (
  id: number,
  userId: number,
  payload: Partial<
    Pick<TeamMember, 'role' | 'department' | 'monthly_limit_quota'>
  >
) => data<TeamMember>(api.patch(`/api/teams/${id}/members/${userId}`, payload))
export const removeTeamMember = (id: number, userId: number) =>
  data<unknown>(api.delete(`/api/teams/${id}/members/${userId}`))
export const fundTeam = (id: number, quota: number, requestId: string) =>
  data<{ balance_quota: number; request_id: string }>(
    api.post(`/api/teams/${id}/fund`, { quota, request_id: requestId })
  )
export const getTeamUsage = async (
  id: number,
  month: string
): Promise<TeamUsage> => {
  const usage = await data<
    Omit<TeamUsage, 'members'> & { members: TeamMember[] | null }
  >(api.get(`/api/teams/${id}/usage`, { params: { month } }))
  return { ...usage, members: usage.members ?? [] }
}
export const getTeamInvites = (id: number) =>
  data<TeamInvite[]>(api.get(`/api/teams/${id}/invites`))
export const createTeamInvite = (
  id: number,
  expiresInSeconds: number,
  maxUses: number
) =>
  data<TeamInvite>(
    api.post(`/api/teams/${id}/invites`, {
      expires_in_seconds: expiresInSeconds,
      max_uses: maxUses,
    })
  )
export const revokeTeamInvite = (id: number, inviteId: number) =>
  data<unknown>(api.delete(`/api/teams/${id}/invites/${inviteId}`))
export const getTeamJoinRequests = (id: number) =>
  data<TeamJoinRequest[]>(api.get(`/api/teams/${id}/join-requests`))
export const decideTeamJoinRequest = (
  id: number,
  requestId: number,
  approve: boolean
) =>
  data<unknown>(
    approve
      ? api.post(`/api/teams/${id}/join-requests/${requestId}/approve`)
      : api.delete(`/api/teams/${id}/join-requests/${requestId}`)
  )
