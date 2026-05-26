import { Create, SimpleForm, TextInput, BooleanInput, SelectInput } from 'react-admin'

export const TemplateCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="title" fullWidth />
      <TextInput source="message" multiline rows={4} fullWidth />
      <SelectInput source="category" choices={[
        { id: 'general', name: 'General' },
        { id: 'greeting', name: 'Greeting' },
        { id: 'order', name: 'Order' },
      ]} />
      <BooleanInput source="is_active" />
    </SimpleForm>
  </Create>
)
