import { useState, useEffect } from 'react'
import { List, Datagrid, TextField, NumberField, DateField, SelectField, SearchInput, SelectInput, useListContext, Button, useNotify, useRefresh, Pagination, useDataProvider } from 'react-admin'
import { Box, Chip, Paper, Typography, Button as MuiButton } from '@mui/material'
import FileDownloadIcon from '@mui/icons-material/FileDownload'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import CancelIcon from '@mui/icons-material/Cancel'
import KeyIcon from '@mui/icons-material/Key'
import LockIcon from '@mui/icons-material/Lock'
import AccessTimeIcon from '@mui/icons-material/AccessTime'
import CheckIcon from '@mui/icons-material/Check'
import WarningIcon from '@mui/icons-material/Warning'

const orderStatuses = [
  { id: 'NEW', name: 'Новый' }, { id: 'WAITING_PAYMENT', name: 'Ожидает оплаты' },
  { id: 'PAYMENT_VERIFICATION', name: 'Проверка платежа' }, { id: 'PAID', name: 'Оплачен' },
  { id: 'WAITING_ACTIVATION', name: 'В очереди' }, { id: 'AWAITING_CREDENTIALS', name: 'Требуются данные' },
  { id: 'CREDENTIALS_RECEIVED', name: 'Данные получены' }, { id: 'AWAITING_2FA', name: 'Ожидает код' },
  { id: 'ACTIVATING', name: 'Активация...' }, { id: 'ACTIVATED', name: 'Активирован' },
  { id: 'KEY_ISSUED', name: 'Ключ выдан' }, { id: 'COMPLETED', name: 'Завершён' },
  { id: 'CANCELLED', name: 'Отменён' }, { id: 'REFUND_REQUESTED', name: 'Запрос возврата' },
  { id: 'REFUNDED', name: 'Возвращён' },
]

const paymentMethods = [
  { id: 'sbp', name: 'СБП' },
  { id: 'card', name: 'Картой' },
  { id: 'crypto', name: 'Криптовалюта' },
]

const statusColors: Record<string, 'success' | 'warning' | 'error' | 'info' | 'default'> = {
  NEW: 'info', WAITING_PAYMENT: 'default', PAYMENT_VERIFICATION: 'warning',
  PAID: 'success', WAITING_ACTIVATION: 'info', AWAITING_CREDENTIALS: 'warning',
  CREDENTIALS_RECEIVED: 'info', CREDENTIALS_INVALID: 'error',
  AWAITING_2FA: 'warning', INVALID_2FA: 'error',
  ACTIVATING: 'warning', ACTIVATED: 'success', KEY_ISSUED: 'success', COMPLETED: 'success',
  CANCELLED: 'error', REFUND_REQUESTED: 'warning', REFUNDED: 'error',
}

const quickFilters = [
  { label: 'Все', value: '', icon: null },
  { label: 'Новые', value: 'NEW', icon: null, color: 'info' },
  { label: 'Ожидают оплаты', value: 'WAITING_PAYMENT', icon: <AccessTimeIcon fontSize="small" />, color: 'default' },
  { label: 'Проверка', value: 'PAYMENT_VERIFICATION', icon: <WarningIcon fontSize="small" />, color: 'warning' },
  { label: 'Оплачены', value: 'PAID', icon: <CheckIcon fontSize="small" />, color: 'success' },
  { label: 'Активация', value: 'WAITING_ACTIVATION,AWAITING_CREDENTIALS,CREDENTIALS_RECEIVED,AWAITING_2FA,ACTIVATING', icon: <LockIcon fontSize="small" />, color: 'warning' },
  { label: 'Завершены', value: 'COMPLETED,KEY_ISSUED,ACTIVATED', icon: <CheckCircleIcon fontSize="small" />, color: 'success' },
  { label: 'Отменены', value: 'CANCELLED,REFUNDED', icon: <CancelIcon fontSize="small" />, color: 'error' },
]

const inlineActionsMap: Record<string, { status: string; label: string; color: string }[]> = {
  PAYMENT_VERIFICATION: [
    { status: 'PAID', label: '✓', color: 'success' },
  ],
  PAID: [
    { status: 'KEY_ISSUED', label: '🔑', color: 'primary' },
    { status: 'WAITING_ACTIVATION', label: 'Взять', color: 'primary' },
  ],
  WAITING_ACTIVATION: [
    { status: 'AWAITING_CREDENTIALS', label: 'Данные', color: 'primary' },
  ],
  CREDENTIALS_RECEIVED: [
    { status: 'AWAITING_2FA', label: '2FA', color: 'primary' },
  ],
  ACTIVATING: [
    { status: 'ACTIVATED', label: '✓', color: 'success' },
  ],
  ACTIVATED: [
    { status: 'COMPLETED', label: '✓', color: 'success' },
  ],
  KEY_ISSUED: [
    { status: 'COMPLETED', label: '✓', color: 'success' },
  ],
  REFUND_REQUESTED: [
    { status: 'REFUNDED', label: 'Возврат', color: 'warning' },
  ],
}

function StatusField({ record }: any) {
  if (!record) return null
  const label = orderStatuses.find(s => s.id === record.status)?.name || record.status
  return <Chip label={label} size="small" color={statusColors[record.status] || 'default'} />
}

function DeliveryField({ record }: any) {
  if (!record) return null
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
      {record.delivery_method === 'key' ? <KeyIcon fontSize="small" sx={{ color: '#4caf50' }} /> : <LockIcon fontSize="small" sx={{ color: '#ff9800' }} />}
      <span>{record.delivery_method === 'key' ? 'Ключ' : 'Активация'}</span>
    </Box>
  )
}

function AmountField({ record }: any) {
  if (!record) return null
  const amount = record.payment_amount || 0
  return (
    <Typography variant="body2" sx={{ fontWeight: 'bold', color: amount > 0 ? '#4caf50' : '#888' }}>
      {amount > 0 ? `${amount.toFixed(2)} ₽` : '—'}
    </Typography>
  )
}

function BulkActions() {
  const { selectedIds } = useListContext()
  const notify = useNotify()
  const refresh = useRefresh()
  const token = localStorage.getItem('token')

  const bulkUpdate = async (status: string) => {
    if (!selectedIds?.length) {
      notify('Ничего не выбрано', { type: 'warning' })
      return
    }
    let success = 0
    for (const id of selectedIds) {
      try {
        await fetch(`/api/v1/admin/orders/${id}/status`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
          body: JSON.stringify({ status, comment: 'Bulk action' }),
        })
        success++
      } catch {}
    }
    notify(`Обновлено: ${success}/${selectedIds.length}`, { type: success > 0 ? 'success' : 'error' })
    refresh()
  }

  if (!selectedIds?.length) return null

  return (
    <Paper sx={{ p: 2, mb: 2, display: 'flex', gap: 1, alignItems: 'center' }}>
      <span>Выбрано: {selectedIds.length}</span>
      <Button label="Подтвердить" onClick={() => bulkUpdate('PAID')}><CheckCircleIcon /></Button>
      <Button label="Отменить" color="error" onClick={() => bulkUpdate('CANCELLED')}><CancelIcon /></Button>
    </Paper>
  )
}

function ExportButton() {
  const notify = useNotify()
  const token = localStorage.getItem('token')

  const handleExport = async () => {
    try {
      const res = await fetch('/api/v1/admin/orders?limit=1000', {
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await res.json()
      const orders = data.data || []

      const headers = ['ID', 'Статус', 'Доставка', 'Оплата', 'Сумма', 'Дата', 'Товар']
      const rows = orders.map((o: any) => [
        o.id,
        o.status,
        o.delivery_method,
        o.payment_method || '',
        o.payment_amount || 0,
        o.created_at,
        o.product?.title || '',
      ])

      const csv = [headers, ...rows].map(r => r.join(',')).join('\n')
      const blob = new Blob(['\ufeff' + csv], { type: 'text/csv;charset=utf-8;' })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `orders_${new Date().toISOString().split('T')[0]}.csv`
      link.click()
      URL.revokeObjectURL(url)
      notify('Экспорт завершён', { type: 'success' })
    } catch {
      notify('Ошибка экспорта', { type: 'error' })
    }
  }

  return (
    <Button label="Экспорт CSV" onClick={handleExport}>
      <FileDownloadIcon />
    </Button>
  )
}

function OrderActions() {
  return (
    <Box sx={{ display: 'flex', gap: 1 }}>
      <ExportButton />
    </Box>
  )
}

function QuickFilters({ currentFilter, onFilterChange }: { currentFilter: string; onFilterChange: (v: string) => void }) {
  const [counts, setCounts] = useState<Record<string, number>>({})

  useEffect(() => {
    fetch('/api/v1/admin/orders?limit=1', {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
      .then(r => r.json())
      .then(data => {
        const total = data.meta?.total || 0
        // Fetch counts per status group
        const groups: Record<string, string[]> = {
          'NEW': ['NEW'],
          'WAITING_PAYMENT': ['WAITING_PAYMENT'],
          'PAYMENT_VERIFICATION': ['PAYMENT_VERIFICATION'],
          'PAID': ['PAID'],
          'activation': ['WAITING_ACTIVATION', 'AWAITING_CREDENTIALS', 'CREDENTIALS_RECEIVED', 'AWAITING_2FA', 'ACTIVATING'],
          'completed': ['COMPLETED', 'KEY_ISSUED', 'ACTIVATED'],
          'cancelled': ['CANCELLED', 'REFUNDED'],
        }
        const c: Record<string, number> = { '': total }
        Object.entries(groups).forEach(([key, statuses]) => {
          c[key] = 0
          statuses.forEach(s => {
            fetch(`/api/v1/admin/orders?status=${s}&limit=1`, {
              headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
            })
              .then(r => r.json())
              .then(d => { c[key] = (c[key] || 0) + (d.meta?.total || 0) })
          })
        })
        setCounts(c)
      })
  }, [])

  return (
    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1, mb: 2 }}>
      {quickFilters.map((f) => {
        const isActive = currentFilter === f.value
        const count = counts[f.value] ?? 0
        return (
          <Chip
            key={f.value || 'all'}
            label={
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                {f.icon}
                <span>{f.label}</span>
                {count > 0 && <span style={{ opacity: 0.7, fontSize: '0.75rem' }}>({count})</span>}
              </Box>
            }
            color={isActive ? (f.color as any) || 'primary' : 'default'}
            variant={isActive ? 'filled' : 'outlined'}
            onClick={() => onFilterChange(f.value)}
            sx={{ cursor: 'pointer', fontWeight: isActive ? 'bold' : 'normal' }}
          />
        )
      })}
    </Box>
  )
}

const filters = [
  <SearchInput key="search" source="search" alwaysOn placeholder="Поиск по ID..." />,
  <SelectInput key="payment_method" source="payment_method" label="Оплата" choices={paymentMethods} />,
  <SelectInput key="delivery_method" source="delivery_method" label="Доставка" choices={[
    { id: 'key', name: 'Ключ' },
    { id: 'activation', name: 'Активация' },
  ]} />,
]

function ClientField({ record }: any) {
  if (!record) return null
  return (
    <Typography variant="body2" sx={{ color: '#fff' }}>
      {record.user?.username || record.user?.first_name || '—'}
    </Typography>
  )
}

function InlineActions({ record }: any) {
  const refresh = useRefresh()
  const notify = useNotify()
  const token = localStorage.getItem('token')

  if (!record) return null

  const actions = inlineActionsMap[record.status] || []
  const mainActions = actions

  if (mainActions.length === 0) return null

  const doAction = async (status: string) => {
    try {
      const res = await fetch(`/api/v1/admin/orders/${record.id}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ status, comment: '' }),
      })
      if (res.ok) {
        notify('Статус обновлён', { type: 'success' })
        refresh()
      } else {
        const err = await res.json()
        const msg = err.error?.message || err.message || 'Ошибка'
        notify(msg, { type: 'error' })
      }
    } catch (e: any) {
      notify('Ошибка сети: ' + e.message, { type: 'error' })
    }
  }

  return (
    <Box sx={{ display: 'flex', gap: 0.5 }}>
      {mainActions.map(a => (
        <MuiButton
          key={a.status}
          size="small"
          variant="outlined"
          color={a.color as any}
          onClick={() => doAction(a.status)}
          sx={{ minWidth: 'auto', px: 1, fontSize: '0.65rem', py: 0.3 }}
        >
          {a.label}
        </MuiButton>
      ))}
    </Box>
  )
}

function OrderListContent() {
  const { filterValues, setFilters } = useListContext()
  const currentStatus = filterValues?.status || ''

  const handleFilterChange = (value: string) => {
    setFilters({ ...filterValues, status: value || undefined }, {})
  }

  return (
    <>
      <QuickFilters currentFilter={currentStatus} onFilterChange={handleFilterChange} />
      <BulkActions />
      <Datagrid rowClick="show" bulkActionButtons={false} sx={{ '& .RaDatagrid-rowCell': { py: 1.5 } }}>
        <TextField source="id" label="ID" sx={{ fontFamily: 'monospace', fontSize: '0.7rem', maxWidth: 100 }} />
        <TextField source="product.title" label="Товар" />
        <ClientField source="user" label="Клиент" />
        <StatusField source="status" label="Статус" />
        <DeliveryField source="delivery_method" label="Доставка" />
        <AmountField source="payment_amount" label="Сумма" />
        <DateField source="created_at" label="Дата" showTime />
        <InlineActions source="actions" label="Действия" />
      </Datagrid>
    </>
  )
}

export const OrderList = () => (
  <List filters={filters} actions={<OrderActions />} pagination={<Pagination rowsPerPageOptions={[10, 25, 50, 100]} />}>
    <OrderListContent />
  </List>
)
