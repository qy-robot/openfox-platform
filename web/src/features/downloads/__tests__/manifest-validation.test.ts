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
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { describe, expect, test } from 'vitest'

import {
  isSafeDownloadUrl,
  parseDownloadManifest,
} from '../lib/download-manifest'
import type { DownloadManifest } from '../types'

function createManifest(): DownloadManifest {
  return {
    schemaVersion: 1,
    product: 'RoboCoding',
    version: null,
    publishedAt: null,
    downloads: [
      {
        id: 'windows-x64',
        status: 'unavailable',
        url: null,
        fileName: null,
        size: null,
        sha256: null,
      },
      {
        id: 'macos-arm64',
        status: 'unavailable',
        url: null,
        fileName: null,
        size: null,
        sha256: null,
      },
      {
        id: 'macos-x64',
        status: 'unavailable',
        url: null,
        fileName: null,
        size: null,
        sha256: null,
      },
      {
        id: 'linux-x64',
        status: 'unavailable',
        url: null,
        fileName: null,
        size: null,
        sha256: null,
      },
    ],
  }
}

describe('download URL safety', () => {
  test.each([
    '/releases/RoboCoding.dmg',
    'https://cdn.example.com/RoboCoding.exe',
  ])('accepts deployable URL %s', (url) => {
    expect(isSafeDownloadUrl(url)).toBe(true)
  })

  test.each([
    '//untrusted.example/RoboCoding.dmg',
    'http://cdn.example.com/RoboCoding.exe',
    'javascript:alert(1)',
    'RoboCoding.AppImage',
    '/\\untrusted.example/RoboCoding.dmg',
    ' https://cdn.example.com/RoboCoding.exe',
  ])('rejects unsafe URL %s', (url) => {
    expect(isSafeDownloadUrl(url)).toBe(false)
  })
})

describe('download availability', () => {
  test('the deployed manifest starts with every build clearly unavailable', () => {
    const manifestPath = resolve(process.cwd(), 'public/downloads.json')
    const manifest = parseDownloadManifest(
      JSON.parse(readFileSync(manifestPath, 'utf8'))
    )

    expect(manifest.downloads).toHaveLength(4)
    expect(
      manifest.downloads.every((target) => target.status === 'unavailable')
    ).toBe(true)
  })

  test('accepts an unavailable target only when it exposes no link', () => {
    expect(parseDownloadManifest(createManifest()).downloads[0].status).toBe(
      'unavailable'
    )
    const linkedUnavailable = createManifest()
    linkedUnavailable.downloads[0].url = '/releases/not-ready.exe'
    expect(() => parseDownloadManifest(linkedUnavailable)).toThrow()
  })

  test('accepts an available target with a verified link and file name', () => {
    const manifest = createManifest()
    manifest.version = '1.0.0'
    manifest.publishedAt = '2026-09-15T10:00:00Z'
    manifest.downloads[0] = {
      ...manifest.downloads[0],
      status: 'available',
      url: '/releases/RoboCoding-1.0.0.exe',
      fileName: 'RoboCoding-1.0.0.exe',
    }

    expect(parseDownloadManifest(manifest).downloads[0]).toMatchObject({
      status: 'available',
      url: '/releases/RoboCoding-1.0.0.exe',
    })
  })

  test('rejects a live download without a named and dated release', () => {
    const manifest = createManifest()
    manifest.downloads[0] = {
      ...manifest.downloads[0],
      status: 'available',
      url: '/releases/RoboCoding-1.0.0.exe',
      fileName: 'RoboCoding-1.0.0.exe',
    }

    expect(() => parseDownloadManifest(manifest)).toThrow()
  })

  test('rejects an available target with an unsafe or absent link', () => {
    const manifest = createManifest()
    manifest.downloads[0] = {
      ...manifest.downloads[0],
      status: 'available',
      url: 'http://untrusted.example/RoboCoding.exe',
      fileName: 'RoboCoding.exe',
    }

    expect(() => parseDownloadManifest(manifest)).toThrow()
  })
})
