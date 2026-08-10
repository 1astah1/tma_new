export function matchesPlatform(productPlatform: string, filterPlatform?: string) {
  if (!filterPlatform) return true
  if (filterPlatform === 'ps') {
    return productPlatform === 'ps4' || productPlatform === 'ps5'
  }
  return productPlatform === filterPlatform
}
