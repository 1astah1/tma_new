import { useState } from 'react'
import {
  List,
  Datagrid,
  TextField,
  SelectField,
  DateField,
  ReferenceField,
  Filter,
  SelectInput,
  useRefresh,
  useNotify,
  Button,
  TopToolbar,
  Create,
  SimpleForm,
  TextInput,
  SelectInput as RASelectInput,
  useListContext,
  Pagination,
  Show,
  SimpleShowLayout,
  ShowButton,
  EditButton,
  DeleteButton,
  Edit,
  useRecordContext,
} from 'react-admin'
import { Box, Dialog, DialogTitle, DialogContent, DialogActions, TextField as MuiTextField, Button as MuiButton } from '@mui/material'
import CloudUploadIcon from '@mui/icons-material/CloudUpload'
import DeleteIcon from '@mui/icons-material/Delete'
import EditIcon from '@mui/icons-material/Edit'
import ContentCopyIcon from '@mui/icons-material/ContentCopy'

const KeyActions = () => (
  <TopToolbar>
    <ImportButton />
  </TopToolbar>
)

const ImportButton = () => {
  const [open, setOpen] = useState(false)
  const [keys, setKeys] = useState('')
  const [productId, setProductId] = useState('')
  const notify = useNotify()
  const refresh = useRefresh()
  const token = localStorage.getItem('token')

  const handleImport = async () => {
    const keyList = keys.split('\n').map(k => k.trim()).filter(Boolean)
    if (!productId || keyList.length === 0) {
      notify('Product ID and keys are required', { type: 'warning' })
      return
    }
    try {
      const res = await fetch('/api/v1/admin/keys/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ product_id: productId, keys: keyList }),
      })
      const data = await res.json()
      notify(`Imported ${data.imported}/${data.total} keys`, { type: 'success' })
      refresh()
      setOpen(false)
      setKeys('')
      setProductId('')
    } catch {
      notify('Import failed', { type: 'error' })
    }
  }

  return (
    <>
      <Button label="Import Keys" onClick={() => setOpen(true)}>
        <CloudUploadIcon />
      </Button>
      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Import Keys</DialogTitle>
        <DialogContent>
          <MuiTextField
            fullWidth
            label="Product ID (UUID)"
            value={productId}
            onChange={(e) => setProductId(e.target.value)}
            margin="normal"
          />
          <MuiTextField
            fullWidth
            multiline
            rows={10}
            label="Keys (one per line)"
            value={keys}
            onChange={(e) => setKeys(e.target.value)}
            margin="normal"
          />
        </DialogContent>
        <DialogActions>
          <MuiButton onClick={() => setOpen(false)}>Cancel</MuiButton>
          <MuiButton onClick={handleImport} variant="contained" color="primary">Import</MuiButton>
        </DialogActions>
      </Dialog>
    </>
  )
}

const KeyBulkDelete = () => {
  const { selectedIds } = useListContext()
  const notify = useNotify()
  const refresh = useRefresh()
  const token = localStorage.getItem('token')

  const handleBulkDelete = async () => {
    if (!selectedIds?.length) {
      notify('No keys selected', { type: 'warning' })
      return
    }
    try {
      await fetch('/api/v1/admin/keys/bulk-delete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ ids: selectedIds }),
      })
      notify(`${selectedIds.length} keys deleted`, { type: 'success' })
      refresh()
    } catch {
      notify('Delete failed', { type: 'error' })
    }
  }

  return (
    <Button
      label="Delete Selected"
      onClick={handleBulkDelete}
      disabled={!selectedIds?.length}
    >
      <DeleteIcon />
    </Button>
  )
}

const KeyFilter = (props: any) => (
  <Filter {...props}>
    <SelectInput source="status" choices={[
      { id: 'available', name: 'Available' },
      { id: 'sold', name: 'Sold' },
      { id: 'reserved', name: 'Reserved' },
    ]} />
  </Filter>
)

const CopyButton = ({ text }: { text: string }) => {
  const notify = useNotify()
  const handleCopy = () => {
    navigator.clipboard.writeText(text)
    notify('Key copied to clipboard', { type: 'success' })
  }
  return (
    <Button label="" onClick={handleCopy} title="Copy key">
      <ContentCopyIcon />
    </Button>
  )
}

export const KeyList = () => (
  <List actions={<KeyActions />} filters={<KeyFilter />} pagination={<Pagination rowsPerPageOptions={[10, 25, 50, 100]} />}>
    <Datagrid bulkActionButtons={<KeyBulkDelete />}>
      <TextField source="id" label="ID" />
      <TextField source="key" label="Key" />
      <ReferenceField source="product_id" reference="products" link="show">
        <TextField source="title" />
      </ReferenceField>
      <SelectField
        source="status"
        choices={[
          { id: 'available', name: 'Available' },
          { id: 'sold', name: 'Sold' },
          { id: 'reserved', name: 'Reserved' },
        ]}
      />
      <DateField source="created_at" showTime />
      <ShowButton />
      <EditButton />
    </Datagrid>
  </List>
)

export const KeyShow = () => (
  <Show>
    <SimpleShowLayout>
      <TextField source="id" />
      <TextField source="key" />
      <ReferenceField source="product_id" reference="products" link="show">
        <TextField source="title" />
      </ReferenceField>
      <SelectField
        source="status"
        choices={[
          { id: 'available', name: 'Available' },
          { id: 'sold', name: 'Sold' },
          { id: 'reserved', name: 'Reserved' },
        ]}
      />
      <TextField source="order_id" label="Order ID" />
      <DateField source="created_at" showTime />
      <Box mt={2}>
        <CopyButton text={useRecordContext()?.key || ''} />
      </Box>
    </SimpleShowLayout>
  </Show>
)

function KeyEditForm() {
  const record = useRecordContext()
  const notify = useNotify()
  const token = localStorage.getItem('token')
  const refresh = useRefresh()

  const handleStatusChange = async (newStatus: string) => {
    if (!record) return
    try {
      await fetch(`/api/v1/admin/keys/${record.id}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ status: newStatus }),
      })
      notify(`Key status changed to ${newStatus}`, { type: 'success' })
      refresh()
    } catch {
      notify('Failed to update status', { type: 'error' })
    }
  }

  return (
    <SimpleForm record={record} toolbar={<EditToolbar />}>
      <TextField source="id" label="Key ID" />
      <TextField source="key" label="Key Value" />
      <ReferenceField source="product_id" reference="products" link="show">
        <TextField source="title" />
      </ReferenceField>
      <SelectField
        source="status"
        choices={[
          { id: 'available', name: 'Available' },
          { id: 'sold', name: 'Sold' },
          { id: 'reserved', name: 'Reserved' },
        ]}
      />
      <TextField source="order_id" label="Assigned Order" />
      <DateField source="created_at" showTime label="Created" />
      <Box mt={2} display="flex" gap={1}>
        <MuiButton variant="outlined" color="success" onClick={() => handleStatusChange('available')}>
          Set Available
        </MuiButton>
        <MuiButton variant="outlined" color="warning" onClick={() => handleStatusChange('reserved')}>
          Set Reserved
        </MuiButton>
        <MuiButton variant="outlined" color="error" onClick={() => handleStatusChange('sold')}>
          Set Sold
        </MuiButton>
      </Box>
    </SimpleForm>
  )
}

const EditToolbar = () => (
  <Box display="flex" justifyContent="space-between" alignItems="center">
    <DeleteButton />
  </Box>
)

export const KeyEdit = () => (
  <Edit>
    <KeyEditForm />
  </Edit>
)

export const KeyCreate = () => (
  <Create>
    <SimpleForm>
      <TextInput source="product_id" />
      <TextInput source="key" multiline rows={3} />
      <RASelectInput
        source="status"
        choices={[
          { id: 'available', name: 'Available' },
          { id: 'sold', name: 'Sold' },
          { id: 'reserved', name: 'Reserved' },
        ]}
      />
    </SimpleForm>
  </Create>
)
