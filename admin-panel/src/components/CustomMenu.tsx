import { Menu, MenuItemLink, useResourceDefinitions } from 'react-admin'
import DashboardIcon from '@mui/icons-material/Dashboard'
import { ListSubheader } from '@mui/material'
import { canAccessResource } from '../utils/roles'

const GROUPS: { label: string; items: string[] }[] = [
  { label: 'Каталог', items: ['catalog-imports', 'products'] },
  { label: 'Витрина', items: ['home-categories', 'home-banners'] },
  { label: 'Продажи', items: ['orders', 'users', 'promos'] },
  { label: 'Контент', items: ['faq-items', 'templates'] },
  { label: 'Система', items: ['settings', 'admins', 'logs'] },
]

export const CustomMenu = () => {
  const resources = useResourceDefinitions()

  return (
    <Menu>
      <MenuItemLink to="/" primaryText="Дашборд" leftIcon={<DashboardIcon />} />
      {GROUPS.map((group) => {
        const visible = group.items.filter((name) => resources[name] && canAccessResource(name))
        if (visible.length === 0) return null
        return (
          <div key={group.label}>
            <ListSubheader sx={{ lineHeight: 2, fontSize: '0.7rem', fontWeight: 700, letterSpacing: 0.5 }}>
              {group.label}
            </ListSubheader>
            {visible.map((name) => {
              const def = resources[name]
              const Icon = def.icon
              return (
                <MenuItemLink
                  key={name}
                  to={`/${name}`}
                  primaryText={def.options?.label || name}
                  leftIcon={Icon ? <Icon /> : undefined}
                />
              )
            })}
          </div>
        )
      })}
    </Menu>
  )
}
