import { useState } from 'react'
import { XboxIcon, PS4Icon, PS5Icon, GamepadIcon, CoinStackIcon, CrownIcon, SearchIcon } from '../icons/PlatformIcons'

interface Props {
  onFilter: (filters: Record<string, string>) => void
}

const platforms = [
  { id: '', label: 'Все', icon: XboxIcon },
  { id: 'ps4', label: 'PS4', icon: PS4Icon },
  { id: 'ps5', label: 'PS5', icon: PS5Icon },
  { id: 'xbox', label: 'Xbox', icon: XboxIcon },
]

const types = [
  { id: '', label: 'Все', icon: CoinStackIcon },
  { id: 'game', label: 'Игры', icon: GamepadIcon },
  { id: 'currency', label: 'Валюта', icon: CoinStackIcon },
  { id: 'subscription', label: 'Подписки', icon: CrownIcon },
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

  return (
    <div className="space-y-6 p-6 bg-slate-950 rounded-2xl border border-slate-800">
      {/* Search Input */}
      <div className="relative">
        <div className="absolute left-4 top-1/2 -translate-y-1/2 text-purple-500">
          <SearchIcon />
        </div>
        <input
          type="text"
          placeholder="Поиск товаров..."
          value={search}
          onChange={(e) => { setSearch(e.target.value); setTimeout(apply, 300) }}
          className="w-full pl-12 pr-4 py-3 rounded-lg border border-slate-700 bg-slate-900 text-slate-300 placeholder-slate-500 text-sm focus:outline-none focus:border-purple-500 transition"
        />
      </div>

      {/* Platforms */}
      <div>
        <div className="text-xs font-semibold text-slate-400 mb-3 tracking-wide">ПЛАТФОРМА</div>
        <div className="flex gap-2 flex-wrap">
          {platforms.map((p) => {
            const Icon = p.icon
            const isActive = platform === p.id
            return (
              <button
                key={p.id}
                onClick={() => { setPlatform(p.id); setTimeout(apply) }}
                className={`flex items-center gap-2 px-4 py-2.5 rounded-xl font-medium transition duration-200 ${
                  isActive
                    ? 'bg-gradient-to-r from-purple-600 to-purple-500 text-white shadow-lg shadow-purple-500/50'
                    : 'bg-slate-800 text-slate-300 border border-slate-700 hover:border-slate-600'
                }`}
              >
                <Icon />
                <span>{p.label}</span>
              </button>
            )
          })}
        </div>
      </div>

      {/* Types */}
      <div>
        <div className="text-xs font-semibold text-slate-400 mb-3 tracking-wide">ТИП</div>
        <div className="flex gap-2 flex-wrap">
          {types.map((t) => {
            const Icon = t.icon
            const isActive = type === t.id
            return (
              <button
                key={t.id}
                onClick={() => { setType(t.id); setTimeout(apply) }}
                className={`flex items-center gap-2 px-4 py-2.5 rounded-xl font-medium transition duration-200 ${
                  isActive
                    ? 'bg-gradient-to-r from-purple-600 to-purple-500 text-white shadow-lg shadow-purple-500/50'
                    : 'bg-slate-800 text-slate-300 border border-slate-700 hover:border-slate-600'
                }`}
              >
                <Icon />
                <span>{t.label}</span>
              </button>
            )
          })}
        </div>
      </div>
    </div>
  )
}
