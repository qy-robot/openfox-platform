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
import { fireEvent, render, screen } from '@testing-library/react'
import { useState } from 'react'
import { expect, test } from 'vitest'

import { DownloadBuilds } from '../components/download-builds'
import { ReleaseStatus } from '../components/release-status'
import type { DownloadManifest } from '../types'

const releasedManifest: DownloadManifest = {
  schemaVersion: 1,
  product: 'RoboCoding',
  version: '1.0.0',
  publishedAt: '2026-09-15T10:00:00Z',
  downloads: [
    {
      id: 'windows-x64',
      status: 'available',
      url: '/releases/RoboCoding-1.0.0.exe',
      fileName: 'RoboCoding-1.0.0.exe',
      size: null,
      sha256: null,
    },
    ...(['macos-arm64', 'macos-x64', 'linux-x64'] as const).map((id) => ({
      id,
      status: 'unavailable' as const,
      url: null,
      fileName: null,
      size: null,
      sha256: null,
    })),
  ],
}

function StaleDataRecovery() {
  const [error, setError] = useState<Error | null>(new Error('offline'))
  return (
    <DownloadBuilds
      manifest={releasedManifest}
      error={error}
      onRetry={() => setError(null)}
    />
  )
}

test('a refresh error hides stale links until retry recovers current data', () => {
  render(<StaleDataRecovery />)

  expect(
    screen.getByText('Downloads are temporarily unavailable')
  ).toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: 'Download' })
  ).not.toBeInTheDocument()

  fireEvent.click(screen.getByRole('button', { name: 'Retry' }))

  const downloadAction = screen.getByRole('button', { name: 'Download' })
  expect(downloadAction).toHaveAttribute(
    'href',
    '/releases/RoboCoding-1.0.0.exe'
  )
  expect(
    screen.queryByText('Downloads are temporarily unavailable')
  ).not.toBeInTheDocument()
})

test('a refresh error hides a stale release version in the status box', () => {
  render(
    <ReleaseStatus manifest={releasedManifest} error={new Error('offline')} />
  )

  expect(screen.getByText('Release status unavailable')).toBeInTheDocument()
  expect(screen.queryByText('1.0.0')).not.toBeInTheDocument()
})
