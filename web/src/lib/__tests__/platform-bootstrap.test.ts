/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { describe, expect, it } from 'vitest'

import { shouldBootstrapPlatform } from '@/lib/platform-bootstrap'

describe('platform bootstrap routing', () => {
  it.each([
    '/account',
    '/account/',
    '/account/login',
    '/account/register',
    '/account/recovery',
    '/account/profile',
    '/account/admin',
    '/account/callback',
  ])('keeps platform initialization off account route %s', (pathname) => {
    expect(shouldBootstrapPlatform(pathname)).toBe(false)
  })

  it.each(['/dashboard', '/sign-in', '/setup', '/accounting'])(
    'keeps platform initialization for platform route %s',
    (pathname) => {
      expect(shouldBootstrapPlatform(pathname)).toBe(true)
    }
  )
})
