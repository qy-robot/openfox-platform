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
import { useQuery } from '@tanstack/react-query'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { DataTableViewModeToggle } from '@/components/data-table'
import { ErrorState } from '@/components/error-state'
import { SectionPageLayout } from '@/components/layout'
import {
  EmptyState,
  LoadingSkeleton,
  ModelCardGrid,
  ModelDetailsDrawer,
  PricingTable,
  SearchBar,
} from '@/features/pricing/components'
import {
  DEFAULT_TOKEN_UNIT,
  VIEW_MODES,
  type ViewMode,
} from '@/features/pricing/constants'
import { usePricingData } from '@/features/pricing/hooks'
import { getUserModels } from '@/lib/api'
import { requireServerSuccess } from '@/lib/server-error-message'

import {
  filterAvailableModels,
  searchAvailableModels,
} from './lib/available-models'

/**
 * Console page listing the models the signed-in user can call, reusing the
 * public model square's catalog data and presentation components narrowed by
 * `/api/user/models`.
 */
export function AvailableModels() {
  const { t } = useTranslation()
  const [searchInput, setSearchInput] = useState('')
  const [viewMode, setViewMode] = useState<ViewMode>(VIEW_MODES.CARD)
  const [selectedModelName, setSelectedModelName] = useState<string | null>(
    null
  )

  const pricing = usePricingData()
  const userModelsQuery = useQuery({
    queryKey: ['user-models'],
    queryFn: async () => requireServerSuccess(await getUserModels()),
    staleTime: 5 * 60 * 1000,
  })

  const availableModels = useMemo(
    () =>
      filterAvailableModels(pricing.models, userModelsQuery.data?.data ?? []),
    [pricing.models, userModelsQuery.data]
  )
  const filteredModels = useMemo(
    () => searchAvailableModels(availableModels, searchInput),
    [availableModels, searchInput]
  )

  const isLoading = pricing.isLoading || userModelsQuery.isLoading
  const hasError = Boolean(pricing.error || userModelsQuery.error)

  const selectedModel = useMemo(
    () =>
      selectedModelName
        ? (availableModels.find(
            (model) => model.model_name === selectedModelName
          ) ?? null)
        : null,
    [availableModels, selectedModelName]
  )

  const renderContent = () => {
    if (hasError) {
      return (
        <ErrorState
          onRetry={() => {
            pricing.refetch()
            userModelsQuery.refetch()
          }}
        />
      )
    }
    if (isLoading) {
      return <LoadingSkeleton viewMode={viewMode} />
    }
    if (filteredModels.length === 0) {
      return (
        <EmptyState
          searchQuery={searchInput}
          hasActiveFilters={searchInput.trim() !== ''}
          onClearFilters={() => setSearchInput('')}
        />
      )
    }
    if (viewMode === VIEW_MODES.TABLE) {
      return (
        <PricingTable
          models={filteredModels}
          priceRate={pricing.priceRate}
          usdExchangeRate={pricing.usdExchangeRate}
          onModelClick={setSelectedModelName}
        />
      )
    }
    return (
      <ModelCardGrid
        models={filteredModels}
        onModelClick={setSelectedModelName}
        priceRate={pricing.priceRate}
        usdExchangeRate={pricing.usdExchangeRate}
      />
    )
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Available Models')}</SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        <div className='flex items-center gap-2'>
          <SearchBar
            value={searchInput}
            onChange={setSearchInput}
            onClear={() => setSearchInput('')}
            placeholder={t('Search model name, provider, endpoint, or tag...')}
            className='w-56 sm:w-72'
          />
          <DataTableViewModeToggle value={viewMode} onChange={setViewMode} />
        </div>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='space-y-4'>
          {!isLoading && !hasError ? (
            <p className='text-muted-foreground text-sm'>
              {t('You have access to {{count}} models', {
                count: availableModels.length,
              })}
            </p>
          ) : null}
          {renderContent()}
        </div>
      </SectionPageLayout.Content>
      {selectedModel && (
        <ModelDetailsDrawer
          open={Boolean(selectedModel)}
          onOpenChange={(open) => {
            if (!open) setSelectedModelName(null)
          }}
          model={selectedModel}
          groupRatio={pricing.groupRatio}
          usableGroup={pricing.usableGroup}
          endpointMap={
            pricing.endpointMap as Record<
              string,
              { path?: string; method?: string }
            >
          }
          autoGroups={pricing.autoGroups}
          priceRate={pricing.priceRate}
          usdExchangeRate={pricing.usdExchangeRate}
          tokenUnit={DEFAULT_TOKEN_UNIT}
        />
      )}
    </SectionPageLayout>
  )
}
