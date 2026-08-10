import { HomeBanner } from '../types/home'
import { COD_MW4_PRODUCT_ID } from './codMw4Pricing'

/** Версия для сброса кэша в Telegram / браузере. */
export const REVIEWS_BANNER_VERSION = '4'

export const REVIEWS_BANNER_IMAGE = `/otzivi.png?v=${REVIEWS_BANNER_VERSION}`
export const REVIEWS_BANNER_LINK = 'https://t.me/shopcoinmint/11'

export const COD_MW4_BANNER_SIZE = { width: 1024, height: 512 }

/** Исходное разрешение файлов в public/ — не уменьшать при деплое. */
export const REVIEWS_BANNER_SIZE = { width: 1942, height: 809 }
export const TG_BANNER_SIZE = { width: 1939, height: 811 }

/** Локальные баннеры на главной. */
export const DEFAULT_HOME_BANNERS: HomeBanner[] = [
  {
    id: 'default-cod-mw4',
    image_url: '/banner-cod-mw4.png',
    link_url: `/product/${COD_MW4_PRODUCT_ID}`,
    width: COD_MW4_BANNER_SIZE.width,
    height: COD_MW4_BANNER_SIZE.height,
  },
  {
    id: 'default-tg',
    image_url: '/tgchanales.png',
    link_url: 'https://t.me/shopcoinmint',
    width: TG_BANNER_SIZE.width,
    height: TG_BANNER_SIZE.height,
  },
  {
    id: 'default-reviews',
    image_url: REVIEWS_BANNER_IMAGE,
    link_url: REVIEWS_BANNER_LINK,
    width: REVIEWS_BANNER_SIZE.width,
    height: REVIEWS_BANNER_SIZE.height,
  },
]

function isReviewsBanner(banner: HomeBanner): boolean {
  const link = banner.link_url?.trim() || ''
  const title = banner.title?.trim().toLowerCase() || ''
  return (
    banner.id === 'default-reviews'
    || link.includes('shopcoinmint/11')
    || title.includes('отзыв')
  )
}

function patchReviewsBanner(banner: HomeBanner): HomeBanner {
  return {
    ...banner,
    image_url: REVIEWS_BANNER_IMAGE,
    link_url: banner.link_url || REVIEWS_BANNER_LINK,
    width: REVIEWS_BANNER_SIZE.width,
    height: REVIEWS_BANNER_SIZE.height,
  }
}

function mergeDefaultBanners(banners: HomeBanner[]): HomeBanner[] {
  const seen = new Set(banners.map((b) => b.id))
  const merged = [...banners]
  for (const banner of DEFAULT_HOME_BANNERS) {
    if (!seen.has(banner.id)) {
      merged.push(banner)
      seen.add(banner.id)
    }
  }
  return merged
}

/** Админские баннеры + встроенные локальные (COD, отзывы и т.д.). */
export function resolveHomeBanners(adminBanners?: HomeBanner[]): HomeBanner[] {
  const admin = (adminBanners || []).filter((b) => b.image_url)
  if (!admin.length) return DEFAULT_HOME_BANNERS

  const patched = admin.map((banner) =>
    isReviewsBanner(banner) ? patchReviewsBanner(banner) : banner,
  )

  if (!patched.some(isReviewsBanner)) {
    patched.push({
      id: 'default-reviews',
      image_url: REVIEWS_BANNER_IMAGE,
      link_url: REVIEWS_BANNER_LINK,
      width: REVIEWS_BANNER_SIZE.width,
      height: REVIEWS_BANNER_SIZE.height,
    })
  }

  return mergeDefaultBanners(patched)
}
