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
      <main className='min-h-[calc(100vh-9rem)]'>
        <section className='relative overflow-hidden px-5 pt-24 pb-14 sm:px-8 md:pt-28 md:pb-20'>
          <div
            aria-hidden='true'
            className='absolute inset-x-0 top-0 -z-10 h-96 bg-[radial-gradient(ellipse_at_50%_-10%,rgba(59,130,246,0.18),transparent_68%)] dark:bg-[radial-gradient(ellipse_at_50%_-10%,rgba(37,99,235,0.25),transparent_68%)]'
          />
          <div className='mx-auto max-w-6xl'>
            <div className='mb-9 flex flex-col gap-4 sm:mb-11 sm:flex-row sm:items-end sm:justify-between'>
              <div>
                <p className='mb-3 text-sm font-medium tracking-wide text-blue-600 dark:text-blue-400'>
                  {t('RoboCoding Desktop')}
                </p>
                <h1 className='text-4xl font-semibold tracking-[-0.045em] sm:text-5xl'>
                  {t('Download')} RoboCoding Desktop
                </h1>
              </div>
              <div className='border-border/60 bg-background/70 inline-flex w-fit items-center gap-2 rounded-full border px-3 py-1.5 text-sm shadow-sm backdrop-blur'>
                <span
                  className='size-1.5 rounded-full bg-blue-500'
                  aria-hidden='true'
                />
                <span className='text-muted-foreground'>
                  {t('Latest version')}
                </span>
                <ReleaseStatus
                  manifest={manifestQuery.data}
                  error={manifestQuery.error}
                />
              </div>
            </div>
            <ReleaseChangelog
              manifest={manifestQuery.data}
              error={manifestQuery.error}
            />
            <DownloadBuilds
              manifest={manifestQuery.data}
              error={manifestQuery.error}
              onRetry={() => manifestQuery.refetch()}
            />
          </div>
        </section>
      </main>
      <Footer />
    </PublicLayout>
  )
}
