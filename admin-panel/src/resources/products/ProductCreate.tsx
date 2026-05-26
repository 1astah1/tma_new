import { useState } from 'react'
import { Create, SimpleForm, TextInput, SelectInput, NumberInput, CheckboxGroupInput, ArrayInput, SimpleFormIterator, useCreate, useNotify, useRedirect, FormDataConsumer } from 'react-admin'
import { Box, Typography, Divider } from '@mui/material'
import { ImageUpload } from '../../components/ImageUpload'

export const ProductCreate = () => {
  const [create] = useCreate()
  const notify = useNotify()
  const redirect = useRedirect()
  const [productType, setProductType] = useState('game')

  const handleSubmit = async (data: any) => {
    try {
      if (!data.title || data.title.trim() === '') {
        notify('Название обязательно', { type: 'warning' })
        return
      }

      if (data.price <= 0) {
        notify('Цена должна быть больше 0', { type: 'warning' })
        return
      }

      if (data.discount_percent < 0 || data.discount_percent > 100) {
        notify('Скидка должна быть от 0 до 100', { type: 'warning' })
        return
      }

      const variants = (data.type === 'currency' || data.type === 'subscription') && data.variants
        ? JSON.stringify(data.variants.filter((v: any) => v.name && v.price))
        : '[]'

      await create('products', {
        data: {
          ...data,
          delivery_methods: data.delivery_methods || ['key'],
          discount_percent: Number(data.discount_percent) || 0,
          variants,
        },
      })
      notify('Товар создан', { type: 'success' })
      redirect('list', 'products')
    } catch (e: any) {
      notify(`Ошибка: ${e.message}`, { type: 'error' })
    }
  }

  return (
    <Create title="Создать товар">
      <SimpleForm onSubmit={handleSubmit}>
        <Box sx={{ mb: 2 }}>
          <Typography variant="h6" gutterBottom>Основная информация</Typography>
          <ImageUpload source="image_url" />
          <TextInput source="title" label="Название" fullWidth required helperText="Название товара" />
          <TextInput source="description" label="Описание" multiline rows={4} fullWidth helperText="Подробное описание товара" />
        </Box>

        <Divider sx={{ my: 2 }} />

        <Box sx={{ mb: 2 }}>
          <Typography variant="h6" gutterBottom>Параметры</Typography>
          <SelectInput source="platform" label="Платформа" choices={[
            { id: 'ps4', name: 'PS4' }, { id: 'ps5', name: 'PS5' }, { id: 'xbox', name: 'Xbox' },
          ]} required fullWidth />
          <SelectInput
            source="type"
            label="Тип"
            choices={[
              { id: 'game', name: 'Игра' },
              { id: 'currency', name: 'Валюта' },
              { id: 'subscription', name: 'Подписка' },
            ]}
            required
            fullWidth
            onChange={(e: any) => setProductType(e.target.value)}
          />
        </Box>

        <Divider sx={{ my: 2 }} />

        <Box sx={{ mb: 2 }}>
          <Typography variant="h6" gutterBottom>Цена и скидки</Typography>
          <NumberInput source="price" label="Цена (₽)" required fullWidth />
          <NumberInput source="discount_percent" label="Скидка %" defaultValue={0} fullWidth helperText="0-100%" />
        </Box>

        <Divider sx={{ my: 2 }} />

        <FormDataConsumer>
          {({ formData }) =>
            (formData.type === 'currency' || formData.type === 'subscription') && (
              <Box sx={{ mb: 2 }}>
                <Typography variant="h6" gutterBottom>Варианты (упаковки)</Typography>
                <ArrayInput source="variants" label="">
                  <SimpleFormIterator inline>
                    <TextInput source="name" label="Название" helperText="Напр. 1000 V-Bucks" />
                    <NumberInput source="price" label="Цена" />
                    <NumberInput source="stock" label="Кол-во" defaultValue={999} />
                  </SimpleFormIterator>
                </ArrayInput>
              </Box>
            )
          }
        </FormDataConsumer>

        <Divider sx={{ my: 2 }} />

        <Box sx={{ mb: 2 }}>
          <Typography variant="h6" gutterBottom>Доставка и статус</Typography>
          <CheckboxGroupInput source="delivery_methods" label="Способы доставки" choices={[
            { id: 'key', name: '🔑 Ключ' }, { id: 'activation', name: '🔐 Активация' },
          ]} defaultValue={['key']} required />
          <SelectInput source="status" label="Статус" choices={[
            { id: 'active', name: '✅ Активен' }, { id: 'inactive', name: '❌ Неактивен' },
          ]} defaultValue="active" required />
        </Box>
      </SimpleForm>
    </Create>
  )
}
