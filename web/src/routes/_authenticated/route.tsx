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
import {
  createFileRoute,
  Outlet,
  redirect,
  useLocation,
} from '@tanstack/react-router'

import { AuthenticatedLayout } from '@/components/layout'
import { isCentralAccountMode } from '@/features/auth/sign-in/central-reauth'
import {
  hasValidAuthentication,
  isExplicitSignOutInProgress,
  resolveAuthentication,
} from '@/lib/auth-session'
import { useAuthStore } from '@/stores/auth-store'

export async function resolveProtectedAuthentication(
  _returnTo: string
): Promise<string | null> {
  if (isCentralAccountMode()) {
    // Central account mode is authenticated by the account-center SSO flow.
    // A cold local route must not probe the legacy refresh endpoint: doing so
    // can hang navigation when no local adapter/cookie is available and can
    // silently restore a session after an explicit central sign-out.
    if (hasValidAuthentication()) return null
    if (isExplicitSignOutInProgress()) return null
    return null
  }
  await resolveAuthentication()
  return null
}

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ location }) => {
    // The root guard may have skipped its refresh because no session hint was
    // present. That skip is an optimization for public pages and must not
    // decide a protected route, so resolve against the server before
    // redirecting. An in-memory session returns without a request.
    const authorizationURL = await resolveProtectedAuthentication(location.href)
    if (authorizationURL) {
      throw redirect({ href: authorizationURL, replace: true })
    }

    const { auth } = useAuthStore.getState()

    if (!auth.user || !auth.accessToken) {
      throw redirect({
        to: '/sign-in',
        search: { redirect: location.href },
      })
    }
  },
  component: ProtectedLayout,
})

function ProtectedLayout() {
  const pathname = useLocation({ select: (location) => location.pathname })
  if (pathname.replace(/\/$/, '') === '/desktop/authorize') {
    return <Outlet />
  }
  return <AuthenticatedLayout />
}
