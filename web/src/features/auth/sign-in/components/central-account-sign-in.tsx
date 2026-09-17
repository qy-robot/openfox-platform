/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

/* oxlint-disable react/iframe-missing-sandbox -- the trusted cross-origin account iframe requires scripts and its own origin for exact postMessage validation */

import { Loader2, LogIn } from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { getServerErrorMessage } from '@/lib/server-error-message'

import {
  clearCentralSignedOut,
  consumeFreshCentralLoginRequest,
  isCentralSignedOut,
} from '../central-reauth'
import {
  cancelCentralSSO,
  finishCentralSSOMessage,
  startCentralSSO,
} from '../central-sso'

const SILENT_AUTHORIZATION_TIMEOUT_MS = 5_000
const INTERACTIVE_AUTHORIZATION_TIMEOUT_MS = 120_000

export function CentralAccountSignIn(props: { redirectTo?: string }) {
  const { t } = useTranslation()
  const [pending, setPending] = useState(false)
  const [authorizationUrl, setAuthorizationUrl] = useState<string | null>(null)
  const [visible, setVisible] = useState(false)
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const requestSequence = useRef(0)
  const returnTo = props.redirectTo || '/dashboard'

  const loadAuthorization = useCallback(
    async (
      options: { prompt?: 'none'; reauthenticate?: boolean } = {},
      showFrame = false
    ) => {
      const sequence = ++requestSequence.current
      cancelCentralSSO()
      setAuthorizationUrl(null)
      setVisible(showFrame)
      setPending(true)
      let url: string
      try {
        url = await startCentralSSO(returnTo, options)
      } catch (error) {
        if (sequence !== requestSequence.current) return
        throw error
      }
      if (sequence !== requestSequence.current) return
      setAuthorizationUrl(url)
    },
    [returnTo]
  )

  const cancelAuthorization = useCallback(() => {
    requestSequence.current += 1
    cancelCentralSSO()
    setAuthorizationUrl(null)
    setPending(false)
    setVisible(false)
  }, [])

  useEffect(() => {
    const freshLogin = consumeFreshCentralLoginRequest()
    if (isCentralSignedOut()) return
    void loadAuthorization(
      freshLogin ? { reauthenticate: true } : { prompt: 'none' },
      freshLogin
    ).catch((error) => {
      toast.error(
        getServerErrorMessage(error, t('Unable to start account sign in'))
      )
      setPending(false)
    })
    return () => {
      requestSequence.current += 1
      cancelCentralSSO()
    }
  }, [loadAuthorization, t])

  useEffect(() => {
    if (!authorizationUrl) return
    const timeout = window.setTimeout(
      () => {
        cancelAuthorization()
        toast.error(t('Account sign in timed out. Please try again.'))
      },
      visible
        ? INTERACTIVE_AUTHORIZATION_TIMEOUT_MS
        : SILENT_AUTHORIZATION_TIMEOUT_MS
    )
    return () => window.clearTimeout(timeout)
  }, [authorizationUrl, cancelAuthorization, t, visible])

  useEffect(() => {
    if (!visible) return
    function cancelOnEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') cancelAuthorization()
    }
    window.addEventListener('keydown', cancelOnEscape)
    return () => window.removeEventListener('keydown', cancelOnEscape)
  }, [cancelAuthorization, visible])

  useEffect(() => {
    async function receive(event: MessageEvent) {
      const frameWindow = iframeRef.current?.contentWindow
      if (!frameWindow) return
      try {
        const currentAuthorizationUrl = authorizationUrl
        if (!currentAuthorizationUrl) return
        const result = await finishCentralSSOMessage(
          event,
          frameWindow,
          currentAuthorizationUrl
        )
        if (result.status === 'ignored') return
        if (result.status === 'interaction_required') {
          await loadAuthorization({}, true)
          return
        }
        window.location.assign(result.returnTo)
      } catch (error) {
        toast.error(
          getServerErrorMessage(error, t('Unable to complete account sign in'))
        )
        setPending(false)
      }
    }
    window.addEventListener('message', receive)
    return () => window.removeEventListener('message', receive)
  }, [authorizationUrl, loadAuthorization, t])

  async function startSignIn() {
    try {
      clearCentralSignedOut()
      await loadAuthorization({}, true)
    } catch (error) {
      toast.error(
        getServerErrorMessage(error, t('Unable to start account sign in'))
      )
      setPending(false)
    }
  }

  return (
    <div className='flex flex-col gap-4'>
      <Alert>
        <LogIn />
        <AlertTitle>{t('Unified account sign in')}</AlertTitle>
        <AlertDescription>
          {t(
            'Use one RoboCoding account to sign in securely across RoboCoding services.'
          )}
        </AlertDescription>
      </Alert>
      <Button disabled={pending} onClick={startSignIn}>
        {pending ? (
          <Loader2 data-icon='inline-start' className='animate-spin' />
        ) : (
          <LogIn data-icon='inline-start' />
        )}
        {t('Sign in with your RoboCoding account')}
      </Button>
      {authorizationUrl && (
        <>
          {/* The cross-origin account iframe needs its scripts and host cookies. */}
          <iframe
            key={authorizationUrl}
            ref={iframeRef}
            title={t('RoboCoding Account sign in')}
            src={authorizationUrl}
            className={
              visible ? 'h-[520px] w-full rounded-lg border' : 'hidden'
            }
            sandbox='allow-forms allow-scripts allow-same-origin'
          />
        </>
      )}
      {visible && (
        <Button type='button' variant='outline' onClick={cancelAuthorization}>
          {t('Cancel')}
        </Button>
      )}
    </div>
  )
}
