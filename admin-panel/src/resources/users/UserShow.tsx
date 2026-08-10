import { useState, useEffect } from 'react'
import { Show, useRecordContext } from 'react-admin'
import { Link } from 'react-router-dom'
import {
  Card, CardContent, Typography, Box, Chip, Grid, Button,
  TextField as MuiTextField, CircularProgress, Table, TableBody, TableCell, TableHead, TableRow,
} from '@mui/material'
import PersonIcon from '@mui/icons-material/Person'
import BlockIcon from '@mui/icons-material/Block'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart'
import NoteIcon from '@mui/icons-material/Note'

const statusColors: Record<string, 'success' | 'warning' | 'error' | 'info' | 'default'> = {
  NEW: 'info', WAITING_PAYMENT: 'default', PAYMENT_VERIFICATION: 'warning',
  PAID: 'success', WAITING_ACTIVATION: 'info', AWAITING_CREDENTIALS: 'warning',
  CREDENTIALS_RECEIVED: 'info', CREDENTIALS_INVALID: 'error',
  AWAITING_2FA: 'warning', INVALID_2FA: 'error',
  ACTIVATING: 'warning', ACTIVATED: 'success', KEY_ISSUED: 'success', COMPLETED: 'success',
  CANCELLED: 'error', REFUND_REQUESTED: 'warning', REFUNDED: 'error',
}

const statusLabels: Record<string, string> = {
  NEW: 'Новый', WAITING_PAYMENT: 'Ожидает оплаты', PAYMENT_VERIFICATION: 'Проверка',
  PAID: 'Оплачен', WAITING_ACTIVATION: 'В очереди', AWAITING_CREDENTIALS: 'Требуются данные',
  CREDENTIALS_RECEIVED: 'Данные получены', CREDENTIALS_INVALID: 'Данные неверны',
  AWAITING_2FA: 'Ожидает код', INVALID_2FA: 'Код неверен',
  ACTIVATING: 'Активация', ACTIVATED: 'Активирован', KEY_ISSUED: 'Ключ выдан', COMPLETED: 'Завершён',
  CANCELLED: 'Отменён', REFUND_REQUESTED: 'Запрос возврата', REFUNDED: 'Возвращён',
}

function formatOrderAmount(value?: number | null) {
  if (value == null) return '—'
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(value)
}

function UserOrders({ userId }: { userId: string }) {
  const [orders, setOrders] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const token = localStorage.getItem('token')

  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch(`/api/v1/admin/orders?user_id=${userId}&limit=100`, {
          headers: { Authorization: `Bearer ${token}` },
        })
        const data = await res.json()
        setOrders(data.data || [])
      } catch {}
      setLoading(false)
    }
    load()
  }, [userId, token])

  if (loading) return <CircularProgress size={24} />
  if (orders.length === 0) {
    return (
      <Typography sx={{ color: '#888' }}>
        Нет заказов. Заявки появятся после нажатия «Оформить с менеджером» в магазине.
      </Typography>
    )
  }

  return (
    <Box sx={{ overflowX: 'auto' }}>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell sx={{ color: '#888' }}>ID</TableCell>
            <TableCell sx={{ color: '#888' }}>Товар</TableCell>
            <TableCell sx={{ color: '#888' }}>Статус</TableCell>
            <TableCell sx={{ color: '#888' }}>Сумма</TableCell>
            <TableCell sx={{ color: '#888' }}>Дата</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {orders.map((order) => (
            <TableRow key={order.id} hover component={Link} to={`/orders/${order.id}/show`} sx={{ textDecoration: 'none', cursor: 'pointer' }}>
              <TableCell sx={{ color: '#90caf9', fontFamily: 'monospace', fontSize: '0.75rem' }}>
                {order.id?.substring(0, 8)}
              </TableCell>
              <TableCell sx={{ color: '#fff' }}>{order.product?.title || '—'}</TableCell>
              <TableCell sx={{ color: '#fff' }}>{statusLabels[order.status] || order.status}</TableCell>
              <TableCell sx={{ color: '#fff' }}>{formatOrderAmount(order.payment_amount)}</TableCell>
              <TableCell sx={{ color: '#fff' }}>
                {order.created_at ? new Date(order.created_at).toLocaleString('ru-RU') : '—'}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Box>
  )
}

function UserDetail() {
  const record = useRecordContext()
  const token = localStorage.getItem('token')
  const [stats, setStats] = useState<any>({})
  const [notes, setNotes] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!record?.id) return
    fetch(`/api/v1/admin/users/${record.id}/stats`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(r => r.json())
      .then(setStats)
      .catch(() => {})
    setNotes(record.admin_notes || '')
  }, [record?.id])

  const handleBan = async () => {
    if (!record?.id) return
    await fetch(`/api/v1/admin/users/${record.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ is_banned: !record.is_banned }),
    })
    window.location.reload()
  }

  const handleSaveNotes = async () => {
    if (!record?.id) return
    setSaving(true)
    await fetch(`/api/v1/admin/users/${record.id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ admin_notes: notes }),
    })
    setSaving(false)
  }

  if (!record) return null

  return (
    <Box sx={{ p: 2 }}>
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                <PersonIcon color="primary" />
                <Typography variant="h6" sx={{ fontWeight: 'bold', color: '#fff' }}>Профиль</Typography>
              </Box>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Typography variant="caption" sx={{ color: '#888' }}>Username</Typography>
                  <Typography sx={{ color: '#fff' }}>{record.username || '—'}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" sx={{ color: '#888' }}>Telegram ID</Typography>
                  <Typography sx={{ color: '#fff', fontFamily: 'monospace' }}>{record.telegram_id}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" sx={{ color: '#888' }}>Имя</Typography>
                  <Typography sx={{ color: '#fff' }}>{record.first_name || '—'}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" sx={{ color: '#888' }}>Последний вход в TMA</Typography>
                  <Typography sx={{ color: '#fff' }}>
                    {record.last_interaction
                      ? new Date(record.last_interaction).toLocaleString('ru-RU')
                      : '—'}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="caption" sx={{ color: '#888' }}>Регистрация</Typography>
                  <Typography sx={{ color: '#fff' }}>
                    {record.created_at ? new Date(record.created_at).toLocaleString('ru-RU') : '—'}
                  </Typography>
                </Grid>
              </Grid>
              <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
                <Chip
                  label={record.is_banned ? 'Заблокирован' : 'Активен'}
                  color={record.is_banned ? 'error' : 'success'}
                  size="small"
                />
                <Button
                  size="small"
                  variant="outlined"
                  color={record.is_banned ? 'success' : 'error'}
                  startIcon={record.is_banned ? <CheckCircleIcon /> : <BlockIcon />}
                  onClick={handleBan}
                >
                  {record.is_banned ? 'Разблокировать' : 'Заблокировать'}
                </Button>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                <ShoppingCartIcon color="primary" />
                <Typography variant="h6" sx={{ fontWeight: 'bold', color: '#fff' }}>Статистика</Typography>
              </Box>
              <Typography variant="h3" sx={{ color: '#fff', fontWeight: 'bold', textAlign: 'center' }}>
                {stats.total_orders || 0}
              </Typography>
              <Typography variant="body2" sx={{ color: '#888', textAlign: 'center' }}>Всего заказов</Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                <NoteIcon color="primary" />
                <Typography variant="h6" sx={{ fontWeight: 'bold', color: '#fff' }}>Заметки админа</Typography>
              </Box>
              <MuiTextField
                fullWidth
                multiline
                rows={3}
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                sx={{ mb: 2 }}
                InputLabelProps={{ style: { color: '#888' } }}
                inputProps={{ style: { color: '#fff' } }}
              />
              <Button variant="contained" onClick={handleSaveNotes} disabled={saving}>
                {saving ? 'Сохранение...' : 'Сохранить'}
              </Button>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12}>
          <Typography variant="h6" sx={{ fontWeight: 'bold', color: '#fff', mb: 2 }}>Заказы пользователя</Typography>
          <UserOrders userId={String(record.id)} />
        </Grid>
      </Grid>
    </Box>
  )
}

export const UserShow = () => (
  <Show>
    <UserDetail />
  </Show>
)
