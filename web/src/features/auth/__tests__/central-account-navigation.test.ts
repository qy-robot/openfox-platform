import { afterEach, expect, it } from 'vitest'

import { ACCOUNT_CENTER_URL } from '@/lib/product-links'

import {
  getAccountRegistrationURL,
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

it('builds a themed account registration URL and preserves only account authorization continuations', () => {
  expect(
    getAccountRegistrationURL({
      accountCenterURL: 'https://account.example.com/base',
      continuation: '/v1/oauth/authorize?client_id=ai&state=one',
      theme: 'dark',
    })
  ).toBe(
    'https://account.example.com/account/register?theme=dark&continue=%2Fv1%2Foauth%2Fauthorize%3Fclient_id%3Dai%26state%3Done'
  )

  expect(
    getAccountRegistrationURL({
      accountCenterURL: 'https://account.example.com',
      continuation: 'https://attacker.example/steal',
      theme: 'light',
    })
  ).toBe('https://account.example.com/account/register?theme=light')
})

it('falls back to the product account origin for unsafe runtime configuration', () => {
  expect(
    getAccountRegistrationURL({
      accountCenterURL: 'http://account.example.com',
      theme: 'dark',
    })
  ).toBe(`${ACCOUNT_CENTER_URL}/account/register?theme=dark`)
})
