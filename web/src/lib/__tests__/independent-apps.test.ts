import { describe, expect, it } from 'vitest'

import { getIndependentAppURL } from '../independent-apps'
import { resolveDomainFamily } from '../product-links'

describe('independent application navigation', () => {
  // jsdom runs on localhost, so sibling entry points resolve to the
  // currently serving domain family (openzrob.com during ICP filing).
  it('sends account management to the account subdomain', () => {
    expect(getIndependentAppURL('/account/profile')).toBe(
      'https://account.openzrob.com/account/profile'
    )
    expect(getIndependentAppURL('/account')).toBe(
      'https://account.openzrob.com/account/profile'
    )
    expect(getIndependentAppURL('/account/admin')).toBe(
      'https://account.openzrob.com/account/admin'
    )
  })

  it('keeps the AI station SSO callback in this application', () => {
    expect(getIndependentAppURL('/account/callback')).toBeNull()
    expect(getIndependentAppURL('/dashboard')).toBeNull()
  })

  it('sends workbenches and public learning to the independent console', () => {
    expect(getIndependentAppURL('/workbench/devices')).toBe(
      'https://dash.openzrob.com/workbench/devices'
    )
    expect(getIndependentAppURL('/workbench/repositories')).toBe(
      'https://dash.openzrob.com/workbench/repositories'
    )
    expect(getIndependentAppURL('/learning/')).toBe(
      'https://dash.openzrob.com/learning'
    )
  })

  it('never copies an unrecognized path into an external destination', () => {
    expect(getIndependentAppURL('/account/https://example.invalid')).toBe(
      'https://account.openzrob.com/account/profile'
    )
    expect(getIndependentAppURL('//example.invalid')).toBeNull()
    expect(getIndependentAppURL('/accounting')).toBeNull()
  })
})

describe('product entry domain families', () => {
  it('derives the entry family from the visited hostname', () => {
    expect(resolveDomainFamily('ai.openzrob.com')).toBe('openzrob.com')
    expect(resolveDomainFamily('www.openfox.work')).toBe('openfox.work')
    expect(resolveDomainFamily('account.openfox.work')).toBe('openfox.work')
  })

  it('falls back to the serving family outside known domains', () => {
    expect(resolveDomainFamily('localhost')).toBe('openzrob.com')
    expect(resolveDomainFamily('openzrob.com.evil.example')).toBe('openzrob.com')
    expect(resolveDomainFamily('notopenzrob.com')).toBe('openzrob.com')
  })
})
