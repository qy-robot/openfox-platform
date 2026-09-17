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
import assert from 'node:assert/strict'

import { describe, test } from 'vitest'

import { ROLE } from '@/lib/roles'

import { canAccessUsageLogsSection } from '../access'
import { resolveLogsViewAccess } from '../components/usage-logs-provider'

describe('usage log route access', () => {
  test('limits regular users to their common usage logs', () => {
    assert.equal(canAccessUsageLogsSection('common', ROLE.USER), true)
    assert.equal(canAccessUsageLogsSection('audit', ROLE.USER), false)
    assert.equal(canAccessUsageLogsSection('task', ROLE.USER), false)
    assert.equal(canAccessUsageLogsSection('drawing', ROLE.USER), false)
  })

  test('allows admins to open audit and task log routes', () => {
    assert.equal(canAccessUsageLogsSection('audit', ROLE.ADMIN), true)
    assert.equal(canAccessUsageLogsSection('task', ROLE.ADMIN), true)
    assert.equal(canAccessUsageLogsSection('drawing', ROLE.SUPER_ADMIN), true)
  })
})

describe('usage log access tier', () => {
  test('keeps users and elevated self views on the self tier', () => {
    assert.equal(resolveLogsViewAccess(ROLE.USER, 'all'), 'self')
    assert.equal(resolveLogsViewAccess(ROLE.ADMIN, 'self'), 'self')
    assert.equal(resolveLogsViewAccess(ROLE.SUPER_ADMIN, 'self'), 'self')
  })

  test('distinguishes admin and root while viewing all logs', () => {
    assert.equal(resolveLogsViewAccess(ROLE.ADMIN, 'all'), 'admin')
    assert.equal(resolveLogsViewAccess(ROLE.SUPER_ADMIN, 'all'), 'root')
  })
})
