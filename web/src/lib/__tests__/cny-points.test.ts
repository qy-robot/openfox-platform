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
import { afterEach, beforeEach, describe, expect, test } from 'vitest'

import {
  formatPointsFromQuota,
  pointsToQuota,
  quotaToEditablePoints,
  quotaToPoints,
} from '@/lib/currency'
import { parseQuotaFromDollars, quotaUnitsToDollars } from '@/lib/format'
import { useSystemConfigStore } from '@/stores/system-config-store'

describe('CNY and points conversion boundary', () => {
  beforeEach(() => {
    useSystemConfigStore.getState().setConfig({
      currency: {
        displayInCurrency: true,
        quotaDisplayType: 'CNY',
        quotaPerUnit: 500_000,
        usdExchangeRate: 7,
        customCurrencySymbol: '¤',
        customCurrencyExchangeRate: 1,
        pointsPerCny: 10,
        quotaPerPoint: 500_000 / 70,
      },
    })
  })

  afterEach(() => {
    useSystemConfigStore.setState(useSystemConfigStore.getInitialState(), true)
  })

  test('round-trips a CNY form amount without exchange-rate conversion', () => {
    const quota = parseQuotaFromDollars(10)
    expect(quota).toBe(5_000_000)
    expect(quotaUnitsToDollars(quota)).toBe(10)
  })

  test('maps 10 CNY to 100 product points regardless of a legacy exchange rate', () => {
    expect(formatPointsFromQuota(5_000_000)).toBe('100')
    expect(quotaToPoints(5_000_000)).toBe(100)
    expect(pointsToQuota(100)).toBe(5_000_000)
  })

  test('keeps an unformatted form value reversible without display rounding', () => {
    const quota = 1_234_567
    expect(pointsToQuota(quotaToEditablePoints(quota))).toBe(quota)
    expect(quotaToEditablePoints(pointsToQuota(2_000))).toBe(2_000)
  })
})
