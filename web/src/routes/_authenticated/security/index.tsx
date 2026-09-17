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
import { createFileRoute, redirect } from '@tanstack/react-router'

import {
  getIdentitySecurityURL,
  usesCentralIdentityManagement,
} from '@/features/auth/central-account-navigation'
import { Security } from '@/features/security'

export const Route = createFileRoute('/_authenticated/security/')({
  beforeLoad: () => {
    if (usesCentralIdentityManagement()) {
      throw redirect({ href: getIdentitySecurityURL(), replace: true })
    }
  },
  component: Security,
})
