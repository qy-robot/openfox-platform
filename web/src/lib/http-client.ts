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
import axios, { type AxiosRequestConfig } from 'axios'
import { t } from 'i18next'

import {
  clearCentralReauthenticationAttempt,
  currentCentralReturnTo,
  isCentralAccountMode,
  requestFreshCentralLogin,
  reserveCentralReauthentication,
} from '@/features/auth/sign-in/central-reauth'
import {
  applyAuthRotation,
  clearAuthentication,
  getFreshAuthHeaders,
  isExplicitSignOutInProgress,
  refreshAuthentication,
} from '@/lib/auth-session'
import { handleServerError, markServerErrorHandled } from '@/lib/handle-server-error'
import {
  getServerErrorMessage,
  safeServerErrorMessage,
} from '@/lib/server-error-message'
import { useAuthStore } from '@/stores/auth-store'

declare module 'axios' {
  export interface AxiosRequestConfig {
    skipBusinessError?: boolean
    skipErrorHandler?: boolean
    disableDuplicate?: boolean
    skipAuthRefresh?: boolean
    authRetry?: boolean
    acceptAuthRotation?: boolean
    singleUseAuthorization?: boolean
  }
}

export type ApiRequestConfig = AxiosRequestConfig

export const api = axios.create({
  baseURL: '',
  withCredentials: true,
  headers: {
    // no-store forbids storage; no-cache also revalidates any older cached response.
    'Cache-Control': 'no-cache, no-store',
  },
})

const inFlightGet = new Map<string, Promise<unknown>>()
const originalGet = api.get.bind(api)
let centralReauthentication: Promise<boolean> | null = null

api.get = ((url: string, config: ApiRequestConfig = {}) => {
  if (config.disableDuplicate) return originalGet(url, config)

  const params = config.params ? JSON.stringify(config.params) : '{}'
  const sessionSID = useAuthStore.getState().auth.session?.sid || 'anonymous'
  const key = `${sessionSID}:${url}?${params}`
  const existingRequest = inFlightGet.get(key)
  if (existingRequest) return existingRequest

  const request = originalGet(url, config).finally(() => {
    inFlightGet.delete(key)
  })
  inFlightGet.set(key, request)
  return request
}) as typeof api.get

function redirectToSignIn(): void {
  if (
    typeof window !== 'undefined' &&
    window.location.pathname !== '/sign-in'
  ) {
    window.location.replace('/sign-in')
  }
}

async function beginCentralReauthentication(
  forceFreshLogin = false
): Promise<boolean> {
  if (!isCentralAccountMode()) return false
  if (centralReauthentication) return centralReauthentication
  if (!reserveCentralReauthentication()) return false

  const attempt = (async () => {
    try {
      if (forceFreshLogin) requestFreshCentralLogin()
      const returnTo = encodeURIComponent(currentCentralReturnTo())
      window.location.assign(`/sign-in?redirect=${returnTo}`)
      return true
    } catch {
      clearCentralReauthenticationAttempt()
      return false
    }
  })()
  centralReauthentication = attempt
  try {
    return await attempt
  } finally {
    if (centralReauthentication === attempt) centralReauthentication = null
  }
}

function reportExpiredSession(
  error: unknown,
  skipErrorHandler?: boolean
): void {
  if (!skipErrorHandler) {
    handleServerError({
      message: t('Session expired!'),
      [safeServerErrorMessage]: true,
      cause: error,
    })
  }
}

api.interceptors.response.use(
  (response) => {
    if (response.config.acceptAuthRotation && response.data?.success === true) {
      applyAuthRotation(response.data.data)
    }

    return response
  },
  async (error) => {
    const config = error?.config as ApiRequestConfig | undefined
    const skipErrorHandler = config?.skipErrorHandler
    const status = error?.response?.status

    if (status === 401) {
      const auth = useAuthStore.getState().auth
      // Revoked background requests may settle after navigation has finished.
      // Keep their rejection, but do not present it as a new user-facing failure.
      // Explicit login/logout calls still own and report their failures.
      if (
        !config?.skipAuthRefresh &&
        config?.headers?.Authorization &&
        (isExplicitSignOutInProgress() ||
          (!auth.accessToken && auth.bootstrapState === 'complete'))
      ) {
        markServerErrorHandled(error)
        throw error
      }

      if (
        isExplicitSignOutInProgress() ||
        (isCentralAccountMode() && window.location.pathname === '/sign-in')
      ) {
        throw error
      }
      if (config && !config.skipAuthRefresh && isCentralAccountMode()) {
        clearAuthentication(false)
        const forceFreshLogin =
          error?.response?.data?.code === 'AUTH_REAUTH_REQUIRED'
        if (!(await beginCentralReauthentication(forceFreshLogin))) {
          reportExpiredSession(error, skipErrorHandler)
          redirectToSignIn()
        }
      } else if (config && !config.skipAuthRefresh && !config.authRetry) {
        config.authRetry = true
        const outcome = await refreshAuthentication()
        if (outcome.kind === 'authenticated') {
          const token = useAuthStore.getState().auth.accessToken
          if (token) {
            config.headers = {
              ...config.headers,
              Authorization: `Bearer ${token}`,
            }
          }
          return api.request(config)
        }

        if (outcome.kind === 'anonymous' || outcome.kind === 'out_of_sync') {
          reportExpiredSession(error, skipErrorHandler)
          redirectToSignIn()
        }
      } else if (config?.authRetry) {
        clearAuthentication(false)
        reportExpiredSession(error, skipErrorHandler)
        redirectToSignIn()
      } else if (!skipErrorHandler) {
        reportExpiredSession(error)
      }
    }
    if (axios.isAxiosError(error)) error.message = getServerErrorMessage(error)
    throw error
  }
)

api.interceptors.request.use(async (config) => {
  if (config.singleUseAuthorization || config.headers.has('X-Security-Proof')) {
    // Refresh before spending a proof/flow, never by replaying its request.
    const explicitlySkipsRefresh = config.skipAuthRefresh === true
    config.skipAuthRefresh = true
    if (explicitlySkipsRefresh) return config
    const auth = useAuthStore.getState().auth
    const refreshBefore = Math.floor(Date.now() / 1000) + 60
    if (
      !explicitlySkipsRefresh &&
      isCentralAccountMode() &&
      (!auth.accessExpiresAt || auth.accessExpiresAt <= refreshBefore)
    ) {
      clearAuthentication(false)
      await beginCentralReauthentication()
      throw axios.AxiosError.from(
        new Error(t('Session expired!')),
        undefined,
        config
      )
    }
    try {
      const headers = await getFreshAuthHeaders()
      for (const [name, value] of Object.entries(headers)) {
        config.headers.set(name, value)
      }
    } catch (error) {
      throw axios.AxiosError.from(error, undefined, config)
    }
    return config
  }
  const accessToken = useAuthStore.getState().auth.accessToken
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  return config
})
