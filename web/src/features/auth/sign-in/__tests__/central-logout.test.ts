import { afterEach, expect, it, vi } from 'vitest'

import { api } from '@/lib/api'

import { logoutCentralBrowserSession } from '../central-logout'

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
})
function setup(enabled = true) {
  vi.spyOn(api, 'get').mockResolvedValue({
    data: {
      success: true,
      data: {
        account_auth_enabled: enabled,
        account_center_url: 'https://account.example.com',
      },
    },
  })
  const request = vi
    .fn()
    .mockResolvedValue(new Response(JSON.stringify({ success: true })))
  vi.stubGlobal('fetch', request)
  return request
}
it('revokes the browser identity cookie without forwarding application credentials', async () => {
  const request = setup()
  await logoutCentralBrowserSession()
  expect(request).toHaveBeenCalledWith(
    'https://account.example.com/v1/auth/browser-logout',
    expect.objectContaining({
      method: 'POST',
      credentials: 'include',
      redirect: 'error',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
    })
  )
})
it('keeps native authentication logout independent of the account service', async () => {
  const request = setup(false)
  await logoutCentralBrowserSession()
  expect(request).not.toHaveBeenCalled()
})
it.each(['AUTH_REQUIRED', 'AUTH_SESSION_INVALID'])(
  'accepts an already absent identity session: %s',
  async (code) => {
    setup().mockResolvedValue(
      new Response(JSON.stringify({ success: false, code }), { status: 401 })
    )
    await expect(logoutCentralBrowserSession()).resolves.toBeUndefined()
  }
)
it.each([
  ['ACCOUNT_UNAVAILABLE', 503],
  ['AUTH_ORIGIN_FORBIDDEN', 403],
  ['OTHER', 401],
])('does not report logout success for %s', async (code, status) => {
  setup().mockResolvedValue(
    new Response(JSON.stringify({ success: false, code }), { status })
  )
  await expect(logoutCentralBrowserSession()).rejects.toThrow()
})
it('does not report success when the account network request fails', async () => {
  setup().mockRejectedValue(new TypeError('Failed to fetch'))
  await expect(logoutCentralBrowserSession()).rejects.toThrow()
})

it('rejects an insecure remote account URL before sending cookies', async () => {
  const request = setup()
  vi.spyOn(api, 'get').mockResolvedValue({
    data: {
      success: true,
      data: {
        account_auth_enabled: true,
        account_center_url: 'http://account.example.com',
      },
    },
  })
  await expect(logoutCentralBrowserSession()).rejects.toThrow()
  expect(request).not.toHaveBeenCalled()
})

it('does not report success for an invalid successful response', async () => {
  setup().mockResolvedValue(new Response(JSON.stringify({ success: false })))
  await expect(logoutCentralBrowserSession()).rejects.toThrow()
})
