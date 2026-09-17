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
import z from 'zod'

import type { DownloadManifest, DownloadTargetId } from '../types'

export const DOWNLOAD_TARGET_IDS = [
  'windows-x64',
  'macos-arm64',
  'macos-x64',
  'linux-x64',
] as const satisfies readonly DownloadTargetId[]

export function isSafeDownloadUrl(value: string): boolean {
  if (value.trim() !== value) {
    return false
  }

  try {
    if (value.startsWith('/')) {
      const baseUrl = new URL('https://download-manifest.invalid')
      return new URL(value, baseUrl).origin === baseUrl.origin
    }
    return /^https:\/\//i.test(value) && new URL(value).protocol === 'https:'
  } catch {
    return false
  }
}

const downloadTargetSchema = z
  .object({
    id: z.enum(DOWNLOAD_TARGET_IDS),
    status: z.enum(['available', 'unavailable']),
    url: z.string().nullable(),
    fileName: z.string().trim().min(1).nullable(),
    size: z.string().trim().min(1).nullable(),
    sha256: z
      .string()
      .regex(/^[a-fA-F0-9]{64}$/)
      .nullable(),
  })
  .strict()
  .superRefine((target, context) => {
    if (target.status === 'available') {
      if (target.url == null || !isSafeDownloadUrl(target.url)) {
        context.addIssue({
          code: 'custom',
          path: ['url'],
          message: 'Available downloads require an HTTPS or root-relative URL',
        })
      }
      if (target.fileName == null) {
        context.addIssue({
          code: 'custom',
          path: ['fileName'],
          message: 'Available downloads require a file name',
        })
      }
      return
    }

    if (target.url != null) {
      context.addIssue({
        code: 'custom',
        path: ['url'],
        message: 'Unavailable downloads cannot expose a URL',
      })
    }
  })

const downloadManifestSchema = z
  .object({
    schemaVersion: z.literal(1),
    product: z.literal('RoboCoding'),
    version: z.string().trim().min(1).nullable(),
    publishedAt: z.iso.datetime().nullable(),
    changelog: z.string().trim().min(1).nullable().optional(),
    downloads: z.array(downloadTargetSchema).length(DOWNLOAD_TARGET_IDS.length),
  })
  .strict()
  .superRefine((manifest, context) => {
    const ids = manifest.downloads.map((target) => target.id)
    for (const targetId of DOWNLOAD_TARGET_IDS) {
      if (ids.filter((id) => id === targetId).length !== 1) {
        context.addIssue({
          code: 'custom',
          path: ['downloads'],
          message: `Download target ${targetId} must appear exactly once`,
        })
      }
    }

    if (manifest.downloads.some((target) => target.status === 'available')) {
      if (manifest.version == null) {
        context.addIssue({
          code: 'custom',
          path: ['version'],
          message: 'A manifest with downloads requires a release version',
        })
      }
      if (manifest.publishedAt == null) {
        context.addIssue({
          code: 'custom',
          path: ['publishedAt'],
          message: 'A manifest with downloads requires a publication time',
        })
      }
    }

    if (manifest.version == null && manifest.changelog != null) {
      context.addIssue({
        code: 'custom',
        path: ['changelog'],
        message: 'A changelog requires a release version',
      })
    }
  })

export function parseDownloadManifest(value: unknown): DownloadManifest {
  return downloadManifestSchema.parse(value)
}
