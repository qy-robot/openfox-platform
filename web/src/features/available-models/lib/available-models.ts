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
import type { PricingModel } from '@/features/pricing/types'

/**
 * Narrow the priced catalog down to the models the current user can actually
 * call (names come from `/api/user/models`). Pricing order is preserved so
 * the page keeps the backend's stable listing.
 */
export function filterAvailableModels(
  models: PricingModel[],
  userModels: string[]
): PricingModel[] {
  const available = new Set(userModels)
  return models.filter((model) => available.has(model.model_name))
}

/** Case-insensitive search across model name, vendor name and description. */
export function searchAvailableModels(
  models: PricingModel[],
  search: string
): PricingModel[] {
  const query = search.trim().toLowerCase()
  if (!query) return models
  return models.filter(
    (model) =>
      model.model_name.toLowerCase().includes(query) ||
      (model.vendor_name ?? '').toLowerCase().includes(query) ||
      (model.description ?? '').toLowerCase().includes(query)
  )
}
