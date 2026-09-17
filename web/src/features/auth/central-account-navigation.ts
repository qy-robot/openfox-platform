import { ACCOUNT_CENTER_URL } from '@/lib/product-links'

import { isCentralAccountMode } from './sign-in/central-reauth'

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
