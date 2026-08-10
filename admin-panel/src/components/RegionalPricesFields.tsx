import { NumberInput, FormDataConsumer } from 'react-admin'
import { Box, Typography, Alert } from '@mui/material'

export function RegionalPricesFields() {
  return (
    <FormDataConsumer>
      {({ formData }) => {
        if (formData?.type !== 'game') return null

        const isPS = formData.platform === 'ps4' || formData.platform === 'ps5'
        const isXbox = formData.platform === 'xbox'

        if (!isPS && !isXbox) return null

        if (formData.prices_edition_catalog) {
          return (
            <Alert severity="info" sx={{ mb: 2 }}>
              У товара настроен каталог изданий (edition_catalog). Региональные цены tr/ua используются только если изданий нет в
              витрине.
            </Alert>
          )
        }

        return (
          <Box sx={{ mb: 2 }}>
            <Typography variant="h6" gutterBottom>
              Цены по регионам (₽)
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
              Эти цены видит покупатель при выборе региона в карточке товара. Поле «Цена» выше — базовая для каталога и заказа.
            </Typography>
            {isPS && (
              <>
                <NumberInput
                  source="prices_tr"
                  label="Турция 🇹🇷 (TR)"
                  fullWidth
                  min={0}
                  helperText="Цена PlayStation Store Турция в рублях"
                />
                <NumberInput
                  source="prices_ua"
                  label="Украина 🇺🇦 (UA)"
                  fullWidth
                  min={0}
                  helperText="Цена PlayStation Store Украина в рублях"
                />
              </>
            )}
            {isXbox && (
              <NumberInput
                source="prices_xbox"
                label="США 🇺🇸 (Xbox)"
                fullWidth
                min={0}
                helperText="Цена Xbox US в рублях"
              />
            )}
          </Box>
        )
      }}
    </FormDataConsumer>
  )
}
