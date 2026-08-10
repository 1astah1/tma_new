import { useState } from 'react'
import { List, Datagrid, TextField, NumberField, SelectField, ChipField, EditButton, ShowButton, TextInput, SelectInput, TopToolbar, CreateButton, Button, useNotify, useRefresh } from 'react-admin'
import { Box, Chip, Typography, Dialog, DialogTitle, DialogContent, DialogActions, Button as MuiButton, LinearProgress } from '@mui/material'
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart'
import CloudDownloadIcon from '@mui/icons-material/CloudDownload'

const statusColors: Record<string, 'success' | 'error' | 'default'> = {
  active: 'success',
  inactive: 'error',
}

function StatusField({ record }: any) {
  if (!record) return null
  const label = record.status === 'active' ? 'Активен' : 'Неактивен'
  return <Chip label={label} size="small" color={statusColors[record.status] || 'default'} />
}

function DiscountField({ record }: any) {
  if (!record) return null
  const discount = record.discount_percent || 0
  if (discount <= 0) return <span style={{ color: '#888' }}>—</span>
  return <span style={{ color: '#4caf50', fontWeight: 'bold' }}>-{discount}%</span>
}

function OrderCountField({ record }: any) {
  if (!record) return null
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
      <ShoppingCartIcon fontSize="small" sx={{ color: '#888' }} />
      <Typography variant="body2" sx={{ color: '#fff' }}>{record.order_count || 0}</Typography>
    </Box>
  )
}

const ImportAllButton = () => {
  const [open, setOpen] = useState(false)
  const [loading, setLoading] = useState(false)
  const [progress, setProgress] = useState<string | null>(null)
  const notify = useNotify()
  const refresh = useRefresh()

  const handleStart = async () => {
    setLoading(true)
    setProgress('Запуск...')
    const token = localStorage.getItem('token')
    try {
      const res = await fetch('/api/v1/admin/catalog-parser/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ full: true, sources: ['xbox', 'ps'] }),
      })
      if (!res.ok) throw new Error(await res.text())
      notify('Импорт запущен', { type: 'success' })
      setOpen(false)
      refresh()
      pollStatus(token!)
    } catch (e: any) {
      notify(`Ошибка: ${e.message}`, { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  const pollStatus = async (token: string) => {
    const poll = setInterval(async () => {
      try {
        const res = await fetch('/api/v1/admin/catalog-parser/status', {
          headers: { Authorization: `Bearer ${token}` },
        })
        if (!res.ok) { clearInterval(poll); return }
        const data = await res.json()
        if (!data.running) {
          clearInterval(poll)
          setProgress(null)
          notify('Импорт завершён', { type: 'success' })
          refresh()
          return
        }
        if (data.current_source || data.current_stage) {
          setProgress(`${data.current_source}: ${data.current_stage} (${data.imported} импортировано)`)
        }
      } catch { clearInterval(poll) }
    }, 3000)
  }

  return (
    <>
      <Button label="Импорт всех игр" onClick={() => setOpen(true)}>
        <CloudDownloadIcon />
      </Button>
      <Dialog open={open} onClose={() => !loading && setOpen(false)}>
        <DialogTitle>Запустить полный импорт игр?</DialogTitle>
        <DialogContent>
          <Typography>Будет выполнен парсинг PS Store и Xbox Store.</Typography>
          {progress && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="body2" sx={{ mb: 1 }}>{progress}</Typography>
              <LinearProgress />
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <MuiButton onClick={() => setOpen(false)} disabled={loading}>Отмена</MuiButton>
          <MuiButton onClick={handleStart} disabled={loading} variant="contained" color="primary">
            {loading ? 'Запуск...' : 'Запустить'}
          </MuiButton>
        </DialogActions>
      </Dialog>
    </>
  )
}

const ProductActions = () => (
  <TopToolbar>
    <ImportAllButton />
    <CreateButton />
  </TopToolbar>
)

const filters = [
  <TextInput key="search" source="search" label="Поиск" alwaysOn placeholder="Название..." />,
  <SelectInput key="platform" source="platform" label="Платформа" choices={[
    { id: 'ps4', name: 'PS4' }, { id: 'ps5', name: 'PS5' }, { id: 'xbox', name: 'Xbox' }, { id: 'pc', name: 'PC' },
  ]} />,
  <SelectInput key="type" source="type" label="Тип" choices={[
    { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
  ]} />,
  <SelectInput key="status" source="status" label="Статус" choices={[
    { id: 'active', name: 'Активен' }, { id: 'inactive', name: 'Неактивен' },
  ]} />,
]

export const ProductList = () => (
  <List filters={filters} filterDefaultValues={{ type: 'game' }} actions={<ProductActions />}>
      <Datagrid rowClick="edit">
        <TextField source="title" label="Название" />
        <ChipField source="platform" label="Платформа" />
        <SelectField source="type" label="Тип" choices={[
          { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
        ]} />
        <NumberField source="price" label="Цена" options={{ style: 'currency', currency: 'RUB' }} />
        <DiscountField source="discount_percent" label="Скидка" />
        <OrderCountField source="order_count" label="Заказы" />
        <StatusField source="status" label="Статус" />
        <Box sx={{ display: 'flex', gap: 0.5 }}>
          <EditButton />
          <ShowButton />
        </Box>
      </Datagrid>
  </List>
)
