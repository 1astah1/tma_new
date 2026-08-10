import { Edit, SimpleForm, TextInput, SelectInput, NumberInput, ArrayInput, SimpleFormIterator, SaveButton, Toolbar, DeleteButton, useUpdate, useNotify, useRedirect, useRecordContext, FormDataConsumer } from 'react-admin'
import { Box, Typography, Divider } from '@mui/material'
import { ProductImportPanel } from '../../components/ProductImportPanel'
import { ImageUpload } from '../../components/ImageUpload'
import { RegionalPricesFields } from '../../components/RegionalPricesFields'
import { GameSectionInput, parseGameSection } from '../../components/GameSectionInput'
import { buildProductPricesPayload, parseProductPrices } from '../../utils/productPrices'

const ProductEditToolbar = () => (
  <Toolbar>
    <SaveButton />
    <DeleteButton />
  </Toolbar>
)

function ProductEditForm() {
  const record = useRecordContext()
  const [update] = useUpdate()
  const notify = useNotify()
  const redirect = useRedirect()

  const handleSubmit = async (data: any) => {
    if (!record) return

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

      const existingPrices = parseProductPrices(record.prices)
      const prices = buildProductPricesPayload(data.platform, {
        prices_tr: data.prices_tr,
        prices_ua: data.prices_ua,
        prices_xbox: data.prices_xbox,
      }, existingPrices)

      const { prices_tr, prices_ua, prices_xbox, prices_edition_catalog, ...rest } = data

      await update('products', {
        id: record.id,
        data: {
          ...rest,
          game_section: data.type === 'game' ? parseGameSection(data.game_section) : '',
          delivery_methods: ['activation'],
          discount_percent: Number(data.discount_percent) || 0,
          variants,
          prices,
        },
      })
      notify('Товар обновлён', { type: 'success' })
      redirect('list', 'products')
    } catch (e: any) {
      notify(`Ошибка: ${e.message}`, { type: 'error' })
    }
  }

  const isVariantType = record?.type === 'currency' || record?.type === 'subscription'
  const parsedVariants = isVariantType && record?.variants ? (typeof record.variants === 'string' ? JSON.parse(record.variants) : record.variants) : []
  const parsedPrices = parseProductPrices(record?.prices)
  const formRecord = {
    ...record,
    variants: parsedVariants,
    prices_tr: parsedPrices.tr ?? '',
    prices_ua: parsedPrices.ua ?? '',
    prices_xbox: parsedPrices.xbox ?? parsedPrices.us ?? '',
    prices_edition_catalog: !!parsedPrices.edition_catalog,
  } as any

  return (
    <SimpleForm record={formRecord} onSubmit={handleSubmit} toolbar={<ProductEditToolbar />}>
      <ProductImportPanel />
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
          { id: 'ps4', name: 'PS4' }, { id: 'ps5', name: 'PS5' }, { id: 'xbox', name: 'Xbox' }, { id: 'pc', name: 'PC' },
        ]} required fullWidth />
        <SelectInput source="type" label="Тип" choices={[
          { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
        ]} required fullWidth />
        <FormDataConsumer>
          {({ formData }) => formData?.type === 'game' ? <GameSectionInput /> : null}
        </FormDataConsumer>
      </Box>

      <Divider sx={{ my: 2 }} />

      <Box sx={{ mb: 2 }}>
        <Typography variant="h6" gutterBottom>Цена и скидки</Typography>
        <NumberInput source="price" label="Базовая цена (₽)" required fullWidth helperText="Для каталога и заказа, если не выбран регион" />
        <NumberInput source="discount_percent" label="Скидка %" fullWidth helperText="0-100%" />
      </Box>

      <RegionalPricesFields />

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
        <Typography variant="body2" sx={{ mb: 2, color: 'text.secondary' }}>
          Продажа только через чат с менеджером
        </Typography>
        <SelectInput source="status" label="Статус" choices={[
          { id: 'active', name: '✅ Активен' }, { id: 'inactive', name: '❌ Неактивен' },
        ]} required />
      </Box>
    </SimpleForm>
  )
}

export const ProductEdit = () => (
  <Edit title="Редактировать товар">
    <ProductEditForm />
  </Edit>
)
