import { useState, useEffect } from 'react'
import { List, useDataProvider, useNotify, useRefresh } from 'react-admin'
import { Card, CardContent, Typography, Box, Grid, Chip, CircularProgress, Paper } from '@mui/material'
import KeyIcon from '@mui/icons-material/Key'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import WarningIcon from '@mui/icons-material/Warning'
import ErrorIcon from '@mui/icons-material/Error'
import InventoryIcon from '@mui/icons-material/Inventory'

function KeyProductListContent() {
  const dataProvider = useDataProvider()
  const notify = useNotify()
  const [products, setProducts] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState<Record<string, any>>({})

  useEffect(() => {
    const load = async () => {
      try {
        const { data } = await dataProvider.getList('products', {
          pagination: { page: 1, perPage: 100 },
          sort: { field: 'created_at', order: 'DESC' },
          filter: {},
        })
        setProducts(data)

        const s: Record<string, any> = {}
        for (const p of data) {
          try {
            const res = await fetch(`/api/v1/admin/products/${p.id}/keys/stats`, {
              headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
            })
            if (res.ok) {
              s[p.id] = await res.json()
            }
          } catch {}
        }
        setStats(s)
      } catch {
        notify('Ошибка загрузки', { type: 'error' })
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', py: 8 }}>
        <CircularProgress />
      </Box>
    )
  }

  return (
    <Grid container spacing={2}>
      {products.map((product) => {
        const st = stats[product.id] || { available: 0, sold: 0, reserved: 0, invalid: 0, total: 0 }
        const hasKeys = product.delivery_methods?.includes('key')
        const noKeys = hasKeys && st.available === 0

        return (
          <Grid item xs={12} sm={6} md={4} key={product.id}>
            <Card
              sx={{
                cursor: hasKeys ? 'pointer' : 'default',
                border: noKeys ? '2px solid #f44336' : '1px solid #333',
                opacity: noKeys ? 0.7 : 1,
                '&:hover': hasKeys ? { borderColor: 'primary.main', bgcolor: '#1a1a2e' } : {},
              }}
              onClick={() => {
                if (hasKeys) {
                  window.location.hash = `#/keys/${product.id}`
                }
              }}
            >
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                  <InventoryIcon color="primary" />
                  <Typography variant="subtitle1" sx={{ fontWeight: 'bold', color: '#fff', flex: 1 }}>
                    {product.title}
                  </Typography>
                  <Chip
                    label={product.platform?.toUpperCase()}
                    size="small"
                    color="default"
                    sx={{ fontSize: '0.65rem' }}
                  />
                </Box>

                {hasKeys ? (
                  <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Typography variant="body2" sx={{ color: '#888' }}>Всего ключей:</Typography>
                      <Typography variant="body2" sx={{ color: '#fff', fontWeight: 'bold' }}>{st.total}</Typography>
                    </Box>
                    <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                      <Chip
                        icon={<CheckCircleIcon />}
                        label={`Доступно: ${st.available}`}
                        size="small"
                        color="success"
                        variant="outlined"
                      />
                      <Chip
                        label={`Продано: ${st.sold}`}
                        size="small"
                        color="default"
                        variant="outlined"
                      />
                      {st.reserved > 0 && (
                        <Chip
                          icon={<WarningIcon />}
                          label={`Резерв: ${st.reserved}`}
                          size="small"
                          color="warning"
                          variant="outlined"
                        />
                      )}
                      {st.invalid > 0 && (
                        <Chip
                          icon={<ErrorIcon />}
                          label={`Брак: ${st.invalid}`}
                          size="small"
                          color="error"
                          variant="outlined"
                        />
                      )}
                    </Box>
                    {noKeys && (
                      <Typography variant="caption" sx={{ color: '#f44336', fontWeight: 'bold' }}>
                        ⚠ Нет доступных ключей
                      </Typography>
                    )}
                  </Box>
                ) : (
                  <Typography variant="body2" sx={{ color: '#888' }}>
                    Ключи не требуются (активация)
                  </Typography>
                )}
              </CardContent>
            </Card>
          </Grid>
        )
      })}
    </Grid>
  )
}

export const KeyProductList = () => (
  <List pagination={false} actions={false} sort={{ field: 'created_at', order: 'DESC' }}>
    <KeyProductListContent />
  </List>
)
