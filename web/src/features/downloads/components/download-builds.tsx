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
import { useTranslation } from 'react-i18next'

import { ErrorState } from '@/components/error-state'
import { LoadingState } from '@/components/loading-state'

import { DOWNLOAD_TARGET_IDS } from '../lib/download-manifest'
import type { DownloadManifest } from '../types'
import { DownloadCard } from './download-card'

type DownloadBuildsProps = {
  manifest: DownloadManifest | undefined
  error: Error | null
  onRetry: () => void
}

export function DownloadBuilds(props: DownloadBuildsProps) {
  const { t } = useTranslation()

  if (props.error != null) {
    return (
      <ErrorState
        className='border-border/60 bg-card/60 min-h-80 rounded-2xl border'
        title={t('Downloads are temporarily unavailable')}
        description={t(
          'The release list could not be loaded. Check your connection and try again.'
        )}
        onRetry={props.onRetry}
      />
    )
  }

  if (props.manifest == null) {
    return (
      <LoadingState
        className='min-h-80'
        message={t('Checking available builds...')}
      />
    )
  }

  const targetMap = new Map(
    props.manifest.downloads.map((target) => [target.id, target])
  )
  return (
    <div className='grid gap-4 md:grid-cols-2 xl:grid-cols-4'>
      {DOWNLOAD_TARGET_IDS.map((targetId) => {
        const target = targetMap.get(targetId)
        if (target == null) return null
        return <DownloadCard key={targetId} target={target} />
      })}
    </div>
  )
}
