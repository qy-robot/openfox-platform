/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { afterEach, expect, it } from 'vitest'

import {
  isCentralSignedOut,
  markCentralSignedOutIfEnabled,
} from '../central-reauth'

afterEach(() => {
  localStorage.removeItem('status')
  sessionStorage.clear()
})

it('preserves a remote sign-out in a central-account tab', () => {
  localStorage.setItem('status', JSON.stringify({ account_auth_enabled: true }))

  markCentralSignedOutIfEnabled()

  expect(isCentralSignedOut()).toBe(true)
})

it('does not add the central marker when local authentication is enabled', () => {
  markCentralSignedOutIfEnabled()

  expect(isCentralSignedOut()).toBe(false)
})
