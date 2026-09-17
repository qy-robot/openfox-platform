import { ACCOUNT_CENTER_URL } from '@/lib/product-links'

import { isCentralAccountMode } from './sign-in/central-reauth'

type ResolvedTheme = 'dark' | 'light'

function accountCenterURL(value: unknown): URL {
  try {
    const url = new URL(
      typeof value === 'string' && value.trim()
        ? value.trim()
        : ACCOUNT_CENTER_URL
    )
    const localHTTP =
      url.protocol === 'http:' &&
      (url.hostname === 'localhost' || url.hostname === '127.0.0.1')
    if (url.protocol === 'https:' || localHTTP) return url
  } catch {
    // Invalid runtime configuration falls back to the product default.
  }
  return new URL(ACCOUNT_CENTER_URL)
}

function accountAuthorizationContinuation(
  value: unknown,
  accountOrigin: string
): string | null {
  if (typeof value !== 'string') return null
  const target = value.trim()
  if (!target || target.includes('\\') || target.startsWith('//')) return null

  try {
    const url = new URL(target, accountOrigin)
    if (
      url.origin !== accountOrigin ||
      url.pathname !== '/v1/oauth/authorize'
    ) {
      return null
    }
    return `${url.pathname}${url.search}${url.hash}`
  } catch {
    return null
  }
}

export function getAccountRegistrationURL(options: {
  accountCenterURL?: unknown
  continuation?: unknown
  theme: ResolvedTheme
}): string {
  const accountURL = accountCenterURL(options.accountCenterURL)
  const registrationURL = new URL('/account/register', accountURL)
  registrationURL.searchParams.set('theme', options.theme)

  const continuation = accountAuthorizationContinuation(
    options.continuation,
    accountURL.origin
  )
  if (continuation) registrationURL.searchParams.set('continue', continuation)

  return registrationURL.toString()
}

export function navigateToAccountRegistration(url: string): void {
  window.location.replace(url)
}

export function getIdentityProfileURL(): string {
  return isCentralAccountMode()
    ? `${ACCOUNT_CENTER_URL}/account/profile`
    : '/profile'
}

export function getIdentitySecurityURL(): string {
  return isCentralAccountMode()
    ? `${ACCOUNT_CENTER_URL}/account/security`
    : '/security'
}

export function usesCentralIdentityManagement(): boolean {
  return isCentralAccountMode()
}

export function getIdentityTeamsURL(): string {
  return '/teams'
}

export function getIdentityAdminURL(): string {
  return '/users'
}
