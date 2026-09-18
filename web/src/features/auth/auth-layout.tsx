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
    <div className='robo-auth-layout relative grid min-h-svh max-w-none lg:grid-cols-2'>
      <Link
        to='/'
        className='absolute top-4 left-4 z-10 flex items-center gap-2 transition-opacity hover:opacity-80 sm:top-8 sm:left-8'
      >
        <div className='relative h-8 w-8'>
          {loading ? (
            <Skeleton className='absolute inset-0 rounded-full' />
          ) : (
            <img
              src={logo}
              alt={t('Logo')}
              className='h-8 w-8 rounded-full object-cover'
            />
          )}
        </div>
        {loading ? (
          <Skeleton className='h-6 w-24' />
        ) : (
          <span>
            <span className='block text-lg font-semibold'>{systemName}</span>
            <span className='text-muted-foreground block text-xs'>
              by擎云机器人
            </span>
          </span>
        )}
      </Link>
      <aside className='bg-accent/60 hidden flex-col justify-center px-16 py-32 lg:flex xl:px-24'>
        <p className='text-primary mb-6 text-sm font-medium'>RoboCoding</p>
        <h1 className='text-4xl leading-snug font-semibold tracking-tight'>
          {t('Learn by building.')}
          <br />
          {t('Move forward with every task.')}
        </h1>
        <p className='text-muted-foreground mt-6 max-w-md text-base leading-8'>
          {t('One account for your desktop workspace and cloud services.')}
        </p>
        <Link
          to='/download'
          className='text-primary mt-10 w-fit text-sm underline underline-offset-4'
        >
          {t('Download desktop app')}
        </Link>
      </aside>
      <div className='flex items-center pt-20 lg:pt-0'>
        <div className='mx-auto flex w-full flex-col justify-center space-y-2 px-4 py-8 sm:w-[480px] sm:p-8'>
          {children}
        </div>
      </div>
    </div>
  )
}
