import { List, Datagrid, TextField, NumberField, DateField, SearchInput, ChipField } from 'react-admin'
import { Chip, Box } from '@mui/material'

function BanField({ record }: any) {
  if (!record) return null
  return (
    <Chip
      label={record.is_banned ? 'Заблокирован' : 'Активен'}
      size="small"
      color={record.is_banned ? 'error' : 'success'}
    />
  )
}

const filters = [<SearchInput key="search" source="search" alwaysOn />]

export const UserList = () => (
  <List filters={filters}>
    <Datagrid rowClick="show">
      <TextField source="username" label="Username" />
      <TextField source="first_name" label="Имя" />
      <NumberField source="telegram_id" label="Telegram ID" />
      <BanField source="is_banned" label="Статус" />
      <DateField source="created_at" label="Регистрация" showTime />
    </Datagrid>
  </List>
)
