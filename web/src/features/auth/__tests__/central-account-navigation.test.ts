import { afterEach, expect, it } from 'vitest'

import { ACCOUNT_CENTER_URL } from '@/lib/product-links'

import {
  getIdentityProfileURL,
  getIdentityTeamsURL,
  getIdentityAdminURL,
  getIdentitySecurityURL,
  usesCentralIdentityManagement,
} from '../central-account-navigation'

afterEach(() => localStorage.removeItem('status'))

it('routes identity management to the account center in central mode', () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))

  expect(getIdentityTeamsURL()).toBe('/teams')
  expect(getIdentityAdminURL()).toBe('/users')
  expect(usesCentralIdentityManagement()).toBe(true)
  expect(getIdentityProfileURL()).toBe(`${ACCOUNT_CENTER_URL}/account/profile`)
  expect(getIdentitySecurityURL()).toBe(
    `${ACCOUNT_CENTER_URL}/account/security`
  )
})

it('retains the platform profile in legacy mode', () => {
  localStorage.setItem(
    'status',
    JSON.stringify({ account_auth_enabled: false })
  )

  expect(usesCentralIdentityManagement()).toBe(false)
  expect(getIdentityTeamsURL()).toBe('/teams')
  expect(getIdentityAdminURL()).toBe('/users')
  expect(getIdentityProfileURL()).toBe('/profile')
  expect(getIdentitySecurityURL()).toBe('/security')
})
