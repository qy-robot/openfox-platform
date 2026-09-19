/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

function productOrigin(value: unknown, fallback: string): string {
  if (typeof value !== 'string' || !value.trim()) {
    return fallback
  }
  try {
    const url = new URL(value.trim())
    if (url.username || url.password || url.protocol !== 'https:') {
      return fallback
    }
    return url.origin
  } catch {
    return fallback
  }
}

export const AI_STATION_URL = productOrigin(
  import.meta.env.VITE_ROBO_AI_URL,
  'https://ai.openfox.work'
)

export const CONSOLE_URL = productOrigin(
  import.meta.env.VITE_ROBO_CONSOLE_URL,
  'https://dash.openfox.work'
)

export const ACCOUNT_CENTER_URL = productOrigin(
  import.meta.env.VITE_ROBO_ACCOUNT_URL,
  'https://account.openfox.work'
)
