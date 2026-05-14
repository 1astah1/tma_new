import { Create, SimpleForm, TextInput, SelectInput, NumberInput, BooleanInput } from 'react-admin'
import { ImageUpload } from '../../components/ImageUpload'

export const ProductCreate = () => (
  <Create>
    <SimpleForm>
      <ImageUpload source="image_url" />
      <TextInput source="title" label="Название" fullWidth required />
      <TextInput source="description" label="Описание" multiline fullWidth />
      <SelectInput source="platform" label="Платформа" choices={[
        { id: 'ps4', name: 'PS4' }, { id: 'ps5', name: 'PS5' }, { id: 'xbox', name: 'Xbox' },
      ]} required />
      <SelectInput source="type" label="Тип" choices={[
        { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
      ]} required />
      <NumberInput source="price" label="Цена" required />
      <TextInput source="image_url" label="URL изображения" fullWidth />
      <BooleanInput source="status" label="Активен" />
    </SimpleForm>
  </Create>
)
