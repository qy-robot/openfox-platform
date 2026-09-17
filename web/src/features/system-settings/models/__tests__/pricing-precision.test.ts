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
import { expect, it } from 'vitest'

import { serializePricingNumber } from '../pricing-format'

it.each([10 / 7, 5 / 7, 0.12345678 / 7, 1e-13, 1e-20, 1.23456789012345])(
  'preserves the stored numeric value %s without display truncation',
  (value) => {
    expect(Number(serializePricingNumber(value))).toBe(value)
  }
)

it('normalizes binary division noise without changing a decimal price', () => {
  expect(serializePricingNumber(0.7 / 7)).toBe('0.1')
})

it.each([Infinity, -Infinity, Number.NaN])(
  'keeps non-finite value %s invalid for downstream validation',
  (value) => {
    expect(Number.isFinite(Number(serializePricingNumber(value)))).toBe(false)
  }
)
