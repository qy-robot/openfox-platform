/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { markCentralSignedOut } from '../central-reauth'
import { CentralAccountSignIn } from '../components/central-account-sign-in'

const {
  cancelCentralSSO,
  finishCentralSSOMessage,
  startCentralSSO,
  toastError,
} = vi.hoisted(() => ({
  cancelCentralSSO: vi.fn(),
  finishCentralSSOMessage: vi.fn(),
  startCentralSSO: vi.fn(),
  toastError: vi.fn(),
}))

vi.mock('../central-sso', () => ({
  cancelCentralSSO,
  finishCentralSSOMessage,
  startCentralSSO,
}))

vi.mock('sonner', () => ({ toast: { error: toastError } }))

afterEach(() => {
  cleanup()
  sessionStorage.clear()
  vi.useRealTimers()
  vi.clearAllMocks()
})

it('replaces the silent iframe before starting interactive authorization', async () => {
  startCentralSSO
    .mockResolvedValueOnce('https://account.openzrob.com/silent')
    .mockResolvedValueOnce('https://account.openzrob.com/interactive')
  finishCentralSSOMessage.mockResolvedValueOnce({
    status: 'interaction_required',
    error: 'login_required',
  })

  render(<CentralAccountSignIn redirectTo='/dashboard' />)

  const silentFrame = await screen.findByTitle('RoboCoding Account sign in')
  expect(silentFrame).toHaveClass('hidden')

  await act(async () => {
    window.dispatchEvent(new MessageEvent('message'))
  })

  await waitFor(() => expect(startCentralSSO).toHaveBeenCalledTimes(2))
  const interactiveFrame = await screen.findByTitle(
    'RoboCoding Account sign in'
  )
  expect(interactiveFrame).not.toBe(silentFrame)
  expect(interactiveFrame).toHaveAttribute(
    'src',
    'https://account.openzrob.com/interactive'
  )
  expect(interactiveFrame).not.toHaveClass('hidden')
})

it('ignores unrelated window messages without surfacing an error', async () => {
  startCentralSSO.mockResolvedValueOnce('https://account.openzrob.com/silent')
  finishCentralSSOMessage.mockResolvedValueOnce({ status: 'ignored' })

  render(<CentralAccountSignIn />)
  await screen.findByTitle('RoboCoding Account sign in')

  await act(async () => {
    window.dispatchEvent(new MessageEvent('message'))
  })

  expect(toastError).not.toHaveBeenCalled()
  expect(startCentralSSO).toHaveBeenCalledTimes(1)
})

it('cancels a stalled authorization attempt after the timeout', async () => {
  vi.useFakeTimers()
  startCentralSSO.mockResolvedValueOnce('https://account.openzrob.com/silent')

  render(<CentralAccountSignIn />)
  await act(async () => {})
  expect(screen.getByTitle('RoboCoding Account sign in')).toBeInTheDocument()

  await act(async () => {
    vi.advanceTimersByTime(5_000)
  })

  expect(screen.queryByTitle('RoboCoding Account sign in')).toBeNull()
  expect(cancelCentralSSO).toHaveBeenCalled()
  expect(toastError).toHaveBeenCalledWith(
    'Account sign in timed out. Please try again.'
  )
})

it('allows two minutes for a visible account login', async () => {
  vi.useFakeTimers()
  sessionStorage.setItem('robocoding:central-fresh-login', '1')
  startCentralSSO.mockResolvedValueOnce(
    'https://account.openzrob.com/interactive'
  )

  render(<CentralAccountSignIn />)
  await act(async () => {})

  await act(async () => {
    vi.advanceTimersByTime(119_999)
  })
  expect(screen.getByTitle('RoboCoding Account sign in')).toBeInTheDocument()

  await act(async () => {
    vi.advanceTimersByTime(1)
  })
  expect(screen.queryByTitle('RoboCoding Account sign in')).toBeNull()
})

it('lets the user cancel an interactive authorization with Escape', async () => {
  sessionStorage.setItem('robocoding:central-fresh-login', '1')
  startCentralSSO.mockResolvedValueOnce(
    'https://account.openzrob.com/interactive'
  )

  render(<CentralAccountSignIn />)
  await screen.findByTitle('RoboCoding Account sign in')
  expect(screen.getByRole('button', { name: 'Cancel' })).toBeVisible()

  fireEvent.keyDown(window, { key: 'Escape' })

  expect(screen.queryByTitle('RoboCoding Account sign in')).toBeNull()
  expect(screen.queryByRole('button', { name: 'Cancel' })).toBeNull()
})

it('stays signed out until the user explicitly starts a new login', async () => {
  markCentralSignedOut()
  startCentralSSO.mockResolvedValueOnce(
    'https://account.openzrob.com/interactive'
  )

  render(<CentralAccountSignIn />)
  await act(async () => {})

  expect(startCentralSSO).not.toHaveBeenCalled()
  fireEvent.click(
    screen.getByRole('button', { name: 'Sign in with your RoboCoding account' })
  )

  await waitFor(() => expect(startCentralSSO).toHaveBeenCalledTimes(1))
  expect(screen.getByTitle('RoboCoding Account sign in')).toBeVisible()
})
