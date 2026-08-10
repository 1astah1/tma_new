import { Admin, Resource } from 'react-admin'
import { dataProvider } from './dataProvider'
import { authProvider } from './authProvider'
import { Dashboard } from './components/Dashboard'
import { LoginPage } from './components/LoginPage'
import { CustomLayout } from './components/CustomLayout'
import { ProductList } from './resources/products/ProductList'
import { ProductEdit } from './resources/products/ProductEdit'
import { ProductCreate } from './resources/products/ProductCreate'
import { ProductShow } from './resources/products/ProductShow'
import { OrderList } from './resources/orders/OrderList'
import { OrderShow } from './resources/orders/OrderShow'
import { UserList } from './resources/users/UserList'
import { UserShow } from './resources/users/UserShow'
import { SettingsEdit } from './resources/settings/SettingsEdit'
import { TemplateList } from './resources/templates/TemplateList'
import { TemplateEdit } from './resources/templates/TemplateEdit'
import { TemplateCreate } from './resources/templates/TemplateCreate'
import { FAQList } from './resources/faq/FAQList'
import { FAQEdit } from './resources/faq/FAQEdit'
import { FAQCreate } from './resources/faq/FAQCreate'
import { AdminList } from './resources/admin-users/AdminList'
import { AdminEdit } from './resources/admin-users/AdminEdit'
import { AdminCreate } from './resources/admin-users/AdminCreate'
import { LogList } from './resources/logs/LogList'
import { PromoList } from './resources/promos/PromoList'
import { PromoCreate } from './resources/promos/PromoCreate'
import { PromoEdit } from './resources/promos/PromoEdit'
import { canAccessResource } from './utils/roles'
import { CatalogImportList } from './resources/catalog-imports/CatalogImportList'
import { HomeCategoriesPage } from './resources/home/HomeCategoriesPage'
import { HomeBannersPage } from './resources/home/HomeBannersPage'
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart'
import InventoryIcon from '@mui/icons-material/Inventory'
import PeopleIcon from '@mui/icons-material/People'
import SettingsIcon from '@mui/icons-material/Settings'
import ArticleIcon from '@mui/icons-material/Article'
import HelpIcon from '@mui/icons-material/Help'
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings'
import HistoryIcon from '@mui/icons-material/History'
import CloudDownloadIcon from '@mui/icons-material/CloudDownload'
import CategoryIcon from '@mui/icons-material/Category'
import ViewCarouselIcon from '@mui/icons-material/ViewCarousel'
import LocalOfferIcon from '@mui/icons-material/LocalOffer'

export default function App() {
  return (
    <Admin
      dataProvider={dataProvider}
      authProvider={authProvider}
      loginPage={LoginPage}
      dashboard={Dashboard}
      layout={CustomLayout}
    >
      {canAccessResource('products') && (
        <Resource name="products" list={ProductList} edit={ProductEdit} create={ProductCreate} show={ProductShow} options={{ label: 'Товары' }} icon={InventoryIcon} />
      )}
      {canAccessResource('catalog-imports') && (
        <Resource name="catalog-imports" list={CatalogImportList} options={{ label: 'Импорт игр' }} icon={CloudDownloadIcon} />
      )}
      {canAccessResource('home-categories') && (
        <Resource name="home-categories" list={HomeCategoriesPage} options={{ label: 'Главная' }} icon={CategoryIcon} />
      )}
      {canAccessResource('home-banners') && (
        <Resource name="home-banners" list={HomeBannersPage} options={{ label: 'Баннеры' }} icon={ViewCarouselIcon} />
      )}
      {canAccessResource('orders') && (
        <Resource name="orders" list={OrderList} show={OrderShow} options={{ label: 'Заказы' }} icon={ShoppingCartIcon} />
      )}
      {canAccessResource('users') && (
        <Resource name="users" list={UserList} show={UserShow} options={{ label: 'Пользователи' }} icon={PeopleIcon} />
      )}
      {canAccessResource('settings') && (
        <Resource name="settings" list={SettingsEdit} options={{ label: 'Настройки' }} icon={SettingsIcon} />
      )}
      {canAccessResource('templates') && (
        <Resource name="templates" list={TemplateList} edit={TemplateEdit} create={TemplateCreate} options={{ label: 'Шаблоны чата' }} icon={ArticleIcon} />
      )}
      {canAccessResource('faq-items') && (
        <Resource name="faq-items" list={FAQList} edit={FAQEdit} create={FAQCreate} options={{ label: 'FAQ' }} icon={HelpIcon} />
      )}
      {canAccessResource('promos') && (
        <Resource name="promos" list={PromoList} edit={PromoEdit} create={PromoCreate} icon={LocalOfferIcon} options={{ label: 'Промокоды' }} />
      )}
      {canAccessResource('admins') && (
        <Resource name="admins" list={AdminList} edit={AdminEdit} create={AdminCreate} options={{ label: 'Админы' }} icon={AdminPanelSettingsIcon} />
      )}
      {canAccessResource('logs') && (
        <Resource name="logs" list={LogList} options={{ label: 'Логи' }} icon={HistoryIcon} />
      )}
    </Admin>
  )
}
