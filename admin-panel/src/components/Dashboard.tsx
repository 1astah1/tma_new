import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Card,
  CardContent,
  Typography,
  Grid,
  Box,
  Paper,
  LinearProgress,
  Button,
  Stack,
  Chip,
} from '@mui/material'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart'
import AttachMoneyIcon from '@mui/icons-material/AttachMoney'
import PeopleIcon from '@mui/icons-material/People'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import WarningIcon from '@mui/icons-material/Warning'
import PendingIcon from '@mui/icons-material/Pending'
import SpeedIcon from '@mui/icons-material/Speed'
import CloudDownloadIcon from '@mui/icons-material/CloudDownload'
import SyncIcon from '@mui/icons-material/Sync'
import ChevronRightIcon from '@mui/icons-material/ChevronRight'

const COLORS = ['#38bdf8', '#4ade80', '#fb923c', '#a78bfa', '#f59e0b', '#ef4444']

const StatCard = ({ title, value, icon, color, subtitle, progress }: any) => (
  <Card sx={{ height: '100%', background: `linear-gradient(135deg, ${color}15 0%, ${color}05 100%)`, border: `1px solid ${color}30` }}>
    <CardContent>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography color="textSecondary" variant="body2">{title}</Typography>
        <Box sx={{ color, fontSize: 28 }}>{icon}</Box>
      </Box>
      <Typography variant="h4" sx={{ fontWeight: 'bold', mb: 1 }}>{value}</Typography>
      {subtitle && <Typography variant="caption" color="textSecondary">{subtitle}</Typography>}
      {progress !== undefined && (
        <Box sx={{ mt: 1.5 }}>
          <LinearProgress variant="determinate" value={Math.min(progress, 100)} sx={{ height: 6, borderRadius: 3, bgcolor: `${color}20`, '& .MuiLinearProgress-bar': { bgcolor: color } }} />
        </Box>
      )}
    </CardContent>
  </Card>
)

const AttentionCard = ({
  title,
  count,
  subtitle,
  color,
  onClick,
}: {
  title: string
  count: number
  subtitle: string
  color: string
  onClick?: () => void
}) => (
  <Card
    onClick={count > 0 ? onClick : undefined}
    sx={{
      height: '100%',
      cursor: count > 0 ? 'pointer' : 'default',
      border: `1px solid ${color}40`,
      bgcolor: count > 0 ? `${color}10` : 'transparent',
      transition: 'transform 0.15s',
      '&:hover': count > 0 ? { transform: 'translateY(-2px)', borderColor: color } : {},
    }}
  >
    <CardContent sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', py: '16px !important' }}>
      <Box>
        <Typography variant="body2" color="text.secondary">{title}</Typography>
        <Typography variant="h4" sx={{ fontWeight: 'bold', color }}>{count}</Typography>
        <Typography variant="caption" color="text.secondary">{subtitle}</Typography>
      </Box>
      {count > 0 ? <ChevronRightIcon sx={{ color }} /> : null}
    </CardContent>
  </Card>
)

export const Dashboard = () => {
  const navigate = useNavigate()
  const [stats, setStats] = useState<any>({})
  const [refreshing, setRefreshing] = useState(false)

  const loadStats = () => {
    fetch('/api/v1/admin/dashboard', {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
      .then((r) => r.json())
      .then(setStats)
      .catch(() => {})
  }

  useEffect(() => {
    loadStats()
  }, [])

  const goList = (resource: string, filter: Record<string, string>) => {
    navigate(`/${resource}?filter=${encodeURIComponent(JSON.stringify(filter))}`)
  }

  const refreshCatalog = async () => {
    setRefreshing(true)
    try {
      await fetch('/api/v1/admin/catalog-imports/refresh-catalog?sync=1', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${localStorage.getItem('token') || ''}`,
        },
        body: '{}',
      })
      loadStats()
    } finally {
      setRefreshing(false)
    }
  }

  const fmt = (n: number) => n.toLocaleString('ru-RU', { minimumFractionDigits: 0, maximumFractionDigits: 0 })
  const fmtMoney = (n: number) => n.toLocaleString('ru-RU', { minimumFractionDigits: 0, maximumFractionDigits: 0 }) + ' ₽'

  const catalog = stats.catalog || {}
  const attentionTotal =
    (stats.pending_orders || 0) +
    (stats.paid_orders || 0) +
    (stats.waiting_payment || 0) +
    (stats.pending_imports || catalog.pending_imports || 0)

  const statusData = [
    { name: 'Новые', value: stats.orders_today || 0, color: '#38bdf8' },
    { name: 'Ожидание оплаты', value: stats.waiting_payment || 0, color: '#fb923c' },
    { name: 'Проверка платежа', value: stats.pending_orders || 0, color: '#f59e0b' },
    { name: 'Оплачены (в работе)', value: stats.paid_orders || stats.active_tasks || 0, color: '#a78bfa' },
    { name: 'Завершённые', value: stats.completed_orders || 0, color: '#4ade80' },
  ].filter((d) => d.value > 0)

  const revenueData = [
    { name: 'Сегодня', revenue: stats.revenue_today || 0 },
    { name: 'Неделя', revenue: stats.revenue_week || 0 },
    { name: 'Месяц', revenue: stats.revenue_month || 0 },
    { name: 'Всего', revenue: stats.total_revenue || 0 },
  ]

  return (
    <div style={{ padding: 20 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 2, mb: 3, flexWrap: 'wrap' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <img src="/favicon.png" alt="COIN MINT" style={{ width: 'auto', height: 48 }} />
          <Typography variant="h4" sx={{ fontWeight: 'bold' }}>COIN MINT</Typography>
        </Box>
        <Stack direction="row" spacing={1}>
          <Button
            variant="outlined"
            size="small"
            startIcon={<CloudDownloadIcon />}
            onClick={() => navigate('/catalog-imports')}
          >
            Мастер каталога
          </Button>
          <Button
            variant="contained"
            size="small"
            startIcon={<SyncIcon />}
            onClick={refreshCatalog}
            disabled={refreshing}
          >
            {refreshing ? 'Обновляем...' : 'Обновить цены'}
          </Button>
        </Stack>
      </Box>

      <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 2 }}>
        Требует внимания {attentionTotal > 0 ? `(${attentionTotal})` : ''}
      </Typography>
      <Grid container spacing={2} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <AttentionCard
            title="Проверка оплаты"
            count={stats.pending_orders || 0}
            subtitle="Подтвердить чек"
            color="#f59e0b"
            onClick={() => goList('orders', { status: 'PAYMENT_VERIFICATION' })}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <AttentionCard
            title="Оплачены"
            count={stats.paid_orders || 0}
            subtitle="Ведут чат с клиентом"
            color="#a78bfa"
            onClick={() => goList('orders', { status: 'PAID' })}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <AttentionCard
            title="Ждут оплаты"
            count={stats.waiting_payment || 0}
            subtitle="Клиент не оплатил"
            color="#fb923c"
            onClick={() => goList('orders', { status: 'WAITING_PAYMENT' })}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <AttentionCard
            title="Импорт без товара"
            count={stats.pending_imports || catalog.pending_imports || 0}
            subtitle="Нужна публикация"
            color="#38bdf8"
            onClick={() => goList('catalog-imports', { status: 'pending' })}
          />
        </Grid>
      </Grid>

      {catalog.game_products != null ? (
        <Stack direction="row" spacing={1} sx={{ mb: 3 }} flexWrap="wrap" useFlexGap>
          <Chip label={`Игр в каталоге: ${catalog.game_products}`} size="small" />
          <Chip label={`Активных: ${catalog.active_game_products || 0}`} size="small" color="success" variant="outlined" />
          {(catalog.inactive_game_products || 0) > 0 ? (
            <Chip
              label={`Скрытых: ${catalog.inactive_game_products}`}
              size="small"
              onClick={() => goList('products', { type: 'game', status: 'inactive' })}
              sx={{ cursor: 'pointer' }}
            />
          ) : null}
        </Stack>
      ) : null}

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Заказов сегодня"
            value={fmt(stats.orders_today || 0)}
            icon={<ShoppingCartIcon />}
            color="#38bdf8"
            subtitle={`Всего: ${fmt(stats.total_orders || 0)}`}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Выручка сегодня"
            value={fmtMoney(stats.revenue_today || 0)}
            icon={<AttachMoneyIcon />}
            color="#4ade80"
            subtitle={`Всего: ${fmtMoney(stats.total_revenue || 0)}`}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Ожидающая выручка"
            value={fmtMoney(stats.pending_revenue || 0)}
            icon={<PendingIcon />}
            color="#fb923c"
            subtitle="Оплаченные + на проверке"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Средний чек"
            value={fmtMoney(stats.avg_check || 0)}
            icon={<SpeedIcon />}
            color="#a78bfa"
            subtitle={`Конверсия: ${(stats.conversion_rate || 0).toFixed(1)}%`}
            progress={stats.conversion_rate || 0}
          />
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Оплачены (в работе)"
            value={fmt(stats.paid_orders || stats.active_tasks || 0)}
            icon={<CheckCircleIcon />}
            color="#a78bfa"
            subtitle={`Ожидают оплаты: ${fmt(stats.waiting_payment || 0)}`}
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Проверка платежей"
            value={fmt(stats.pending_orders || 0)}
            icon={<WarningIcon />}
            color="#f59e0b"
            subtitle="Требуют подтверждения"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Завершённые заказы"
            value={fmt(stats.completed_orders || 0)}
            icon={<CheckCircleIcon />}
            color="#4ade80"
            subtitle="Успешно обработаны"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Пользователей"
            value={fmt(stats.total_users || 0)}
            icon={<PeopleIcon />}
            color="#38bdf8"
            subtitle="Всего зарегистрировано"
          />
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2, fontWeight: 'bold' }}>Выручка по периодам</Typography>
              <ResponsiveContainer width="100%" height={250}>
                <BarChart data={revenueData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                  <XAxis dataKey="name" stroke="#888" />
                  <YAxis stroke="#888" tickFormatter={(v) => `${fmt(v)} ₽`} />
                  <Tooltip formatter={(value) => `${fmtMoney(Number(value || 0))}`} />
                  <Bar dataKey="revenue" fill="#4ade80" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2, fontWeight: 'bold' }}>Распределение заказов</Typography>
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie
                    data={statusData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percent }) => `${name}: ${((percent || 0) * 100).toFixed(0)}%`}
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {statusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2, fontWeight: 'bold' }}>Топ товаров</Typography>
              {stats.top_products?.length > 0 ? (
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                  {stats.top_products.slice(0, 8).map((row: any) => (
                    <Paper
                      key={row.product_id}
                      sx={{ p: 1.5, display: 'flex', justifyContent: 'space-between', cursor: 'pointer' }}
                      onClick={() => navigate(`/products/${row.product_id}`)}
                    >
                      <Box>
                        <Typography variant="body2" sx={{ fontWeight: 'medium' }}>{row.title}</Typography>
                        <Typography variant="caption" color="textSecondary">{row.platform} · {row.orders_count} заказов</Typography>
                      </Box>
                      <Typography variant="body2" sx={{ fontWeight: 'bold', color: '#4caf50' }}>{fmtMoney(row.revenue || 0)}</Typography>
                    </Paper>
                  ))}
                </Box>
              ) : (
                <Typography color="textSecondary">Нет данных</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2, fontWeight: 'bold' }}>Платформы и промо</Typography>
              <Box sx={{ mb: 2 }}>
                <Typography variant="caption" color="textSecondary">Использований промокодов</Typography>
                <Typography variant="h5" sx={{ fontWeight: 'bold' }}>{fmt(stats.promo_usage_total || 0)}</Typography>
              </Box>
              {stats.platform_stats?.length > 0 ? (
                <ResponsiveContainer width="100%" height={180}>
                  <BarChart data={stats.platform_stats.map((p: any) => ({ name: p.platform, revenue: p.revenue }))}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                    <XAxis dataKey="name" stroke="#888" />
                    <YAxis stroke="#888" tickFormatter={(v) => `${fmt(v)} ₽`} />
                    <Tooltip formatter={(value) => fmtMoney(Number(value || 0))} />
                    <Bar dataKey="revenue" fill="#38bdf8" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <Typography color="textSecondary">Нет данных по платформам</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2, fontWeight: 'bold' }}>Последние заказы</Typography>
              {stats.recent_orders?.length > 0 ? (
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                  {stats.recent_orders.map((order: any) => (
                    <Paper
                      key={order.id}
                      sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer' }}
                      onClick={() => navigate(`/orders/${order.id}/show`)}
                    >
                      <Box>
                        <Typography variant="body2" sx={{ fontWeight: 'medium' }}>{order.product_title || 'Товар'}</Typography>
                        <Typography variant="caption" color="textSecondary">#{order.id?.substring(0, 8)}</Typography>
                      </Box>
                      <Box sx={{ textAlign: 'right' }}>
                        <Typography variant="body2" sx={{ fontWeight: 'bold', color: '#4caf50' }}>
                          {order.payment_amount ? `${Number(order.payment_amount).toFixed(2)} ₽` : '—'}
                        </Typography>
                        <Typography variant="caption" color="textSecondary">{order.status}</Typography>
                      </Box>
                    </Paper>
                  ))}
                </Box>
              ) : (
                <Typography color="textSecondary">Нет заказов</Typography>
              )}
            </CardContent>
          </Card>
        </Grid>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2, fontWeight: 'bold' }}>Экономика</Typography>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                    <AttachMoneyIcon sx={{ color: '#4ade80' }} />
                    <Box>
                      <Typography variant="h6">{fmtMoney(stats.revenue_week || 0)}</Typography>
                      <Typography variant="caption" color="textSecondary">Выручка за неделю</Typography>
                    </Box>
                  </Box>
                </Grid>
                <Grid item xs={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                    <AttachMoneyIcon sx={{ color: '#38bdf8' }} />
                    <Box>
                      <Typography variant="h6">{fmtMoney(stats.revenue_month || 0)}</Typography>
                      <Typography variant="caption" color="textSecondary">Выручка за месяц</Typography>
                    </Box>
                  </Box>
                </Grid>
                <Grid item xs={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                    <CheckCircleIcon sx={{ color: '#4ade80' }} />
                    <Box>
                      <Typography variant="h6">{(stats.conversion_rate || 0).toFixed(1)}%</Typography>
                      <Typography variant="caption" color="textSecondary">Конверсия</Typography>
                    </Box>
                  </Box>
                </Grid>
                <Grid item xs={6}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                    <SpeedIcon sx={{ color: '#a78bfa' }} />
                    <Box>
                      <Typography variant="h6">{fmtMoney(stats.avg_check || 0)}</Typography>
                      <Typography variant="caption" color="textSecondary">Средний чек</Typography>
                    </Box>
                  </Box>
                </Grid>
              </Grid>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </div>
  )
}
