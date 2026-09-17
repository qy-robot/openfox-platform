/* Copyright (C) 2023-2026 QuantumNous; licensed under GNU AGPLv3 or later. */
import { useMutation, useQuery } from '@tanstack/react-query'
import { CheckCircle2, Laptop, XCircle } from 'lucide-react'
import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { SystemBrand } from '@/components/layout/components/system-brand'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { handleServerError } from '@/lib/handle-server-error'

import { decideDesktopAuthorization, getDesktopAuthorization } from './api'

export function DesktopAuthorizationPage(props: { userCode: string }) {
  const { t } = useTranslation()
  const [decision, setDecision] = useState<'approved' | 'denied' | null>(null)
  const authorization = useQuery({
    queryKey: ['desktop-authorization', props.userCode],
    queryFn: () => getDesktopAuthorization(props.userCode),
    retry: false,
  })
  const decide = useMutation({
    mutationFn: (approve: boolean) =>
      decideDesktopAuthorization(props.userCode, approve),
    onSuccess: (_, approve) => setDecision(approve ? 'approved' : 'denied'),
    onError: (error) => handleServerError(error),
  })
  let authorizationContent: ReactNode
  if (decision) {
    authorizationContent = (
      <Alert>
        {decision === 'approved' ? <CheckCircle2 /> : <XCircle />}
        <AlertTitle>
          {decision === 'approved'
            ? t('Desktop access approved')
            : t('Desktop access denied')}
        </AlertTitle>
        <AlertDescription>
          {t('This browser page can now be closed.')}
        </AlertDescription>
      </Alert>
    )
  } else if (authorization.data) {
    authorizationContent = (
      <>
        <div className='bg-muted/60 grid grid-cols-1 gap-3 rounded-xl p-4 sm:grid-cols-2'>
          <div>
            <div className='text-muted-foreground text-xs'>{t('Device')}</div>
            <div className='mt-1 font-medium break-words'>
              {authorization.data.device_name}
            </div>
          </div>
          <div>
            <div className='text-muted-foreground text-xs'>
              {t('Verification code')}
            </div>
            <div className='mt-1 font-mono text-lg font-bold tracking-widest'>
              {authorization.data.user_code}
            </div>
          </div>
        </div>
        <Alert>
          <AlertTitle>{t('Review before approving')}</AlertTitle>
          <AlertDescription>
            {t(
              'Approving signs this desktop app into your account. Do not approve a code sent by someone else.'
            )}
          </AlertDescription>
        </Alert>
        <div className='grid grid-cols-2 gap-3'>
          <Button
            variant='outline'
            disabled={decide.isPending}
            onClick={() => decide.mutate(false)}
          >
            <XCircle />
            {t('Deny')}
          </Button>
          <Button
            disabled={decide.isPending}
            onClick={() => decide.mutate(true)}
          >
            <CheckCircle2 />
            {t('Approve')}
          </Button>
        </div>
      </>
    )
  } else {
    authorizationContent = (
      <div className='text-muted-foreground py-10 text-center'>
        {authorization.isError
          ? t('This authorization request is invalid or expired.')
          : t('Loading authorization request...')}
      </div>
    )
  }
  return (
    <main className='bg-background flex min-h-svh flex-col items-center justify-center gap-6 px-4 py-10'>
      <SystemBrand variant='inline' />
      <Card className='w-full max-w-md'>
        <CardHeader className='text-center'>
          <Laptop className='text-primary mx-auto size-10' />
          <CardTitle>
            {decision
              ? t('Authorization complete')
              : t('Confirm desktop sign-in')}
          </CardTitle>
          <CardDescription>
            {decision
              ? t('You can return to the robocoding desktop app.')
              : t(
                  'Only continue if this code and device match the desktop app in front of you.'
                )}
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-5'>{authorizationContent}</CardContent>
      </Card>
    </main>
  )
}
