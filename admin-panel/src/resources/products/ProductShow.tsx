import { Show, SimpleShowLayout, TextField, NumberField, SelectField, ChipField, DateField, EditButton, useRecordContext } from 'react-admin'
import { Box, Typography, Chip, Stack } from '@mui/material'
import { ProductImportPanel } from '../../components/ProductImportPanel'
import { parseProductPrices } from '../../utils/productPrices'

function RegionalPricesShow() {
  const record = useRecordContext()
  if (!record || record.type !== 'game') return null
  const prices = parseProductPrices(record.prices)
  const isPS = record.platform === 'ps4' || record.platform === 'ps5'
  const isXbox = record.platform === 'xbox'
  if (!isPS && !isXbox) return null
  if (prices.edition_catalog) return null

  const items: { label: string; value?: number }[] = []
  if (isPS) {
    items.push({ label: 'Турция 🇹🇷', value: prices.tr })
    items.push({ label: 'Украина 🇺🇦', value: prices.ua })
  }
  if (isXbox) {
    items.push({ label: 'США 🇺🇸', value: prices.xbox ?? prices.us })
  }

  if (!items.some((i) => i.value != null)) return null

  return (
    <Box sx={{ mb: 2 }}>
      <Typography variant="subtitle2" color="text.secondary" gutterBottom>Цены по регионам</Typography>
      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
        {items.map((item) => (
          <Chip
            key={item.label}
            label={item.value != null ? `${item.label}: ${Math.round(item.value)} ₽` : `${item.label}: —`}
            size="small"
            variant="outlined"
          />
        ))}
      </Stack>
    </Box>
  )
}

function VariantsField({ record }: any) {
  if (!record) return null
  try {
    const variants = typeof record.variants === 'string' ? JSON.parse(record.variants) : record.variants
    if (!Array.isArray(variants) || variants.length === 0) return <Typography variant="body2" color="text.secondary">Нет вариантов</Typography>
    return (
      <Box sx={{ mt: 1 }}>
        {variants.map((v: any, i: number) => (
          <Box key={i} sx={{ display: 'flex', gap: 2, mb: 1, p: 1, bgcolor: '#1a1a1a', borderRadius: 1 }}>
            <Typography variant="body2"><strong>{v.name}</strong></Typography>
            <Typography variant="body2">{v.price} ₽</Typography>
            <Typography variant="body2" color="text.secondary">Кол-во: {v.stock}</Typography>
          </Box>
        ))}
      </Box>
    )
  } catch {
    return <Typography variant="body2" color="text.secondary">Ошибка parsing вариантов</Typography>
  }
}

export const ProductShow = () => (
  <Show actions={<EditButton />}>
    <SimpleShowLayout>
      <ProductImportPanel />
      <TextField source="id" label="ID" />
      <TextField source="title" label="Название" />
      <TextField source="description" label="Описание" />
      <ChipField source="platform" label="Платформа" />
      <SelectField source="type" label="Тип" choices={[
        { id: 'game', name: 'Игра' }, { id: 'currency', name: 'Валюта' }, { id: 'subscription', name: 'Подписка' },
      ]} />
      <NumberField source="price" label="Базовая цена" options={{ style: 'currency', currency: 'RUB' }} />
      <RegionalPricesShow />
      <NumberField source="discount_percent" label="Скидка %" />
      <ChipField source="delivery_methods" label="Способы доставки" />
      <ChipField source="status" label="Статус" />
      <VariantsField source="variants" label="Варианты" />
      <DateField source="created_at" label="Создан" showTime />
      <DateField source="updated_at" label="Обновлён" showTime />
    </SimpleShowLayout>
  </Show>
)
