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
import { Fragment } from 'react'
import { useTranslation } from 'react-i18next'

import { useStatus } from '@/hooks/use-status'
import { useSystemConfig } from '@/hooks/use-system-config'
import { cn } from '@/lib/utils'

interface FooterProps {
  logo?: string
  name?: string
  columns?: { title: string; links: { text: string; href: string }[] }[]
  copyright?: string
  className?: string
}

function LegalLinks(props: { leadingSeparator?: boolean }) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const items: { key: string; label: string; href: string }[] = []
  if (status?.user_agreement_enabled) {
    items.push({
      key: 'user-agreement',
      label: t('User Agreement'),
      href: '/user-agreement',
    })
  }
  if (status?.privacy_policy_enabled) {
    items.push({
      key: 'privacy-policy',
      label: t('Privacy Policy'),
      href: '/privacy-policy',
    })
  }
  if (items.length === 0) {
    return null
  }
  return (
    <>
      {items.map((item, index) => (
        <Fragment key={item.key}>
          {(props.leadingSeparator || index > 0) && (
            <span aria-hidden='true' className='text-muted-foreground/30'>
              ·
            </span>
          )}
          <Link
            to={item.href}
            className='hover:text-foreground transition-colors duration-200'
          >
            {item.label}
          </Link>
        </Fragment>
      ))}
    </>
  )
}

export function Footer(props: FooterProps) {
  const { t } = useTranslation()
  const { systemName, logo, footerHtml } = useSystemConfig()
  return (
    <footer
      className={cn(
        'robo-public-footer border-border bg-card border-t',
        props.className
      )}
    >
      <div className='mx-auto max-w-6xl px-6 py-12'>
        <div className='flex flex-col justify-between gap-8 sm:flex-row'>
          <div>
            <Link to='/' className='inline-flex items-center gap-3'>
              <img
                src={logo || props.logo}
                alt=''
                className='size-10 rounded-xl'
              />
              <span>
                <span className='block text-base font-semibold'>
                  {systemName || props.name}
                </span>
              </span>
            </Link>
          </div>
          <nav
            aria-label={t('Footer navigation')}
            className='flex flex-wrap items-start gap-x-7 gap-y-4 text-base'
          >
            <Link to='/download' className='hover:text-primary'>
              {t('Download')}
            </Link>
            <Link to='/dashboard' className='hover:text-primary'>
              {t('Console')}
            </Link>
            <Link to='/about' className='hover:text-primary'>
              {t('About')}
            </Link>
          </nav>
        </div>
        {props.columns?.map((column) => (
          <div key={column.title} className='mt-6 text-base'>
            <p className='font-medium'>{t(column.title)}</p>
            <ul className='mt-2 flex flex-wrap gap-4'>
              {column.links.map((link) => (
                <li key={link.href}>
                  <a
                    href={link.href}
                    className='text-muted-foreground hover:text-primary'
                  >
                    {t(link.text)}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        ))}
        {footerHtml && (
          <div
            className='custom-footer text-muted-foreground mt-6 text-base'
            dangerouslySetInnerHTML={{ __html: footerHtml }}
          />
        )}
        <div className='border-border text-muted-foreground mt-8 flex flex-wrap items-center justify-between gap-4 border-t pt-6 text-sm'>
          <span>
            © {new Date().getFullYear()} {systemName}.{' '}
            {props.copyright ?? t('All rights reserved.')}
          </span>
          <div className='flex flex-wrap items-center gap-3'>
            <LegalLinks />
            <Link to='/licenses' className='hover:text-primary'>
              {t('Open-source licenses')}
            </Link>
          </div>
        </div>
      </div>
    </footer>
  )
}
