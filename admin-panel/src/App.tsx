import { Admin, Resource, CustomRoutes } from 'react-admin'
import { Route } from 'react-router-dom'
import { dataProvider } from './dataProvider'
import { authProvider } from './authProvider'
import { Dashboard } from './components/Dashboard'
import { LoginPage } from './components/LoginPage'
import { ProductList } from './resources/products/ProductList'
import { ProductEdit } from './resources/products/ProductEdit'
import { ProductCreate } from './resources/products/ProductCreate'
import { ProductShow } from './resources/products/ProductShow'
import { OrderList } from './resources/orders/OrderList'
import { OrderShow } from './resources/orders/OrderShow'
import { UserList } from './resources/users/UserList'
import { UserShow } from './resources/users/UserShow'
import { SettingsEdit } from './resources/settings/SettingsEdit'
import { KeyList, KeyCreate, KeyShow, KeyEdit } from './resources/keys/KeyList'
import { KeyProductList } from './resources/keys/KeyProductList'
import { KeyProductDetail } from './resources/keys/KeyProductDetail'
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
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart'
import InventoryIcon from '@mui/icons-material/Inventory'
import PeopleIcon from '@mui/icons-material/People'
import SettingsIcon from '@mui/icons-material/Settings'
import KeyIcon from '@mui/icons-material/Key'
import ArticleIcon from '@mui/icons-material/Article'
import HelpIcon from '@mui/icons-material/Help'
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings'
import HistoryIcon from '@mui/icons-material/History'

export default function App() {
  return (
    <Admin
      dataProvider={dataProvider}
      authProvider={authProvider}
      loginPage={LoginPage}
      dashboard={Dashboard}
    >
      <Resource name="products" list={ProductList} edit={ProductEdit} create={ProductCreate} show={ProductShow} icon={InventoryIcon} />
      <Resource name="orders" list={OrderList} show={OrderShow} icon={ShoppingCartIcon} />
      <Resource name="users" list={UserList} show={UserShow} icon={PeopleIcon} />
      <Resource name="settings" list={SettingsEdit} icon={SettingsIcon} />
      <Resource name="keys" list={KeyProductList} icon={KeyIcon} />
      <Resource name="templates" list={TemplateList} edit={TemplateEdit} create={TemplateCreate} icon={ArticleIcon} />
      <Resource name="faq-items" list={FAQList} edit={FAQEdit} create={FAQCreate} icon={HelpIcon} />
      <Resource name="admins" list={AdminList} edit={AdminEdit} create={AdminCreate} icon={AdminPanelSettingsIcon} />
      <Resource name="logs" list={LogList} icon={HistoryIcon} />
      <CustomRoutes>
        <Route path="keys/:id" element={<KeyProductDetail />} />
      </CustomRoutes>
    </Admin>
  )
}
