import { ProductEditionCatalog } from '../types/product'

export const COD_MW4_PRODUCT_ID = '63c124f6-af26-4cac-b5da-3be19905398a'
export const COD_MW4_TITLE_KEY = 'call of duty: modern warfare 4'
export const COD_MW4_DISPLAY_TITLE = 'Call of Duty®: Modern Warfare® 4 (PS/XBOX/PC)'

/** Ручные цены предзаказа Call of Duty: Modern Warfare 4. */
export const COD_MW4_EDITION_CATALOG: ProductEditionCatalog = {
  ps_tr: [
    { id: 'standard', name: 'Standard Edition', price: 6300 },
    { id: 'vault', name: 'Vault Edition', price: 7800, discount_label: '−10%' },
  ],
  ps_ua: [
    { id: 'standard', name: 'Standard Edition', price: 7300 },
    { id: 'vault', name: 'Vault Edition', price: 8900, discount_label: '−10%' },
  ],
  xbox: [
    { id: 'standard', name: 'Standard Edition', price: 5000 },
    { id: 'vault', name: 'Vault Edition', price: 6700, discount_label: '−10%' },
  ],
}
