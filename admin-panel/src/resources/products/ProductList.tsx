import { List, Datagrid, TextField, NumberField, SelectField, ChipField, EditButton, ShowButton, FilterForm, TextInput, SelectInput, NumberInput, TopToolbar, CreateButton } from 'react-admin'
import { Box, Chip, Typography } from '@mui/material'
import KeyIcon from '@mui/icons-material/Key'
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart'

const statusColors: Record<string, 'success' | 'error' | 'default'> = {
  active: 'success',
  inactive: 'error',
}

function StatusField({ record }: any) {
  if (!record) return null
  const label = record.status === 'active' ? 'Активен' : 'Неактивен'
  return <Chip label={label} size="small" color={statusColors[record.status] || 'default'} />
}

function DiscountField({ record }: any) {
  if (!record) return null
  const discount = record.discount_percent || 0
  if (discount <= 0) return <span style={{ color: '#888' }}>—</span>
  return <span style={{ color: '#4caf50', fontWeight: 'bold' }}>-{discount}%</span>
}

function KeysCountField({ record }: any) {
  if (!record) return null
  const hasKeys = record.delivery_methods?.includes('key')
  if (!hasKeys) return <span style={{ color: '#888' }}>—</span>
  const count = record.available_keys ?? 0
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
      <KeyIcon fontSize="small" sx={{ color: count > 0 ? '#4caf50' : '#f44336' }} />
      <Typography variant="body2" sx={{ color: count > 0 ? '#fff' : '#f44336', fontWeight: 'bold' }}>
        {count}
      </Typography>
    </Box>
  )
}

function OrderCountField({ record }: any) {
  if (!record) return null
  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
      <ShoppingCartIcon fontSize="small" sx={{ color: '#888' }} />
      <Typography variant="body2" sx={{ color: '#fff' }}>{record.order_count || 0}</Typography>
    </Box>
  )
}

const ProductActions = () => (
  <TopToolbar>
    <CreateButton />
  </TopToolbar>
)

const filters = [
  <TextInput key="search" source="search" label="Поиск" alwaysOn placeholder="Название..." />,
  <SelectInput key="platform" source="platform" label="Платформа" choices={[
    { id: 'ps4', name: 'PS4' }, { id: 'ps5', name: 'PS5' }, { id: 'xbox', name: 'Xbox' },
  ]} />,
  <SelectInput key="type" source="type" label="Тип" choices={[
    { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
  ]} />,
  <SelectInput key="status" source="status" label="Статус" choices={[
    { id: 'active', name: 'Активен' }, { id: 'inactive', name: 'Неактивен' },
  ]} />,
]

export const ProductList = () => (
  <List filters={filters} actions={<ProductActions />}>
      <Datagrid rowClick="edit">
        <TextField source="title" label="Название" />
        <ChipField source="platform" label="Платформа" />
        <SelectField source="type" label="Тип" choices={[
          { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
        ]} />
        <NumberField source="price" label="Цена" options={{ style: 'currency', currency: 'RUB' }} />
        <DiscountField source="discount_percent" label="Скидка" />
        <KeysCountField source="available_keys" label="Ключи" />
        <OrderCountField source="order_count" label="Заказы" />
        <StatusField source="status" label="Статус" />
        <Box sx={{ display: 'flex', gap: 0.5 }}>
          <EditButton />
          <ShowButton />
        </Box>
      </Datagrid>
  </List>
)
