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
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'

import { getDownloadManifest } from './api'
import { DownloadBuilds } from './components/download-builds'
import { ReleaseChangelog, ReleaseStatus } from './components/release-status'

export function Downloads() {
  const { t } = useTranslation()
  const manifestQuery = useQuery({
    queryKey: ['desktop-download-manifest'],
    queryFn: getDownloadManifest,
    retry: 1,
    staleTime: 5 * 60 * 1000,
  })

  return (
    <PublicLayout showMainContainer={false}>
      <main className='robo-download'>
        <section className='robo-download-hero'>
          <div className='robo-download-hero-grid' aria-hidden='true' />
          <div className='robo-download-hero-inner'>
            <p className='robo-download-kicker'>
              <span aria-hidden='true' />
              {t('RoboCoding Desktop')}
            </p>
            <h1>
              {t('Download')} <em>RoboCoding.</em>
            </h1>
            <div className='robo-download-hero-meta'>
              <span>{t('Latest version')}</span>
              <strong>
                <ReleaseStatus
                  manifest={manifestQuery.data}
                  error={manifestQuery.error}
                />
              </strong>
            </div>
          </div>
        </section>
        <section
          className='robo-download-builds'
          aria-labelledby='download-builds-title'
        >
          <div className='robo-download-section-heading'>
            <div>
              <p className='robo-download-eyebrow'>
                {t('Choose your environment')}
              </p>
              <h2 id='download-builds-title'>
                {t('Start where the work is.')}
              </h2>
            </div>
          </div>
          <DownloadBuilds
            manifest={manifestQuery.data}
            error={manifestQuery.error}
            onRetry={() => manifestQuery.refetch()}
          />
        </section>
        <ReleaseChangelog
          manifest={manifestQuery.data}
          error={manifestQuery.error}
        />
      </main>
      <Footer className='robo-footer' />
    </PublicLayout>
  )
}
