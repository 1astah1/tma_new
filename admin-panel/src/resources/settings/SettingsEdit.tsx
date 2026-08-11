import { useState, useEffect } from 'react'
import { Card, CardContent, Typography, TextField, Button, Grid, Snackbar, Tabs, Tab, Box, Alert } from '@mui/material'

type PricingFormulas = {
  try_rub_manual: number
  turkey_markup: number
  ukraine_markup: number
  xbox_usd_multiplier: number
  min_price_rub: number
}

const defaultPricing: PricingFormulas = {
  try_rub_manual: 0,
  turkey_markup: 2.2,
  ukraine_markup: 2.3,
  xbox_usd_multiplier: 80,
  min_price_rub: 149,
}

async function loadSetting(key: string) {
  const res = await fetch(`/api/v1/admin/settings?key=${key}`, {
    headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
  })
  if (!res.ok) return null
  const d = await res.json()
  return d.value ?? ''
}

async function saveSetting(key: string, value: unknown) {
  await fetch('/api/v1/admin/settings', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${localStorage.getItem('token')}`,
    },
    body: JSON.stringify({ key, value }),
  })
}

function parsePricing(raw: unknown): PricingFormulas {
  if (!raw) return { ...defaultPricing }
  try {
    const data = typeof raw === 'string' ? JSON.parse(raw) : raw
    return {
      try_rub_manual: Number(data.try_rub_manual) || 0,
      turkey_markup: Number(data.turkey_markup) || defaultPricing.turkey_markup,
      ukraine_markup: Number(data.ukraine_markup) || defaultPricing.ukraine_markup,
      xbox_usd_multiplier: Number(data.xbox_usd_multiplier) || defaultPricing.xbox_usd_multiplier,
      min_price_rub: Number(data.min_price_rub) || defaultPricing.min_price_rub,
    }
  } catch {
    return { ...defaultPricing }
  }
}

export const SettingsEdit = () => {
  const [tab, setTab] = useState(0)
  const [payment, setPayment] = useState<any>({ sbp: {}, card: {}, crypto: {} })
  const [shop, setShop] = useState({
    support_url: 'https://t.me/coin_mint_chat',
    reviews_url: 'https://t.me/coin_mint_reviews',
    manager_url: 'https://t.me/KromkaJQ',
    shop_rules: '',
    bot_username: '',
    ps_store_proxy: '',
  })
  const [pricing, setPricing] = useState<PricingFormulas>({ ...defaultPricing })
  const [preview, setPreview] = useState<{ try_rub_rate?: number; examples?: Record<string, number> }>({})
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [snack, setSnack] = useState('')

  const loadPreview = () => {
    fetch('/api/v1/admin/settings/pricing-preview', {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
      .then((r) => (r.ok ? r.json() : {}))
      .then(setPreview)
      .catch(() => {})
  }

  useEffect(() => {
    Promise.all([
      loadSetting('payment_details'),
      loadSetting('support_url'),
      loadSetting('reviews_url'),
      loadSetting('manager_url'),
      loadSetting('shop_rules'),
      loadSetting('bot_username'),
      loadSetting('ps_store_proxy'),
      loadSetting('pricing_formulas'),
    ])
      .then(([paymentRaw, supportUrl, reviewsUrl, managerUrl, shopRules, botUsername, psProxy, pricingRaw]) => {
        try {
          setPayment(typeof paymentRaw === 'string' ? JSON.parse(paymentRaw) : paymentRaw || { sbp: {}, card: {}, crypto: {} })
        } catch {
          setPayment({ sbp: {}, card: {}, crypto: {} })
        }
        setShop({
          support_url: String(supportUrl || 'https://t.me/coin_mint_chat'),
          reviews_url: String(reviewsUrl || 'https://t.me/coin_mint_reviews'),
          manager_url: String(managerUrl || 'https://t.me/KromkaJQ'),
          shop_rules: String(shopRules || ''),
          bot_username: String(botUsername || ''),
          ps_store_proxy: String(psProxy || ''),
        })
        setPricing(parsePricing(pricingRaw))
      })
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (tab === 1) loadPreview()
  }, [tab, pricing])

  const saveShop = async () => {
    setSaving(true)
    await saveSetting('support_url', shop.support_url)
    await saveSetting('reviews_url', shop.reviews_url)
    await saveSetting('manager_url', shop.manager_url)
    await saveSetting('shop_rules', shop.shop_rules)
    await saveSetting('bot_username', shop.bot_username)
    await saveSetting('ps_store_proxy', shop.ps_store_proxy)
    setSaving(false)
    setSnack('Настройки магазина сохранены')
  }

  const savePricing = async () => {
    setSaving(true)
    await saveSetting('pricing_formulas', pricing)
    setSaving(false)
    setSnack('Формулы цен сохранены. Запустите «Обновить цены» в импорте каталога.')
    loadPreview()
  }

  const savePayment = async () => {
    setSaving(true)
    await saveSetting('payment_details', payment)
    setSaving(false)
    setSnack('Реквизиты сохранены')
  }

  const setPaymentField = (section: string, field: string, value: string) => {
    setPayment((prev: any) => ({ ...prev, [section]: { ...prev[section], [field]: value } }))
  }

  const setPricingNum = (field: keyof PricingFormulas, value: string) => {
    const num = value === '' ? 0 : Number(value)
    setPricing((prev) => ({ ...prev, [field]: Number.isFinite(num) ? num : prev[field] }))
  }

  if (loading) return <div>Загрузка...</div>

  return (
    <div style={{ padding: 20, maxWidth: 900 }}>
      <Typography variant="h4" gutterBottom>Настройки</Typography>

      <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 3 }}>
        <Tab label="Магазин" />
        <Tab label="Формулы цен" />
        <Tab label="Оплата" />
      </Tabs>

      {tab === 0 && (
        <>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Ссылки и контент</Typography>
              <Grid container spacing={2}>
                <Grid item xs={12}>
                  <TextField fullWidth label="Поддержка (Telegram)" value={shop.support_url} onChange={(e) => setShop({ ...shop, support_url: e.target.value })} />
                </Grid>
                <Grid item xs={12}>
                  <TextField fullWidth label="Отзывы" value={shop.reviews_url} onChange={(e) => setShop({ ...shop, reviews_url: e.target.value })} />
                </Grid>
                <Grid item xs={12} md={6}>
                  <TextField
                    fullWidth
                    label="Личка менеджера (заявки на покупку)"
                    helperText="Сюда уходит заявка с готовым текстом, когда покупатель жмёт «Купить»"
                    value={shop.manager_url}
                    onChange={(e) => setShop({ ...shop, manager_url: e.target.value })}
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField fullWidth label="Username бота (без @)" value={shop.bot_username} onChange={(e) => setShop({ ...shop, bot_username: e.target.value })} />
                </Grid>
                <Grid item xs={12}>
                  <TextField fullWidth multiline minRows={6} label="Правила магазина" value={shop.shop_rules} onChange={(e) => setShop({ ...shop, shop_rules: e.target.value })} />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    label="Прокси PS Store"
                    placeholder="http://127.0.0.1:8080"
                    value={shop.ps_store_proxy}
                    onChange={(e) => setShop({ ...shop, ps_store_proxy: e.target.value })}
                    helperText="HTTP_PROXY/HTTPS_PROXY в env имеют приоритет"
                  />
                </Grid>
              </Grid>
            </CardContent>
          </Card>
          <Button variant="contained" onClick={saveShop} disabled={saving}>{saving ? 'Сохранение...' : 'Сохранить'}</Button>
        </>
      )}

      {tab === 1 && (
        <>
          <Alert severity="info" sx={{ mb: 2 }}>
            После изменения формул откройте «Импорт игр» → «Обновить цены», чтобы пересчитать каталог.
          </Alert>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Коэффициенты</Typography>
              <Grid container spacing={2}>
                <Grid item xs={12} sm={6}>
                  <TextField
                    fullWidth
                    type="number"
                    label="TRY → RUB (ручной курс)"
                    value={pricing.try_rub_manual || ''}
                    onChange={(e) => setPricingNum('try_rub_manual', e.target.value)}
                    helperText="0 = курс ЦБ РФ автоматически"
                  />
                </Grid>
                <Grid item xs={12} sm={6}>
                  <TextField
                    fullWidth
                    type="number"
                    label="Минимальная цена, ₽"
                    value={pricing.min_price_rub}
                    onChange={(e) => setPricingNum('min_price_rub', e.target.value)}
                    inputProps={{ min: 1, step: 1 }}
                  />
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="subtitle2" color="text.secondary" sx={{ mt: 1 }}>PlayStation Турция</Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    Номинал TRY (250, 500, 750…) × множитель
                  </Typography>
                  <TextField
                    fullWidth
                    type="number"
                    label="Множитель Турция"
                    value={pricing.turkey_markup}
                    onChange={(e) => setPricingNum('turkey_markup', e.target.value)}
                    inputProps={{ min: 0.1, step: 0.1 }}
                  />
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="subtitle2" color="text.secondary" sx={{ mt: 1 }}>PlayStation Украина</Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    Цена UAH × множитель
                  </Typography>
                  <TextField
                    fullWidth
                    type="number"
                    label="Множитель Украина"
                    value={pricing.ukraine_markup}
                    onChange={(e) => setPricingNum('ukraine_markup', e.target.value)}
                    inputProps={{ min: 0.1, step: 0.1 }}
                  />
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="subtitle2" color="text.secondary" sx={{ mt: 1 }}>Xbox США</Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    Цена USD × множитель
                  </Typography>
                  <TextField
                    fullWidth
                    type="number"
                    label="Множитель Xbox USD"
                    value={pricing.xbox_usd_multiplier}
                    onChange={(e) => setPricingNum('xbox_usd_multiplier', e.target.value)}
                    inputProps={{ min: 1, step: 1 }}
                  />
                </Grid>
              </Grid>
            </CardContent>
          </Card>

          {preview.examples && (
            <Card sx={{ mb: 2 }}>
              <CardContent>
                <Typography variant="h6" gutterBottom>Примеры (текущие формулы)</Typography>
                <Typography variant="body2">Курс TRY: {preview.try_rub_rate?.toFixed(2) ?? '—'} ₽</Typography>
                <Box component="ul" sx={{ mt: 1, pl: 2 }}>
                  <li>500 TRY (TR) → {preview.examples.turkey_500_try?.toFixed(0)} ₽</li>
                  <li>1000 UAH (UA) → {preview.examples.ukraine_1000_uah?.toFixed(0)} ₽</li>
                  <li>10 USD (Xbox) → {preview.examples.xbox_10_usd?.toFixed(0)} ₽</li>
                </Box>
              </CardContent>
            </Card>
          )}

          <Button variant="contained" onClick={savePricing} disabled={saving}>{saving ? 'Сохранение...' : 'Сохранить формулы'}</Button>
        </>
      )}

      {tab === 2 && (
        <>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>СБП</Typography>
              <Grid container spacing={2}>
                <Grid item xs={6}><TextField fullWidth label="Телефон" value={payment.sbp?.phone || ''} onChange={(e) => setPaymentField('sbp', 'phone', e.target.value)} /></Grid>
                <Grid item xs={6}><TextField fullWidth label="Банк" value={payment.sbp?.bank || ''} onChange={(e) => setPaymentField('sbp', 'bank', e.target.value)} /></Grid>
                <Grid item xs={6}><TextField fullWidth label="Получатель" value={payment.sbp?.receiver || ''} onChange={(e) => setPaymentField('sbp', 'receiver', e.target.value)} /></Grid>
              </Grid>
            </CardContent>
          </Card>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Карта</Typography>
              <TextField fullWidth label="Номер карты" value={payment.card?.number || ''} onChange={(e) => setPaymentField('card', 'number', e.target.value)} sx={{ mb: 2 }} />
              <TextField fullWidth label="Банк" value={payment.card?.bank || ''} onChange={(e) => setPaymentField('card', 'bank', e.target.value)} />
            </CardContent>
          </Card>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>Криптовалюта</Typography>
              <TextField fullWidth label="Binance ID" value={payment.crypto?.binance || ''} onChange={(e) => setPaymentField('crypto', 'binance', e.target.value)} sx={{ mb: 2 }} />
              <TextField fullWidth label="Bybit ID" value={payment.crypto?.bybit || ''} onChange={(e) => setPaymentField('crypto', 'bybit', e.target.value)} sx={{ mb: 2 }} />
              <TextField fullWidth label="TRC20" value={payment.crypto?.trc20 || ''} onChange={(e) => setPaymentField('crypto', 'trc20', e.target.value)} />
            </CardContent>
          </Card>
          <Button variant="contained" onClick={savePayment} disabled={saving}>{saving ? 'Сохранение...' : 'Сохранить'}</Button>
        </>
      )}

      <Snackbar open={!!snack} autoHideDuration={4000} onClose={() => setSnack('')} message={snack} />
    </div>
  )
}
