import { List, Datagrid, TextField, NumberField, DateField, SelectField, ReferenceField, FilterForm, SelectInput, SearchInput } from 'react-admin'

const orderStatuses = [
  { id: 'NEW', name: 'Новый' }, { id: 'WAITING_PAYMENT', name: 'Ожидает оплаты' },
  { id: 'PAYMENT_VERIFICATION', name: 'Проверка платежа' }, { id: 'PAID', name: 'Оплачен' },
  { id: 'WAITING_ACTIVATION', name: 'В очереди' }, { id: 'AWAITING_CREDENTIALS', name: 'Требуются данные' },
  { id: 'CREDENTIALS_RECEIVED', name: 'Данные получены' }, { id: 'AWAITING_2FA', name: 'Ожидает код' },
  { id: 'ACTIVATING', name: 'Активация...' }, { id: 'ACTIVATED', name: 'Активирован' },
  { id: 'KEY_ISSUED', name: 'Ключ выдан' }, { id: 'COMPLETED', name: 'Завершён' },
  { id: 'CANCELLED', name: 'Отменён' }, { id: 'REFUNDED', name: 'Возвращён' },
]

const filters = [
  <SearchInput key="search" source="search" alwaysOn />,
  <SelectInput key="status" source="status" label="Статус" choices={orderStatuses} />,
]

export const OrderList = () => (
  <List filters={filters}>
    <Datagrid rowClick="show">
      <TextField source="id" label="ID" />
      <TextField source="delivery_method" label="Метод" />
      <SelectField source="status" label="Статус" choices={orderStatuses} />
      <NumberField source="payment_amount" label="Сумма" options={{ style: 'currency', currency: 'RUB' }} />
      <DateField source="created_at" label="Дата" showTime />
    </Datagrid>
  </List>
)
