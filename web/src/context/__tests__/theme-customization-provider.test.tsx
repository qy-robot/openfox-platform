/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.
*/
import { act, cleanup, renderHook, waitFor } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  ThemeCustomizationProvider,
  useThemeCustomization,
} from '../theme-customization-provider'

const state = vi.hoisted(() => ({ centralAccountEnabled: false }))

vi.mock('@/hooks/use-status', () => ({
  useStatus: () => ({
    status: { account_auth_enabled: state.centralAccountEnabled },
  }),
}))

function wrapper(props: { children: ReactNode }) {
  return (
    <ThemeCustomizationProvider>{props.children}</ThemeCustomizationProvider>
  )
}

function setPresetCookie(value: string) {
  document.cookie = `theme_preset=${value}; path=/`
}

afterEach(() => {
  cleanup()
  state.centralAccountEnabled = false
  document.cookie = 'theme_preset=; path=/; max-age=0'
  document.body.removeAttribute('data-theme-preset')
  document.body.removeAttribute('data-theme-font')
})

describe('ThemeCustomizationProvider', () => {
  it('locks a stored warm preset to the company default in central account mode', async () => {
    setPresetCookie('anthropic')
    state.centralAccountEnabled = true

    const { result, rerender } = renderHook(useThemeCustomization, { wrapper })

    await waitFor(() => {
      expect(result.current.customization.preset).toBe('default')
      expect(document.body).not.toHaveAttribute('data-theme-preset')
      expect(document.body).toHaveAttribute('data-theme-font', 'sans')
    })
    result.current.setPreset('rose-garden')
    expect(document.cookie).toContain('theme_preset=anthropic')

    state.centralAccountEnabled = false
    rerender()

    await waitFor(() => {
      expect(result.current.customization.preset).toBe('anthropic')
      expect(document.body).toHaveAttribute('data-theme-preset', 'anthropic')
    })
  })

  it('clears a saved preset when resetting preferences in central account mode', async () => {
    setPresetCookie('anthropic')
    state.centralAccountEnabled = true
    const { result, rerender } = renderHook(useThemeCustomization, { wrapper })
    act(() => result.current.resetCustomization())
    expect(document.cookie).not.toContain('theme_preset=')
    state.centralAccountEnabled = false
    rerender()
    await waitFor(() => {
      expect(result.current.customization.preset).toBe('default')
      expect(document.body).not.toHaveAttribute('data-theme-preset')
    })
  })

  it('keeps a stored preset selectable for a standalone installation', async () => {
    setPresetCookie('lake-view')

    const { result } = renderHook(useThemeCustomization, { wrapper })

    await waitFor(() => {
      expect(result.current.customization.preset).toBe('lake-view')
      expect(document.body).toHaveAttribute('data-theme-preset', 'lake-view')
    })
    expect(result.current.isPresetLocked).toBe(false)
  })
})
