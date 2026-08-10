import { List, Datagrid, TextField, NumberField, BooleanField, DateField, EditButton, DeleteButton } from 'react-admin'

export const PromoList = () => (
  <List perPage={25} sort={{ field: 'created_at', order: 'DESC' }}>
    <Datagrid>
      <TextField source="code" label="Код" />
      <NumberField source="discount_percent" label="Скидка %" />
      <NumberField source="discount_fixed" label="Фикс. ₽" />
      <NumberField source="used_count" label="Использовано" />
      <NumberField source="usage_limit" label="Лимит" />
      <BooleanField source="is_active" label="Активен" />
      <DateField source="valid_until" label="До" showTime />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
)
