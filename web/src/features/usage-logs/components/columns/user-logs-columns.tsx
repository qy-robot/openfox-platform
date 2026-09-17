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
import type { ColumnDef } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'

import { formatPointsFromQuota } from '@/lib/currency'
import { formatTimestampToDate } from '@/lib/format'

import type { UsageLog } from '../../data/schema'
import { getLogTypeConfig } from '../../lib/utils'

/** A deliberately small record of personal consumption, without billing internals. */
export function useUserLogsColumns(): ColumnDef<UsageLog>[] {
  const { t } = useTranslation()
  return [
    {
      accessorKey: 'created_at',
      header: t('Time'),
      cell: ({ row }) => formatTimestampToDate(row.original.created_at),
      enableHiding: false,
    },
    {
      accessorKey: 'model_name',
      header: t('Model'),
      cell: ({ row }) => (
        <span className='break-all'>{row.original.model_name || '—'}</span>
      ),
      enableHiding: false,
    },
    {
      accessorKey: 'quota',
      header: t('Points used'),
      cell: ({ row }) =>
        formatPointsFromQuota(row.original.quota, {
          maximumFractionDigits: 8,
        }),
      enableHiding: false,
    },
    {
      id: 'source',
      header: t('Source'),
      cell: ({ row }) => {
        const log = row.original
        if (log.type !== 2) return t(getLogTypeConfig(log.type).label)
        if (log.token_name === 'robocoding desktop') {
          return t('Robo Coding client usage')
        }
        return log.token_name ? t('API usage') : t('Model usage')
      },
      enableHiding: false,
    },
  ]
}
