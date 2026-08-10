import { Edit, SimpleForm, TextInput, BooleanInput, SelectArrayInput } from 'react-admin'

const roles = [
  { id: 'super_admin', name: 'Супер-админ' },
  { id: 'game_manager', name: 'Каталог и контент' },
  { id: 'support', name: 'Поддержка' },
  { id: 'finance', name: 'Финансы' },
]

export const AdminEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="username" fullWidth />
      <TextInput source="telegram_id" label="Telegram ID" type="number" fullWidth />
      <SelectArrayInput source="roles" choices={roles} fullWidth />
      <BooleanInput source="is_active" />
    </SimpleForm>
  </Edit>
)
