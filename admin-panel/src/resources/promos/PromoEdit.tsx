import { Edit, SimpleForm, TextInput, NumberInput, BooleanInput, DateTimeInput } from 'react-admin'

export const PromoEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="code" label="Промокод" fullWidth />
      <NumberInput source="discount_percent" label="Скидка %" min={0} max={100} />
      <NumberInput source="discount_fixed" label="Фиксированная скидка ₽" min={0} />
      <NumberInput source="usage_limit" label="Лимит использований" />
      <NumberInput source="used_count" label="Использовано" disabled />
      <DateTimeInput source="valid_until" label="Действует до" />
      <BooleanInput source="is_active" label="Активен" />
    </SimpleForm>
  </Edit>
)
