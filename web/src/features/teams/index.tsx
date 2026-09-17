/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.
*/
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link, useNavigate } from '@tanstack/react-router'
import { Building2, Copy, Plus, UserPlus, Users } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { EmptyState } from '@/components/empty-state'
import { SectionPageLayout } from '@/components/layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Progress } from '@/components/ui/progress'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { copyToClipboard } from '@/lib/copy-to-clipboard'
import {
  formatPointsFromQuota,
  pointsToQuota,
  quotaToEditablePoints,
} from '@/lib/currency'
import { handleServerError } from '@/lib/handle-server-error'

import {
  createTeam,
  createTeamInvite,
  decideTeamJoinRequest,
  fundTeam,
  getMyTeams,
  getTeam,
  getTeamInvites,
  getTeamJoinRequests,
  getTeamMembers,
  getTeamUsage,
  joinTeam,
  removeTeamMember,
  revokeTeamInvite,
  updateTeamMember,
} from './api'
import type { TeamMember } from './types'

const sections = ['overview', 'members', 'usage', 'settings'] as const
export type TeamSection = (typeof sections)[number]

function Points(props: { quota: number; className?: string }) {
  const { t } = useTranslation()
  return (
    <span className={props.className}>
      {t('{{count}} points', {
        count: formatPointsFromQuota(props.quota),
      })}
    </span>
  )
}

function MemberLimitInput(props: {
  member: TeamMember
  disabled: boolean
  onSave: (quota: number) => void
}) {
  const { t } = useTranslation()
  const [value, setValue] = useState(
    String(quotaToEditablePoints(props.member.monthly_limit_quota))
  )
  const [dirty, setDirty] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!dirty) {
      setValue(String(quotaToEditablePoints(props.member.monthly_limit_quota)))
    }
  }, [dirty, props.member.monthly_limit_quota])

  const id = `member-limit-${props.member.user_id}`
  const errorId = `${id}-error`
  return (
    <div className='space-y-1.5'>
      <Label htmlFor={id}>{t('Monthly point limit')}</Label>
      <Input
        id={id}
        aria-label={t('Monthly point limit for {{name}}', {
          name: props.member.username,
        })}
        aria-invalid={Boolean(error) || undefined}
        aria-describedby={error ? errorId : undefined}
        disabled={props.disabled}
        type='number'
        min={0}
        value={value}
        onChange={(event) => {
          setValue(event.target.value)
          setDirty(true)
          setError('')
        }}
        onBlur={() => {
          if (!dirty) return
          const points = Number(value)
          if (value.trim() === '' || !Number.isFinite(points) || points < 0) {
            setError(t('Enter a finite point limit of zero or greater.'))
            return
          }
          const quota = pointsToQuota(points)
          setDirty(false)
          setError('')
          if (quota !== props.member.monthly_limit_quota) props.onSave(quota)
        }}
      />
      {error && (
        <p id={errorId} role='alert' className='text-destructive text-xs'>
          {error}
        </p>
      )}
    </div>
  )
}

export function TeamsHome() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [code, setCode] = useState('')
  const teams = useQuery({ queryKey: ['teams', 'self'], queryFn: getMyTeams })
  const create = useMutation({
    mutationFn: () => createTeam(name.trim()),
    onSuccess: async (team) => {
      await queryClient.invalidateQueries({ queryKey: ['teams'] })
      void navigate({
        to: '/teams/$teamId/$section',
        params: { teamId: String(team.id), section: 'overview' },
      })
    },
    onError: (error) => handleServerError(error),
  })
  const join = useMutation({
    mutationFn: () => joinTeam(code.trim()),
    onSuccess: async () => {
      toast.success(t('Join request sent'))
      setCode('')
      await queryClient.invalidateQueries({ queryKey: ['teams'] })
    },
    onError: (error) => handleServerError(error),
  })

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Teams')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto grid max-w-6xl gap-4 lg:grid-cols-[minmax(0,1fr)_340px]'>
          <div className='space-y-3'>
            {teams.data?.map((team) => (
              <Link
                key={team.id}
                to='/teams/$teamId/$section'
                params={{ teamId: String(team.id), section: 'overview' }}
                className='block'
              >
                <Card className='hover:bg-muted/30 transition-colors'>
                  <CardHeader>
                    <CardTitle className='flex items-center justify-between gap-3'>
                      <span className='flex items-center gap-2'>
                        <Building2 className='text-primary size-4' />
                        {team.name}
                      </span>
                      <Badge variant='secondary'>{t(team.role)}</Badge>
                    </CardTitle>
                    <CardDescription>
                      {team.department || t('No department')}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className='grid gap-3 sm:grid-cols-3'>
                    <div>
                      <div className='text-muted-foreground text-xs'>
                        {t('Team balance')}
                      </div>
                      <Points
                        quota={team.balance_quota}
                        className='text-lg font-semibold tabular-nums'
                      />
                    </div>
                    <div>
                      <div className='text-muted-foreground text-xs'>
                        {t('Used this month')}
                      </div>
                      <Points
                        quota={team.monthly_used_quota}
                        className='text-lg font-semibold tabular-nums'
                      />
                    </div>
                    <div>
                      <div className='text-muted-foreground text-xs'>
                        {t('My monthly limit')}
                      </div>
                      <Points
                        quota={team.monthly_limit_quota}
                        className='text-lg font-semibold tabular-nums'
                      />
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
            {!teams.isLoading && teams.data?.length === 0 && (
              <EmptyState
                icon={Users}
                title={t('No teams yet')}
                description={t('Create a team or join one with a team code.')}
                bordered
              />
            )}
          </div>
          <div className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Create a team')}</CardTitle>
                <CardDescription>
                  {t(
                    'Invite people and pay for their usage from one shared balance.'
                  )}
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-3'>
                <Label htmlFor='team-name'>{t('Team name')}</Label>
                <Input
                  id='team-name'
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
                <Button
                  className='w-full'
                  disabled={!name.trim() || create.isPending}
                  onClick={() => create.mutate()}
                >
                  <Plus />
                  {t('Create team')}
                </Button>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>{t('Join a team')}</CardTitle>
                <CardDescription>
                  {t(
                    'Enter a team code. An administrator will review your request.'
                  )}
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-3'>
                <Label htmlFor='team-code'>{t('Team code')}</Label>
                <Input
                  id='team-code'
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                />
                <Button
                  variant='outline'
                  className='w-full'
                  disabled={!code.trim() || join.isPending}
                  onClick={() => join.mutate()}
                >
                  <UserPlus />
                  {t('Request to join')}
                </Button>
              </CardContent>
            </Card>
          </div>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}

export function TeamWorkspace(props: { teamId: number; section: TeamSection }) {
  const { t } = useTranslation()
  const team = useQuery({
    queryKey: ['teams', props.teamId],
    queryFn: () => getTeam(props.teamId),
  })
  const labels: Record<TeamSection, string> = {
    overview: t('Overview'),
    members: t('Members'),
    usage: t('Usage'),
    settings: t('Settings'),
  }
  const visibleSections =
    team.data?.role === 'member' ? sections.slice(0, 1) : sections
  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {team.data?.name || t('Team')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='mx-auto max-w-6xl space-y-5'>
          <nav
            className='flex gap-1 overflow-x-auto border-b'
            aria-label={t('Team navigation')}
          >
            {visibleSections.map((section) => (
              <Link
                key={section}
                to='/teams/$teamId/$section'
                params={{ teamId: String(props.teamId), section }}
                className={`px-3 py-2 text-sm font-medium ${props.section === section ? 'border-primary text-foreground border-b-2' : 'text-muted-foreground'}`}
              >
                {labels[section]}
              </Link>
            ))}
          </nav>
          {team.data && props.section === 'overview' && (
            <TeamOverview team={team.data} />
          )}
          {team.data &&
            team.data.role !== 'member' &&
            props.section === 'members' && (
              <TeamMembers teamId={props.teamId} canManage />
            )}
          {team.data &&
            team.data.role !== 'member' &&
            props.section === 'usage' && (
              <TeamUsageView teamId={props.teamId} />
            )}
          {team.data &&
            team.data.role !== 'member' &&
            props.section === 'settings' && <TeamSettings team={team.data} />}
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}

function TeamOverview(props: {
  team: {
    id: number
    balance_quota: number
    member_count: number
    join_code?: string
    role: string
  }
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [points, setPoints] = useState(100)
  const fund = useMutation({
    mutationFn: (input: { points: number; requestId: string }) =>
      fundTeam(props.team.id, pointsToQuota(input.points), input.requestId),
    onSuccess: async () => {
      toast.success(t('Team balance updated'))
      await queryClient.invalidateQueries({ queryKey: ['teams'] })
    },
    onError: (error) => handleServerError(error),
  })
  return (
    <div className='grid gap-4 md:grid-cols-3'>
      <Card>
        <CardHeader>
          <CardDescription>{t('Available team balance')}</CardDescription>
          <CardTitle className='text-3xl tabular-nums'>
            <Points quota={props.team.balance_quota} />
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader>
          <CardDescription>{t('Members')}</CardDescription>
          <CardTitle className='text-3xl tabular-nums'>
            {props.team.member_count}
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader>
          <CardDescription>{t('Team code')}</CardDescription>
          <CardTitle className='font-mono'>
            {props.team.join_code || t('Hidden')}
          </CardTitle>
        </CardHeader>
      </Card>
      {props.team.role !== 'member' && (
        <Card className='md:col-span-3'>
          <CardHeader>
            <CardTitle>{t('Add team balance')}</CardTitle>
            <CardDescription>
              {t(
                'Transfer points from your personal balance to this team. Transfers belong to the team after completion.'
              )}
            </CardDescription>
          </CardHeader>
          <CardContent className='flex flex-col gap-2 sm:flex-row'>
            <Input
              aria-label={t('Points to transfer')}
              type='number'
              min={1}
              value={points}
              onChange={(e) => setPoints(Number(e.target.value))}
            />
            <Button
              disabled={points <= 0 || fund.isPending}
              onClick={() =>
                fund.mutate({
                  points,
                  requestId: globalThis.crypto.randomUUID(),
                })
              }
            >
              {t('Transfer {{count}} points', { count: points })}
            </Button>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

function TeamMembers(props: { teamId: number; canManage: boolean }) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const members = useQuery({
    queryKey: ['teams', props.teamId, 'members'],
    queryFn: () => getTeamMembers(props.teamId),
  })
  const joinRequests = useQuery({
    queryKey: ['teams', props.teamId, 'join-requests'],
    queryFn: () => getTeamJoinRequests(props.teamId),
    enabled: props.canManage,
  })
  const decideJoin = useMutation({
    mutationFn: (input: { requestId: number; approve: boolean }) =>
      decideTeamJoinRequest(props.teamId, input.requestId, input.approve),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ['teams', props.teamId, 'join-requests'],
        }),
        queryClient.invalidateQueries({
          queryKey: ['teams', props.teamId, 'members'],
        }),
      ])
    },
    onError: (error) => handleServerError(error),
  })
  const update = useMutation({
    mutationFn: (input: {
      member: TeamMember
      role?: TeamMember['role']
      department?: string
      monthlyLimitQuota?: number
    }) =>
      updateTeamMember(props.teamId, input.member.user_id, {
        role: input.role,
        department: input.department,
        monthly_limit_quota: input.monthlyLimitQuota,
      }),
    onSuccess: async () =>
      queryClient.invalidateQueries({
        queryKey: ['teams', props.teamId, 'members'],
      }),
    onError: (error) => handleServerError(error),
  })
  const remove = useMutation({
    mutationFn: (userId: number) => removeTeamMember(props.teamId, userId),
    onSuccess: async () =>
      queryClient.invalidateQueries({
        queryKey: ['teams', props.teamId, 'members'],
      }),
    onError: (error) => handleServerError(error),
  })
  return (
    <div className='space-y-4'>
      {props.canManage && (joinRequests.data?.length || 0) > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>{t('Join requests')}</CardTitle>
            <CardDescription>
              {t('Review people who entered the team code.')}
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-2'>
            {joinRequests.data?.map((request) => (
              <div
                key={request.id}
                className='flex items-center gap-3 rounded-lg border p-3'
              >
                <div className='min-w-0 flex-1'>
                  <div className='font-medium'>
                    {t('User {{id}}', { id: request.user_id })}
                  </div>
                  <div className='text-muted-foreground text-xs'>
                    {t('Pending approval')}
                  </div>
                </div>
                <Button
                  size='sm'
                  variant='outline'
                  onClick={() =>
                    decideJoin.mutate({ requestId: request.id, approve: false })
                  }
                >
                  {t('Reject')}
                </Button>
                <Button
                  size='sm'
                  onClick={() =>
                    decideJoin.mutate({ requestId: request.id, approve: true })
                  }
                >
                  {t('Approve')}
                </Button>
              </div>
            ))}
          </CardContent>
        </Card>
      )}
      <Card>
        <CardHeader>
          <CardTitle>{t('Members')}</CardTitle>
          <CardDescription>
            {t(
              'Set roles, departments, and monthly team spending limits. New requests stop at the limit; requests already in progress settle their actual usage.'
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-3'>
          {members.data?.map((member) => {
            const percent =
              member.monthly_limit_quota > 0
                ? Math.min(
                    100,
                    (member.monthly_used_quota / member.monthly_limit_quota) *
                      100
                  )
                : 0
            return (
              <div
                key={member.user_id}
                className='grid gap-3 rounded-lg border p-3 md:grid-cols-[minmax(140px,1fr)_140px_140px_150px_auto] md:items-center'
              >
                <div>
                  <div className='font-medium'>
                    {member.display_name || member.username}
                  </div>
                  <div className='text-muted-foreground text-xs'>
                    <Points quota={member.monthly_used_quota} /> /{' '}
                    <Points quota={member.monthly_limit_quota} />
                  </div>
                  <Progress value={percent} className='mt-2 h-1.5' />
                </div>
                <div className='space-y-1.5'>
                  <Label htmlFor={`member-department-${member.user_id}`}>
                    {t('Department')}
                  </Label>
                  <Input
                    id={`member-department-${member.user_id}`}
                    aria-label={t('Department for {{name}}', {
                      name: member.username,
                    })}
                    disabled={!props.canManage}
                    defaultValue={member.department}
                    onBlur={(e) => {
                      if (e.target.value !== member.department) {
                        update.mutate({ member, department: e.target.value })
                      }
                    }}
                  />
                </div>
                <MemberLimitInput
                  member={member}
                  disabled={!props.canManage}
                  onSave={(quota) =>
                    update.mutate({ member, monthlyLimitQuota: quota })
                  }
                />
                <div className='space-y-1.5'>
                  <Label>{t('Role')}</Label>
                  <Select
                    value={member.role}
                    disabled={!props.canManage || member.role === 'owner'}
                    onValueChange={(value) =>
                      update.mutate({
                        member,
                        role: value as TeamMember['role'],
                      })
                    }
                  >
                    <SelectTrigger
                      aria-label={t('Role for {{name}}', {
                        name: member.username,
                      })}
                      className='w-full'
                    >
                      <SelectValue>{t(member.role)}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value='member'>{t('member')}</SelectItem>
                      <SelectItem value='admin'>{t('admin')}</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                {props.canManage && member.role !== 'owner' ? (
                  <Button
                    variant='ghost'
                    size='sm'
                    onClick={() => remove.mutate(member.user_id)}
                  >
                    {t('Remove')}
                  </Button>
                ) : (
                  <span />
                )}
              </div>
            )
          })}
        </CardContent>
      </Card>
    </div>
  )
}

function TeamUsageView(props: { teamId: number }) {
  const { t } = useTranslation()
  const month = new Date().toISOString().slice(0, 7)
  const usage = useQuery({
    queryKey: ['teams', props.teamId, 'usage', month],
    queryFn: () => getTeamUsage(props.teamId, month),
  })
  return (
    <div className='space-y-4'>
      <Card>
        <CardHeader>
          <CardDescription>{t('Team usage this month')}</CardDescription>
          <CardTitle className='text-3xl'>
            <Points quota={usage.data?.total_usage_quota || 0} />
          </CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardContent className='space-y-3'>
          {usage.data?.members.map((member) => (
            <div
              key={member.user_id}
              className='flex items-center justify-between border-b py-2 last:border-0'
            >
              <div>
                <div className='font-medium'>
                  {member.display_name || member.username}
                </div>
                <div className='text-muted-foreground text-xs'>
                  {member.department || t('No department')}
                </div>
              </div>
              <Points
                quota={member.monthly_used_quota}
                className='font-mono font-medium'
              />
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}

function TeamSettings(props: {
  team: { id: number; join_code?: string; role: string }
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [expiresDays, setExpiresDays] = useState(7)
  const [maxUses, setMaxUses] = useState(20)
  const [freshInvite, setFreshInvite] = useState<Awaited<
    ReturnType<typeof createTeamInvite>
  > | null>(null)
  const invites = useQuery({
    queryKey: ['teams', props.team.id, 'invites'],
    queryFn: () => getTeamInvites(props.team.id),
    enabled: props.team.role !== 'member',
  })
  const create = useMutation({
    mutationFn: () =>
      createTeamInvite(props.team.id, expiresDays * 24 * 60 * 60, maxUses),
    onSuccess: async (invite) => {
      setFreshInvite(invite)
      queryClient.invalidateQueries({
        queryKey: ['teams', props.team.id, 'invites'],
      })
    },
    onError: (error) => handleServerError(error),
  })
  const revoke = useMutation({
    mutationFn: (id: number) => revokeTeamInvite(props.team.id, id),
    onSuccess: async () =>
      queryClient.invalidateQueries({
        queryKey: ['teams', props.team.id, 'invites'],
      }),
    onError: (error) => handleServerError(error),
  })
  return (
    <div className='space-y-4'>
      <Card>
        <CardHeader>
          <CardTitle>{t('Join code')}</CardTitle>
          <CardDescription>
            {t(
              'People who enter this code send a request that an administrator must approve.'
            )}
          </CardDescription>
        </CardHeader>
        <CardContent className='flex items-center gap-2'>
          <code className='bg-muted flex-1 rounded-md px-3 py-2'>
            {props.team.join_code ||
              t('Only administrators can view this code')}
          </code>
          {props.team.join_code && (
            <Button
              variant='outline'
              size='icon'
              aria-label={t('Copy team code')}
              onClick={() => copyToClipboard(props.team.join_code || '')}
            >
              <Copy />
            </Button>
          )}
        </CardContent>
      </Card>
      {props.team.role !== 'member' && (
        <Card>
          <CardHeader>
            <CardTitle>{t('Invite links')}</CardTitle>
            <CardDescription>
              {t('Links expire after 7 days and can be revoked at any time.')}
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-3'>
            <div className='grid gap-3 sm:grid-cols-2'>
              <div className='space-y-1.5'>
                <Label htmlFor='invite-days'>{t('Valid for days')}</Label>
                <Input
                  id='invite-days'
                  type='number'
                  min={1}
                  max={90}
                  value={expiresDays}
                  onChange={(event) =>
                    setExpiresDays(Number(event.target.value))
                  }
                />
              </div>
              <div className='space-y-1.5'>
                <Label htmlFor='invite-uses'>{t('Maximum uses')}</Label>
                <Input
                  id='invite-uses'
                  type='number'
                  min={1}
                  max={1000}
                  value={maxUses}
                  onChange={(event) => setMaxUses(Number(event.target.value))}
                />
              </div>
            </div>
            <Button onClick={() => create.mutate()} disabled={create.isPending}>
              <Plus />
              {t('Create invite link')}
            </Button>
            {[
              ...(freshInvite ? [freshInvite] : []),
              ...(invites.data || []).filter(
                (invite) => invite.id !== freshInvite?.id
              ),
            ].map((invite) => {
              const url = invite.token
                ? `${window.location.origin}/teams/join?token=${encodeURIComponent(invite.token)}`
                : ''
              return (
                <div
                  key={invite.id}
                  className='flex flex-col gap-2 rounded-lg border p-3 sm:flex-row sm:items-center'
                >
                  <div className='min-w-0 flex-1'>
                    <code className='block truncate text-xs'>
                      {url || t('Existing invite link')}
                    </code>
                    <div className='text-muted-foreground mt-1 text-xs'>
                      {t('{{used}} of {{max}} uses', {
                        used: invite.used_count,
                        max: invite.max_uses,
                      })}
                    </div>
                  </div>
                  {url && (
                    <Button
                      variant='outline'
                      size='sm'
                      onClick={() => copyToClipboard(url)}
                    >
                      {t('Copy')}
                    </Button>
                  )}
                  <Button
                    variant='ghost'
                    size='sm'
                    onClick={() => revoke.mutate(invite.id)}
                  >
                    {t('Revoke')}
                  </Button>
                </div>
              )
            })}
          </CardContent>
        </Card>
      )}
    </div>
  )
}
