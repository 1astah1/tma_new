import { useState } from 'react'
import { Show, useRecordContext, Button, useRefresh } from 'react-admin'
import { Typography, Box, Card, CardMedia, Dialog, DialogTitle, DialogContent, DialogActions, TextField as MuiTextField, Divider, Chip, Grid, Accordion, AccordionSummary, AccordionDetails, Alert, Button as MuiButton, Paper } from '@mui/material'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import CancelIcon from '@mui/icons-material/Cancel'
import AssignmentIcon from '@mui/icons-material/Assignment'
import LoginIcon from '@mui/icons-material/Login'
import DoneAllIcon from '@mui/icons-material/DoneAll'
import ReceiptIcon from '@mui/icons-material/Receipt'
import PersonIcon from '@mui/icons-material/Person'
import InventoryIcon from '@mui/icons-material/Inventory'
import HistoryIcon from '@mui/icons-material/History'
import WarningIcon from '@mui/icons-material/Warning'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import MonetizationOnIcon from '@mui/icons-material/MonetizationOn'
import ArrowForwardIcon from '@mui/icons-material/ArrowForward'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import ErrorIcon from '@mui/icons-material/Error'
import { OrderChat } from './OrderChat'

const orderStatuses = [
  { id: 'NEW', name: 'Новый' },
  { id: 'WAITING_PAYMENT', name: 'Ожидает оплаты' },
  { id: 'PAYMENT_VERIFICATION', name: 'Проверка платежа' },
  { id: 'PAID', name: 'Оплачен' },
  { id: 'COMPLETED', name: 'Завершён' },
  { id: 'CANCELLED', name: 'Отменён' },
  { id: 'REFUND_REQUESTED', name: 'Запрос возврата' },
  { id: 'REFUNDED', name: 'Возвращён' },
  // legacy
  { id: 'WAITING_ACTIVATION', name: 'В работе (legacy)' },
  { id: 'KEY_ISSUED', name: 'Выдан (legacy)' },
  { id: 'ACTIVATED', name: 'Активирован (legacy)' },
]

const statusLabels: Record<string, string> = {}
orderStatuses.forEach(s => { statusLabels[s.id] = s.name })

const statusColors: Record<string, 'success' | 'warning' | 'error' | 'info' | 'default'> = {
  NEW: 'info',
  WAITING_PAYMENT: 'default',
  PAYMENT_VERIFICATION: 'warning',
  PAID: 'success',
  COMPLETED: 'success',
  CANCELLED: 'error',
  REFUND_REQUESTED: 'warning',
  REFUNDED: 'error',
}

const legacyComplete = { status: 'COMPLETED', label: 'Завершить заказ', icon: <DoneAllIcon />, color: 'success' as const, variant: 'contained' as const }
const legacyCancel = { status: 'CANCELLED', label: 'Отменить', icon: <CancelIcon />, color: 'error' as const, variant: 'outlined' as const }

const statusActions: Record<string, { status: string; label: string; icon: JSX.Element; color: 'primary' | 'error' | 'warning' | 'success'; variant?: 'contained' | 'outlined' }[]> = {
  WAITING_PAYMENT: [
    { status: 'PAYMENT_VERIFICATION', label: 'Проверить оплату', icon: <PlayArrowIcon />, color: 'primary', variant: 'contained' },
    { status: 'CANCELLED', label: 'Отменить', icon: <CancelIcon />, color: 'error', variant: 'outlined' },
  ],
  PAYMENT_VERIFICATION: [
    { status: 'PAID', label: 'Подтвердить оплату', icon: <CheckCircleIcon />, color: 'success', variant: 'contained' },
    { status: 'CANCELLED', label: 'Отклонить', icon: <CancelIcon />, color: 'error', variant: 'outlined' },
  ],
  PAID: [
    { status: 'COMPLETED', label: 'Завершить заказ', icon: <DoneAllIcon />, color: 'success', variant: 'contained' },
    { status: 'CANCELLED', label: 'Отменить', icon: <CancelIcon />, color: 'error', variant: 'outlined' },
  ],
  REFUND_REQUESTED: [
    { status: 'REFUNDED', label: 'Обработать возврат', icon: <MonetizationOnIcon />, color: 'warning', variant: 'contained' },
    { status: 'PAID', label: 'Отклонить возврат', icon: <CancelIcon />, color: 'error', variant: 'outlined' },
  ],
  WAITING_ACTIVATION: [legacyComplete, legacyCancel],
  AWAITING_CREDENTIALS: [legacyComplete, legacyCancel],
  CREDENTIALS_RECEIVED: [legacyComplete, legacyCancel],
  AWAITING_2FA: [legacyComplete, legacyCancel],
  ACTIVATING: [legacyComplete, legacyCancel],
  ACTIVATED: [legacyComplete, legacyCancel],
  KEY_ISSUED: [legacyComplete, legacyCancel],
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  const handleCopy = () => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }
  return (
    <Button size="small" onClick={handleCopy} sx={{ minWidth: 'auto', p: 0.5 }}>
      <ContentCopyIcon fontSize="small" sx={{ color: copied ? 'success.main' : 'text.secondary' }} />
    </Button>
  )
}

function ActionCard() {
  const record = useRecordContext()
  const refresh = useRefresh()
  const token = localStorage.getItem('token')
  const [cancelOpen, setCancelOpen] = useState(false)
  const [cancelReason, setCancelReason] = useState('')
  const [commentOpen, setCommentOpen] = useState(false)
  const [commentAction, setCommentAction] = useState<any>(null)
  const [commentText, setCommentText] = useState('')
  if (!record) return null

  const order = record.order || record
  const actions = statusActions[order.status] || []
  if (actions.length === 0) return null

  const updateStatus = async (status: string, comment?: string) => {
    try {
      const res = await fetch(`/api/v1/admin/orders/${order.id}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ status, comment: comment || '' }),
      })
      if (!res.ok) {
        const err = await res.json()
        const msg = err.message || err.error || 'Не удалось изменить статус'
        alert('Ошибка: ' + msg)
        return
      }
      refresh()
    } catch (e) {
      alert('Ошибка сети: ' + e)
    }
  }

  const handleAction = (a: typeof statusActions[keyof typeof statusActions][0]) => {
    if (a.status === 'CANCELLED') setCancelOpen(true)
    else updateStatus(a.status, a.label)
  }

  const statusColorMap: Record<string, string> = {
    info: '#2196f3', warning: '#ff9800', success: '#4caf50', error: '#f44336', default: '#9e9e9e',
  }
  const currentColor = statusColorMap[statusColors[order.status] || 'default']

  return (
    <>
      <Paper sx={{ p: 3, mb: 3, border: `2px solid ${currentColor}40`, bgcolor: `${currentColor}08`, borderRadius: 2 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2 }}>
          <Chip
            label={statusLabels[order.status] || order.status}
            color={statusColors[order.status] || 'default'}
            size="medium"
            sx={{ fontWeight: 'bold', fontSize: '0.9rem', px: 1 }}
          />
          <Typography variant="body2" sx={{ color: '#888' }}>Доступные действия:</Typography>
        </Box>
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1.5 }}>
          {actions.map((a) => (
            <MuiButton
              key={a.status}
              variant={a.variant || 'contained'}
              color={a.color}
              size="large"
              startIcon={a.icon}
              onClick={() => handleAction(a)}
              sx={{ fontWeight: 'bold', px: 3, py: 1.2, fontSize: '0.9rem' }}
            >
              {a.label}
            </MuiButton>
          ))}
        </Box>
      </Paper>

      <Dialog open={cancelOpen} onClose={() => setCancelOpen(false)}>
        <DialogTitle sx={{ color: '#000' }}>Причина отмены</DialogTitle>
        <DialogContent>
          <MuiTextField autoFocus fullWidth label="Причина" value={cancelReason} onChange={(e) => setCancelReason(e.target.value)} sx={{ mt: 1 }} InputLabelProps={{ style: { color: '#888' } }} inputProps={{ style: { color: '#000' } }} />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCancelOpen(false)}>Назад</Button>
          <Button color="error" onClick={() => { updateStatus('CANCELLED', cancelReason); setCancelOpen(false) }}>Отменить заказ</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={commentOpen} onClose={() => setCommentOpen(false)}>
        <DialogTitle sx={{ color: '#000' }}>{commentAction?.label || 'Комментарий'}</DialogTitle>
        <DialogContent>
          <MuiTextField autoFocus fullWidth label="Причина / Комментарий" value={commentText} onChange={(e) => setCommentText(e.target.value)} sx={{ mt: 1 }} InputLabelProps={{ style: { color: '#888' } }} inputProps={{ style: { color: '#000' } }} />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCommentOpen(false)}>Назад</Button>
          <Button color="error" onClick={() => { if (commentAction) updateStatus(commentAction.status, commentText); setCommentOpen(false) }}>Подтвердить</Button>
        </DialogActions>
      </Dialog>
    </>
  )
}

function CustomerInfo() {
  const record = useRecordContext()
  const order = record?.order || record
  if (!order?.user) return null

  return (
    <Card sx={{ p: 2, mb: 2 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
        <PersonIcon color="primary" />
        <Typography variant="subtitle1" sx={{ fontWeight: 'bold', color: '#fff' }}>Клиент</Typography>
      </Box>
      <Divider sx={{ mb: 2 }} />
      <Grid container spacing={2}>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Имя</Typography>
          <Typography sx={{ color: '#fff' }}>{order.user.username || '—'}</Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Telegram ID</Typography>
          <Typography sx={{ color: '#fff' }}>{order.user.telegram_id}</Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>First Name</Typography>
          <Typography sx={{ color: '#fff' }}>{order.user.first_name || '—'}</Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Дата регистрации</Typography>
          <Typography sx={{ color: '#fff' }}>{order.user.created_at ? new Date(order.user.created_at).toLocaleDateString('ru-RU') : '—'}</Typography>
        </Grid>
      </Grid>
    </Card>
  )
}

function ProductInfo() {
  const record = useRecordContext()
  const order = record?.order || record
  if (!order?.product) return null

  return (
    <Card sx={{ p: 2, mb: 2 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
        <InventoryIcon color="primary" />
        <Typography variant="subtitle1" sx={{ fontWeight: 'bold', color: '#fff' }}>Товар</Typography>
      </Box>
      <Divider sx={{ mb: 2 }} />
      <Box sx={{ display: 'flex', gap: 2 }}>
        {order.product.image_url && (
          <CardMedia component="img" image={order.product.image_url} sx={{ width: 80, height: 80, borderRadius: 2, objectFit: 'cover' }} />
        )}
        <Box sx={{ flex: 1 }}>
          <Typography variant="body1" sx={{ fontWeight: 'medium', color: '#fff' }}>{order.product.title}</Typography>
          <Typography variant="caption" sx={{ color: '#888' }}>{order.product.platform} • {order.product.type}</Typography>
          {order.variant_id && (
            <Typography variant="caption" sx={{ color: '#888', display: 'block' }}>Вариант: {order.variant_id}</Typography>
          )}
        </Box>
      </Box>
    </Card>
  )
}

function ReceiptDisplay() {
  const record = useRecordContext()
  const order = record?.order || record
  if (!order?.payment_receipt_url) return null

  const isImage = order.payment_receipt_url.match(/\.(png|jpg|jpeg|gif|webp)/i)

  return (
    <Card sx={{ p: 2, mb: 2 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
        <ReceiptIcon color="primary" />
        <Typography variant="subtitle1" sx={{ fontWeight: 'bold', color: '#fff' }}>Чек об оплате</Typography>
      </Box>
      <Divider sx={{ mb: 2 }} />
      {isImage ? (
        <Card sx={{ maxWidth: 500, mx: 'auto' }}>
          <CardMedia component="img" image={order.payment_receipt_url} alt="Receipt" />
        </Card>
      ) : (
        <Box sx={{ textAlign: 'center', py: 2 }}>
          <Typography variant="body2" sx={{ color: '#888' }} gutterBottom>Чек в формате PDF</Typography>
          <MuiButton href={order.payment_receipt_url} target="_blank" variant="outlined" startIcon={<ReceiptIcon />}>
            Скачать чек
          </MuiButton>
        </Box>
      )}
    </Card>
  )
}

function AssignedAdminInfo() {
  const record = useRecordContext()
  const admin = record?.assigned_admin
  if (!admin) return null

  return (
    <Card sx={{ p: 2, mb: 2, bgcolor: '#e3f2fd', border: '1px solid #bbdefb' }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <PersonIcon sx={{ color: 'primary.main' }} />
        <Typography variant="body2" sx={{ color: '#000' }}>
          <b>Назначен:</b> {admin.username || admin.telegram_id}
        </Typography>
      </Box>
    </Card>
  )
}

function StatusHistory() {
  const record = useRecordContext()
  const history = record?.history || []
  if (!history.length) return null

  return (
    <Accordion sx={{ mb: 2, bgcolor: '#1a1a2e', color: '#fff' }}>
      <AccordionSummary expandIcon={<ExpandMoreIcon sx={{ color: '#fff' }} />}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <HistoryIcon color="primary" />
          <Typography variant="subtitle1" sx={{ fontWeight: 'bold', color: '#fff' }}>История изменений</Typography>
          <Chip label={history.length} size="small" color="primary" sx={{ ml: 1 }} />
        </Box>
      </AccordionSummary>
      <AccordionDetails>
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          {history.map((h: any, i: number) => (
            <Box key={h.id || i} sx={{ display: 'flex', gap: 2, alignItems: 'flex-start' }}>
              <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
                <Box sx={{ width: 12, height: 12, borderRadius: '50%', bgcolor: 'primary.main', mt: 0.5 }} />
                {i < history.length - 1 && <Box sx={{ width: 2, flex: 1, bgcolor: '#444', mt: 0.5 }} />}
              </Box>
              <Box sx={{ flex: 1, pb: 2 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, flexWrap: 'wrap' }}>
                  {h.old_status && (
                    <Chip label={statusLabels[h.old_status] || h.old_status} size="small" color={statusColors[h.old_status] || 'default'} variant="outlined" />
                  )}
                  <Typography variant="body2" sx={{ color: '#888' }}>→</Typography>
                  <Chip label={statusLabels[h.new_status] || h.new_status} size="small" color={statusColors[h.new_status] || 'default'} />
                </Box>
                <Typography variant="caption" sx={{ color: '#888', display: 'block', mt: 0.5 }}>
                  {new Date(h.created_at).toLocaleString('ru-RU')}
                  {h.changed_by_type === 'admin' ? ' • Админ' : h.changed_by_type === 'user' ? ' • Пользователь' : ' • Система'}
                </Typography>
                {h.comment && (
                  <Typography variant="body2" sx={{ mt: 0.5, fontStyle: 'italic', color: '#ccc' }}>
                    {h.comment}
                  </Typography>
                )}
              </Box>
            </Box>
          ))}
        </Box>
      </AccordionDetails>
    </Accordion>
  )
}

function RefundAlert() {
  const record = useRecordContext()
  const order = record?.order || record
  if (order?.status !== 'REFUND_REQUESTED') return null

  return (
    <Alert severity="warning" sx={{ mb: 2 }} icon={<WarningIcon />}>
      <Typography variant="body2" sx={{ fontWeight: 'medium', color: '#000' }}>Запрошен возврат средств</Typography>
      <Typography variant="caption" sx={{ color: '#333' }}>Необходимо обработать запрос возврата</Typography>
    </Alert>
  )
}

function OrderDetailFields() {
  const record = useRecordContext()
  const order = record?.order || record
  if (!order) return null

  return (
    <Card sx={{ p: 2, mb: 2 }}>
      <Typography variant="subtitle1" sx={{ fontWeight: 'bold', mb: 2, color: '#fff' }}>Детали заказа</Typography>
      <Divider sx={{ mb: 2 }} />
      <Grid container spacing={2}>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>ID</Typography>
          <Typography variant="body2" sx={{ fontFamily: 'monospace', color: '#fff' }}>{order.id?.substring(0, 8)}...</Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Статус</Typography>
          <Box sx={{ mt: 0.5 }}>
            <Chip label={statusLabels[order.status] || order.status} color={statusColors[order.status] || 'default'} size="small" />
          </Box>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Доставка</Typography>
          <Typography sx={{ color: '#fff' }}>💬 Через менеджера</Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Оплата</Typography>
          <Typography sx={{ color: '#fff' }}>{order.payment_method || '—'}</Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Количество</Typography>
          <Typography sx={{ color: '#fff' }}>{order.quantity || 1}</Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Сумма</Typography>
          <Typography sx={{ fontWeight: 'bold', color: '#4caf50', fontSize: '1.1rem' }}>
            {order.payment_amount ? `${order.payment_amount.toFixed(2)} ₽` : '—'}
          </Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Создан</Typography>
          <Typography sx={{ color: '#fff' }}>{order.created_at ? new Date(order.created_at).toLocaleString('ru-RU') : '—'}</Typography>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#888' }}>Обновлён</Typography>
          <Typography sx={{ color: '#fff' }}>{order.updated_at ? new Date(order.updated_at).toLocaleString('ru-RU') : '—'}</Typography>
        </Grid>
        {order.cancelled_reason && (
          <Grid item xs={12}>
            <Typography variant="caption" sx={{ color: '#888' }}>Причина отмены</Typography>
            <Typography sx={{ color: '#f44336' }}>{order.cancelled_reason}</Typography>
          </Grid>
        )}
      </Grid>
    </Card>
  )
}

function CredentialsDisplay() {
  const record = useRecordContext()
  const account = record?.account
  if (!account) return null

  return (
    <Card sx={{ p: 2, mb: 2, bgcolor: '#fff3e0', border: '1px solid #ffe0b2' }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
        <PersonIcon sx={{ color: 'warning.main' }} />
        <Typography variant="subtitle1" sx={{ fontWeight: 'bold', color: '#000' }}>Данные аккаунта клиента</Typography>
      </Box>
      <Divider sx={{ mb: 2 }} />
      <Grid container spacing={2}>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#666' }}>Логин</Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography sx={{ fontFamily: 'monospace', fontWeight: 'medium', color: '#000', fontSize: '0.95rem' }}>{account.login}</Typography>
            <CopyButton text={account.login} />
          </Box>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#666' }}>Пароль</Typography>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Typography sx={{ fontFamily: 'monospace', fontWeight: 'medium', color: '#000', fontSize: '0.95rem' }}>{account.password}</Typography>
            <CopyButton text={account.password} />
          </Box>
        </Grid>
        <Grid item xs={6}>
          <Typography variant="caption" sx={{ color: '#666' }}>Платформа</Typography>
          <Typography sx={{ color: '#000' }}>{account.platform}</Typography>
        </Grid>
        {account.two_factor_code && (
          <Grid item xs={6}>
            <Typography variant="caption" sx={{ color: '#666' }}>2FA код</Typography>
            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
              <Typography sx={{ fontFamily: 'monospace', fontWeight: 'bold', color: '#d32f2f', fontSize: '1.1rem' }}>{account.two_factor_code}</Typography>
              <CopyButton text={account.two_factor_code} />
            </Box>
          </Grid>
        )}
        {account.notes && (
          <Grid item xs={12}>
            <Typography variant="caption" sx={{ color: '#666' }}>Заметки</Typography>
            <Typography sx={{ color: '#000' }}>{account.notes}</Typography>
          </Grid>
        )}
      </Grid>
    </Card>
  )
}

function OrderShowContent() {
  const record = useRecordContext()
  const order = record?.order || record
  const chatEnabled = order?.status && !['NEW', 'WAITING_PAYMENT', 'PAYMENT_VERIFICATION'].includes(order.status)

  return (
    <>
      <ActionCard />
      <RefundAlert />
      <CustomerInfo />
      <ProductInfo />
      <OrderDetailFields />
      {chatEnabled && <OrderChat />}
      <ReceiptDisplay />
      <StatusHistory />
    </>
  )
}

export const OrderShow = () => (
  <Show>
    <OrderShowContent />
  </Show>
)
