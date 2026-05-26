import { useEffect, useState } from 'react'
import { Card, CardContent, Typography, Grid, Box, Paper, LinearProgress } from '@mui/material'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, BarChart, Bar } from 'recharts'
import TrendingUpIcon from '@mui/icons-material/TrendingUp'
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart'
import AttachMoneyIcon from '@mui/icons-material/AttachMoney'
import KeyIcon from '@mui/icons-material/Key'
import PeopleIcon from '@mui/icons-material/People'
import WarningIcon from '@mui/icons-material/Warning'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'
import PendingIcon from '@mui/icons-material/Pending'
import SpeedIcon from '@mui/icons-material/Speed'

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

export const Dashboard = () => {
  const [stats, setStats] = useState<any>({})

  useEffect(() => {
    fetch('/api/v1/admin/dashboard', {
      headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
    })
      .then((r) => r.json())
      .then(setStats)
      .catch(() => {})
  }, [])

  const fmt = (n: number) => n.toLocaleString('ru-RU', { minimumFractionDigits: 0, maximumFractionDigits: 0 })
  const fmtMoney = (n: number) => n.toLocaleString('ru-RU', { minimumFractionDigits: 0, maximumFractionDigits: 0 }) + ' ₽'

  const statusData = [
    { name: 'Новые', value: stats.orders_today || 0, color: '#38bdf8' },
    { name: 'Ожидание оплаты', value: stats.waiting_payment || 0, color: '#fb923c' },
    { name: 'Проверка платежа', value: stats.pending_orders || 0, color: '#f59e0b' },
    { name: 'Активные задачи', value: stats.active_tasks || 0, color: '#a78bfa' },
    { name: 'Завершённые', value: (stats.total_orders || 0) - (stats.orders_today || 0) - (stats.waiting_payment || 0) - (stats.pending_orders || 0) - (stats.active_tasks || 0), color: '#4ade80' },
  ].filter(d => d.value > 0)

  const revenueData = [
    { name: 'Сегодня', revenue: stats.revenue_today || 0 },
    { name: 'Неделя', revenue: stats.revenue_week || 0 },
    { name: 'Месяц', revenue: stats.revenue_month || 0 },
    { name: 'Всего', revenue: stats.total_revenue || 0 },
  ]

  return (
    <div style={{ padding: 20 }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 4 }}>
        <img src="/favicon.png" alt="COIN MINT" style={{ width: 'auto', height: 48 }} />
        <Typography variant="h4" sx={{ fontWeight: 'bold' }}>COIN MINT Dashboard</Typography>
      </Box>

      {/* Revenue & Orders Row */}
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

      {/* Tasks & Keys Row */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} sm={6} md={3}>
          <StatCard
            title="Активные задачи"
            value={fmt(stats.active_tasks || 0)}
            icon={<TrendingUpIcon />}
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
            title="Свободных ключей"
            value={fmt(stats.available_keys || 0)}
            icon={<KeyIcon />}
            color="#fb923c"
            subtitle={`Продано: ${fmt(stats.sold_keys || 0)}`}
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

      {/* Charts Row */}
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
                  <Tooltip formatter={(value: number) => `${fmtMoney(value)}`} />
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
                    label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
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

      {/* Bottom Row */}
      <Grid container spacing={3}>
        <Grid item xs={12} md={6}>
          <Card>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2, fontWeight: 'bold' }}>Последние заказы</Typography>
              {stats.recent_orders?.length > 0 ? (
                <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                  {stats.recent_orders.map((order: any) => (
                    <Paper key={order.id} sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
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
