import { Product } from './product'



export interface HomeBanner {

  id: string

  image_url: string

  link_url?: string

  title?: string

  /** Натуральная ширина файла — для чёткого рендера без апскейла. */

  width?: number

  /** Натуральная высота файла. */

  height?: number

}



export interface HomeCategoryListItem {

  id: string

  title: string

  image_url: string

  product_count: number

  catalog_type?: string

}



export interface HomeCategoryDetail {

  id: string

  title: string

  image_url: string

  catalog_type?: string

  products: Product[]

}

