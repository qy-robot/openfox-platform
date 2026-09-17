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

import { normalizeCreemProducts } from '../use-topup-info'

it('normalizes CNY Creem products and rejects foreign or unknown currencies', () => {
  expect(
    normalizeCreemProducts([
      {
        name: 'CNY package',
        productId: 'cny-package',
        price: 12.5,
        quota: 6_250_000,
        currency: 'cny',
      },
      {
        name: 'Legacy USD package',
        productId: 'usd-package',
        price: 2,
        quota: 1_000_000,
        currency: 'USD',
      },
      {
        name: 'Missing currency',
        productId: 'unknown-package',
        price: 2,
        quota: 1_000_000,
      },
    ])
  ).toEqual([
    {
      name: 'CNY package',
      productId: 'cny-package',
      price: 12.5,
      quota: 6_250_000,
      currency: 'CNY',
    },
  ])
})
