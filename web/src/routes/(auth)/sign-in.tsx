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
import { z } from 'zod'

import { sanitizeAuthRedirect } from '@/features/auth/lib/auth-redirect'
import { SignIn } from '@/features/auth/sign-in'
import { isCentralAccountMode } from '@/features/auth/sign-in/central-reauth'
import {
  hasValidAuthentication,
  resolveAuthentication,
} from '@/lib/auth-session'
import { ensureStatus } from '@/lib/status-query'

const searchSchema = z.object({
  redirect: z.string().optional(),
})

export async function resolveSignInAuthentication(): Promise<boolean> {
  if (isCentralAccountMode()) return hasValidAuthentication()
  return (await resolveAuthentication()).kind === 'authenticated'
}

export const Route = createFileRoute('/(auth)/sign-in')({
  component: SignIn,
  validateSearch: searchSchema,
  beforeLoad: async ({ context, search }) => {
    // 登录表单形态由 status 的 account_auth_enabled 决定；先等共享 status
    // 缓存就绪再渲染，避免本地/中央登录表单在 status 返回前后切换闪烁。
    // 已有缓存快照时立即返回并后台刷新；拉取失败不阻塞登录页。
    await ensureStatus(context.queryClient).catch(() => null)
    // 根 guard 可能因为没有会话提示而跳过了 refresh。此处必须回源确认，
    // 否则持有有效 Refresh Cookie 的用户会被要求重新输入密码。
    if (await resolveSignInAuthentication()) {
      const target =
        sanitizeAuthRedirect(search?.redirect, window.location.origin) ??
        '/dashboard'
      throw redirect({ href: target, replace: true })
    }
  },
})
