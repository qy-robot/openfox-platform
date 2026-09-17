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

import type { DownloadManifest } from '../types'

export function ReleaseStatus(props: {
  manifest: DownloadManifest | undefined
  error: Error | null
}) {
  const { t } = useTranslation()

  let label = t('Checking releases...')
  if (props.error != null) {
    label = t('Release status unavailable')
  } else if (props.manifest != null) {
    label = props.manifest.version ?? t('No release published yet')
  }

  return <p className='text-lg font-medium'>{label}</p>
}
