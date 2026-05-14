import { List, Datagrid, TextField, NumberField, DateField, SearchInput } from 'react-admin'

const filters = [<SearchInput key="search" source="search" alwaysOn />]

export const UserList = () => (
  <List filters={filters}>
    <Datagrid rowClick="show">
      <TextField source="username" label="Username" />
      <NumberField source="telegram_id" label="Telegram ID" />
      <DateField source="created_at" label="Регистрация" />
    </Datagrid>
  </List>
)
