import { useEffect, useState } from 'react'
import { Card, CardContent, Typography, Grid } from '@mui/material'

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

  return (
    <div style={{ padding: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
        <img src="/favicon.png" alt="COIN MINT" style={{ width: 'auto', height: 48 }} />
        <Typography variant="h4">COIN MINT</Typography>
      </div>
      <Grid container spacing={2}>
        <Grid item xs={12} sm={6} md={3}>
          <Card><CardContent>
            <Typography variant="h5">{stats.orders_today || 0}</Typography>
            <Typography color="textSecondary">Заказов сегодня</Typography>
          </CardContent></Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card><CardContent>
            <Typography variant="h5">{stats.revenue_today || 0} ₽</Typography>
            <Typography color="textSecondary">Выручка сегодня</Typography>
          </CardContent></Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card><CardContent>
            <Typography variant="h5">{stats.active_tasks || 0}</Typography>
            <Typography color="textSecondary">Активные задачи</Typography>
          </CardContent></Card>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Card><CardContent>
            <Typography variant="h5">{stats.available_keys || 0}</Typography>
            <Typography color="textSecondary">Свободных ключей</Typography>
          </CardContent></Card>
        </Grid>
      </Grid>
    </div>
  )
}
