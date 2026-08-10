import { useEffect, useState } from 'react'
import { Button, useNotify } from 'react-admin'
import {
  Box,
  Card,
  CardContent,
  Checkbox,
  Chip,
  FormControlLabel,
  IconButton,
  Stack,
  TextField,
  Typography,
  Button as MuiButton,
} from '@mui/material'
import SaveIcon from '@mui/icons-material/Save'
import AddIcon from '@mui/icons-material/Add'
import DeleteIcon from '@mui/icons-material/Delete'
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward'
import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward'

const apiUrl = '/api/v1/admin'

type HomeBanner = {
  id: string
  image_url: string
  link_url: string
  title: string
  sort_order: number
  active: boolean
}

function authHeaders() {
  const headers = new Headers({ 'Content-Type': 'application/json' })
  const token = localStorage.getItem('token')
  if (token) headers.set('Authorization', `Bearer ${token}`)
  return headers
}

function newBanner(sortOrder: number): HomeBanner {
  return {
    id: crypto.randomUUID(),
    image_url: '',
    link_url: '',
    title: '',
    sort_order: sortOrder,
    active: true,
  }
}

export function HomeBannersPage() {
  const notify = useNotify()
  const [banners, setBanners] = useState<HomeBanner[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const res = await fetch(`${apiUrl}/settings?key=home_banners`, { headers: authHeaders() })
      if (res.ok) {
        const json = await res.json()
        const parsed = JSON.parse(json.value || '[]') as HomeBanner[]
        setBanners([...parsed].sort((a, b) => a.sort_order - b.sort_order))
      } else {
        setBanners([])
      }
    } catch (e: any) {
      notify(e.message || 'Не удалось загрузить баннеры', { type: 'error' })
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load().catch(() => setLoading(false))
  }, [])

  const updateBanner = (id: string, patch: Partial<HomeBanner>) => {
    setBanners((prev) => prev.map((b) => (b.id === id ? { ...b, ...patch } : b)))
  }

  const addBanner = () => {
    setBanners((prev) => [...prev, newBanner(prev.length)])
  }

  const removeBanner = (id: string) => {
    setBanners((prev) => prev.filter((b) => b.id !== id).map((b, index) => ({ ...b, sort_order: index })))
  }

  const moveBanner = (id: string, direction: -1 | 1) => {
    setBanners((prev) => {
      const index = prev.findIndex((b) => b.id === id)
      if (index < 0) return prev
      const target = index + direction
      if (target < 0 || target >= prev.length) return prev
      const next = [...prev]
      ;[next[index], next[target]] = [next[target], next[index]]
      return next.map((b, i) => ({ ...b, sort_order: i }))
    })
  }

  const uploadImage = async (id: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    const headers = new Headers()
    const token = localStorage.getItem('token')
    if (token) headers.set('Authorization', `Bearer ${token}`)
    const res = await fetch(`${apiUrl}/upload`, { method: 'POST', headers, body: form })
    if (!res.ok) throw new Error('Не удалось загрузить изображение')
    const json = await res.json()
    updateBanner(id, { image_url: json.url })
  }

  const save = async () => {
    setSaving(true)
    try {
      const payload = banners.map((b, index) => ({ ...b, sort_order: index }))
      const res = await fetch(`${apiUrl}/settings`, {
        method: 'PUT',
        headers: authHeaders(),
        body: JSON.stringify({ key: 'home_banners', value: payload }),
      })
      if (!res.ok) throw new Error('Не удалось сохранить баннеры')
      notify('Баннеры сохранены', { type: 'success' })
    } catch (e: any) {
      notify(e.message, { type: 'error' })
    } finally {
      setSaving(false)
    }
  }

  return (
    <Box sx={{ p: 2 }}>
      <Typography variant="h5" sx={{ mb: 1 }}>Баннеры на главной</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
        Загружайте PNG/JPG <b>не меньше 1940×810 px</b> и <b>без сжатия</b> — файлы отдаются как есть.
        Маленькие картинки (1024 px и меньше) будут размытыми на телефоне.
      </Typography>
      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Stack direction="row" spacing={1} alignItems="center" sx={{ flexWrap: 'wrap' }}>
            <Button label="Добавить баннер" onClick={addBanner} startIcon={<AddIcon />} />
            <Button label="Сохранить" onClick={save} disabled={saving} startIcon={<SaveIcon />} />
            <Chip label={`Баннеров: ${banners.length}`} />
          </Stack>
        </CardContent>
      </Card>

      {loading ? (
        <Typography>Загрузка...</Typography>
      ) : (
        <Stack spacing={2}>
          {banners.map((banner) => (
            <Card key={banner.id} variant="outlined">
              <CardContent>
                <Stack spacing={2}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Typography fontWeight={700} sx={{ flex: 1 }}>Баннер</Typography>
                    <IconButton size="small" onClick={() => moveBanner(banner.id, -1)}>
                      <ArrowUpwardIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" onClick={() => moveBanner(banner.id, 1)}>
                      <ArrowDownwardIcon fontSize="small" />
                    </IconButton>
                    <IconButton size="small" color="error" onClick={() => removeBanner(banner.id)}>
                      <DeleteIcon fontSize="small" />
                    </IconButton>
                  </Stack>

                  {banner.image_url ? (
                    <Box component="img" src={banner.image_url} alt="" sx={{ width: '100%', maxHeight: 180, objectFit: 'cover', borderRadius: 2 }} />
                  ) : null}

                  <TextField
                    label="Заголовок (необязательно)"
                    value={banner.title}
                    onChange={(e) => updateBanner(banner.id, { title: e.target.value })}
                    fullWidth
                  />
                  <TextField
                    label="URL картинки"
                    value={banner.image_url}
                    onChange={(e) => updateBanner(banner.id, { image_url: e.target.value })}
                    fullWidth
                  />
                  <MuiButton component="label" variant="outlined">
                    Загрузить картинку
                    <input
                      hidden
                      type="file"
                      accept="image/*"
                      onChange={(e) => {
                        const file = e.target.files?.[0]
                        if (!file) return
                        uploadImage(banner.id, file)
                          .then(() => notify('Картинка загружена', { type: 'success' }))
                          .catch((err) => notify(err.message, { type: 'error' }))
                      }}
                    />
                  </MuiButton>
                  <TextField
                    label="Ссылка при клике"
                    value={banner.link_url}
                    onChange={(e) => updateBanner(banner.id, { link_url: e.target.value })}
                    fullWidth
                    helperText="Например /catalog?type=game, /category/uuid, /product/uuid"
                  />
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={banner.active}
                        onChange={(e) => updateBanner(banner.id, { active: e.target.checked })}
                      />
                    }
                    label="Показывать на главной"
                  />
                </Stack>
              </CardContent>
            </Card>
          ))}
          {!banners.length ? (
            <Typography color="text.secondary">Баннеров пока нет. Нажмите «Добавить баннер».</Typography>
          ) : null}
        </Stack>
      )}
    </Box>
  )
}
