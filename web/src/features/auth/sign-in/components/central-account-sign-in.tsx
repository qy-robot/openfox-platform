/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { LoadingState } from '@/components/loading-state'
import { Button } from '@/components/ui/button'
import { getServerErrorMessage } from '@/lib/server-error-message'

import {
  clearCentralSignedOut,
  consumeFreshCentralLoginRequest,
  isCentralSignedOut,
} from '../central-reauth'
import { startCentralSSO } from '../central-sso'

/** The backend-provided authorization URL must be a plain redirect request. */
function isAccountAuthorizationURL(value: string): boolean {
  try {
    const url = new URL(value)
    const localHTTP =
      url.protocol === 'http:' &&
      (url.hostname === 'localhost' || url.hostname === '127.0.0.1')
    if (
      (url.protocol !== 'https:' && !localHTTP) ||
      url.username ||
      url.password
    ) {
      return false
    }
    return (
      url.pathname === '/v1/oauth/authorize' &&
      !url.searchParams.has('response_mode') &&
      url.searchParams.has('code_challenge') &&
      url.searchParams.has('state')
    )
  } catch {
    return false
  }
}

export function CentralAccountSignIn(props: { redirectTo?: string }) {
  const { t } = useTranslation()
  const [error, setError] = useState<string | null>(null)
  const [starting, setStarting] = useState(false)
  const redirecting = useRef(false)
  const returnTo = props.redirectTo || '/dashboard'

  const beginRedirect = useCallback(async () => {
    if (redirecting.current) return
    redirecting.current = true
    setStarting(true)
    setError(null)
    try {
      const reauthenticate = consumeFreshCentralLoginRequest()
      const authorizationUrl = await startCentralSSO(returnTo, {
        reauthenticate,
      })
      if (!isAccountAuthorizationURL(authorizationUrl)) {
        throw new Error('The account authorization URL is invalid.')
      }
      clearCentralSignedOut()
      window.location.assign(authorizationUrl)
    } catch (caught) {
      redirecting.current = false
      setStarting(false)
      setError(
        getServerErrorMessage(caught, t('Unable to start account sign in'))
      )
    }
  }, [returnTo, t])

  useEffect(() => {
    // An explicit sign-out must not immediately drag the user back into the
    // account login page; any other entry redirects straight to the IdP.
    if (isCentralSignedOut()) return
    // oxlint-disable-next-line react/set-state-in-effect -- mount navigates the whole page to the IdP; the pending state must render before the redirect
    void beginRedirect()
  }, [beginRedirect])

  if (starting) {
    return (
      <LoadingState message={t('Redirecting you to the account center...')} />
    )
  }

  return (
    <div className='w-full space-y-4'>
      {error ? (
        <p className='text-destructive text-sm'>{error}</p>
      ) : (
        <p className='text-muted-foreground text-base'>
          {t('Sign in with your OpenFox account')}
        </p>
      )}
      <Button
        type='button'
        className='w-full justify-center'
        onClick={() => void beginRedirect()}
      >
        {t('Continue to account center')}
      </Button>
    </div>
  )
}
