import { useState } from 'react'

interface Props {
  onFilter: (filters: Record<string, string>) => void
}

const platforms = [
  { id: '', label: 'Все' },
  { id: 'ps4', label: 'PS4' },
  { id: 'ps5', label: 'PS5' },
  { id: 'xbox', label: 'Xbox' },
]

const types = [
  { id: '', label: 'Все' },
  { id: 'game', label: 'Игры' },
  { id: 'currency', label: 'Валюта' },
  { id: 'subscription', label: 'Подписки' },
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
    <div className="space-y-3 p-4 bg-tg-secondary rounded-xl">
      <input
        type="text"
        placeholder="🔍 Поиск товаров..."
        value={search}
        onChange={(e) => { setSearch(e.target.value); setTimeout(apply, 300) }}
        className="w-full px-3 py-2 rounded-lg border border-gray-200 bg-white text-sm"
      />
      <div>
        <div className="text-xs text-tg-hint mb-1">Платформа</div>
        <div className="flex gap-1 flex-wrap">
          {platforms.map((p) => (
            <button
              key={p.id}
              onClick={() => { setPlatform(p.id); setTimeout(apply) }}
              className={`px-3 py-1 rounded-full text-xs font-medium transition ${
                platform === p.id ? 'bg-tg-button text-white' : 'bg-white text-gray-600 border border-gray-200'
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>
      <div>
        <div className="text-xs text-tg-hint mb-1">Тип</div>
        <div className="flex gap-1 flex-wrap">
          {types.map((t) => (
            <button
              key={t.id}
              onClick={() => { setType(t.id); setTimeout(apply) }}
              className={`px-3 py-1 rounded-full text-xs font-medium transition ${
                type === t.id ? 'bg-tg-button text-white' : 'bg-white text-gray-600 border border-gray-200'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
