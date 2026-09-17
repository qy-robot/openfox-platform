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
import { beforeEach, describe, expect, test } from 'vitest'

import {
  DEFAULT_LOGO,
  DEFAULT_SYSTEM_NAME,
  normalizeSystemName,
} from '@/lib/constants'
import { mapStatusDataToConfig, readCachedStatus } from '@/lib/status-query'
import { useSystemConfigStore } from '@/stores/system-config-store'

beforeEach(() => {
  window.localStorage.clear()
  useSystemConfigStore.setState(useSystemConfigStore.getInitialState(), true)
})

describe('RoboCoding brand defaults', () => {
  test.each([
    'New API',
    'new-api',
    'newapi',
    'robocoding',
    'RoboCodingAI',
    '  NEW API  ',
  ])('normalizes the legacy default %s', (legacyName) => {
    expect(normalizeSystemName(legacyName)).toBe(DEFAULT_SYSTEM_NAME)
    expect(mapStatusDataToConfig({ system_name: legacyName })).toMatchObject({
      systemName: DEFAULT_SYSTEM_NAME,
      logo: DEFAULT_LOGO,
    })
  })

  test('preserves a deliberate custom system name and logo', () => {
    expect(
      mapStatusDataToConfig({
        system_name: '  Acme Robotics  ',
        logo: '  https://cdn.example.com/acme.svg  ',
      })
    ).toMatchObject({
      systemName: 'Acme Robotics',
      logo: 'https://cdn.example.com/acme.svg',
    })
  })

  test('normalizes a legacy cached status before first render', () => {
    window.localStorage.setItem(
      'status',
      JSON.stringify({ system_name: 'New API', logo: '/logo.png' })
    )

    expect(readCachedStatus()).toMatchObject({
      system_name: DEFAULT_SYSTEM_NAME,
      logo: DEFAULT_LOGO,
    })
  })

  test('normalizes a legacy persisted config during hydration', async () => {
    window.localStorage.setItem(
      'system-config-storage',
      JSON.stringify({
        state: {
          config: { systemName: 'RoboCodingAI', logo: '' },
          loadedLogoUrl: '',
        },
        version: 0,
      })
    )

    await useSystemConfigStore.persist.rehydrate()

    expect(useSystemConfigStore.getState().config.systemName).toBe(
      DEFAULT_SYSTEM_NAME
    )
    expect(useSystemConfigStore.getState().config.logo).toBe(DEFAULT_LOGO)
  })
})
