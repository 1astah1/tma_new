import { useState } from 'react'
import { Show, SimpleShowLayout, TextField, NumberField, DateField, SelectField, useRecordContext, Button, TopToolbar } from 'react-admin'
import { Typography, Box, Card, CardMedia, Dialog, DialogTitle, DialogContent, DialogActions, TextField as MuiTextField } from '@mui/material'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import CancelIcon from '@mui/icons-material/Cancel'
import KeyIcon from '@mui/icons-material/Key'
import AssignmentIcon from '@mui/icons-material/Assignment'
import LoginIcon from '@mui/icons-material/Login'
import DoneAllIcon from '@mui/icons-material/DoneAll'
import VisibilityIcon from '@mui/icons-material/Visibility'

const orderStatuses = [
  { id: 'NEW', name: 'Новый' }, { id: 'WAITING_PAYMENT', name: 'Ожидает оплаты' },
  { id: 'PAYMENT_VERIFICATION', name: 'Проверка платежа' }, { id: 'PAID', name: 'Оплачен' },
  { id: 'WAITING_ACTIVATION', name: 'В очереди' }, { id: 'AWAITING_CREDENTIALS', name: 'Требуются данные' },
  { id: 'CREDENTIALS_RECEIVED', name: 'Данные получены' }, { id: 'AWAITING_2FA', name: 'Ожидает код' },
  { id: 'ACTIVATING', name: 'Активация...' }, { id: 'ACTIVATED', name: 'Активирован' },
  { id: 'KEY_ISSUED', name: 'Ключ выдан' }, { id: 'COMPLETED', name: 'Завершён' },
  { id: 'CANCELLED', name: 'Отменён' }, { id: 'REFUNDED', name: 'Возвращён' },
]

const statusActions: Record<string, { status: string; label: string; icon: JSX.Element; color: 'primary' | 'error' | 'warning' }[]> = {
  PAYMENT_VERIFICATION: [
    { status: 'PAID', label: 'Подтвердить оплату', icon: <CheckCircleIcon />, color: 'primary' },
    { status: 'CANCELLED', label: 'Отклонить', icon: <CancelIcon />, color: 'error' },
  ],
  PAID: [
    { status: 'KEY_ISSUED', label: 'Выдать ключ', icon: <KeyIcon />, color: 'primary' },
    { status: 'CANCELLED', label: 'Отменить', icon: <CancelIcon />, color: 'error' },
  ],
  WAITING_ACTIVATION: [
    { status: 'AWAITING_CREDENTIALS', label: 'Взять задачу', icon: <AssignmentIcon />, color: 'primary' },
    { status: 'CANCELLED', label: 'Отменить', icon: <CancelIcon />, color: 'error' },
  ],
  CREDENTIALS_RECEIVED: [
    { status: 'AWAITING_2FA', label: 'Готов войти', icon: <LoginIcon />, color: 'primary' },
    { status: 'CANCELLED', label: 'Отменить', icon: <CancelIcon />, color: 'error' },
  ],
  ACTIVATING: [
    { status: 'ACTIVATED', label: 'Завершить активацию', icon: <DoneAllIcon />, color: 'primary' },
  ],
  ACTIVATED: [
    { status: 'COMPLETED', label: 'Завершить заказ', icon: <DoneAllIcon />, color: 'primary' },
  ],
}

function OrderActionsBar() {
  const record = useRecordContext()
  const token = localStorage.getItem('token')
  const [cancelOpen, setCancelOpen] = useState(false)
  const [cancelReason, setCancelReason] = useState('')
  const [creds, setCreds] = useState<{ login: string; password: string } | null>(null)

  if (!record) return null

  const actions = statusActions[record.status] || []

  const updateStatus = async (status: string, comment?: string) => {
    await fetch(`/api/v1/admin/orders/${record.id}/status`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
      body: JSON.stringify({ status, comment: comment || '' }),
    })
    window.location.reload()
  }

  const decryptCredentials = async () => {
    try {
      const res = await fetch(`/api/v1/admin/orders/${record.id}/decrypt-credentials`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      })
      const data = await res.json()
      setCreds(data)
    } catch { alert('Ошибка расшифровки') }
  }

  return (
    <>
      <TopToolbar>
        {actions.map((a) => (
          <Button key={a.status} color={a.color} onClick={() => {
            if (a.status === 'CANCELLED') setCancelOpen(true)
            else updateStatus(a.status, a.label)
          }}>
            {a.icon} {a.label}
          </Button>
        ))}
        {record.delivery_method === 'activation' && ['AWAITING_CREDENTIALS','CREDENTIALS_RECEIVED','AWAITING_2FA','ACTIVATING'].includes(record.status) && (
          <Button color="warning" onClick={decryptCredentials}><VisibilityIcon /> Расшифровать данные</Button>
        )}
      </TopToolbar>

      {creds && (
        <Box sx={{ p: 2, mx: 2, mb: 2, bgcolor: '#fff3e0', borderRadius: 2, border: '1px solid #ffe0b2' }}>
          <Typography variant="subtitle2" gutterBottom>🔓 Данные аккаунта:</Typography>
          <Typography><b>Логин:</b> {creds.login}</Typography>
          <Typography><b>Пароль:</b> {creds.password}</Typography>
        </Box>
      )}

      <Dialog open={cancelOpen} onClose={() => setCancelOpen(false)}>
        <DialogTitle>Причина отмены</DialogTitle>
        <DialogContent>
          <MuiTextField autoFocus fullWidth label="Причина" value={cancelReason} onChange={(e) => setCancelReason(e.target.value)} sx={{ mt: 1 }} />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCancelOpen(false)}>Назад</Button>
          <Button color="error" onClick={() => { updateStatus('CANCELLED', cancelReason); setCancelOpen(false) }}>Отменить заказ</Button>
        </DialogActions>
      </Dialog>
    </>
  )
}

function ReceiptBlock() {
  const record = useRecordContext()
  if (!record?.payment_receipt_url) return null
  return (
    <Box sx={{ mt: 2, mx: 2 }}>
      <Typography variant="h6" gutterBottom>🧾 Чек об оплате</Typography>
      {record.payment_receipt_url.match(/\.(png|jpg|jpeg|gif|webp)/i) ? (
        <Card sx={{ maxWidth: 400 }}><CardMedia component="img" image={record.payment_receipt_url} alt="Receipt" /></Card>
      ) : (
        <a href={record.payment_receipt_url} target="_blank" rel="noreferrer">Скачать чек (PDF)</a>
      )}
    </Box>
  )
}

export const OrderShow = () => (
  <Show actions={<OrderActionsBar />}>
    <>
      <SimpleShowLayout>
        <TextField source="id" label="ID" />
        <SelectField source="status" label="Статус" choices={orderStatuses} />
        <TextField source="delivery_method" label="Метод доставки" />
        <TextField source="payment_method" label="Метод оплаты" />
        <NumberField source="payment_amount" label="Сумма" options={{ style: 'currency', currency: 'RUB' }} />
        <DateField source="created_at" label="Создан" showTime />
        <DateField source="updated_at" label="Обновлён" showTime />
        <TextField source="cancelled_reason" label="Причина отмены" />
      </SimpleShowLayout>
      <ReceiptBlock />
    </>
  </Show>
)
