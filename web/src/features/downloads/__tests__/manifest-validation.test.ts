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
    product: 'OpenFox',
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
    '/releases/OpenFox.dmg',
    'https://cdn.example.com/OpenFox.exe',
  ])('accepts deployable URL %s', (url) => {
    expect(isSafeDownloadUrl(url)).toBe(true)
  })

  test.each([
    '//untrusted.example/OpenFox.dmg',
    'http://cdn.example.com/OpenFox.exe',
    'javascript:alert(1)',
    'OpenFox.AppImage',
    '/\\untrusted.example/OpenFox.dmg',
    ' https://cdn.example.com/OpenFox.exe',
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
      url: '/releases/OpenFox-1.0.0.exe',
      fileName: 'OpenFox-1.0.0.exe',
    }

    expect(parseDownloadManifest(manifest).downloads[0]).toMatchObject({
      status: 'available',
      url: '/releases/OpenFox-1.0.0.exe',
    })
  })

  test('accepts an optional or nullable release changelog', () => {
    expect(parseDownloadManifest(createManifest()).changelog).toBeUndefined()

    const noRelease = createManifest()
    noRelease.changelog = null
    expect(parseDownloadManifest(noRelease).changelog).toBeNull()

    const released = createManifest()
    released.version = '1.0.0'
    released.changelog = '新增桌面安装包\n修复升级流程'
    expect(parseDownloadManifest(released).changelog).toBe(
      '新增桌面安装包\n修复升级流程'
    )
  })

  test('rejects an empty changelog or changelog without a release', () => {
    const empty = createManifest()
    empty.version = '1.0.0'
    empty.changelog = '   '
    expect(() => parseDownloadManifest(empty)).toThrow()

    const unpublished = createManifest()
    unpublished.changelog = '不应展示的更新日志'
    expect(() => parseDownloadManifest(unpublished)).toThrow()
  })

  test('rejects a live download without a named and dated release', () => {
    const manifest = createManifest()
    manifest.downloads[0] = {
      ...manifest.downloads[0],
      status: 'available',
      url: '/releases/OpenFox-1.0.0.exe',
      fileName: 'OpenFox-1.0.0.exe',
    }

    expect(() => parseDownloadManifest(manifest)).toThrow()
  })

  test('rejects an available target with an unsafe or absent link', () => {
    const manifest = createManifest()
    manifest.downloads[0] = {
      ...manifest.downloads[0],
      status: 'available',
      url: 'http://untrusted.example/OpenFox.exe',
      fileName: 'OpenFox.exe',
    }

    expect(() => parseDownloadManifest(manifest)).toThrow()
  })
})
