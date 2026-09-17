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
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { Building2, ChevronRight, Plus } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { getMyTeams } from '@/features/teams/api'
import { formatPointsFromQuota } from '@/lib/currency'

export function TeamsCard() {
  const { t } = useTranslation()
  const teams = useQuery({ queryKey: ['teams', 'self'], queryFn: getMyTeams })
  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2'>
          <Building2 className='text-primary size-4' />
          {t('Teams and payment')}
        </CardTitle>
        <CardDescription>
          {t(
            'Your personal account stays active when you create or join teams.'
          )}
        </CardDescription>
      </CardHeader>
      <CardContent className='space-y-2'>
        {teams.data?.map((team) => (
          <Link
            key={team.id}
            to='/teams/$teamId/$section'
            params={{ teamId: String(team.id), section: 'overview' }}
            className='hover:bg-muted flex items-center gap-3 rounded-lg border p-3'
          >
            <div className='min-w-0 flex-1'>
              <div className='truncate font-medium'>{team.name}</div>
              <div className='text-muted-foreground text-xs'>
                {t('{{count}} points available', {
                  count: formatPointsFromQuota(team.balance_quota),
                })}
              </div>
            </div>
            <ChevronRight className='text-muted-foreground size-4' />
          </Link>
        ))}
        <Button
          variant='outline'
          className='mt-2 w-full'
          render={<Link to='/teams' />}
        >
          <Plus />
          {t('Create or join a team')}
        </Button>
      </CardContent>
    </Card>
  )
}
