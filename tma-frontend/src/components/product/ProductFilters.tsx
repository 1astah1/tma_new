import { useState } from 'react'

interface Props {
  onFilter: (filters: Record<string, string>) => void
}

const platforms = [
  { id: '', label: 'Все', icon: '/icons/all.png' },
  { id: 'ps4', label: 'PS4', icon: '/icons/Icon_ps.png' },
  { id: 'ps5', label: 'PS5', icon: '/icons/Icon_ps.png' },
  { id: 'xbox', label: 'Xbox', icon: '/icons/Icon_xbox.png' },
]

const types = [
  { id: '', label: 'Все', icon: '/icons/all.png' },
  { id: 'game', label: 'Игры', icon: '/icons/game.png' },
  { id: 'currency', label: 'Валюта', icon: '/icons/valuta.png' },
  { id: 'subscription', label: 'Подписки', icon: '/icons/Sub.png' },
]

export function ProductFilters({ onFilter }: Props) {
  const [platform, setPlatform] = useState('')
  const [type, setType] = useState('')
  const [search, setSearch] = useState('')

  const apply = () => {
    const f: Record<string, string> = {}
    if (platform) f.platform = platform
    if (type) f.type = type
    if (search) f.search = search
    onFilter(f)
  }

  const FilterButton = ({
    label,
    icon,
    isActive,
    onClick,
  }: {
    label: string
    icon: string
    isActive: boolean
    onClick: () => void
  }) => (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 px-4 py-3 rounded-2xl text-sm font-medium transition-all duration-300 ${
        isActive
          ? 'bg-gradient-to-br from-violet-600 to-purple-700 text-white shadow-lg shadow-purple-500/30'
          : 'bg-white/5 backdrop-blur-sm border border-white/10 text-gray-300 hover:bg-white/10'
      }`}
    >
      <img src={icon} alt="" className="w-5 h-5" />
      {label}
    </button>
  )

  return (
    <div className="space-y-6 p-5 bg-tg-secondary/50 backdrop-blur-xl rounded-3xl border border-white/5">
      <div className="relative">
        <div className="absolute left-4 top-1/2 -translate-y-1/2">
          <img src="/icons/search.svg" alt="" className="w-5 h-5" />
        </div>
        <input
          type="text"
          placeholder="Поиск товаров..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setTimeout(apply, 300) }}
          className="w-full pl-12 pr-4 py-3.5 rounded-2xl bg-white/5 backdrop-blur-sm border border-white/10 text-sm text-white placeholder-gray-400 focus:outline-none focus:border-violet-500/50 transition-colors"
        />
      </div>

      <div>
        <div className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Платформа</div>
        <div className="flex gap-2 flex-wrap">
          {platforms.map((p) => (
            <FilterButton
              key={p.id}
              label={p.label}
              icon={p.icon}
              isActive={platform === p.id}
              onClick={() => { setPlatform(p.id); setTimeout(apply) }}
            />
          ))}
        </div>
      </div>

      <div>
        <div className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Тип</div>
        <div className="flex gap-2 flex-wrap">
          {types.map((t) => (
            <FilterButton
              key={t.id}
              label={t.label}
              icon={t.icon}
              isActive={type === t.id}
              onClick={() => { setType(t.id); setTimeout(apply) }}
            />
          ))}
        </div>
      </div>
    </div>
  )
}
