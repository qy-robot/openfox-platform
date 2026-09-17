/* Copyright (C) 2023-2026 QuantumNous; licensed under GNU AGPLv3 or later. */
import { createFileRoute, notFound } from '@tanstack/react-router'

import { TeamWorkspace, type TeamSection } from '@/features/teams'

const validSections = new Set<TeamSection>([
  'overview',
  'members',
  'usage',
  'settings',
])

export const Route = createFileRoute('/_authenticated/teams/$teamId/$section')({
  component: RouteComponent,
})

function RouteComponent() {
  const params = Route.useParams()
  const teamId = Number(params.teamId)
  if (
    !Number.isInteger(teamId) ||
    teamId <= 0 ||
    !validSections.has(params.section as TeamSection)
  ) {
    throw notFound()
  }
  return (
    <TeamWorkspace teamId={teamId} section={params.section as TeamSection} />
  )
}
