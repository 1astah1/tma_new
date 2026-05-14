import { useState } from 'react'
import { useInput } from 'react-admin'
import { Button, Box, Typography } from '@mui/material'

export const ImageUpload = (props: { source: string }) => {
  const { id, field } = useInput({ source: props.source })
  const [uploading, setUploading] = useState(false)
  const preview = field?.value || ''

  const handleChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setUploading(true)
    const form = new FormData()
    form.append('file', file)

    try {
      const res = await fetch('/api/v1/admin/upload', {
        method: 'POST',
        headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
        body: form,
      })
      const data = await res.json()
      field.onChange(data.url)
    } catch (err) {
      console.error('Upload failed', err)
    } finally {
      setUploading(false)
    }
  }

  return (
    <Box sx={{ mb: 2 }}>
      <Typography variant="caption" display="block" gutterBottom>Изображение</Typography>
      {preview && (
        <Box sx={{ mb: 1 }}>
          <img src={preview} alt="preview" style={{ maxWidth: 200, maxHeight: 150, borderRadius: 8 }} />
        </Box>
      )}
      <Button variant="outlined" component="label" disabled={uploading} size="small">
        {uploading ? 'Загрузка...' : 'Загрузить'}
        <input type="file" hidden accept="image/*" onChange={handleChange} />
      </Button>
    </Box>
  )
}
