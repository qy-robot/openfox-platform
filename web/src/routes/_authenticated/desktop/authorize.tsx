/* Copyright (C) 2023-2026 QuantumNous; licensed under GNU AGPLv3 or later. */
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'

import { DesktopAuthorizationPage } from '@/features/desktop-authorization'

export const Route = createFileRoute('/_authenticated/desktop/authorize')({
  validateSearch: z.object({ user_code: z.string().min(4).max(32) }),
  component: RouteComponent,
})

function RouteComponent() {
  const search = Route.useSearch()
  return <DesktopAuthorizationPage userCode={search.user_code} />
}
