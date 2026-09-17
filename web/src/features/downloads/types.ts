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
export type DownloadTargetId =
  | 'windows-x64'
  | 'macos-arm64'
  | 'macos-x64'
  | 'linux-x64'

export type DownloadTarget = {
  id: DownloadTargetId
  status: 'available' | 'unavailable'
  url: string | null
  fileName: string | null
  size: string | null
  sha256: string | null
}

export type DownloadManifest = {
  schemaVersion: 1
  product: 'RoboCoding'
  version: string | null
  publishedAt: string | null
  downloads: DownloadTarget[]
}
