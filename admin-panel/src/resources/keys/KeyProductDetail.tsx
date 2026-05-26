import { useState, useEffect } from 'react'
import { useNotify, useRedirect } from 'react-admin'
import { useParams } from 'react-router-dom'
import {
  Card, CardContent, Typography, Box, Chip, Button, TextField,
  Dialog, DialogTitle, DialogContent, DialogActions, Table,
  TableBody, TableCell, TableHead, TableRow, TablePagination,
  CircularProgress, Paper, Divider, Select, MenuItem, InputLabel, FormControl, Grid
} from '@mui/material'
import ArrowBackIcon from '@mui/icons-material/ArrowBack'
import KeyIcon from '@mui/icons-material/Key'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import WarningIcon from '@mui/icons-material/Warning'
import ErrorIcon from '@mui/icons-material/Error'
import CloudUploadIcon from '@mui/icons-material/CloudUpload'
import DeleteIcon from '@mui/icons-material/Delete'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'
import EditIcon from '@mui/icons-material/Edit'
import LockOpenIcon from '@mui/icons-material/LockOpen'

export function KeyProductDetail() {
  const { id } = useParams<{ id: string }>()
  const notify = useNotify()
  const redirect = useRedirect()
  const token = localStorage.getItem('token')

  const [product, setProduct] = useState<any>(null)
  const [keys, setKeys] = useState<any[]>([])
  const [stats, setStats] = useState<any>({ available: 0, sold: 0, reserved: 0, invalid: 0, total: 0 })
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(0)
  const [perPage, setPerPage] = useState(25)
  const [total, setTotal] = useState(0)
  const [statusFilter, setStatusFilter] = useState('')

  const [importOpen, setImportOpen] = useState(false)
  const [importText, setImportText] = useState('')
  const [importing, setImporting] = useState(false)

  const [editOpen, setEditOpen] = useState(false)
  const [editKey, setEditKey] = useState<any>(null)
  const [editValue, setEditValue] = useState('')

  const loadData = async () => {
    if (!id) return
    setLoading(true)
    try {
      const res = await fetch(`/api/v1/admin/products/${id}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (res.ok) setProduct(await res.json())

      const params = new URLSearchParams({ page: String(page + 1), limit: String(perPage) })
      if (statusFilter) params.set('status', statusFilter)

      const keysRes = await fetch(`/api/v1/admin/products/${id}/keys?${params}`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (keysRes.ok) {
        const data = await keysRes.json()
        setKeys(data.data || [])
        setTotal(data.meta?.total || 0)
      }

      const statsRes = await fetch(`/api/v1/admin/products/${id}/keys/stats`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (statsRes.ok) setStats(await statsRes.json())
    } catch {
      notify('Ошибка загрузки', { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { loadData() }, [id, page, perPage, statusFilter])

  const handleImport = async () => {
    if (!importText.trim()) {
      notify('Введите ключи', { type: 'warning' })
      return
    }
    setImporting(true)
    const keyList = importText.split('\n').map(k => k.trim()).filter(Boolean)
    try {
      const res = await fetch('/api/v1/admin/keys/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ product_id: id, keys: keyList }),
      })
      const data = await res.json()
      notify(`Импортировано: ${data.imported}/${data.total}`, { type: 'success' })
      setImportOpen(false)
      setImportText('')
      loadData()
    } catch {
      notify('Ошибка импорта', { type: 'error' })
    } finally {
      setImporting(false)
    }
  }

  const handleDelete = async (keyId: string) => {
    try {
      await fetch(`/api/v1/admin/keys/${keyId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
      })
      notify('Ключ удалён', { type: 'success' })
      loadData()
    } catch {
      notify('Ошибка удаления', { type: 'error' })
    }
  }

  const handleRelease = async (keyId: string) => {
    try {
      await fetch(`/api/v1/admin/keys/${keyId}/release`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      })
      notify('Ключ освобождён', { type: 'success' })
      loadData()
    } catch {
      notify('Ошибка', { type: 'error' })
    }
  }

  const handleStatusChange = async (keyId: string, status: string) => {
    try {
      await fetch(`/api/v1/admin/keys/${keyId}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ status }),
      })
      notify('Статус обновлён', { type: 'success' })
      loadData()
    } catch {
      notify('Ошибка', { type: 'error' })
    }
  }

  const handleEdit = async () => {
    if (!editKey || !editValue.trim()) return
    try {
      await fetch(`/api/v1/admin/keys/${editKey.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ key: editValue.trim() }),
      })
      notify('Ключ обновлён', { type: 'success' })
      setEditOpen(false)
      loadData()
    } catch {
      notify('Ошибка', { type: 'error' })
    }
  }

  const copyKey = (text: string) => {
    navigator.clipboard.writeText(text)
    notify('Скопировано', { type: 'success' })
  }

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  const statusColors: Record<string, 'success' | 'warning' | 'error' | 'default'> = {
    available: 'success', sold: 'default', reserved: 'warning', invalid: 'error',
  }
  const statusLabels: Record<string, string> = {
    available: 'Доступен', sold: 'Продан', reserved: 'Резерв', invalid: 'Брак',
  }

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
        <Button startIcon={<ArrowBackIcon />} onClick={() => redirect('/keys')}>
          Назад
        </Button>
        <Typography variant="h5" sx={{ fontWeight: 'bold', color: '#fff' }}>
          Ключи: {product?.title}
        </Typography>
      </Box>

      {/* Stats */}
      <Grid container spacing={2} sx={{ mb: 3 }}>
        <Grid item xs={6} sm={2.4}>
          <Card sx={{ bgcolor: '#4caf5015', border: '1px solid #4caf5030' }}>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <CheckCircleIcon sx={{ color: '#4caf50', fontSize: 32 }} />
              <Typography variant="h4" sx={{ color: '#fff', fontWeight: 'bold' }}>{stats.available}</Typography>
              <Typography variant="caption" sx={{ color: '#888' }}>Доступно</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6} sm={2.4}>
          <Card sx={{ bgcolor: '#2196f315', border: '1px solid #2196f330' }}>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <KeyIcon sx={{ color: '#2196f3', fontSize: 32 }} />
              <Typography variant="h4" sx={{ color: '#fff', fontWeight: 'bold' }}>{stats.sold}</Typography>
              <Typography variant="caption" sx={{ color: '#888' }}>Продано</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6} sm={2.4}>
          <Card sx={{ bgcolor: '#ff980015', border: '1px solid #ff980030' }}>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <WarningIcon sx={{ color: '#ff9800', fontSize: 32 }} />
              <Typography variant="h4" sx={{ color: '#fff', fontWeight: 'bold' }}>{stats.reserved}</Typography>
              <Typography variant="caption" sx={{ color: '#888' }}>Резерв</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6} sm={2.4}>
          <Card sx={{ bgcolor: '#f4433615', border: '1px solid #f4433630' }}>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <ErrorIcon sx={{ color: '#f44336', fontSize: 32 }} />
              <Typography variant="h4" sx={{ color: '#fff', fontWeight: 'bold' }}>{stats.invalid}</Typography>
              <Typography variant="caption" sx={{ color: '#888' }}>Брак</Typography>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={6} sm={2.4}>
          <Card sx={{ bgcolor: '#9c27b015', border: '1px solid #9c27b030' }}>
            <CardContent sx={{ textAlign: 'center', py: 2 }}>
              <KeyIcon sx={{ color: '#9c27b0', fontSize: 32 }} />
              <Typography variant="h4" sx={{ color: '#fff', fontWeight: 'bold' }}>{stats.total}</Typography>
              <Typography variant="caption" sx={{ color: '#888' }}>Всего</Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Actions */}
      <Box sx={{ display: 'flex', gap: 2, mb: 2, alignItems: 'center' }}>
        <Button variant="contained" startIcon={<CloudUploadIcon />} onClick={() => setImportOpen(true)}>
          Импорт ключей
        </Button>
        <FormControl size="small" sx={{ minWidth: 150 }}>
          <InputLabel sx={{ color: '#888' }}>Статус</InputLabel>
          <Select
            value={statusFilter}
            label="Статус"
            onChange={(e) => setStatusFilter(e.target.value)}
            sx={{ color: '#fff', '& .MuiOutlinedInput-notchedOutline': { borderColor: '#444' } }}
          >
            <MenuItem value="">Все</MenuItem>
            <MenuItem value="available">Доступно</MenuItem>
            <MenuItem value="sold">Продано</MenuItem>
            <MenuItem value="reserved">Резерв</MenuItem>
            <MenuItem value="invalid">Брак</MenuItem>
          </Select>
        </FormControl>
      </Box>

      {/* Keys Table */}
      <Paper sx={{ overflow: 'auto' }}>
        <Table size="small">
          <TableHead>
            <TableRow>
              <TableCell sx={{ color: '#888' }}>Ключ</TableCell>
              <TableCell sx={{ color: '#888' }}>Статус</TableCell>
              <TableCell sx={{ color: '#888' }}>Заказ</TableCell>
              <TableCell sx={{ color: '#888' }}>Дата</TableCell>
              <TableCell sx={{ color: '#888' }}>Действия</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {keys.map((key) => (
              <TableRow key={key.id} sx={{ '&:hover': { bgcolor: '#1a1a2e' } }}>
                <TableCell>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace', color: '#fff', maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {key.key || '••••••••'}
                    </Typography>
                    {key.key && (
                      <Button size="small" onClick={() => copyKey(key.key)} sx={{ minWidth: 'auto', p: 0.5 }}>
                        <ContentCopyIcon fontSize="small" />
                      </Button>
                    )}
                  </Box>
                </TableCell>
                <TableCell>
                  <Chip
                    label={statusLabels[key.status] || key.status}
                    size="small"
                    color={statusColors[key.status] || 'default'}
                  />
                </TableCell>
                <TableCell>
                  {key.order_id ? (
                    <Typography variant="caption" sx={{ fontFamily: 'monospace', color: '#888' }}>
                      {key.order_id.substring(0, 8)}...
                    </Typography>
                  ) : '—'}
                </TableCell>
                <TableCell>
                  <Typography variant="caption" sx={{ color: '#888' }}>
                    {new Date(key.created_at).toLocaleDateString('ru-RU')}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Box sx={{ display: 'flex', gap: 0.5 }}>
                    {key.status === 'sold' && (
                      <Button size="small" onClick={() => handleRelease(key.id)} title="Освободить">
                        <LockOpenIcon fontSize="small" />
                      </Button>
                    )}
                    <Button size="small" onClick={() => { setEditKey(key); setEditValue(key.key || ''); setEditOpen(true) }} title="Редактировать">
                      <EditIcon fontSize="small" />
                    </Button>
                    <Button size="small" onClick={() => handleDelete(key.id)} title="Удалить" color="error">
                      <DeleteIcon fontSize="small" />
                    </Button>
                  </Box>
                </TableCell>
              </TableRow>
            ))}
            {keys.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} sx={{ textAlign: 'center', color: '#888', py: 4 }}>
                  Ключи не найдены
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        <TablePagination
          component="div"
          count={total}
          page={page}
          onPageChange={(_, p) => setPage(p)}
          rowsPerPage={perPage}
          onRowsPerPageChange={(e) => { setPerPage(Number(e.target.value)); setPage(0) }}
          labelRowsPerPage="Строк:"
          sx={{ color: '#888' }}
        />
      </Paper>

      {/* Import Dialog */}
      <Dialog open={importOpen} onClose={() => setImportOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle sx={{ color: '#000' }}>Импорт ключей</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            multiline
            rows={10}
            label="Ключи (по одному на строку)"
            value={importText}
            onChange={(e) => setImportText(e.target.value)}
            sx={{ mt: 1 }}
            InputLabelProps={{ style: { color: '#888' } }}
            inputProps={{ style: { color: '#000' } }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setImportOpen(false)}>Отмена</Button>
          <Button onClick={handleImport} variant="contained" disabled={importing}>
            {importing ? 'Импорт...' : 'Импорт'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onClose={() => setEditOpen(false)}>
        <DialogTitle sx={{ color: '#000' }}>Редактировать ключ</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            label="Ключ"
            value={editValue}
            onChange={(e) => setEditValue(e.target.value)}
            sx={{ mt: 1 }}
            InputLabelProps={{ style: { color: '#888' } }}
            inputProps={{ style: { color: '#000' } }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setEditOpen(false)}>Отмена</Button>
          <Button onClick={handleEdit} variant="contained">Сохранить</Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
