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
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'

type AuthLayoutProps = {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <div className='openfox-auth-layout relative grid min-h-svh max-w-none overflow-hidden lg:grid-cols-[minmax(0,46%)_minmax(0,54%)]'>
      <Link
        to='/'
        className='openfox-auth-brand focus-visible:ring-ring absolute top-6 left-6 z-10 flex items-center gap-3 rounded-xl transition-opacity outline-none hover:opacity-80 focus-visible:ring-2 sm:top-10 sm:left-10'
      >
        <div className='border-border relative size-10 overflow-hidden rounded-xl border bg-white p-1'>
          {loading ? (
            <Skeleton className='absolute inset-0 rounded-xl' />
          ) : (
            <img
              src={logo}
              alt={t('Logo')}
              className='size-full object-contain'
            />
          )}
        </div>
        {loading ? (
          <Skeleton className='h-6 w-24' />
        ) : (
          <span className='text-lg font-semibold tracking-[-0.03em]'>
            {systemName}
          </span>
        )}
      </Link>
      <aside className='openfox-auth-scene border-border relative hidden flex-col justify-center overflow-hidden border-r px-16 py-32 lg:flex xl:px-24'>
        <div className='openfox-auth-arc' aria-hidden='true' />
        <img
          src={logo}
          alt=''
          aria-hidden='true'
          className='openfox-auth-mark'
        />
        <div className='relative z-10 max-w-[31rem]'>
          <span
            className='mb-5 block text-xl leading-none text-[var(--brand-star)]'
            aria-hidden='true'
          >
            ✦
          </span>
          <h1 className='text-[clamp(2.25rem,3vw,3.75rem)] leading-[1.18] font-semibold tracking-[-0.045em]'>
            {t('Models and API access.')}
          </h1>
          <p className='dark:text-muted-foreground mt-8 max-w-md text-base leading-8 text-[#506b7c]'>
            {t('Browse models, manage API keys and review your wallet.')}
          </p>
          <Link
            to='/download'
            className='border-primary/40 text-primary hover:border-primary focus-visible:outline-ring mt-12 inline-flex w-fit items-center border-b pb-1 text-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-offset-4'
          >
            {t('Download desktop app')}
          </Link>
        </div>
      </aside>
      <main className='bg-card flex min-w-0 items-center pt-24 lg:pt-0'>
        <div className='mx-auto flex w-full max-w-[29rem] flex-col justify-center space-y-2 px-7 py-12 sm:px-8'>
          {children}
        </div>
      </main>
    </div>
  )
}
