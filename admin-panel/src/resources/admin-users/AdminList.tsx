import { List, Datagrid, TextField, BooleanField, EditButton, DeleteButton, Filter, TextInput } from 'react-admin'

const AdminFilter = (props: any) => (
  <Filter {...props}>
    <TextInput source="username" />
  </Filter>
)

export const AdminList = () => (
  <List filters={<AdminFilter />}>
    <Datagrid>
      <TextField source="telegram_id" />
      <TextField source="username" />
      <TextField source="roles" />
      <BooleanField source="is_active" label="Active" />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
)
