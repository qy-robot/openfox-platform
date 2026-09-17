/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

/** Account-center and SSO callback routes use account APIs, not platform boot APIs. */
export function shouldBootstrapPlatform(pathname: string): boolean {
  return pathname !== '/account' && !pathname.startsWith('/account/')
}
