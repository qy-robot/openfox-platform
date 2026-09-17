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
import { formatPricingNumber } from '@/features/system-settings/models/pricing-format'
import type { CurrencyConfig } from '@/stores/system-config-store'

export type PricingCurrency = {
  label: string
  symbol: string
  exchangeRate: number
}

export const CNY_PRICING_CURRENCY: PricingCurrency = {
  label: 'CNY',
  symbol: '¥',
  exchangeRate: 1,
}

/** Compatibility alias for callers that have not yet renamed the import. */
export const USD_PRICING_CURRENCY = CNY_PRICING_CURRENCY

export function getSitePricingCurrency(
  _config: CurrencyConfig
): PricingCurrency {
  return CNY_PRICING_CURRENCY
}

export function isValidPricingCurrency(
  currency: PricingCurrency | null
): currency is PricingCurrency {
  return (
    currency !== null &&
    Number.isFinite(currency.exchangeRate) &&
    currency.exchangeRate > 0
  )
}

export function formatPricingAmount(
  value: string | number,
  currency = CNY_PRICING_CURRENCY
): string {
  if (value === '') return ''
  const amount = Number(value)
  if (!Number.isFinite(amount)) return '—'
  return `${currency.symbol}${formatPricingNumber(amount)}`
}
