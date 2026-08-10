import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button, useNotify } from 'react-admin'
import {
  Box,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  FormControlLabel,
  IconButton,
  MenuItem,
  Radio,
  RadioGroup,
  Stack,
  Tab,
  Tabs,
  TextField,
  Typography,
  Button as MuiButton,
  Alert,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import DragIndicatorIcon from '@mui/icons-material/DragIndicator'
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward'
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import { ProductPool } from './productPool'
import {
  DEFAULT_FEED_SECTIONS,
  DEFAULT_TILES,
  FEED_SECTION_HINTS,
  FEED_SECTION_LABELS,
  FeedSectionKey,
  HomeCategory,
  ProductRow,
} from './types'

const apiUrl = '/api/v1/admin'
const FEED_KEYS: FeedSectionKey[] = ['preorders', 'new_releases', 'popular']
type CategoryMode = 'products' | 'catalog'

function authHeaders() {
  const headers = new Headers({ 'Content-Type': 'application/json' })
  const token = localStorage.getItem('token')
  if (token) headers.set('Authorization', `Bearer ${token}`)
  return headers
}

function parseIds(raw: unknown): string[] {
  if (!raw) return []
  try {
    const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
    return Array.isArray(parsed) ? parsed.filter((id) => typeof id === 'string') : []
  } catch {
    return []
  }
}

function isFeedItem(c: HomeCategory): boolean {
  return c.kind === 'feed_section' || !!c.section_key
    || c.id.startsWith('home-feed-') || c.id === 'home-section-popular'
}

function feedKeyFromItem(c: HomeCategory): FeedSectionKey | null {
  if (c.section_key && FEED_KEYS.includes(c.section_key)) return c.section_key
  if (c.id === 'home-feed-preorders') return 'preorders'
  if (c.id === 'home-feed-new') return 'new_releases'
  if (c.id === 'home-feed-popular' || c.id === 'home-section-popular') return 'popular'
  return null
}

function newTile(sortOrder: number): HomeCategory {
  return {
    id: crypto.randomUUID(),
    title: 'Новая плитка',
    image_url: '',
    product_ids: [],
    catalog_type: 'game',
    kind: 'tile',
    sort_order: sortOrder,
  }
}

export function HomeCategoriesPage() {
  const notify = useNotify()
  const [feedSections, setFeedSections] = useState<Record<FeedSectionKey, HomeCategory>>(DEFAULT_FEED_SECTIONS)
  const [activeFeed, setActiveFeed] = useState<FeedSectionKey>('preorders')
  const [tiles, setTiles] = useState<HomeCategory[]>([])
  const [products, setProducts] = useState<ProductRow[]>([])
  const [productMap, setProductMap] = useState<Record<string, ProductRow>>({})
  const [feedSearch, setFeedSearch] = useState('')
  const [tileSearch, setTileSearch] = useState('')
  const [activeTileId, setActiveTileId] = useState<string | null>(null)
  const [dragId, setDragId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [productsLoading, setProductsLoading] = useState(false)
  const [productsError, setProductsError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [tileEditMode, setTileEditMode] = useState<CategoryMode>('catalog')
  const [feedModes, setFeedModes] = useState<Record<FeedSectionKey, 'auto' | 'products'>>({
    preorders: 'auto',
    new_releases: 'auto',
    popular: 'auto',
  })

  const activeFeedSection = feedSections[activeFeed]
  const feedMode = feedModes[activeFeed]
  const activeTile = tiles.find((c) => c.id === activeTileId) || null
  const tileMode = activeTile?.product_ids.length ? 'products' : tileEditMode

  const loadProducts = useCallback(async (query: string) => {
    setProductsLoading(true)
    setProductsError(null)
    try {
      const params = new URLSearchParams({ type: 'game', status: 'active', limit: '100', search: query.trim() })
      const res = await fetch(`${apiUrl}/products?${params}`, { headers: authHeaders() })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json?.error?.message || `Ошибка (${res.status})`)
      const rows: ProductRow[] = json.data || []
      setProducts(rows)
      setProductMap((prev) => {
        const next = { ...prev }
        rows.forEach((p) => { next[p.id] = p })
        return next
      })
    } catch (e: any) {
      setProducts([])
      setProductsError(e.message || 'Не удалось загрузить игры')
    } finally {
      setProductsLoading(false)
    }
  }, [])

  const loadSelectedMeta = async (ids: string[]) => {
    const missing = ids.filter((id) => !productMap[id]).slice(0, 40)
    await Promise.all(missing.map(async (id) => {
      try {
        const res = await fetch(`${apiUrl}/products/${id}`, { headers: authHeaders() })
        if (!res.ok) return
        const p = await res.json()
        if (p?.id) setProductMap((prev) => ({ ...prev, [p.id]: p }))
      } catch { /* ignore */ }
    }))
  }

  const load = async () => {
    setLoading(true)
    try {
      const [settingRes, popularRes] = await Promise.all([
        fetch(`${apiUrl}/settings?key=home_categories`, { headers: authHeaders() }),
        fetch(`${apiUrl}/settings?key=popular_product_ids`, { headers: authHeaders() }),
      ])

      let all: HomeCategory[] = []
      if (settingRes.ok) {
        const json = await settingRes.json()
        const parsed = typeof json.value === 'string' ? JSON.parse(json.value) : json.value
        if (Array.isArray(parsed)) all = parsed
      }

      const legacyPopular = popularRes.ok ? parseIds((await popularRes.json()).value) : []
      const feeds = { ...DEFAULT_FEED_SECTIONS }
      const feedItems = all.filter(isFeedItem)
      for (const key of FEED_KEYS) {
        const found = feedItems.find((c) => feedKeyFromItem(c) === key)
        if (found) feeds[key] = { ...feeds[key], ...found, kind: 'feed_section', section_key: key }
      }
      if (!feeds.popular.product_ids.length && legacyPopular.length) {
        feeds.popular = { ...feeds.popular, product_ids: legacyPopular }
      }

      let tileList = all.filter((c) => !isFeedItem(c))
      if (!tileList.length) tileList = DEFAULT_TILES.map((c) => ({ ...c }))

      setFeedSections(feeds)
      setFeedModes({
        preorders: feeds.preorders.product_ids.length ? 'products' : 'auto',
        new_releases: feeds.new_releases.product_ids.length ? 'products' : 'auto',
        popular: feeds.popular.product_ids.length ? 'products' : 'auto',
      })
      setTiles([...tileList].sort((a, b) => a.sort_order - b.sort_order))
      setActiveTileId(tileList[0]?.id || null)

      const allIds = [...feeds.preorders.product_ids, ...feeds.new_releases.product_ids, ...feeds.popular.product_ids]
      await loadSelectedMeta(allIds)
    } catch (e: any) {
      notify(e.message || 'Ошибка загрузки', { type: 'error' })
      setFeedSections(DEFAULT_FEED_SECTIONS)
      setTiles(DEFAULT_TILES.map((c) => ({ ...c })))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load().then(() => loadProducts('')).catch(() => setLoading(false))
  }, [])

  useEffect(() => {
    setFeedSearch('')
  }, [activeFeed])

  useEffect(() => {
    const timer = setTimeout(() => loadProducts(feedSearch), feedSearch ? 350 : 0)
    return () => clearTimeout(timer)
  }, [feedSearch, activeFeed, loadProducts])

  useEffect(() => {
    if (!activeTile || tileMode !== 'products') return
    const timer = setTimeout(() => loadProducts(tileSearch), tileSearch ? 350 : 0)
    return () => clearTimeout(timer)
  }, [activeTile?.id, tileMode, tileSearch, loadProducts])

  const updateFeed = (key: FeedSectionKey, patch: Partial<HomeCategory>) => {
    setFeedSections((prev) => ({ ...prev, [key]: { ...prev[key], ...patch } }))
  }

  const setFeedMode = (key: FeedSectionKey, mode: 'auto' | 'products') => {
    setFeedModes((prev) => ({ ...prev, [key]: mode }))
    if (mode === 'auto') updateFeed(key, { product_ids: [] })
  }

  const updateTile = (id: string, patch: Partial<HomeCategory>) => {
    setTiles((prev) => prev.map((c) => (c.id === id ? { ...c, ...patch } : c)))
  }

  const mergeProduct = (productId: string, p?: ProductRow) => {
    if (p) setProductMap((prev) => ({ ...prev, [productId]: p }))
  }

  const save = async () => {
    setSaving(true)
    try {
      const feedPayload = FEED_KEYS.map((key, index) => ({
        ...feedSections[key],
        kind: 'feed_section' as const,
        section_key: key,
        sort_order: index,
        product_ids: feedSections[key].product_ids || [],
      }))
      const tilePayload = tiles.map((c, index) => ({
        ...c,
        kind: 'tile' as const,
        sort_order: index,
        catalog_type: c.product_ids.length ? undefined : (c.catalog_type || 'game'),
        product_ids: c.product_ids || [],
      }))
      const payload = [...feedPayload, ...tilePayload]

      const [catRes, popRes] = await Promise.all([
        fetch(`${apiUrl}/settings`, {
          method: 'PUT',
          headers: authHeaders(),
          body: JSON.stringify({ key: 'home_categories', value: payload }),
        }),
        fetch(`${apiUrl}/settings`, {
          method: 'PUT',
          headers: authHeaders(),
          body: JSON.stringify({ key: 'popular_product_ids', value: feedSections.popular.product_ids }),
        }),
      ])
      if (!catRes.ok || !popRes.ok) throw new Error('Не удалось сохранить')
      notify('Главная сохранена', { type: 'success' })
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const feedCounts = useMemo(() => ({
    preorders: feedSections.preorders.product_ids.length,
    new_releases: feedSections.new_releases.product_ids.length,
    popular: feedSections.popular.product_ids.length,
  }), [feedSections])

  const catalogLabel = useMemo(() => ({ game: 'Игры', currency: 'Валюта', subscription: 'Подписки' } as Record<string, string>), [])

  return (
    <Box sx={{ p: 2, maxWidth: 1200 }}>
      <Typography variant="h5" sx={{ mb: 0.5 }}>Главная: секции и плитки</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        Управляйте блоками <b>Предзаказы</b>, <b>Новинки</b> и <b>Популярное</b> на главной. Игры — из импортированного каталога.
      </Typography>

      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Stack direction="row" spacing={1} alignItems="center" sx={{ flexWrap: 'wrap' }}>
            <Button label="Сохранить" onClick={save} disabled={saving} startIcon={<SaveIcon />} />
            <Chip label={`Предзаказы: ${feedCounts.preorders || 'авто'}`} size="small" />
            <Chip label={`Новинки: ${feedCounts.new_releases || 'авто'}`} size="small" />
            <Chip label={`Популярное: ${feedCounts.popular || 'авто'}`} size="small" color="primary" variant="outlined" />
          </Stack>
        </CardContent>
      </Card>

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 6 }}><CircularProgress /></Box>
      ) : (
        <Stack spacing={3}>
          <Card>
            <Tabs value={activeFeed} onChange={(_, v) => setActiveFeed(v)} variant="fullWidth">
              {FEED_KEYS.map((key) => (
                <Tab
                  key={key}
                  value={key}
                  label={`${FEED_SECTION_LABELS[key]}${feedSections[key].product_ids.length ? ` (${feedSections[key].product_ids.length})` : ''}`}
                />
              ))}
            </Tabs>
            <CardContent>
              <Typography fontWeight={700} sx={{ mb: 1 }}>{FEED_SECTION_LABELS[activeFeed]}</Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                {FEED_SECTION_HINTS[activeFeed]}
              </Typography>

              <RadioGroup
                row
                value={feedMode === 'products' ? 'products' : 'auto'}
                onChange={(e) => setFeedMode(activeFeed, e.target.value as 'auto' | 'products')}
                sx={{ mb: 2 }}
              >
                <FormControlLabel value="auto" control={<Radio />} label="Автоматически из каталога" />
                <FormControlLabel value="products" control={<Radio />} label="Свой список игр" />
              </RadioGroup>

              {feedMode === 'products' ? (
                <ProductPool
                  title={`Игры в «${FEED_SECTION_LABELS[activeFeed]}»`}
                  selectedIds={activeFeedSection.product_ids}
                  productMap={productMap}
                  products={products}
                  search={feedSearch}
                  onSearchChange={setFeedSearch}
                  onSearch={() => loadProducts(feedSearch)}
                  loading={productsLoading}
                  error={productsError}
                  onAdd={(id) => {
                    if (activeFeedSection.product_ids.includes(id)) return
                    updateFeed(activeFeed, { product_ids: [...activeFeedSection.product_ids, id] })
                    mergeProduct(id, productMap[id] || products.find((p) => p.id === id))
                  }}
                  onRemove={(id) => updateFeed(activeFeed, { product_ids: activeFeedSection.product_ids.filter((x) => x !== id) })}
                />
              ) : (
                <Alert severity="info">
                  Сейчас игры подбираются автоматически. Переключите на «Свой список», чтобы выбрать вручную.
                </Alert>
              )}
            </CardContent>
          </Card>

          <Accordion>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Typography fontWeight={600}>Плитки категорий (Игры / Валюта / Подписки)</Typography>
            </AccordionSummary>
            <AccordionDetails>
              <Stack spacing={2}>
                <Button label="Добавить плитку" onClick={() => {
                  const t = newTile(tiles.length)
                  setTiles((prev) => [...prev, t])
                  setActiveTileId(t.id)
                }} startIcon={<AddIcon />} />

                <Box sx={{ display: 'grid', gridTemplateColumns: { xs: '1fr', md: '280px 1fr' }, gap: 2 }}>
                  <Stack spacing={1}>
                    {tiles.map((cat, index) => (
                      <Card
                        key={cat.id}
                        draggable
                        onDragStart={() => setDragId(cat.id)}
                        onDragOver={(e) => e.preventDefault()}
                        onDrop={() => {
                          if (!dragId || dragId === cat.id) return
                          setTiles((prev) => {
                            const from = prev.findIndex((c) => c.id === dragId)
                            const to = prev.findIndex((c) => c.id === cat.id)
                            if (from < 0 || to < 0) return prev
                            const next = [...prev]
                            const [item] = next.splice(from, 1)
                            next.splice(to, 0, item)
                            return next.map((c, i) => ({ ...c, sort_order: i }))
                          })
                          setDragId(null)
                        }}
                        variant={cat.id === activeTileId ? 'elevation' : 'outlined'}
                        sx={{ cursor: 'grab', borderColor: cat.id === activeTileId ? 'primary.main' : undefined }}
                        onClick={() => setActiveTileId(cat.id)}
                      >
                        <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 1, py: '8px !important' }}>
                          <DragIndicatorIcon fontSize="small" />
                          <Chip size="small" label={index + 1} />
                          <Box sx={{ flex: 1, minWidth: 0 }}>
                            <Typography fontWeight={700} noWrap fontSize={13}>{cat.title}</Typography>
                            <Typography variant="caption" color="text.secondary" noWrap>
                              {cat.product_ids.length ? `${cat.product_ids.length} игр` : `→ ${catalogLabel[cat.catalog_type || 'game']}`}
                            </Typography>
                          </Box>
                          <IconButton size="small" onClick={(e) => { e.stopPropagation(); setTiles((p) => { const i = p.findIndex((x) => x.id === cat.id); if (i <= 0) return p; const n = [...p]; [n[i - 1], n[i]] = [n[i], n[i - 1]]; return n }) }}><ArrowUpwardIcon fontSize="small" /></IconButton>
                          <IconButton size="small" onClick={(e) => { e.stopPropagation(); setTiles((p) => { const i = p.findIndex((x) => x.id === cat.id); if (i < 0 || i >= p.length - 1) return p; const n = [...p]; [n[i], n[i + 1]] = [n[i + 1], n[i]]; return n }) }}><ArrowDownwardIcon fontSize="small" /></IconButton>
                          <IconButton size="small" color="error" onClick={(e) => { e.stopPropagation(); setTiles((p) => p.filter((x) => x.id !== cat.id)); if (activeTileId === cat.id) setActiveTileId(tiles[0]?.id || null) }}><DeleteIcon fontSize="small" /></IconButton>
                        </CardContent>
                      </Card>
                    ))}
                  </Stack>

                  {activeTile ? (
                    <Card variant="outlined">
                      <CardContent>
                        <Stack spacing={2}>
                          <TextField label="Название" value={activeTile.title} onChange={(e) => updateTile(activeTile.id, { title: e.target.value })} fullWidth />
                          <TextField label="URL картинки" value={activeTile.image_url} onChange={(e) => updateTile(activeTile.id, { image_url: e.target.value })} fullWidth />
                          <Divider />
                          <RadioGroup value={tileMode} onChange={(e) => {
                            const m = e.target.value as CategoryMode
                            setTileEditMode(m)
                            if (m === 'catalog') updateTile(activeTile.id, { product_ids: [], catalog_type: activeTile.catalog_type || 'game' })
                          }}>
                            <FormControlLabel value="catalog" control={<Radio />} label="Ссылка на каталог" />
                            <FormControlLabel value="products" control={<Radio />} label="Свой список" />
                          </RadioGroup>
                          {tileMode === 'catalog' ? (
                            <TextField select label="Раздел" value={activeTile.catalog_type || 'game'} onChange={(e) => updateTile(activeTile.id, { catalog_type: e.target.value, product_ids: [] })} fullWidth>
                              <MenuItem value="game">Игры</MenuItem>
                              <MenuItem value="currency">Валюта</MenuItem>
                              <MenuItem value="subscription">Подписки</MenuItem>
                            </TextField>
                          ) : (
                            <ProductPool
                              title="Игры в плитке"
                              selectedIds={activeTile.product_ids}
                              productMap={productMap}
                              products={products}
                              search={tileSearch}
                              onSearchChange={setTileSearch}
                              onSearch={() => loadProducts(tileSearch)}
                              loading={productsLoading}
                              error={productsError}
                              onAdd={(id) => {
                                if (activeTile.product_ids.includes(id)) return
                                updateTile(activeTile.id, { product_ids: [...activeTile.product_ids, id] })
                                mergeProduct(id, productMap[id] || products.find((p) => p.id === id))
                              }}
                              onRemove={(id) => updateTile(activeTile.id, { product_ids: activeTile.product_ids.filter((x) => x !== id) })}
                            />
                          )}
                        </Stack>
                      </CardContent>
                    </Card>
                  ) : null}
                </Box>
              </Stack>
            </AccordionDetails>
          </Accordion>
        </Stack>
      )}
    </Box>
  )
}
