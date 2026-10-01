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
import { describe, expect, it } from 'vitest'

import type { PricingModel } from '@/features/pricing/types'

import {
  filterAvailableModels,
  searchAvailableModels,
} from '../lib/available-models'

let nextId = 0
function pricedModel(
  modelName: string,
  extras: Partial<PricingModel> = {}
): PricingModel {
  nextId += 1
  return {
    id: nextId,
    model_name: modelName,
    quota_type: 0,
    model_ratio: 1,
    completion_ratio: 1,
    enable_groups: ['default'],
    ...extras,
  }
}

const catalog = [
  pricedModel('gpt-4o', { vendor_name: 'OpenAI' }),
  pricedModel('claude-sonnet-4', {
    vendor_name: 'Anthropic',
    description: 'Balanced coding model',
  }),
  pricedModel('deepseek-v3', { vendor_name: 'DeepSeek' }),
]

describe('filterAvailableModels', () => {
  it('keeps only priced models present in the user model list, preserving catalog order', () => {
    const result = filterAvailableModels(catalog, [
      'deepseek-v3',
      'gpt-4o',
      'unlisted-model',
    ])
    expect(result.map((model) => model.model_name)).toEqual([
      'gpt-4o',
      'deepseek-v3',
    ])
  })

  it('returns no models when the user list is empty', () => {
    expect(filterAvailableModels(catalog, [])).toEqual([])
  })

  it('returns no models when the catalog is empty', () => {
    expect(filterAvailableModels([], ['gpt-4o'])).toEqual([])
  })

  it('drops user models that have no pricing metadata instead of synthesizing entries', () => {
    const result = filterAvailableModels(catalog, ['gpt-4o', 'mystery-model'])
    expect(result.map((model) => model.model_name)).toEqual(['gpt-4o'])
  })
})

describe('searchAvailableModels', () => {
  it('returns all models for blank or whitespace-only queries', () => {
    expect(searchAvailableModels(catalog, '')).toHaveLength(3)
    expect(searchAvailableModels(catalog, '   ')).toHaveLength(3)
  })

  it('matches model names case-insensitively', () => {
    expect(
      searchAvailableModels(catalog, 'GPT-4O').map((model) => model.model_name)
    ).toEqual(['gpt-4o'])
  })

  it('matches vendor names and descriptions', () => {
    expect(
      searchAvailableModels(catalog, 'anthropic').map(
        (model) => model.model_name
      )
    ).toEqual(['claude-sonnet-4'])
    expect(
      searchAvailableModels(catalog, 'coding').map((model) => model.model_name)
    ).toEqual(['claude-sonnet-4'])
  })

  it('returns no models when nothing matches', () => {
    expect(searchAvailableModels(catalog, 'llama')).toEqual([])
  })
})
