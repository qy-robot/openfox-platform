/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { useNavigate } from '@tanstack/react-router'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { ErrorState } from '@/components/error-state'
import { LoadingState } from '@/components/loading-state'
import { getServerErrorMessage } from '@/lib/server-error-message'

import { finishCentralSSO } from './central-sso'

export function CentralSSOCallback(props: { code?: string; state?: string }) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!props.code || !props.state) return
    let active = true
    void finishCentralSSO(props.code, props.state)
      .then((returnTo) => {
        if (active) return navigate({ href: returnTo, replace: true })
      })
      .catch((reason) => {
        if (active) {
          setError(
            getServerErrorMessage(reason, t('Unable to finish account sign in'))
          )
        }
      })
    return () => {
      active = false
    }
  }, [navigate, props.code, props.state, t])

  const callbackError =
    !props.code || !props.state
      ? t('The account callback is incomplete.')
      : error
  if (callbackError) {
    return (
      <ErrorState
        title={t('Account sign in failed')}
        description={callbackError}
      />
    )
  }
  return <LoadingState message={t('Finishing account sign in...')} />
}
