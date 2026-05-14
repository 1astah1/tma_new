import { Login } from 'react-admin'
import { Box, Typography } from '@mui/material'

export const LoginPage = () => (
  <Login>
    <Box sx={{ textAlign: 'center', mb: 3 }}>
      <img src="/favicon.png" alt="COIN MINT" style={{ width: 'auto', height: 64, marginBottom: 12 }} />
      <Typography variant="h5" fontWeight="bold">COIN MINT</Typography>
      <Typography variant="body2" color="text.secondary">Admin Panel</Typography>
    </Box>
  </Login>
)
