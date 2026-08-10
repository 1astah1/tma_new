import { formatPriceOrManager } from '../../utils/format'
import {
  EditionPlatformKey,
  EDITION_PLATFORM_LABELS,
  editionRegionLabel,
  listEditionPlatforms,
} from '../../utils/productEditionPricing'
import { ProductEditionCatalog, ProductEditionOption } from '../../types/product'

type EditionPricingPanelProps = {
  catalog: ProductEditionCatalog
  platform: EditionPlatformKey
  editionId: string | null
  onPlatformChange: (platform: EditionPlatformKey) => void
  onEditionChange: (edition: ProductEditionOption) => void
}

export function EditionPricingPanel({
  catalog,
  platform,
  editionId,
  onPlatformChange,
  onEditionChange,
}: EditionPricingPanelProps) {
  const platforms = listEditionPlatforms(catalog)
  const editions = catalog[platform] ?? []
  const selected = editions.find((item) => item.id === editionId) ?? editions[0] ?? null

  return (
    <div className="space-y-3">
      <div className="-mx-1 flex gap-2 overflow-x-auto pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {platforms.map((key) => {
          const meta = EDITION_PLATFORM_LABELS[key]
          const active = platform === key
          return (
            <button
              key={key}
              type="button"
              onClick={() => onPlatformChange(key)}
              className={`min-w-[108px] shrink-0 rounded-xl border px-3 py-2.5 text-left transition-colors ${
                active
                  ? 'border-amber-500/50 bg-amber-500/10'
                  : 'border-white/10 bg-white/[0.04] hover:bg-white/[0.07]'
              }`}
            >
              <div className="text-[10px] font-bold uppercase tracking-[0.08em] text-white/40">
                {key === 'xbox' ? 'XBOX/PC' : meta.title.toUpperCase()}
              </div>
              <div className="mt-0.5 text-sm font-bold text-white">{meta.region}</div>
            </button>
          )
        })}
      </div>

      <div className="rounded-2xl border border-white/10 bg-black/25 p-3">
        <div className="mb-2.5 text-[11px] font-bold uppercase tracking-[0.14em] text-white/35">
          {editionRegionLabel(platform)}
        </div>
        <div className="space-y-2">
          {editions.map((edition) => {
            const active = (selected?.id ?? editionId) === edition.id
            return (
              <button
                key={edition.id}
                type="button"
                onClick={() => onEditionChange(edition)}
                className={`flex w-full items-center justify-between gap-3 rounded-xl border px-3 py-3 text-left transition ${
                  active
                    ? 'border-amber-500/45 bg-amber-500/10'
                    : 'border-white/10 bg-white/[0.03] hover:border-white/18'
                }`}
              >
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-bold text-white">{edition.name}</div>
                  {edition.discount_label ? (
                    <div className="mt-0.5 text-[11px] font-medium text-amber-300/75">
                      с учётом скидки {edition.discount_label.replace('−', '')}
                    </div>
                  ) : null}
                </div>
                <div className="shrink-0 text-base font-black text-white sm:text-lg">
                  {formatPriceOrManager(edition.price)}
                </div>
              </button>
            )
          })}
        </div>
      </div>
    </div>
  )
}
