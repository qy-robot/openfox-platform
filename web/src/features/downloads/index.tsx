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
import {
  Box,
  CheckCircle2,
  Cpu,
  RefreshCw,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'
import { Badge } from '@/components/ui/badge'

import { getDownloadManifest } from './api'
import { DownloadBuilds } from './components/download-builds'
import { ReleaseStatus } from './components/release-status'

const DOWNLOAD_BENEFITS: Array<{
  icon: LucideIcon
  title: string
  description: string
}> = [
  {
    icon: Cpu,
    title: 'Match your computer',
    description: 'Pick Windows, your Mac chip, or Linux.',
  },
  {
    icon: CheckCircle2,
    title: 'Clear release status',
    description: 'See which builds are ready and which are on the way.',
  },
  {
    icon: RefreshCw,
    title: 'One place for every update',
    description: 'Return here for future desktop releases.',
  },
]

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
      <main>
        <section className='border-border/50 relative overflow-hidden border-b px-5 pt-28 pb-16 sm:px-8 md:pt-36 md:pb-24'>
          <div
            aria-hidden='true'
            className='absolute inset-0 -z-10 bg-[radial-gradient(circle_at_18%_18%,rgba(59,130,246,0.16),transparent_33%),radial-gradient(circle_at_85%_5%,rgba(125,211,252,0.12),transparent_28%)] dark:bg-[radial-gradient(circle_at_18%_18%,rgba(37,99,235,0.22),transparent_34%),radial-gradient(circle_at_85%_5%,rgba(14,116,144,0.14),transparent_30%)]'
          />
          <div className='mx-auto grid max-w-6xl items-end gap-10 lg:grid-cols-[1.35fr_0.65fr]'>
            <div>
              <Badge
                className='mb-6 border-blue-500/20 bg-blue-500/8 text-blue-700 dark:text-blue-300'
                variant='outline'
              >
                {t('RoboCoding Desktop')}
              </Badge>
              <h1 className='max-w-4xl text-4xl leading-[1.04] font-semibold tracking-[-0.045em] text-balance sm:text-6xl lg:text-7xl'>
                {t('Build with your robot, from your own desktop.')}
              </h1>
              <p className='text-muted-foreground mt-7 max-w-2xl text-base leading-7 text-pretty sm:text-lg'>
                {t(
                  'RoboCoding brings robot development, AI assistance, and your project workspace into one focused application.'
                )}
              </p>
            </div>
            <div className='border-border/60 bg-background/65 rounded-2xl border p-5 shadow-sm backdrop-blur'>
              <div className='mb-5 flex items-center justify-between gap-3'>
                <span className='text-muted-foreground text-xs font-semibold tracking-[0.18em] uppercase'>
                  {t('Release channel')}
                </span>
                <span className='size-2 rounded-full bg-blue-500 shadow-[0_0_0_6px_rgba(59,130,246,0.1)]' />
              </div>
              <ReleaseStatus
                manifest={manifestQuery.data}
                error={manifestQuery.error}
              />
              <p className='text-muted-foreground mt-2 text-sm leading-6'>
                {t(
                  'Choose the build for your computer below. New releases appear here when they are ready.'
                )}
              </p>
            </div>
          </div>
        </section>

        <section className='px-5 py-14 sm:px-8 md:py-20'>
          <div className='mx-auto max-w-6xl'>
            <div className='mb-9 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between'>
              <div>
                <p className='text-sm font-semibold text-blue-600 dark:text-blue-400'>
                  {t('Desktop builds')}
                </p>
                <h2 className='mt-2 text-2xl font-semibold tracking-tight sm:text-3xl'>
                  {t('Choose your platform')}
                </h2>
              </div>
              <p className='text-muted-foreground max-w-lg text-sm leading-6 sm:text-right'>
                {t(
                  'Ready builds include a download button. Upcoming versions stay clearly marked.'
                )}
              </p>
            </div>
            <DownloadBuilds
              manifest={manifestQuery.data}
              error={manifestQuery.error}
              onRetry={() => manifestQuery.refetch()}
            />
          </div>
        </section>

        <section className='border-border/50 bg-muted/30 border-y px-5 py-12 sm:px-8'>
          <div className='mx-auto grid max-w-6xl gap-8 md:grid-cols-3'>
            {DOWNLOAD_BENEFITS.map((benefit) => {
              const Icon = benefit.icon
              return (
                <div key={benefit.title} className='flex gap-4'>
                  <div className='mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400'>
                    <Icon className='size-4' aria-hidden='true' />
                  </div>
                  <div>
                    <h3 className='font-medium'>{t(benefit.title)}</h3>
                    <p className='text-muted-foreground mt-1 text-sm leading-6'>
                      {t(benefit.description)}
                    </p>
                  </div>
                </div>
              )
            })}
          </div>
        </section>

        <section className='px-5 py-14 text-center sm:px-8 md:py-20'>
          <Box
            className='mx-auto size-7 text-blue-600 dark:text-blue-400'
            aria-hidden='true'
          />
          <h2 className='mt-4 text-xl font-semibold'>
            {t('Need help choosing a build?')}
          </h2>
          <p className='text-muted-foreground mx-auto mt-2 max-w-xl text-sm leading-6'>
            {t(
              'On macOS, open About This Mac to check whether your chip is Apple Silicon or Intel. Windows and Linux releases currently target 64-bit computers.'
            )}
          </p>
          <p className='text-muted-foreground/70 mt-8 text-xs'>
            RoboCoding · by擎云机器人
          </p>
        </section>
      </main>
      <Footer />
    </PublicLayout>
  )
}
