export type HomeCategoryKind = 'tile' | 'feed_section'

export type FeedSectionKey = 'preorders' | 'new_releases' | 'popular'

export type HomeCategory = {
  id: string
  title: string
  image_url: string
  product_ids: string[]
  catalog_type?: string
  kind?: HomeCategoryKind
  section_key?: FeedSectionKey
  sort_order: number
}

export type ProductRow = {
  id: string
  title: string
  type: string
  platform?: string
  price?: number
  image_url?: string
}

export const FEED_SECTION_IDS: Record<FeedSectionKey, string> = {
  preorders: 'home-feed-preorders',
  new_releases: 'home-feed-new',
  popular: 'home-feed-popular',
}

export const DEFAULT_FEED_SECTIONS: Record<FeedSectionKey, HomeCategory> = {
  preorders: {
    id: FEED_SECTION_IDS.preorders,
    title: 'Предзаказы',
    image_url: '',
    product_ids: [],
    kind: 'feed_section',
    section_key: 'preorders',
    sort_order: 0,
  },
  new_releases: {
    id: FEED_SECTION_IDS.new_releases,
    title: 'Новинки',
    image_url: '',
    product_ids: [],
    kind: 'feed_section',
    section_key: 'new_releases',
    sort_order: 1,
  },
  popular: {
    id: FEED_SECTION_IDS.popular,
    title: 'Популярное',
    image_url: '',
    product_ids: [],
    kind: 'feed_section',
    section_key: 'popular',
    sort_order: 2,
  },
}

export const DEFAULT_TILES: HomeCategory[] = [
  { id: 'default-game', title: 'Игры', image_url: '/Игры.png', catalog_type: 'game', product_ids: [], kind: 'tile', sort_order: 0 },
  { id: 'default-currency', title: 'Валюта', image_url: '/Валюты.png', catalog_type: 'currency', product_ids: [], kind: 'tile', sort_order: 1 },
  { id: 'default-subscription', title: 'Подписки', image_url: '/Подписка.png', catalog_type: 'subscription', product_ids: [], kind: 'tile', sort_order: 2 },
]

export const FEED_SECTION_LABELS: Record<FeedSectionKey, string> = {
  preorders: 'Предзаказы',
  new_releases: 'Новинки',
  popular: 'Популярное',
}

export const FEED_SECTION_HINTS: Record<FeedSectionKey, string> = {
  preorders: 'Секция «Предзаказы» на главной. Пустой список = автоматически из каталога (раздел preorder).',
  new_releases: 'Секция «Новинки» на главной. Пустой список = автоматически из каталога (раздел new).',
  popular: 'Секция «Популярное» на главной. Пустой список = авто: хиты PS/Xbox, шутеры и ваши продажи. Ручной список — приоритет над авто.',
}
