/* Copyright (C) 2023-2026 QuantumNous; licensed under GNU AGPLv3 or later. */
import { createFileRoute } from '@tanstack/react-router'

import { TeamsHome } from '@/features/teams'

export const Route = createFileRoute('/_authenticated/teams/')({
  component: TeamsHome,
})
