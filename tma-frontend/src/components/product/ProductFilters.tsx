import { useEffect, useRef, useState } from 'react'
import { homePlatformSections } from '../../constants/homeSections'

export type FilterValue = {
  platform?: string
  type?: string
  search?: string
  section?: string
}

interface Props {
  mode: 'home' | 'catalog'
  value: FilterValue
  onChange: (filters: Record<string, string>) => void
}

const platforms = [
  { id: '', label: 'Все', icon: '/icons/all.png' },
  { id: 'ps', label: 'PS', icon: '/icons/Icon_ps.png' },
  { id: 'xbox', label: 'Xbox', icon: '/icons/Icon_xbox.png' },
]

const types = [
  { id: '', label: 'Все', icon: '/icons/all.png' },
  { id: 'game', label: 'Игры', icon: '/icons/game.png' },
  // { id: 'currency', label: 'Валюта', icon: '/icons/valuta.png' },
  // { id: 'subscription', label: 'Подписки', icon: '/icons/Sub.png' },
]

function buildFilters(platform: string, type: string, search: string, section = '') {
  const f: Record<string, string> = {}
  if (platform) f.platform = platform
  if (type) f.type = type
  if (section) f.section = section
  if (search.trim()) f.search = search.trim()
  return f
}

function SearchInput({
  value,
  onChange,
}: {
  value: string
  onChange: (value: string) => void
}) {
  return (
    <div className="relative">
      <img
        src="/icons/search.svg"
        alt=""
        className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 opacity-40"
      />
      <input
        type="search"
        enterKeyHint="search"
        placeholder="Поиск..."
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-10 w-full rounded-xl border border-white/[0.08] bg-white/[0.04] pl-9 pr-9 text-base text-white placeholder:text-white/30 focus:border-amber-500/40 focus:outline-none"
      />
      {value ? (
        <button
          type="button"
          aria-label="Очистить"
          onClick={() => onChange('')}
          className="absolute right-2 top-1/2 flex h-6 w-6 -translate-y-1/2 items-center justify-center rounded-full text-white/40 hover:bg-white/10 hover:text-white"
        >
          ×
        </button>
      ) : null}
    </div>
  )
}

function SegmentButton({
  label,
  icon,
  active,
  onClick,
}: {
  label: string
  icon: string
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex flex-1 items-center justify-center gap-1.5 rounded-lg py-2.5 text-sm font-semibold transition-colors ${
        active ? 'bg-amber-600 text-white shadow-sm' : 'text-white/45 hover:text-white/70'
      }`}
    >
      <img src={icon} alt="" className="h-4 w-4 object-contain" />
      {label}
    </button>
  )
}

function SectionButton({
  label,
  icon,
  active,
  onClick,
}: {
  label: string
  icon: string
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-center gap-2 rounded-xl border px-3 py-2.5 text-left text-sm font-medium transition-colors ${
        active
          ? 'border-amber-500/60 bg-amber-600/20 text-white'
          : 'border-white/[0.06] bg-white/[0.03] text-white/50 hover:border-white/12 hover:text-white/70'
      }`}
    >
      <img src={icon} alt="" className="h-4 w-4 shrink-0 object-contain opacity-80" />
      <span className="leading-tight">{label}</span>
    </button>
  )
}

export function ProductFilters({ mode, value, onChange }: Props) {
  const platform = value.platform ?? ''
  const type = value.type ?? ''
  const section = value.section ?? 'game'
  const [searchInput, setSearchInput] = useState(value.search ?? '')
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    setSearchInput(value.search ?? '')
  }, [value.search])

  const emit = (nextPlatform: string, nextType: string, nextSearch: string, nextSection = '') => {
    onChange(buildFilters(nextPlatform, nextType, nextSearch, nextSection))
  }

  const buildHomeFilters = (nextPlatform: string, nextSearch: string, nextSection = section) => {
    const f: Record<string, string> = {}
    if (nextPlatform) {
      f.platform = nextPlatform
      f.section = nextSection || 'game'
    }
    if (nextSearch.trim()) f.search = nextSearch.trim()
    return f
  }

  const buildCatalogPlatformFilters = (nextPlatform: string, nextSearch: string, nextSection = 'game') => {
    const f: Record<string, string> = {}
    if (nextPlatform) {
      f.platform = nextPlatform
      f.section = nextSection
    } else if (type) {
      f.type = type
    }
    if (nextSearch.trim()) f.search = nextSearch.trim()
    return f
  }

  const setPlatform = (id: string) => {
    if (mode === 'home') {
      onChange(buildHomeFilters(id, searchInput, id ? 'game' : ''))
      return
    }
    if (id) {
      onChange(buildCatalogPlatformFilters(id, searchInput, 'game'))
      return
    }
    emit('', type, searchInput)
  }

  const setSection = (id: string) => {
    if (mode === 'home') {
      onChange(buildHomeFilters(platform, searchInput, id))
      return
    }
    onChange(buildCatalogPlatformFilters(platform, searchInput, id))
  }

  const setType = (id: string) => {
    emit(platform, id, searchInput)
  }

  const onSearchChange = (next: string) => {
    setSearchInput(next)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    debounceRef.current = setTimeout(() => {
      if (mode === 'home') {
        onChange(buildHomeFilters(platform, next, platform ? section : ''))
        return
      }
      if (platform) {
        onChange(buildCatalogPlatformFilters(platform, next, section))
        return
      }
      emit(platform, type, next)
    }, 300)
  }

  const showSections = !!platform

  return (
    <div className="space-y-3 rounded-2xl border border-white/[0.06] bg-white/[0.02] p-3">
      <SearchInput value={searchInput} onChange={onSearchChange} />

      <div className="flex gap-1 rounded-xl bg-black/20 p-1">
        {platforms.map((p) => (
          <SegmentButton
            key={p.id || 'all'}
            label={p.label}
            icon={p.icon}
            active={platform === p.id}
            onClick={() => setPlatform(p.id)}
          />
        ))}
      </div>

      {showSections ? (
        <div className="grid grid-cols-2 gap-2">
          {homePlatformSections.map((item) => (
            <SectionButton
              key={item.id}
              label={item.label}
              icon={item.icon}
              active={section === item.id}
              onClick={() => setSection(item.id)}
            />
          ))}
        </div>
      ) : mode === 'catalog' ? (
        <div className="grid grid-cols-2 gap-2">
          {types.map((t) => (
            <SectionButton
              key={t.id || 'all-types'}
              label={t.label}
              icon={t.icon}
              active={type === t.id}
              onClick={() => setType(t.id)}
            />
          ))}
        </div>
      ) : null}
    </div>
  )
}
