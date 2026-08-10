import { SelectInput } from 'react-admin'

export const GAME_SECTION_CATALOG = 'catalog'

const choices = [
  { id: GAME_SECTION_CATALOG, name: 'Игры' },
  { id: 'new', name: 'Новинки' },
  { id: 'preorder', name: 'Предзаказы' },
]

export function formatGameSection(value?: string | null) {
  return value && value !== '' ? value : GAME_SECTION_CATALOG
}

export function parseGameSection(value?: string | null) {
  return value === GAME_SECTION_CATALOG ? '' : (value ?? '')
}

export function GameSectionInput() {
  return (
    <SelectInput
      source="game_section"
      label="Раздел на главной"
      choices={choices}
      format={formatGameSection}
      parse={parseGameSection}
      fullWidth
      helperText="«Игры» — обычный каталог. «Новинки» и «Предзаказы» — секции на главной."
    />
  )
}
