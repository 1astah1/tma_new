import { useEffect, useState } from 'react'
import { useNotify, useRefresh } from 'react-admin'
import {
  Box,
  Button,
  Card,
  CardContent,
  Step,
  StepLabel,
  Stepper,
  Typography,
  Stack,
  LinearProgress,
  Alert,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from '@mui/material'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import CheckCircleIcon from '@mui/icons-material/CheckCircle'

const apiUrl = '/api/v1/admin'
const STEPS = ['Импорт', 'Публикация', 'Синхронизация', 'Готово']

function authHeaders() {
  const headers = new Headers({ 'Content-Type': 'application/json' })
  const token = localStorage.getItem('token')
  if (token) headers.set('Authorization', `Bearer ${token}`)
  return headers
}

type WizardResult = {
  imported?: number
  published?: number
  linked_existing?: number
  products_synced?: number
  products_hidden?: number
  enriched?: number
}

type Props = {
  advancedPanel?: React.ReactNode
}

export function CatalogWizard({ advancedPanel }: Props) {
  const notify = useNotify()
  const refresh = useRefresh()
  const [activeStep, setActiveStep] = useState(0)
  const [running, setRunning] = useState(false)
  const [source, setSource] = useState<'ps' | 'xbox' | 'both'>('both')
  const [log, setLog] = useState<string[]>([])
  const [result, setResult] = useState<WizardResult>({})

  const pushLog = (msg: string) => setLog((prev) => [...prev, msg])

  const runFullCycle = async () => {
    setRunning(true)
    setActiveStep(0)
    setLog([])
    setResult({})
    const acc: WizardResult = {}

    try {
      setActiveStep(0)
      if (source === 'xbox' || source === 'both') {
        pushLog('Импорт Xbox...')
        const xboxRes = await fetch(`${apiUrl}/catalog-imports/import-xbox`, {
          method: 'POST',
          headers: authHeaders(),
          body: '{}',
        })
        const xboxJson = await xboxRes.json().catch(() => ({}))
        if (!xboxRes.ok) throw new Error(xboxJson.error?.message || 'Ошибка импорта Xbox')
        acc.imported = (acc.imported || 0) + (xboxJson.imported || 0)
        acc.published = (acc.published || 0) + (xboxJson.published || 0)
        acc.products_synced = (acc.products_synced || 0) + (xboxJson.products_synced || 0)
        pushLog(`Xbox: +${xboxJson.imported || 0} импорт, опубликовано ${xboxJson.published || 0}`)
      }

      if (source === 'ps' || source === 'both') {
        pushLog('Импорт PS Store (фон)...')
        const psRes = await fetch(`${apiUrl}/catalog-imports/import-ps`, {
          method: 'POST',
          headers: authHeaders(),
          body: '{}',
        })
        const psJson = await psRes.json().catch(() => ({}))
        if (!psRes.ok) throw new Error(psJson.error?.message || 'Ошибка импорта PS')
        if (psJson.status === 'started') {
          pushLog('PS Store: импорт запущен в фоне')
        } else {
          acc.imported = (acc.imported || 0) + (psJson.imported || 0)
          pushLog(`PS: +${psJson.imported || 0} импорт`)
        }
      }

      setActiveStep(1)
      pushLog('Публикация ожидающих...')
      const pubRes = await fetch(`${apiUrl}/catalog-imports/republish-pending`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const pubJson = await pubRes.json().catch(() => ({}))
      if (!pubRes.ok) throw new Error(pubJson.error?.message || 'Ошибка публикации')
      acc.published = (acc.published || 0) + (pubJson.published || 0)
      acc.linked_existing = (acc.linked_existing || 0) + (pubJson.linked_existing || 0)
      pushLog(`Опубликовано: ${pubJson.published || 0}, связано: ${pubJson.linked_existing || 0}`)

      setActiveStep(2)
      pushLog('Синхронизация цен и дат...')
      const syncRes = await fetch(`${apiUrl}/catalog-imports/refresh-catalog?sync=1`, {
        method: 'POST',
        headers: authHeaders(),
        body: '{}',
      })
      const syncJson = await syncRes.json().catch(() => ({}))
      if (!syncRes.ok) throw new Error(syncJson.error?.message || 'Ошибка синхронизации')
      acc.products_synced = (acc.products_synced || 0) + (syncJson.products_synced || 0)
      acc.products_hidden = (acc.products_hidden || 0) + (syncJson.products_hidden || 0)
      acc.enriched = (acc.enriched || 0) + (syncJson.enriched || 0)
      acc.published = (acc.published || 0) + (syncJson.published || 0)
      pushLog(
        `Синхр.: цен ${syncJson.products_synced || 0}, скрыто ${syncJson.products_hidden || 0}, опубликовано ${syncJson.published || 0}`,
      )

      setResult(acc)
      setActiveStep(3)
      notify('Мастер каталога завершён', { type: 'success' })
      refresh()
    } catch (e: any) {
      notify(e.message, { type: 'error' })
      pushLog(`Ошибка: ${e.message}`)
    } finally {
      setRunning(false)
    }
  }

  useEffect(() => {
    if (!running && activeStep === 3) return
  }, [running, activeStep])

  return (
    <Card sx={{ mb: 2 }}>
      <CardContent>
        <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 2 }}>
          Мастер каталога
        </Typography>
        <Stepper activeStep={activeStep} alternativeLabel sx={{ mb: 3 }}>
          {STEPS.map((label) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>

        <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
          Один сценарий: импорт из магазинов → публикация в товары → синхронизация цен.
        </Typography>

        <Stack direction="row" spacing={1} sx={{ mb: 2 }} flexWrap="wrap" useFlexGap>
          {(['ps', 'xbox', 'both'] as const).map((opt) => (
            <Button
              key={opt}
              size="small"
              variant={source === opt ? 'contained' : 'outlined'}
              onClick={() => setSource(opt)}
              disabled={running}
            >
              {opt === 'ps' ? 'PS Store' : opt === 'xbox' ? 'Xbox' : 'PS + Xbox'}
            </Button>
          ))}
        </Stack>

        <Button
          variant="contained"
          color="primary"
          startIcon={running ? undefined : <PlayArrowIcon />}
          onClick={runFullCycle}
          disabled={running}
          sx={{ mb: 2 }}
        >
          {running ? 'Выполняется...' : 'Запустить полный цикл'}
        </Button>

        {running && <LinearProgress sx={{ mb: 2 }} />}

        {log.length > 0 && (
          <Box sx={{ mb: 2, p: 1.5, bgcolor: '#111', borderRadius: 1, maxHeight: 140, overflow: 'auto' }}>
            {log.map((line, i) => (
              <Typography key={i} variant="caption" display="block" color="text.secondary">
                {line}
              </Typography>
            ))}
          </Box>
        )}

        {activeStep === 3 && !running && (
          <Alert severity="success" icon={<CheckCircleIcon />} sx={{ mb: 2 }}>
            Импортировано: {result.imported ?? '—'} · Опубликовано: {result.published ?? '—'} ·
            Синхронизировано цен: {result.products_synced ?? '—'}
            {result.products_hidden ? ` · Скрыто: ${result.products_hidden}` : ''}
          </Alert>
        )}

        {advancedPanel ? (
          <Accordion disableGutters elevation={0} sx={{ bgcolor: 'transparent', '&:before': { display: 'none' } }}>
            <AccordionSummary expandIcon={<ExpandMoreIcon />}>
              <Typography variant="body2" color="text.secondary">
                Расширенные действия
              </Typography>
            </AccordionSummary>
            <AccordionDetails sx={{ px: 0 }}>{advancedPanel}</AccordionDetails>
          </Accordion>
        ) : null}
      </CardContent>
    </Card>
  )
}
