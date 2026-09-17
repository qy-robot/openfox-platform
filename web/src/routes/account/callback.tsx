/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'

import { CentralSSOCallback } from '@/features/auth/sign-in/central-sso-callback'

const searchSchema = z.object({
  code: z.string().optional(),
  state: z.string().optional(),
})

export const Route = createFileRoute('/account/callback')({
  validateSearch: searchSchema,
  component: AccountCallbackRoute,
})

function AccountCallbackRoute() {
  const search = Route.useSearch()
  return <CentralSSOCallback code={search.code} state={search.state} />
}
