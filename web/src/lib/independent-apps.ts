import { ACCOUNT_CENTER_URL, CONSOLE_URL } from './product-links'

/** Compatibility navigation only; application code lives in the independent services. */
export function getIndependentAppURL(pathname: string): string | null {
  if (pathname === '/account/callback') return null
  if (pathname === '/account' || pathname.startsWith('/account/')) {
    const section = pathname.split('/')[2]
    const allowed = [
      'login',
      'register',
      'recovery',
      'profile',
      'security',
      'creator',
      'github',
      'admin',
    ]
    return `${ACCOUNT_CENTER_URL}/account/${allowed.includes(section) ? section : 'profile'}`
  }
  if (pathname === '/workbench' || pathname.startsWith('/workbench/')) {
    const kind = pathname.split('/')[2]
    const allowed = ['skills', 'devices', 'repositories']
    return `${CONSOLE_URL}/workbench/${allowed.includes(kind) ? kind : 'skills'}`
  }
  if (pathname === '/learning' || pathname === '/learning/') {
    return `${CONSOLE_URL}/learning`
  }
  return null
}
