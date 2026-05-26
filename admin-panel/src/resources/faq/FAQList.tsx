import { List, Datagrid, TextField, BooleanField, EditButton, DeleteButton, Filter, TextInput, NumberInput } from 'react-admin'

const FAQFilter = (props: any) => (
  <Filter {...props}>
    <TextInput source="question" />
  </Filter>
)

export const FAQList = () => (
  <List filters={<FAQFilter />}>
    <Datagrid>
      <TextField source="question" />
      <TextField source="sort_order" />
      <BooleanField source="is_active" label="Active" />
      <EditButton />
      <DeleteButton />
    </Datagrid>
  </List>
)
