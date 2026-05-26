import { Edit, SimpleForm, TextInput, BooleanInput, NumberInput } from 'react-admin'

export const FAQEdit = () => (
  <Edit>
    <SimpleForm>
      <TextInput source="question" fullWidth />
      <TextInput source="answer" multiline rows={6} fullWidth />
      <NumberInput source="sort_order" />
      <BooleanInput source="is_active" />
    </SimpleForm>
  </Edit>
)
