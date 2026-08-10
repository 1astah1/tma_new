export function getAdminRoles(): string[] {
  try {
    const admin = localStorage.getItem('admin')
    if (!admin) return []
    return JSON.parse(admin).roles || []
  } catch {
    return []
  }
}

export function canAccessResource(resource: string): boolean {
  const roles = getAdminRoles()
  if (roles.length === 0) return true
  if (roles.includes('super_admin') || roles.includes('game_manager')) return true

  const financeOnly = ['orders', 'settings', 'logs', 'promos']
  const supportOnly = ['orders', 'users', 'faq-items', 'templates', 'logs']

  if (roles.includes('finance')) {
    return financeOnly.includes(resource)
  }
  if (roles.includes('support')) {
    return supportOnly.includes(resource)
  }
  return true
}
