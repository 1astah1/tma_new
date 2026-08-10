import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { formatCardPriceOrManager, MANAGER_PRICE_LABEL } from '../../utils/format'
import { isPreorderProduct } from '../../utils/productPreorder'
import { getProductDisplayTitle, getUnifiedProductTitle } from '../../utils/productDisplayTitle'
import {
  CatalogCard,
  PlatformFamily,
  getCatalogCardFamilies,
  getCatalogCardPrice,
  platformFamily,
} from '../../utils/groupProducts'
import { useWishlist } from '../../stores/wishlistStore'

const familyLabels: Record<PlatformFamily, string> = {
  ps: 'PS',
  xbox: 'Xbox',
  pc: 'PC',
  other: 'Game',
}

const familyColors: Record<PlatformFamily, string> = {
  ps: 'bg-indigo-500/20 text-indigo-300',
  xbox: 'bg-green-500/20 text-green-300',
  pc: 'bg-cyan-500/20 text-cyan-300',
  other: 'bg-white/10 text-white/60',
}

export function ProductCard({ card }: { card: CatalogCard }) {
  const nav = useNavigate()
  const { toggleItem, isWishlisted } = useWishlist()
  const product = card.primary
  const preorder = card.variants.some((item) => isPreorderProduct(item))
  const hasDiscount = !preorder && product.discount_percent > 0
  const priceInfo = getCatalogCardPrice(card)
  const families = getCatalogCardFamilies(card)
  const badgeFamilies = families.length > 0 ? families : [platformFamily(product.platform)]
  const multiPlatform = badgeFamilies.length > 1
  const wishlisted = card.variants.some((item) => isWishlisted(item.id))
  const [useBlurredBg, setUseBlurredBg] = useState(false)

  const displayTitle = multiPlatform
    ? getUnifiedProductTitle(product, true)
    : getProductDisplayTitle(product)

  const priceLabel = formatCardPriceOrManager(priceInfo.price, priceInfo.from || multiPlatform)
  const priceNeedsManager = priceLabel === MANAGER_PRICE_LABEL

  const openProduct = () => nav(`/product/${card.id}`)

  return (
    <div
      onClick={openProduct}
      className="group relative flex h-full cursor-pointer flex-col overflow-hidden rounded-3xl border border-white/10 bg-[#161616] shadow-[0_14px_35px_rgba(0,0,0,0.28)] transition-all duration-300 hover:-translate-y-0.5 hover:border-amber-400/40 hover:shadow-[0_18px_45px_rgba(201,168,76,0.22)]"
    >
      <div className="relative aspect-[3/4] shrink-0 overflow-hidden bg-[#0c0c0c]">
        {product.image_url ? (
          <>
            {useBlurredBg && (
              <img
                src={product.image_url}
                alt=""
                className="absolute inset-0 h-full w-full scale-[1.2] object-cover opacity-50 blur-2xl"
              />
            )}
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_18%,rgba(201,168,76,0.28),transparent_42%),linear-gradient(180deg,rgba(255,255,255,0.05),rgba(0,0,0,0.18))]" />
            <img
              src={product.image_url}
              alt={product.title}
              onLoad={(event) => {
                const img = event.currentTarget
                setUseBlurredBg(img.naturalWidth < 400 || img.naturalHeight < 400)
              }}
              className="relative z-10 h-full w-full object-contain transition-transform duration-300 group-hover:scale-[1.025]"
            />
          </>
        ) : (
          <div className="flex h-full w-full items-center justify-center bg-gradient-to-br from-amber-900/40 to-slate-950 text-5xl">
            {product.type === 'game' ? '🎮' : product.type === 'currency' ? '💰' : '📦'}
          </div>
        )}
        <div className="pointer-events-none absolute inset-x-0 bottom-0 z-20 h-16 bg-gradient-to-t from-[#161616] via-[#161616]/70 to-transparent" />
        <div className="pointer-events-none absolute bottom-2 right-2 z-20 rounded-lg bg-black/50 px-1.5 py-1 shadow-lg shadow-black/40 backdrop-blur-sm">
          <img src="/wiht_logo.png" alt="" className="h-8 w-auto drop-shadow-md" />
        </div>

        {preorder && (
          <span className="absolute left-2.5 top-2.5 z-30 rounded-full bg-sky-500 px-2 py-1 text-[10px] font-black text-white shadow">
            ПРЕДЗАКАЗ
          </span>
        )}

        {hasDiscount && !preorder && (
          <span className="absolute left-2.5 top-2.5 z-30 rounded-full bg-red-500 px-2 py-1 text-[10px] font-bold text-white shadow">
            -{product.discount_percent}%
          </span>
        )}

        <button
          onClick={(e) => {
            e.stopPropagation()
            toggleItem(product.id)
          }}
          className="absolute right-2.5 top-2.5 z-30 flex h-8 w-8 items-center justify-center rounded-full border border-white/10 bg-black/45 backdrop-blur-md"
        >
          <span className={`text-sm ${wishlisted ? '' : 'grayscale'}`}>❤️</span>
        </button>
      </div>

      <div className="relative z-30 flex flex-1 flex-col bg-gradient-to-b from-[#161616] to-[#111111] p-3">
        <div className="mb-2 flex flex-wrap gap-1">
          {badgeFamilies.map((family) => (
            <span
              key={family}
              className={`inline-block rounded-md px-1.5 py-0.5 text-[10px] font-bold ${familyColors[family]}`}
            >
              {familyLabels[family]}
            </span>
          ))}
        </div>

        <h3 className="mb-3 line-clamp-2 min-h-[2.5rem] flex-1 text-[15px] font-bold leading-tight tracking-[-0.01em] text-white">
          {displayTitle}
        </h3>

        <div className="mt-auto flex flex-col gap-2">
          <div className="min-h-[1.125rem]">
            <span
              className={`block leading-tight ${
                priceNeedsManager
                  ? 'text-[11px] font-medium text-white/45'
                  : 'text-sm font-black tracking-tight text-amber-200'
              }`}
            >
              {priceLabel}
            </span>
          </div>

          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation()
              openProduct()
            }}
            className={`w-full rounded-xl px-3 py-2 text-xs font-black text-white shadow-lg transition-colors ${
              preorder
                ? 'bg-sky-500 shadow-sky-500/25 hover:bg-sky-400'
                : 'bg-amber-500 shadow-amber-500/25 hover:bg-amber-400'
            }`}
          >
            Купить
          </button>
        </div>
      </div>
    </div>
  )
}

export function productToCard(product: import('../../types/product').Product): CatalogCard {
  return { id: product.id, primary: product, variants: [product] }
}
