import { useState } from 'react'
import { useInput, useNotify } from 'react-admin'
import { Button, Box, Typography, LinearProgress } from '@mui/material'
import CloudUploadIcon from '@mui/icons-material/CloudUpload'
import DeleteIcon from '@mui/icons-material/Delete'

export const ImageUpload = (props: { source: string }) => {
  const { id, field } = useInput({ source: props.source })
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState(0)
  const notify = useNotify()
  const preview = field?.value || ''

  const handleChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    if (!file.type.startsWith('image/')) {
      notify('Только изображения', { type: 'warning' })
      return
    }

    if (file.size > 5 * 1024 * 1024) {
      notify('Файл слишком большой (макс. 5МБ)', { type: 'warning' })
      return
    }

    setUploading(true)
    setProgress(0)
    const form = new FormData()
    form.append('file', file)

    try {
      const xhr = new XMLHttpRequest()
      xhr.open('POST', '/api/v1/admin/upload')
      xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('token')}`)

      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          setProgress(Math.round((event.loaded / event.total) * 100))
        }
      }

      xhr.onload = () => {
        if (xhr.status === 200) {
          const data = JSON.parse(xhr.responseText)
          field.onChange(data.url)
          notify('Изображение загружено', { type: 'success' })
        } else {
          notify('Ошибка загрузки', { type: 'error' })
        }
        setUploading(false)
      }

      xhr.onerror = () => {
        notify('Ошибка сети', { type: 'error' })
        setUploading(false)
      }

      xhr.send(form)
    } catch (err) {
      notify('Ошибка загрузки', { type: 'error' })
      setUploading(false)
    }
  }

  const handleRemove = () => {
    field.onChange('')
    notify('Изображение удалено', { type: 'info' })
  }

  return (
    <Box sx={{ mb: 2 }}>
      <Typography variant="caption" display="block" gutterBottom>Изображение</Typography>
      {preview && (
        <Box sx={{ mb: 1, position: 'relative', display: 'inline-block' }}>
          <img src={preview} alt="preview" style={{ maxWidth: 200, maxHeight: 150, borderRadius: 8, border: '1px solid #333' }} />
          <Button
            size="small"
            color="error"
            onClick={handleRemove}
            sx={{ position: 'absolute', top: 4, right: 4, minWidth: 'auto', p: 0.5, bgcolor: 'rgba(0,0,0,0.6)', '&:hover': { bgcolor: 'rgba(0,0,0,0.8)' } }}
          >
            <DeleteIcon fontSize="small" />
          </Button>
        </Box>
      )}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <Button variant="outlined" component="label" disabled={uploading} size="small" startIcon={<CloudUploadIcon />}>
          {uploading ? 'Загрузка...' : 'Загрузить'}
          <input type="file" hidden accept="image/*" onChange={handleChange} />
        </Button>
        {uploading && (
          <Box sx={{ flex: 1, maxWidth: 200 }}>
            <LinearProgress variant="determinate" value={progress} />
            <Typography variant="caption" color="text.secondary">{progress}%</Typography>
          </Box>
        )}
      </Box>
    </Box>
  )
}
