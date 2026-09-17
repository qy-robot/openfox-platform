import { describe, expect, it } from 'vitest'

import { getIndependentAppURL } from '../independent-apps'

describe('independent application navigation', () => {
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
