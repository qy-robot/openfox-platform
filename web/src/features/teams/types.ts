/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.
*/
export type TeamRole = 'owner' | 'admin' | 'member'

export interface TeamMembership {
  id: number
  name: string
  role: TeamRole
  department: string
  balance_quota: number
  monthly_limit_quota: number
  monthly_used_quota: number
  month: string
}

export interface TeamDetail {
  id: number
  name: string
  join_code?: string
  balance_quota: number
  role: TeamRole
  member_count: number
  created_time: number
}

export interface TeamMember {
  user_id: number
  username: string
  display_name?: string
  role: TeamRole
  department: string
  monthly_limit_quota: number
  monthly_used_quota: number
}

export interface TeamInvite {
  id: number
  token?: string
  expires_at: number
  max_uses: number
  used_count: number
}

export interface TeamUsage {
  month: string
  total_usage_quota: number
  members: TeamMember[]
}

export interface TeamJoinRequest {
  id: number
  user_id: number
  status: string
  created_time: number
}

export interface ApiEnvelope<T> {
  success: boolean
  message?: string
  data?: T
}
