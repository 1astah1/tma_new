import { Edit, SimpleForm, TextInput, BooleanInput, SelectArrayInput } from 'react-admin'

const roles = [
  { id: 'super_admin', name: 'Super Admin' },
  { id: 'game_manager', name: 'Game Manager' },
  { id: 'activation_admin', name: 'Activation Admin' },
  { id: 'support', name: 'Support' },
  { id: 'finance', name: 'Finance' },
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
