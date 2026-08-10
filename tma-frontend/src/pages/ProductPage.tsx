import { useEffect, useMemo, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useProduct } from '../hooks/useProducts'
import { useCreateOrder } from '../hooks/useOrders'
import { useToast } from '../components/ui/Toast'
import { useCart } from '../stores/cartStore'
import { useGoBack } from '../hooks/useGoBack'
import { Header } from '../components/layout/Header'
import { Button, Loader } from '../components/ui/Button'
import { EditionPricingPanel } from '../components/product/EditionPricingPanel'
import { formatPrice, formatPriceOrManager } from '../utils/format'
import { openManagerOrderChat } from '../utils/managerChat'
import { formatReleaseDate, isPreorderProduct } from '../utils/productPreorder'
import {
  EditionPlatformKey,
  defaultEditionPlatform,
  editionRegionLabel,
  findEdition,
  getEditionCatalog,
  listEditionPlatforms,
  resolveEditionCatalogForProduct,
  shouldShowFullEditionCatalog,
} from '../utils/productEditionPricing'
import { getProductDisplayTitle, getUnifiedProductTitle } from '../utils/productDisplayTitle'
import { platformLabels, typeLabels, platformColors, ProductVariant } from '../types/product'

const CHAT_DELIVERY = 'activation' as const

function isPlayStationPlatform(platform: string) {
  return platform === 'ps4' || platform === 'ps5'
}

function getXboxUsPrice(prices?: { xbox?: number; us?: number }) {
  if (!prices) return null
  return prices.xbox ?? prices.us ?? null
}

export function ProductPage() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
  const goBack = useGoBack('/catalog')
  const toast = useToast()
  const { data: product, isLoading } = useProduct(id!)
  const createOrder = useCreateOrder()
  const { addItem, getItemCount } = useCart()
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [imageRatio, setImageRatio] = useState<number | null>(null)
  const [selectedRegion, setSelectedRegion] = useState<'tr' | 'ua'>('tr')
  const [editionPlatform, setEditionPlatform] = useState<EditionPlatformKey>('ps_tr')
  const [selectedEditionId, setSelectedEditionId] = useState<string | null>(null)

  const rawEditionCatalog = product
    ? getEditionCatalog(product.id, product.prices, product.title_key, product.platform)
    : null
  const editionCatalog = rawEditionCatalog && product
    ? resolveEditionCatalogForProduct(rawEditionCatalog, product)
    : null
  const editionPlatforms = editionCatalog ? listEditionPlatforms(editionCatalog) : []
  const isMultiPlatformCatalog = product
    ? shouldShowFullEditionCatalog(product, rawEditionCatalog)
    : false

  useEffect(() => {
    if (!editionCatalog || !editionPlatforms.length) return
    const preferred = defaultEditionPlatform(product!, editionPlatforms)
    if (preferred && !editionPlatforms.includes(editionPlatform)) {
      setEditionPlatform(preferred)
    }
    const activePlatform = editionPlatforms.includes(editionPlatform)
      ? editionPlatform
      : preferred ?? editionPlatforms[0]
    const editions = editionCatalog[activePlatform] ?? []
    if (!selectedEditionId && editions[0]) {
      setSelectedEditionId(editions[0].id)
    }
  }, [editionCatalog, editionPlatform, editionPlatforms, product, selectedEditionId])

  const selectedEdition = useMemo(() => {
    if (!editionCatalog || !selectedEditionId) return null
    return findEdition(editionCatalog, editionPlatform, selectedEditionId)
  }, [editionCatalog, editionPlatform, selectedEditionId])

  const preorder = product ? isPreorderProduct(product) : false
  const hasDiscount = product ? !preorder && product.discount_percent > 0 : false
  const basePrice = product ? (hasDiscount ? product.price * (1 - product.discount_percent / 100) : product.price) : 0
  const variantPrice = selectedVariant ? selectedVariant.price : 0
  const gamePrice = selectedEdition?.price ?? basePrice
  const totalPrice = editionCatalog && selectedEdition
    ? selectedEdition.price
    : product?.type === 'currency' || product?.type === 'subscription'
      ? (variantPrice || basePrice) * quantity
      : gamePrice
  const variants: ProductVariant[] = product && Array.isArray(product.variants) ? product.variants : []
  const hasVariants = !!product && variants.length > 0 && (product.type === 'currency' || product.type === 'subscription')
  const isPlayStation = product ? isPlayStationPlatform(product.platform) : false
  const isXbox = product?.platform === 'xbox'
  const xboxUsPrice = product ? getXboxUsPrice(product.prices) : null

  useEffect(() => {
    setSelectedVariant(null)
    setQuantity(1)
    setSelectedEditionId(null)
    setEditionPlatform('ps_tr')
    setImageRatio(null)
  }, [id])

  const handleAddToCart = () => {
    if (!product) return
    if (editionCatalog && !selectedEdition) {
      toast.toast('Выберите издание', 'error')
      return
    }
    addItem({
      productId: product.id,
      variantId: selectedVariant?.id,
      quantity: (product.type === 'currency' || product.type === 'subscription') ? quantity : 1,
      title: selectedEdition
        ? `${getProductDisplayTitle(product)} — ${selectedEdition.name} (${editionRegionLabel(editionPlatform)})`
        : getProductDisplayTitle(product),
      price: selectedEdition?.price ?? (selectedVariant ? selectedVariant.price : basePrice),
      image: product.image_url || undefined,
      deliveryMethod: CHAT_DELIVERY,
    })
    toast.toast(`Добавлено в корзину (${getItemCount()})`, 'success')
  }

  const handleBuy = async () => {
    if (!product) return
    if (hasVariants && !selectedVariant) {
      toast.toast('Выберите вариант товара', 'error')
      return
    }
    if (editionCatalog && !selectedEdition) {
      toast.toast('Выберите издание', 'error')
      return
    }

    const qty = (product.type === 'currency' || product.type === 'subscription') ? quantity : 1
    const variantId = selectedVariant?.id
      ?? (selectedEdition ? `${editionPlatform}:${selectedEdition.id}` : undefined)

    try {
      const order = await createOrder.mutateAsync({
        productId: product.id,
        variantId,
        quantity: qty,
      })
      nav(`/orders/${order.id}`)
      openManagerOrderChat(order.id)
    } catch {
      toast.toast('Не удалось создать заказ. Попробуйте ещё раз.', 'error')
    }
  }

  if (isLoading) return <div className="flex justify-center py-20"><Loader /></div>
  if (!product) return <div className="p-4 text-center text-[var(--tg-hint)]">Товар не найден</div>

  const displayTitle = product ? getUnifiedProductTitle(product, isMultiPlatformCatalog) : ''

  const mediaAspectRatio = imageRatio ? `${imageRatio}` : '0.75'

  return (
    <div className="pb-page-bar">
      <Header title="" onBack={goBack} />
      <div className="space-y-4 p-4">
        <div className="relative overflow-hidden rounded-3xl border border-white/10 bg-[#0c0c0c] shadow-[0_18px_45px_rgba(0,0,0,0.35)]" style={{ aspectRatio: mediaAspectRatio }}>
          {product.image_url ? (
            <>
              <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_18%,rgba(201,168,76,0.32),transparent_42%),linear-gradient(180deg,rgba(255,255,255,0.05),rgba(0,0,0,0.18))]" />
              <img
                src={product.image_url}
                alt={displayTitle}
                onLoad={(event) => {
                  const img = event.currentTarget
                  const ratio = img.naturalWidth / Math.max(img.naturalHeight, 1)
                  setImageRatio(Math.min(Math.max(ratio, 0.58), 1.85))
                }}
                className="relative z-10 h-full w-full object-contain"
              />
            </>
          ) : (
            <div className="flex h-full items-center justify-center text-6xl">
              {product.type === 'game' ? '🎮' : product.type === 'currency' ? '💰' : '📦'}
            </div>
          )}
          {hasDiscount && !editionCatalog ? (
            <span className="absolute left-3 top-3 z-20 rounded-full bg-red-500 px-3 py-1 text-sm font-black text-white shadow-lg">
              -{product.discount_percent}%
            </span>
          ) : null}
        </div>

        <div className="rounded-3xl border border-white/10 bg-[#161616] p-4 shadow-[0_12px_30px_rgba(0,0,0,0.24)]">
          <div className="mb-3">
            <span className="rounded-md bg-white/7 px-2 py-1 text-xs font-medium text-white/55">
              {typeLabels[product.type]}
            </span>
          </div>

          <h1 className="mb-2 break-words text-xl font-black leading-tight tracking-[-0.02em] text-white sm:text-2xl">
            {displayTitle}
          </h1>

          {preorder ? (
            <div className="mb-4 space-y-1">
              <div className="text-xs font-bold uppercase tracking-[0.14em] text-sky-300">Предзаказ</div>
              <div className="text-sm font-medium text-sky-100/90">{formatReleaseDate(product.release_date)}</div>
            </div>
          ) : null}

          {editionCatalog && editionPlatforms.length > 0 ? (
            <>
              {!isMultiPlatformCatalog ? (
                <div className="mb-3 flex items-center gap-2">
                  <span className={`rounded-md px-2 py-1 text-xs font-bold ${platformColors[product.platform]}`}>
                    {platformLabels[product.platform]}
                  </span>
                </div>
              ) : null}
              <EditionPricingPanel
              catalog={editionCatalog}
              platform={editionPlatform}
              editionId={selectedEditionId}
              onPlatformChange={(nextPlatform) => {
                setEditionPlatform(nextPlatform)
                const first = editionCatalog[nextPlatform]?.[0]
                setSelectedEditionId(first?.id ?? null)
              }}
              onEditionChange={(edition) => setSelectedEditionId(edition.id)}
            />
            </>
          ) : (
            <>
              {!editionCatalog ? (
                <div className="mb-3 flex items-center gap-2">
                  <span className={`rounded-md px-2 py-1 text-xs font-bold ${platformColors[product.platform]}`}>
                    {platformLabels[product.platform]}
                  </span>
                </div>
              ) : null}

              <div className="flex items-end justify-between gap-3">
                <div className="min-w-0 flex-1">
                  {product.prices && Object.keys(product.prices).length > 0 ? (
                    <div className="space-y-2">
                      {isPlayStation ? (
                        <>
                          <div className="flex items-center gap-1.5">
                            <button
                              onClick={() => setSelectedRegion('tr')}
                              className={`rounded-lg px-2.5 py-1 text-xs font-bold transition-colors ${
                                selectedRegion === 'tr'
                                  ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                                  : 'bg-white/5 text-white/40 border border-white/10 hover:bg-white/10'
                              }`}
                            >
                              🇹🇷 TR
                            </button>
                            <button
                              onClick={() => setSelectedRegion('ua')}
                              className={`rounded-lg px-2.5 py-1 text-xs font-bold transition-colors ${
                                selectedRegion === 'ua'
                                  ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                                  : 'bg-white/5 text-white/40 border border-white/10 hover:bg-white/10'
                              }`}
                            >
                              🇺🇦 UA
                            </button>
                          </div>
                          {selectedRegion === 'tr' && product.prices.tr != null && (
                            <div className="flex items-baseline gap-2">
                              <span className="text-xs font-medium text-white/45">Турция</span>
                              <span className="text-2xl font-black leading-none text-amber-200">
                                {formatPriceOrManager(product.prices.tr)}
                              </span>
                            </div>
                          )}
                          {selectedRegion === 'ua' && product.prices.ua != null && (
                            <div className="flex items-baseline gap-2">
                              <span className="text-xs font-medium text-white/45">Украина</span>
                              <span className="text-2xl font-black leading-none text-amber-200">
                                {formatPriceOrManager(product.prices.ua)}
                              </span>
                            </div>
                          )}
                          {selectedRegion === 'ua' && product.prices.ua == null && (
                            <div className="flex items-center gap-2 rounded-xl border border-amber-500/20 bg-amber-500/10 px-3 py-2">
                              <span className="text-sm font-bold text-amber-300">🇺🇦 UA</span>
                              <span className="text-xs text-amber-200/70">Цену на этот регион можно узнать у менеджера</span>
                            </div>
                          )}
                        </>
                      ) : isXbox && xboxUsPrice != null ? (
                        <div className="flex items-baseline gap-2">
                          <span className="text-xs font-medium text-white/45">🇺🇸 US</span>
                          <span className="text-2xl font-black leading-none text-amber-200">
                            {formatPriceOrManager(xboxUsPrice)}
                          </span>
                        </div>
                      ) : (
                        <div className="text-2xl font-black leading-none text-amber-200">
                          {formatPriceOrManager(basePrice)}
                        </div>
                      )}
                    </div>
                  ) : (
                    <div className="text-2xl font-black leading-none text-amber-200">
                      {formatPriceOrManager(basePrice)}
                    </div>
                  )}
                </div>
                <div className="rounded-2xl bg-white/7 px-3 py-2 text-right text-xs font-medium text-white/55">
                  Через менеджера
                </div>
              </div>
            </>
          )}
        </div>

        {product.description && (
          <div className="rounded-3xl border border-white/10 bg-[#161616]/80 p-4">
            <div className="mb-2 text-xs font-bold uppercase tracking-[0.14em] text-white/35">Описание</div>
            <p className="text-sm leading-relaxed text-white/62">{product.description}</p>
          </div>
        )}

        {hasVariants && (
          <div className="rounded-3xl border border-white/10 bg-[#161616]/80 p-4">
            <label className="mb-3 block text-sm font-bold text-white">Выберите вариант</label>
            <div className="grid grid-cols-2 gap-2">
              {variants.map((v) => (
                <button
                  key={v.id}
                  onClick={() => { setSelectedVariant(v); setQuantity(1) }}
                  className={`rounded-2xl border p-3 text-left transition ${
                    selectedVariant?.id === v.id
                      ? 'border-amber-400 bg-amber-500/14'
                      : 'border-white/10 bg-black/12 hover:border-white/20'
                  }`}
                >
                  <div className="font-bold text-sm text-white">{v.name}</div>
                  <div className="mt-1 text-xs text-white/45">{formatPriceOrManager(v.price)}</div>
                </button>
              ))}
            </div>
          </div>
        )}

        {(product.type === 'currency' || product.type === 'subscription') && (selectedVariant || !hasVariants) && (
          <div className="rounded-3xl border border-white/10 bg-[#161616]/80 p-4">
            <label className="mb-3 block text-sm font-bold text-white">Количество</label>
            <div className="flex items-center gap-3">
              <button
                onClick={() => setQuantity(Math.max(1, quantity - 1))}
                className="flex h-11 w-11 items-center justify-center rounded-2xl bg-white/8 text-xl font-bold hover:bg-white/12"
              >
                −
              </button>
              <span className="w-12 text-center text-xl font-black text-white">{quantity}</span>
              <button
                onClick={() => setQuantity(quantity + 1)}
                className="flex h-11 w-11 items-center justify-center rounded-2xl bg-white/8 text-xl font-bold hover:bg-white/12"
              >
                +
              </button>
            </div>
          </div>
        )}

        {(product.type === 'currency' || product.type === 'subscription') && (
          <div className="rounded-3xl border border-white/10 bg-[#161616]/80 p-4">
            <div className="mb-1 flex justify-between text-sm">
              <span className="text-white/45">Цена за единицу</span>
              <span>{formatPriceOrManager(selectedVariant ? selectedVariant.price : basePrice)}</span>
            </div>
            <div className="mb-2 flex justify-between text-sm">
              <span className="text-white/45">Количество</span>
              <span>× {quantity}</span>
            </div>
            <div className="flex justify-between border-t border-white/10 pt-2 text-lg font-black">
              <span>Итого</span>
              <span className="text-amber-200">{formatPriceOrManager(totalPrice)}</span>
            </div>
          </div>
        )}

        <div className="rounded-3xl border border-amber-500/20 bg-amber-500/10 p-4 text-sm text-white/70">
          Оформление и выдача — только через чат с менеджером в Telegram.
        </div>

        <div className="sticky bottom-above-nav z-40 -mx-1 flex flex-col gap-2 rounded-3xl border border-white/10 bg-[#111111]/92 p-2 shadow-[0_-12px_35px_rgba(0,0,0,0.32)] backdrop-blur-xl sm:flex-row">
          <Button size="md" onClick={handleAddToCart} className="w-full min-w-0 flex-1 text-sm sm:text-base" variant="secondary">
            В корзину
          </Button>
          <Button
            size="md"
            onClick={handleBuy}
            loading={createOrder.isPending}
            className="w-full min-w-0 flex-1 text-sm sm:text-base"
          >
            {preorder ? 'ПРЕДЗАКАЗ' : 'Оформить с менеджером'}
          </Button>
        </div>
      </div>
    </div>
  )
}
