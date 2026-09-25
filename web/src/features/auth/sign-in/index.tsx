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
import { Link, useSearch } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { useTheme } from '@/context/theme-provider'
import { useStatus } from '@/hooks/use-status'

import { AuthLayout } from '../auth-layout'
import { getAccountRegistrationURL } from '../central-account-navigation'
import { TermsFooter } from '../components/terms-footer'
import { CentralAccountSignIn } from './components/central-account-sign-in'
import { UserAuthForm } from './components/user-auth-form'

export function SignIn() {
  const { t } = useTranslation()
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })
  const { status } = useStatus()
  const { resolvedTheme } = useTheme()
  const centralAccountEnabled = status?.account_auth_enabled === true
  const registrationURL = getAccountRegistrationURL({
    accountCenterURL: status?.account_center_url,
    theme: resolvedTheme,
  })

  return (
    <AuthLayout>
      <div className='w-full space-y-9'>
        <div className='space-y-3'>
          <h2 className='text-3xl font-semibold tracking-[-0.04em] sm:text-[2.125rem]'>
            {t('Sign in')}
          </h2>
          {!status?.self_use_mode_enabled &&
            status?.register_enabled !== false && (
              <p className='text-muted-foreground text-left text-base'>
                {t("Don't have an account?")}{' '}
                {centralAccountEnabled ? (
                  <a
                    href={registrationURL}
                    className='hover:text-primary font-medium underline underline-offset-4'
                  >
                    {t('Sign up')}
                  </a>
                ) : (
                  <Link
                    to='/sign-up'
                    className='hover:text-primary font-medium underline underline-offset-4'
                  >
                    {t('Sign up')}
                  </Link>
                )}
                .
              </p>
            )}
        </div>

        {centralAccountEnabled ? (
          <CentralAccountSignIn redirectTo={redirect} />
        ) : (
          <UserAuthForm redirectTo={redirect} />
        )}

        <TermsFooter
          variant='sign-in'
          status={status}
          className='text-center'
        />
      </div>
    </AuthLayout>
  )
}
