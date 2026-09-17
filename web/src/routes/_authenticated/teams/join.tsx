/* Copyright (C) 2023-2026 QuantumNous; licensed under GNU AGPLv3 or later. */
import { useMutation } from '@tanstack/react-query'
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { CheckCircle2, Users } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { joinTeamInvite } from '@/features/teams/api'
import { handleServerError } from '@/lib/handle-server-error'

export const Route = createFileRoute('/_authenticated/teams/join')({
  validateSearch: z.object({ token: z.string().min(1) }),
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const search = Route.useSearch()
  const navigate = useNavigate()
  const join = useMutation({
    mutationFn: () => joinTeamInvite(search.token),
    onSuccess: (team) =>
      void navigate({
        to: '/teams/$teamId/$section',
        params: { teamId: String(team.team_id), section: 'overview' },
      }),
    onError: (error) => handleServerError(error),
  })
  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Team invitation')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <Card className='mx-auto mt-12 max-w-lg'>
          <CardHeader className='text-center'>
            <Users className='text-primary mx-auto size-10' />
            <CardTitle>{t('Join this team?')}</CardTitle>
            <CardDescription>
              {t(
                'Your personal account stays separate. You can choose team or personal payment for each API key.'
              )}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              className='w-full'
              disabled={join.isPending}
              onClick={() => join.mutate()}
            >
              <CheckCircle2 />
              {join.isPending ? t('Joining...') : t('Accept invitation')}
            </Button>
          </CardContent>
        </Card>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
