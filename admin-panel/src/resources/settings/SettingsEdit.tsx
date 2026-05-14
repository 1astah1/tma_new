import { useState, useEffect } from 'react'
import { Card, CardContent, Typography, TextField, Button, Grid, Snackbar } from '@mui/material'

export const SettingsEdit = () => {
  const [data, setData] = useState<any>({ sbp: {}, card: {}, crypto: {} })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [snack, setSnack] = useState('')

  useEffect(() => {
    fetch('/api/v1/admin/settings?key=payment_details', {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
      .then((r) => r.json())
      .then((d) => {
        try { setData(JSON.parse(d.value)) } catch { setData(d) }
      })
      .finally(() => setLoading(false))
  }, [])

  const save = async () => {
    setSaving(true)
    await fetch('/api/v1/admin/settings', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${localStorage.getItem('token')}`,
      },
      body: JSON.stringify({ key: 'payment_details', value: data }),
    })
    setSaving(false)
    setSnack('Сохранено!')
  }

  const set = (section: string, field: string, value: string) => {
    setData((prev: any) => ({ ...prev, [section]: { ...prev[section], [field]: value } }))
  }

  if (loading) return <div>Загрузка...</div>

  return (
    <div style={{ padding: 20, maxWidth: 800 }}>
      <Typography variant="h4" gutterBottom>Платёжные реквизиты</Typography>

      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>🏦 СБП</Typography>
          <Grid container spacing={2}>
            <Grid item xs={6}><TextField fullWidth label="Номер телефона" value={data.sbp?.phone || ''} onChange={(e) => set('sbp', 'phone', e.target.value)} /></Grid>
            <Grid item xs={6}><TextField fullWidth label="Банк" value={data.sbp?.bank || ''} onChange={(e) => set('sbp', 'bank', e.target.value)} /></Grid>
            <Grid item xs={6}><TextField fullWidth label="Получатель" value={data.sbp?.receiver || ''} onChange={(e) => set('sbp', 'receiver', e.target.value)} /></Grid>
          </Grid>
        </CardContent>
      </Card>

      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>💳 Карта</Typography>
          <TextField fullWidth label="Номер карты" value={data.card?.number || ''} onChange={(e) => set('card', 'number', e.target.value)} sx={{ mb: 2 }} />
          <TextField fullWidth label="Банк" value={data.card?.bank || ''} onChange={(e) => set('card', 'bank', e.target.value)} />
        </CardContent>
      </Card>

      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>₿ Криптовалюта</Typography>
          <TextField fullWidth label="Binance ID" value={data.crypto?.binance || ''} onChange={(e) => set('crypto', 'binance', e.target.value)} sx={{ mb: 2 }} />
          <TextField fullWidth label="Bybit ID" value={data.crypto?.bybit || ''} onChange={(e) => set('crypto', 'bybit', e.target.value)} sx={{ mb: 2 }} />
          <TextField fullWidth label="TRC20 Адрес" value={data.crypto?.trc20 || ''} onChange={(e) => set('crypto', 'trc20', e.target.value)} />
        </CardContent>
      </Card>

      <Button variant="contained" onClick={save} disabled={saving}>{saving ? 'Сохранение...' : 'Сохранить'}</Button>
      <Snackbar open={!!snack} autoHideDuration={3000} onClose={() => setSnack('')} message={snack} />
    </div>
  )
}
