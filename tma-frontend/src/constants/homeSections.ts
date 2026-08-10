export const homePlatformSections = [
  { id: 'game', label: 'Игры', icon: '/icons/game.png' },
  { id: 'new', label: 'Новинки', icon: '/icons/game.png' },
  { id: 'preorder', label: 'Предзаказы', icon: '/icons/Sub.png' },
] as const

export type HomePlatformSection = typeof homePlatformSections[number]['id']

export const homeSectionLabels: Record<HomePlatformSection, string> = {
  game: 'Игры',
  new: 'Новинки',
  preorder: 'Предзаказы',
}
