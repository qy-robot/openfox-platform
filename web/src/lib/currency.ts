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
/**
 * ============================================================================
 * Currency Formatting Library
 * ============================================================================
 *
 * This module formats the platform's native CNY amounts and exact quota units.
 * Retail prices, balances, payments, and ledgers are all denominated in CNY.
 *
 * ## Key Concepts
 *
 * 1. **CNY**: The only runtime billing and display currency.
 * 2. **Quota**: Exact integer ledger units; 500,000 quota equals 1 CNY.
 * 3. **Points**: Product points; 10 points equal 1 CNY.
 * 4. **Legacy fields**: USD/custom display and exchange-rate settings remain in
 *    API types for compatibility only. They never change runtime amounts.
 *
 * ## When to Use Each Function
 *
 * - `formatCurrencyFromUSD()` and `formatBillingCurrencyFromUSD()` are legacy
 *   export names. Their input is now CNY and neither applies an exchange rate.
 * - `formatLocalCurrencyAmount()` formats a CNY payment amount directly.
 * - `formatQuotaWithCurrency()` converts quota to CNY using `quotaPerUnit`.
 */
import {
  useSystemConfigStore,
  DEFAULT_CURRENCY_CONFIG,
  type CurrencyConfig,
  type CurrencyDisplayType,
} from '@/stores/system-config-store'

export interface CurrencyFormatOptions {
  /** Fraction digits to use when |value| >= 1 */
  digitsLarge?: number
  /** Fraction digits to use when |value| < 1 */
  digitsSmall?: number
  /** Whether to abbreviate thousands with k suffix */
  abbreviate?: boolean
  /** Minimal absolute value to display when rounding would produce zero */
  minimumNonZero?: number
  /**
   * Use locale-aware compact notation for large values (e.g. "$28万" in zh,
   * "$280K" in en). The currency symbol is preserved.
   */
  compact?: boolean
  /** Whether to include the currency/custom symbol. Token displays are unchanged. */
  showSymbol?: boolean
  /** Locale used for number formatting (defaults to the runtime locale) */
  locale?: Intl.LocalesArgument | undefined
}

type ResolvedCurrencyFormatOptions = Omit<
  Required<CurrencyFormatOptions>,
  'locale'
> & {
  locale: Intl.LocalesArgument | undefined
}

type DisplayMeta =
  | {
      kind: 'currency'
      symbol: string
      currencyCode: string
      exchangeRate: number
    }
  | {
      kind: 'custom'
      symbol: string
      exchangeRate: number
    }
  | {
      kind: 'tokens'
      /** Number of quota units per CNY */
      quotaPerUnit: number
    }

const DEFAULT_FORMAT_OPTIONS: ResolvedCurrencyFormatOptions = {
  digitsLarge: 2,
  digitsSmall: 4,
  abbreviate: true,
  minimumNonZero: 0,
  compact: false,
  showSymbol: true,
  locale: undefined,
}

const DISPLAY_TYPE_VALUES = ['USD', 'CNY', 'TOKENS', 'CUSTOM'] as const
type DisplayTypeLiteral = (typeof DISPLAY_TYPE_VALUES)[number]

export function isCurrencyDisplayType(
  value: unknown
): value is CurrencyDisplayType {
  return (
    typeof value === 'string' &&
    DISPLAY_TYPE_VALUES.includes(value as DisplayTypeLiteral)
  )
}

export function parseCurrencyDisplayType(
  _value: unknown,
  _fallback: CurrencyDisplayType = 'CNY'
): CurrencyDisplayType {
  return 'CNY'
}

function getConfig(): CurrencyConfig {
  const { config } = useSystemConfigStore.getState()
  const currency = config?.currency ?? DEFAULT_CURRENCY_CONFIG
  return {
    ...DEFAULT_CURRENCY_CONFIG,
    ...currency,
    quotaPerUnit:
      currency?.quotaPerUnit && currency.quotaPerUnit > 0
        ? currency.quotaPerUnit
        : DEFAULT_CURRENCY_CONFIG.quotaPerUnit,
    usdExchangeRate:
      currency?.usdExchangeRate && currency.usdExchangeRate > 0
        ? currency.usdExchangeRate
        : DEFAULT_CURRENCY_CONFIG.usdExchangeRate,
    customCurrencyExchangeRate:
      currency?.customCurrencyExchangeRate &&
      currency.customCurrencyExchangeRate > 0
        ? currency.customCurrencyExchangeRate
        : DEFAULT_CURRENCY_CONFIG.customCurrencyExchangeRate,
    customCurrencySymbol:
      currency?.customCurrencySymbol?.trim() ||
      DEFAULT_CURRENCY_CONFIG.customCurrencySymbol,
    pointsPerCny:
      currency?.pointsPerCny && currency.pointsPerCny > 0
        ? currency.pointsPerCny
        : DEFAULT_CURRENCY_CONFIG.pointsPerCny,
    quotaPerPoint:
      currency?.quotaPerPoint && currency.quotaPerPoint > 0
        ? currency.quotaPerPoint
        : DEFAULT_CURRENCY_CONFIG.quotaPerPoint,
  }
}

/** Format exact backend quota units as robocoding product points. */
export function formatPointsFromQuota(
  quota: number,
  options: {
    maximumFractionDigits?: number
    locale?: Intl.LocalesArgument
  } = {}
): string {
  const points = quotaToPoints(quota)
  return new Intl.NumberFormat(options.locale, {
    maximumFractionDigits: options.maximumFractionDigits ?? 1,
  }).format(points)
}

/** Convert exact backend quota units to an unformatted point value for forms. */
export function quotaToPoints(quota: number): number {
  const quotaPerPoint = getConfig().quotaPerUnit / 10
  return Number.isFinite(quota) ? quota / quotaPerPoint : 0
}

/** Prefer a concise point value only when converting it back preserves quota. */
export function quotaToEditablePoints(quota: number): number {
  const points = quotaToPoints(quota)
  const nearestInteger = Math.round(points)
  return pointsToQuota(nearestInteger) === quota ? nearestInteger : points
}

/** Convert user-entered points back into exact integer quota units. */
export function pointsToQuota(points: number): number {
  if (!Number.isFinite(points) || points <= 0) return 0
  return Math.round(points * (getConfig().quotaPerUnit / 10))
}

function getDisplayMeta(_config: CurrencyConfig): DisplayMeta {
  return {
    kind: 'currency',
    symbol: '¥',
    currencyCode: 'CNY',
    exchangeRate: 1,
  }
}

function getBillingDisplayMeta(config: CurrencyConfig): DisplayMeta {
  const meta = getDisplayMeta(config)
  if (meta.kind === 'tokens') {
    return {
      kind: 'currency',
      symbol: '¥',
      currencyCode: 'CNY',
      exchangeRate: 1,
    }
  }
  return meta
}

function mergeOptions(
  options?: CurrencyFormatOptions
): ResolvedCurrencyFormatOptions {
  if (!options) return DEFAULT_FORMAT_OPTIONS
  return {
    digitsLarge: options.digitsLarge ?? DEFAULT_FORMAT_OPTIONS.digitsLarge,
    digitsSmall: options.digitsSmall ?? DEFAULT_FORMAT_OPTIONS.digitsSmall,
    abbreviate: options.abbreviate ?? DEFAULT_FORMAT_OPTIONS.abbreviate,
    minimumNonZero:
      options.minimumNonZero ?? DEFAULT_FORMAT_OPTIONS.minimumNonZero,
    compact: options.compact ?? DEFAULT_FORMAT_OPTIONS.compact,
    showSymbol: options.showSymbol ?? DEFAULT_FORMAT_OPTIONS.showSymbol,
    locale: options.locale ?? DEFAULT_FORMAT_OPTIONS.locale,
  }
}

function getFractionDigits(
  value: number,
  digitsLarge: number,
  digitsSmall: number
): number {
  return Math.abs(value) >= 1 ? digitsLarge : digitsSmall
}

/** Return the configured fraction digits for a plain currency value. */
export function getCurrencyFractionDigits(
  value: number,
  options?: CurrencyFormatOptions
): number {
  const merged = mergeOptions(options)
  return getFractionDigits(value, merged.digitsLarge, merged.digitsSmall)
}

function removeTrailingZeros(str: string): string {
  if (!str.includes('.')) return str
  return str.replace(/(\.[0-9]*?)0+$/, '$1').replace(/\.$/, '')
}

function formatNumberWithSuffix(
  value: number,
  digitsLarge: number,
  digitsSmall: number,
  abbreviate: boolean
): string {
  const abs = Math.abs(value)
  if (abbreviate && abs >= 1000) {
    const result = value / 1000
    return `${removeTrailingZeros(result.toFixed(1))}k`
  }

  const digits = getFractionDigits(value, digitsLarge, digitsSmall)
  return removeTrailingZeros(value.toFixed(digits))
}

function adjustForMinimum(
  value: number,
  digits: number,
  minimumNonZero: number
): number {
  if (value === 0) return value

  const threshold = minimumNonZero > 0 ? minimumNonZero : Math.pow(10, -digits)
  const abs = Math.abs(value)
  if (abs > 0 && abs < threshold) {
    return value > 0 ? threshold : -threshold
  }
  return value
}

function formatCurrencyValue(
  value: number,
  options: ResolvedCurrencyFormatOptions,
  meta: DisplayMeta
): string {
  if (meta.kind === 'tokens') {
    if (options.compact) {
      return new Intl.NumberFormat(options.locale, {
        notation: 'compact',
        maximumFractionDigits: 1,
      }).format(value)
    }
    return formatNumberWithSuffix(
      value,
      options.digitsLarge,
      options.digitsSmall,
      options.abbreviate
    )
  }

  const digits = getFractionDigits(
    value,
    options.digitsLarge,
    options.digitsSmall
  )
  const adjustedValue = adjustForMinimum(value, digits, options.minimumNonZero)

  if (meta.kind === 'currency') {
    if (!options.showSymbol) {
      return new Intl.NumberFormat(options.locale, {
        notation: options.compact ? 'compact' : 'standard',
        minimumFractionDigits: 0,
        maximumFractionDigits: options.compact ? 1 : digits,
      }).format(adjustedValue)
    }

    const formatted = new Intl.NumberFormat(options.locale, {
      style: 'currency',
      currency: meta.currencyCode,
      currencyDisplay: 'narrowSymbol',
      notation: options.compact ? 'compact' : 'standard',
      minimumFractionDigits: 0,
      maximumFractionDigits: options.compact ? 1 : digits,
    }).format(adjustedValue)
    return formatted
  }

  const decimal = new Intl.NumberFormat(options.locale, {
    notation: options.compact ? 'compact' : 'standard',
    minimumFractionDigits: 0,
    maximumFractionDigits: options.compact ? 1 : digits,
  }).format(adjustedValue)

  return options.showSymbol ? `${meta.symbol} ${decimal}` : decimal
}

/**
 * Get the current currency configuration and display metadata.
 *
 * @returns Object containing config and display metadata
 *
 * @internal
 * This is primarily for internal use. Most consumers should use the
 * higher-level formatting functions instead.
 */
export function getCurrencyDisplay() {
  const config = getConfig()
  const meta = getDisplayMeta(config)
  return { config, meta }
}

/**
 * Format a CNY amount. The export name is retained for source compatibility;
 * no USD conversion or configurable exchange rate is applied.
 *
 * @param amountCNY - Amount in CNY
 * @param options - Optional formatting configuration
 * @returns Formatted string with currency symbol or token count
 *
 * @example
 * formatCurrencyFromUSD(10) → "¥10"
 */
export function formatCurrencyFromUSD(
  amountCNY: number | null | undefined,
  options?: CurrencyFormatOptions
): string {
  if (amountCNY == null || Number.isNaN(amountCNY)) return '-'

  const { config, meta } = getCurrencyDisplay()
  const merged = mergeOptions(options)

  if (meta.kind === 'tokens') {
    const tokens = amountCNY * config.quotaPerUnit
    if (merged.compact) {
      return new Intl.NumberFormat(merged.locale, {
        notation: 'compact',
        maximumFractionDigits: 1,
      }).format(tokens)
    }
    return formatNumberWithSuffix(
      tokens,
      0,
      merged.digitsSmall,
      merged.abbreviate
    )
  }

  const value = amountCNY

  return formatCurrencyValue(value, merged, meta)
}

/**
 * Format CNY for billing/payment contexts. The export name is retained for
 * source compatibility and does not imply or perform USD conversion.
 *
 * @param amountCNY - Amount in CNY
 * @param options - Optional formatting configuration
 * @returns Formatted string with currency symbol (never tokens)
 *
 * @example
 * formatBillingCurrencyFromUSD(10) → "¥10"
 */
export function formatBillingCurrencyFromUSD(
  amountCNY: number | null | undefined,
  options?: CurrencyFormatOptions
): string {
  if (amountCNY == null || Number.isNaN(amountCNY)) return '-'

  const { config } = getCurrencyDisplay()
  const meta = getBillingDisplayMeta(config)
  const merged = mergeOptions(options)
  return formatCurrencyValue(amountCNY, merged, meta)
}

/**
 * Format raw quota values (token units) to display currency.
 *
 * Converts raw quota units to CNY using `quotaPerUnit`.
 *
 * @param quota - Raw quota amount in token units (e.g., 5000000)
 * @param options - Optional formatting configuration
 * @returns Formatted string with currency symbol or token count
 *
 * @example
 * // With quotaPerUnit: 500000
 * formatQuotaWithCurrency(5000000) → "¥10"
 */
export function formatQuotaWithCurrency(
  quota: number | null | undefined,
  options?: CurrencyFormatOptions
): string {
  if (quota == null || Number.isNaN(quota)) return '-'

  const { config } = getCurrencyDisplay()
  const amountCNY = quota / config.quotaPerUnit
  return formatCurrencyFromUSD(amountCNY, options)
}

/**
 * Get the current currency label for UI display.
 *
 * Returns a simple string label representing the current display currency.
 * Useful for labels, tooltips, and UI text.
 *
 * @returns The fixed currency label "CNY"
 *
 * @example
 * getCurrencyLabel() → "CNY"
 *
 * @remarks
 * Use this for:
 * - Currency selector labels
 * - Table column headers
 * - Form field labels
 */
export function getCurrencyLabel(): string {
  return 'CNY'
}

/**
 * Check if currency display is enabled (not in token-only mode).
 *
 * @returns Always true while CNY is the fixed display currency
 *
 * @remarks
 * Use this to conditionally show currency-specific UI elements
 */
export function isCurrencyDisplayEnabled(): boolean {
  const { meta } = getCurrencyDisplay()
  return meta.kind !== 'tokens'
}

/**
 * Format an amount already denominated in CNY without conversion.
 *
 * @param amount - Amount in CNY
 * @param options - Optional formatting configuration
 * @returns Formatted string with appropriate currency symbol
 *
 * @example
 * formatLocalCurrencyAmount(50) → "¥50"
 */
export function formatLocalCurrencyAmount(
  amount: number | null | undefined,
  options?: CurrencyFormatOptions
): string {
  if (amount == null || Number.isNaN(amount)) return '-'

  const { config } = getCurrencyDisplay()
  const meta = getBillingDisplayMeta(config)
  const merged = mergeOptions(options)

  return formatCurrencyValue(amount, merged, meta)
}
