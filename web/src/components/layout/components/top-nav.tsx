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
import { Link, useRouterState } from '@tanstack/react-router'
import { Menu } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'

import type { TopNavLink } from '../types'

type TopNavProps = React.HTMLAttributes<HTMLElement> & {
  links: TopNavLink[]
}

/**
 * 顶部导航栏组件
 * 在大屏幕显示水平导航，在小屏幕显示下拉菜单
 */
export function TopNav({ className, links, ...props }: TopNavProps) {
  const { t } = useTranslation()
  const pathname = useRouterState({
    select: (state) => state.location.pathname,
  })
  // 规范化链接，确保所有可选属性都有默认值
  const normalizedLinks = useMemo(
    () =>
      links.map((link) => ({
        disabled: false,
        external: false,
        ...link,
        isActive:
          link.isActive ??
          (!link.external &&
            (pathname === link.href ||
              (link.href !== '/' && pathname.startsWith(`${link.href}/`)))),
      })),
    [links, pathname]
  )

  return (
    <>
      {/* 移动端下拉菜单 */}
      <div className='lg:hidden'>
        <DropdownMenu modal={false}>
          <DropdownMenuTrigger
            aria-label={t('Toggle navigation menu')}
            render={
              <Button
                size='icon'
                variant='outline'
                className='size-9 rounded-xl'
              />
            }
          >
            <Menu />
          </DropdownMenuTrigger>
          <DropdownMenuContent side='bottom' align='start'>
            <DropdownMenuGroup>
              {normalizedLinks.map(
                ({ title, href, isActive, disabled, external }) => {
                  if (disabled) {
                    return (
                      <DropdownMenuItem
                        key={`${title}-${href}`}
                        disabled
                        className='py-2 text-sm'
                      >
                        {title}
                      </DropdownMenuItem>
                    )
                  }
                  return (
                    <DropdownMenuItem
                      key={`${title}-${href}`}
                      className='py-2 text-sm'
                      render={
                        external ? (
                          <a
                            href={href}
                            target='_blank'
                            rel='noopener noreferrer'
                            aria-current={isActive ? 'page' : undefined}
                            className={!isActive ? 'text-muted-foreground' : ''}
                          >
                            {title}
                          </a>
                        ) : (
                          <Link
                            to={href}
                            className={!isActive ? 'text-muted-foreground' : ''}
                            aria-current={isActive ? 'page' : undefined}
                          >
                            {title}
                          </Link>
                        )
                      }
                    />
                  )
                }
              )}
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* 桌面端水平导航 */}
      <nav
        className={cn('hidden items-center gap-1 lg:flex', className)}
        {...props}
      >
        {normalizedLinks.map(({ title, href, isActive, disabled, external }) =>
          external ? (
            <a
              key={`${title}-${href}`}
              href={href}
              target='_blank'
              rel='noopener noreferrer'
              aria-current={isActive ? 'page' : undefined}
              aria-disabled={disabled || undefined}
              tabIndex={disabled ? -1 : undefined}
              className={cn(
                'hover:bg-accent hover:text-accent-foreground rounded-lg px-2.5 py-2 text-sm font-medium whitespace-nowrap transition-colors',
                isActive
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground',
                disabled && 'pointer-events-none opacity-50'
              )}
            >
              {title}
            </a>
          ) : (
            <Link
              key={`${title}-${href}`}
              to={href}
              disabled={disabled}
              aria-current={isActive ? 'page' : undefined}
              className={cn(
                'hover:bg-accent hover:text-accent-foreground rounded-lg px-2.5 py-2 text-sm font-medium whitespace-nowrap transition-colors',
                isActive
                  ? 'bg-accent text-accent-foreground'
                  : 'text-muted-foreground',
                disabled && 'opacity-50'
              )}
            >
              {title}
            </Link>
          )
        )}
      </nav>
    </>
  )
}
