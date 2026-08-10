import {
  Alert,
  Box,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Stack,
  TextField,
  Typography,
  Button as MuiButton,
} from '@mui/material'
import SearchIcon from '@mui/icons-material/Search'
import type { ProductRow } from './types'

type ProductPoolProps = {
  title: string
  hint?: string
  selectedIds: string[]
  productMap: Record<string, ProductRow>
  products: ProductRow[]
  search: string
  onSearchChange: (value: string) => void
  onSearch: () => void
  loading: boolean
  error: string | null
  onAdd: (productId: string) => void
  onRemove: (productId: string) => void
}

export function ProductPool({
  title,
  hint,
  selectedIds,
  productMap,
  products,
  search,
  onSearchChange,
  onSearch,
  loading,
  error,
  onAdd,
  onRemove,
}: ProductPoolProps) {
  const pool = products.filter((p) => !selectedIds.includes(p.id))

  return (
    <Stack spacing={2}>
      <Typography fontWeight={600}>{title}</Typography>
      {hint ? <Alert severity="info">{hint}</Alert> : null}

      <Typography variant="subtitle2">Выбрано: {selectedIds.length}</Typography>
      <Stack direction="row" spacing={1} sx={{ flexWrap: 'wrap', gap: 1 }}>
        {selectedIds.map((pid) => {
          const p = productMap[pid]
          return (
            <Chip
              key={pid}
              label={p?.title || `${pid.slice(0, 8)}…`}
              onDelete={() => onRemove(pid)}
            />
          )
        })}
        {!selectedIds.length ? (
          <Typography variant="caption" color="text.secondary">
            Пока ничего не выбрано — добавьте из списка ниже
          </Typography>
        ) : null}
      </Stack>

      <Stack direction="row" spacing={1}>
        <TextField
          label="Поиск по названию"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          size="small"
          fullWidth
          placeholder="FIFA, Call of Duty, GTA..."
          InputProps={{ startAdornment: <SearchIcon fontSize="small" sx={{ mr: 1, color: 'text.secondary' }} /> }}
          onKeyDown={(e) => { if (e.key === 'Enter') onSearch() }}
        />
        <MuiButton variant="outlined" onClick={onSearch} disabled={loading}>
          Найти
        </MuiButton>
      </Stack>

      {error ? <Alert severity="error">{error}</Alert> : null}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', py: 3 }}>
          <CircularProgress size={28} />
        </Box>
      ) : (
        <Stack spacing={1} sx={{ maxHeight: 360, overflow: 'auto', pr: 0.5 }}>
          {pool.map((product) => (
            <Card key={product.id} variant="outlined">
              <CardContent sx={{ display: 'flex', alignItems: 'center', gap: 1.5, py: '8px !important' }}>
                {product.image_url ? (
                  <Box component="img" src={product.image_url} alt="" sx={{ width: 32, height: 44, objectFit: 'contain', flexShrink: 0 }} />
                ) : (
                  <Box sx={{ width: 32, height: 44, bgcolor: 'action.hover', borderRadius: 0.5, flexShrink: 0 }} />
                )}
                <Box sx={{ flex: 1, minWidth: 0 }}>
                  <Typography fontWeight={600} noWrap fontSize={14}>{product.title}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {product.platform?.toUpperCase() || '—'} · {Math.round(product.price || 0)} ₽
                  </Typography>
                </Box>
                <MuiButton size="small" variant="contained" onClick={() => onAdd(product.id)}>
                  Добавить
                </MuiButton>
              </CardContent>
            </Card>
          ))}
          {!pool.length ? (
            <Typography color="text.secondary" sx={{ py: 2, textAlign: 'center' }}>
              {search.trim() ? 'Ничего не найдено' : 'Загрузка каталога... нажмите «Найти»'}
            </Typography>
          ) : null}
        </Stack>
      )}
    </Stack>
  )
}
