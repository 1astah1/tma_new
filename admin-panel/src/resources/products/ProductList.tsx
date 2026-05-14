import { List, Datagrid, TextField, NumberField, SelectField, ChipField, EditButton, FilterForm, TextInput, SelectInput, NumberInput } from 'react-admin'

const filters = [
  <TextInput key="search" source="search" label="Поиск" alwaysOn />,
  <SelectInput key="platform" source="platform" label="Платформа" choices={[
    { id: 'ps4', name: 'PS4' }, { id: 'ps5', name: 'PS5' }, { id: 'xbox', name: 'Xbox' },
  ]} />,
  <SelectInput key="type" source="type" label="Тип" choices={[
    { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
  ]} />,
]

export const ProductList = () => (
  <List filters={filters}>
    <Datagrid rowClick="edit">
      <TextField source="title" label="Название" />
      <ChipField source="platform" label="Платформа" />
      <SelectField source="type" label="Тип" choices={[
        { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
      ]} />
      <NumberField source="price" label="Цена" options={{ style: 'currency', currency: 'RUB' }} />
      <ChipField source="status" label="Статус" />
      <EditButton />
    </Datagrid>
  </List>
)
