import { Create, SimpleForm, TextInput, NumberInput, BooleanInput, DateTimeInput } from 'react-admin'

export const PromoCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="code" label="Промокод" fullWidth required />
      <NumberInput source="discount_percent" label="Скидка %" min={0} max={100} />
      <NumberInput source="discount_fixed" label="Фиксированная скидка ₽" min={0} />
      <NumberInput source="usage_limit" label="Лимит использований" />
      <DateTimeInput source="valid_until" label="Действует до" />
      <BooleanInput source="is_active" label="Активен" defaultValue />
    </SimpleForm>
  </Create>
)
