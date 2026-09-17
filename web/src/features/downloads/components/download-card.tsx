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
import { Download } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { IconType } from 'react-icons'
import { FaApple, FaLinux, FaWindows } from 'react-icons/fa'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

import type { DownloadTarget, DownloadTargetId } from '../types'

type TargetPresentation = {
  title: string
  architecture: string
  helper: string
  icon: IconType
}

const TARGET_PRESENTATIONS: Record<DownloadTargetId, TargetPresentation> = {
  'windows-x64': {
    title: 'Windows',
    architecture: '64-bit',
    helper: 'For 64-bit Windows computers',
    icon: FaWindows,
  },
  'macos-arm64': {
    title: 'macOS',
    architecture: 'Apple Silicon',
    helper: 'For Macs with M-series chips',
    icon: FaApple,
  },
  'macos-x64': {
    title: 'macOS',
    architecture: 'Intel',
    helper: 'For Intel-based Macs',
    icon: FaApple,
  },
  'linux-x64': {
    title: 'Linux',
    architecture: '64-bit',
    helper: 'For 64-bit Linux computers',
    icon: FaLinux,
  },
}

export function DownloadCard(props: { target: DownloadTarget }) {
  const { t } = useTranslation()
  const presentation = TARGET_PRESENTATIONS[props.target.id]
  const Icon = presentation.icon
  const available = props.target.status === 'available'

  return (
    <Card className='border-border/60 bg-card/80 relative min-h-64 overflow-hidden shadow-[0_18px_55px_-38px_rgba(37,99,235,0.55)] transition-shadow hover:shadow-[0_22px_60px_-32px_rgba(37,99,235,0.6)]'>
      <div
        aria-hidden='true'
        className='absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-blue-500/70 to-transparent'
      />
      <CardHeader className='gap-3 pt-2'>
        <div className='flex size-12 items-center justify-center rounded-xl border border-blue-500/15 bg-blue-500/8 text-blue-600 dark:text-blue-400'>
          <Icon className='size-6' aria-hidden='true' />
        </div>
        <CardTitle className='text-xl tracking-tight'>
          {t(presentation.title)}
        </CardTitle>
        <CardAction>
          <Badge variant={available ? 'default' : 'secondary'}>
            {available ? t('Available') : t('Coming soon')}
          </Badge>
        </CardAction>
      </CardHeader>
      <CardContent className='flex flex-1 flex-col gap-5'>
        <div>
          <p className='font-medium'>{t(presentation.architecture)}</p>
          <p className='text-muted-foreground mt-1 text-sm'>
            {t(presentation.helper)}
          </p>
        </div>
        <div className='text-muted-foreground mt-auto flex min-h-5 items-center gap-2 text-xs'>
          {available ? (
            <>
              <span>{props.target.fileName}</span>
              {props.target.size != null && <span>· {props.target.size}</span>}
            </>
          ) : (
            <span>{t('This build has not been published yet.')}</span>
          )}
        </div>
        {available && props.target.url != null ? (
          <Button
            size='lg'
            className='h-10 w-full bg-blue-600 text-white hover:bg-blue-500 dark:bg-blue-500 dark:hover:bg-blue-400'
            render={
              <a
                href={props.target.url}
                download={props.target.fileName ?? undefined}
                rel='noopener noreferrer'
              />
            }
          >
            <Download aria-hidden='true' />
            {t('Download')}
          </Button>
        ) : (
          <Button
            size='lg'
            className='h-10 w-full'
            variant='secondary'
            disabled
          >
            {t('Unavailable')}
          </Button>
        )}
      </CardContent>
    </Card>
  )
}
