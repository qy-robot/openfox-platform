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
import { describe, expect, it } from 'vitest'

import {
  formatCreemPrice,
  formatPaymentAmount,
  formatTopupEntitlement,
} from '../format'

describe('top-up history currency formatting', () => {
  it('derives the CNY entitlement from exact credited quota', () => {
    expect(
      formatTopupEntitlement({
        amount: 10,
        credited_quota: 5_000_000,
        currency: 'USD',
      })
    ).toContain('10')
    expect(
      formatTopupEntitlement({
        amount: 10,
        credited_quota: 5_000_000,
        currency: 'USD',
      })
    ).toContain('¥')
  })

  it('does not infer the legacy entitlement unit from payment currency', () => {
    const formatted = formatTopupEntitlement(
      { amount: 1, currency: 'CNY' },
      'Legacy unit unknown'
    )
    expect(formatted).toBe('1 (Legacy unit unknown)')
    expect(formatted).not.toContain('¥')
  })

  it('marks a legacy amount with no audited currency as unknown', () => {
    expect(
      formatTopupEntitlement({ amount: 3.5 }, 'Unknown audited currency')
    ).toBe('3.5 (Unknown audited currency)')
  })

  it('keeps an exact zero credited quota instead of falling back to legacy amount', () => {
    const formatted = formatTopupEntitlement({
      amount: 9,
      credited_quota: 0,
      currency: 'USD',
    })
    expect(formatted).toContain('¥0')
    expect(formatted).not.toContain('9')
  })

  it('preserves the audited payment currency', () => {
    expect(formatPaymentAmount(7.3, 'CNY')).toContain('¥')
    expect(formatPaymentAmount(1, 'USD')).toContain('$')
  })

  it('formats validated Creem products in CNY', () => {
    expect(formatCreemPrice(12.5, 'CNY')).toBe('¥12.50')
  })
})
