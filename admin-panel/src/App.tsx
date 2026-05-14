import { Admin, Resource, CustomRoutes } from 'react-admin'
import { Route } from 'react-router-dom'
import { dataProvider } from './dataProvider'
import { authProvider } from './authProvider'
import { Dashboard } from './components/Dashboard'
import { LoginPage } from './components/LoginPage'
import { ProductList } from './resources/products/ProductList'
import { ProductEdit } from './resources/products/ProductEdit'
import { ProductCreate } from './resources/products/ProductCreate'
import { OrderList } from './resources/orders/OrderList'
import { OrderShow } from './resources/orders/OrderShow'
import { UserList } from './resources/users/UserList'
import { SettingsEdit } from './resources/settings/SettingsEdit'
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart'
import InventoryIcon from '@mui/icons-material/Inventory'
import PeopleIcon from '@mui/icons-material/People'
import SettingsIcon from '@mui/icons-material/Settings'

export default function App() {
  return (
    <Admin
      dataProvider={dataProvider}
      authProvider={authProvider}
      loginPage={LoginPage}
      dashboard={Dashboard}
    >
      <Resource name="products" list={ProductList} edit={ProductEdit} create={ProductCreate} icon={InventoryIcon} />
      <Resource name="orders" list={OrderList} show={OrderShow} icon={ShoppingCartIcon} />
      <Resource name="users" list={UserList} icon={PeopleIcon} />
      <Resource name="settings" list={SettingsEdit} icon={SettingsIcon} />
    </Admin>
  )
}
