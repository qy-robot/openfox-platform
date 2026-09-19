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
 * Application-wide constants
 */

// System Configuration Defaults
export const DEFAULT_SYSTEM_NAME = 'OpenFox'
export const DEFAULT_SYSTEM_SUBTITLE = ''
export const DEFAULT_LOGO = '/robocodingai-logo.png'

const LEGACY_DEFAULT_SYSTEM_NAMES = new Set([
  'new api',
  'new-api',
  'newapi',
  'robocoding',
  'robocodingai',
])

/** Replace upstream and provisional defaults while preserving custom branding. */
export function normalizeSystemName(value: unknown): string {
  if (typeof value !== 'string') return DEFAULT_SYSTEM_NAME

  const name = value.trim()
  if (!name || LEGACY_DEFAULT_SYSTEM_NAMES.has(name.toLowerCase())) {
    return DEFAULT_SYSTEM_NAME
  }
  return name
}

/** Keep deliberate custom logos and fall back to the bundled OpenFox mark. */
export function normalizeSystemLogo(value: unknown): string {
  if (typeof value !== 'string') return DEFAULT_LOGO
  const logo = value.trim()
  return !logo || logo === '/logo.png' ? DEFAULT_LOGO : logo
}

// LocalStorage Keys
export const STORAGE_KEYS = {
  SYSTEM_NAME: 'system_name',
  LOGO: 'logo',
  FOOTER_HTML: 'footer_html',
} as const
