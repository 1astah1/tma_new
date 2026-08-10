import { useState } from 'react'
import { useRecordContext, useRefresh, useNotify, Link } from 'react-admin'
import { Box, Button, Card, CardContent, Chip, Stack, Typography } from '@mui/material'
import SyncIcon from '@mui/icons-material/Sync'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'

const apiUrl = '/api/v1/admin'

type CatalogImportInfo = {
  id: string
  source: string
  external_id: string
  title: string
  status: string
  original_price_rub?: number
  updated_at?: string
}

function authHeaders() {
  const headers = new Headers({ 'Content-Type': 'application/json' })
  const token = localStorage.getItem('token')
  if (token) headers.set('Authorization', `Bearer ${token}`)
  return headers
}

const sourceLabels: Record<string, string> = {
  ps_store: 'PS Store',
  xbox_store: 'Xbox Store',
}

export function ProductImportPanel() {
  const record = useRecordContext()
  const refresh = useRefresh()
  const notify = useNotify()
  const [syncing, setSyncing] = useState(false)

  const imp: CatalogImportInfo | null = record?.catalog_import || null
  if (!record || record.type !== 'game') return null

  const syncPrice = async () => {
    if (!record.id) return
    setSyncing(true)
    try {
      const res = await fetch(`${apiUrl}/products/${record.id}/sync-from-import`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const json = await res.json().catch(() => ({}))
      if (!res.ok) throw new Error(json.error?.message || json.message || 'Ошибка синхронизации')
      notify(`Цена обновлена: ${Math.round(json.price || 0)} ₽`, { type: 'success' })
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setSyncing(false)
    }
  }

  if (!imp) {
    return (
      <Card sx={{ mb: 2, border: '1px dashed #555' }}>
        <CardContent>
          <Typography variant="subtitle2" color="text.secondary">
            Нет связи с импортом — товар создан вручную или импорт удалён
          </Typography>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card sx={{ mb: 2, bgcolor: '#0d1b2a', border: '1px solid #1e3a5f' }}>
      <CardContent>
        <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 1 }}>
          Источник цены (импорт)
        </Typography>
        <Stack spacing={1}>
          <Typography variant="body2">
            {sourceLabels[imp.source] || imp.source} · {imp.title}
          </Typography>
          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
            <Chip label={imp.status === 'approved' ? 'Утверждено' : imp.status} size="small" color="primary" variant="outlined" />
            {imp.original_price_rub ? (
              <Chip label={`${Math.round(imp.original_price_rub)} ₽ в импорте`} size="small" />
            ) : null}
            {record.price && imp.original_price_rub && Math.abs(record.price - imp.original_price_rub) > 1 ? (
              <Chip label="Цена расходится" size="small" color="warning" />
            ) : null}
          </Stack>
          <Stack direction="row" spacing={1} flexWrap="wrap">
            <Button
              size="small"
              variant="outlined"
              startIcon={<SyncIcon />}
              onClick={syncPrice}
              disabled={syncing}
            >
              {syncing ? 'Обновляем...' : 'Обновить цену из стора'}
            </Button>
            <Button
              size="small"
              component={Link}
              to={`/catalog-imports?filter=${encodeURIComponent(JSON.stringify({ search: imp.title?.slice(0, 40) }))}`}
              startIcon={<OpenInNewIcon />}
            >
              Открыть в импорте
            </Button>
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  )
}
