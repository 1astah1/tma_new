import { List, Datagrid, TextField, BooleanField, EditButton, DeleteButton, Filter, TextInput, SelectInput } from 'react-admin'

const TemplateFilter = (props: any) => (
  <Filter {...props}>
    <SelectInput source="category" choices={[
      { id: 'general', name: 'General' },
      { id: 'greeting', name: 'Greeting' },
      { id: 'order', name: 'Order' },
    ]} alwaysOn />
    <TextInput source="title" />
  </Filter>
)

export const TemplateList = () => (
  <List filters={<TemplateFilter />}>
    <Datagrid>
      <TextField source="title" />
      <TextField source="category" />
      <BooleanField source="is_active" label="Active" />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
)
