/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

// openzrob.com / openfox.work 双域名族并行对外（ICP 备案期间 openzrob.com 为生效入口）。
// 同级应用入口按当前访问域名动态派生，不构建期写死，保证从任一族访问时跳转留在同族。
const DOMAIN_FAMILIES = ['openzrob.com', 'openfox.work'] as const
const FALLBACK_DOMAIN_FAMILY = 'openzrob.com'

export function resolveDomainFamily(hostname: string): string {
  return (
    DOMAIN_FAMILIES.find(
      (family) => hostname === family || hostname.endsWith(`.${family}`)
    ) ?? FALLBACK_DOMAIN_FAMILY
  )
}

function siblingOrigin(subdomain: string): string {
  const hostname = typeof window === 'undefined' ? '' : window.location.hostname
  return `https://${subdomain}.${resolveDomainFamily(hostname)}`
}

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
  siblingOrigin('ai')
)

export const CONSOLE_URL = productOrigin(
  import.meta.env.VITE_ROBO_CONSOLE_URL,
  siblingOrigin('dash')
)

export const ACCOUNT_CENTER_URL = productOrigin(
  import.meta.env.VITE_ROBO_ACCOUNT_URL,
  siblingOrigin('account')
)
