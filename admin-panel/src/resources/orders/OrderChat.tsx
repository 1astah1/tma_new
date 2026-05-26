import { useState, useEffect, useRef } from 'react'
import { useRecordContext } from 'react-admin'
import { Card, Box, Typography, TextField, Button, Divider, CircularProgress, Dialog, DialogTitle, DialogContent, DialogActions, Chip, IconButton, Accordion, AccordionSummary, AccordionDetails } from '@mui/material'
import SendIcon from '@mui/icons-material/Send'
import ArticleIcon from '@mui/icons-material/Article'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'

const categoryLabels: Record<string, string> = {
  general: 'Общие',
  greeting: 'Приветствия',
  order: 'Заказы',
}

export function OrderChat() {
  const record = useRecordContext()
  const order = record?.order || record
  const token = localStorage.getItem('token')
  const [messages, setMessages] = useState<any[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [fetching, setFetching] = useState(false)
  const [templates, setTemplates] = useState<any[]>([])
  const [templateDialogOpen, setTemplateDialogOpen] = useState(false)
  const [templateLoading, setTemplateLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const pollRef = useRef<number>()
  const msgCountRef = useRef(0)

  const orderId = order?.id

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const fetchMessages = async () => {
    if (!orderId) return
    try {
      const res = await fetch(`/api/v1/admin/orders/${orderId}/chat`, {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (res.ok) {
        const data = await res.json()
        if (data.length > msgCountRef.current) {
          msgCountRef.current = data.length
          scrollToBottom()
        }
        setMessages(data)
      }
    } catch {}
  }

  useEffect(() => {
    if (!orderId) return
    setFetching(true)
    fetchMessages().finally(() => setFetching(false))
    pollRef.current = window.setInterval(fetchMessages, 3000)
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [orderId])

  const sendMessage = async (message: string) => {
    if (!message.trim() || loading || !orderId) return
    setLoading(true)
    try {
      const res = await fetch(`/api/v1/admin/orders/${orderId}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ message: message.trim() }),
      })
      if (res.ok) {
        setInput('')
        fetchMessages()
      }
    } catch {}
    setLoading(false)
  }

  const sendTemplate = async (template: any) => {
    setInput(template.message)
    setTemplateDialogOpen(false)
  }

  const loadTemplates = async () => {
    setTemplateLoading(true)
    setTemplateDialogOpen(true)
    try {
      const res = await fetch('/api/v1/admin/templates', {
        headers: { Authorization: `Bearer ${token}` },
      })
      if (res.ok) {
        const data = await res.json()
        setTemplates(data.data || [])
      }
    } catch {}
    setTemplateLoading(false)
  }

  const handleSend = () => {
    sendMessage(input)
  }

  const formatTime = (date: string) => {
    return new Date(date).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
  }

  if (!orderId) return null

  return (
    <Card sx={{ mb: 3 }}>
      <Box sx={{ p: 2, borderBottom: '1px solid #333', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Typography variant="subtitle1" sx={{ fontWeight: 'bold', color: '#fff' }}>💬 Чат с клиентом</Typography>
        <Button
          size="small"
          startIcon={<ArticleIcon />}
          onClick={loadTemplates}
          sx={{ color: '#aaa' }}
        >
          Шаблоны
        </Button>
      </Box>

      <Box sx={{ height: 300, overflowY: 'auto', p: 2, display: 'flex', flexDirection: 'column', gap: 1.5 }}>
        {fetching && messages.length === 0 && (
          <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
            <CircularProgress size={24} />
          </Box>
        )}
        {!fetching && messages.length === 0 && (
          <Typography variant="body2" sx={{ color: '#888', textAlign: 'center', py: 4 }}>
            Сообщений пока нет
          </Typography>
        )}
        {messages.map((msg) => (
          <Box key={msg.id} sx={{ display: 'flex', justifyContent: msg.sender_type === 'admin' ? 'flex-end' : 'flex-start' }}>
            <Box sx={{
              maxWidth: '80%',
              p: 1.5,
              borderRadius: 2,
              bgcolor: msg.sender_type === 'admin' ? 'primary.main' : '#2a2a3e',
              color: msg.sender_type === 'admin' ? '#fff' : '#fff',
              borderBottomRightRadius: msg.sender_type === 'admin' ? 0 : 2,
              borderBottomLeftRadius: msg.sender_type === 'admin' ? 2 : 0,
            }}>
              <Typography variant="body2">{msg.message}</Typography>
              <Typography variant="caption" sx={{ color: msg.sender_type === 'admin' ? 'rgba(255,255,255,0.6)' : '#888', display: 'block', mt: 0.5 }}>
                {formatTime(msg.created_at)}
              </Typography>
            </Box>
          </Box>
        ))}
        <div ref={messagesEndRef} />
      </Box>

      <Divider />
      <Box sx={{ p: 2, display: 'flex', gap: 1 }}>
        <TextField
          fullWidth
          size="small"
          placeholder="Сообщение..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          sx={{
            '& .MuiOutlinedInput-root': { color: '#fff' },
            '& .MuiOutlinedInput-notchedOutline': { borderColor: '#444' },
          }}
        />
        <Button
          variant="contained"
          onClick={handleSend}
          disabled={!input.trim() || loading}
          sx={{ minWidth: 'auto', px: 2 }}
        >
          {loading ? <CircularProgress size={20} /> : <SendIcon />}
        </Button>
      </Box>

      <Dialog open={templateDialogOpen} onClose={() => setTemplateDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle sx={{ color: '#000' }}>Шаблоны сообщений</DialogTitle>
        <DialogContent>
          {templateLoading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          ) : templates.length === 0 ? (
            <Typography sx={{ color: '#666', textAlign: 'center', py: 2 }}>Нет шаблонов</Typography>
          ) : (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
              {Object.entries(categoryLabels).map(([cat, label]) => {
                const catTemplates = templates.filter(t => t.category === cat)
                if (catTemplates.length === 0) return null
                return (
                  <Accordion key={cat} defaultExpanded sx={{ bgcolor: '#f5f5f5' }}>
                    <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                      <Typography sx={{ color: '#000', fontWeight: 'medium' }}>{label}</Typography>
                      <Chip label={catTemplates.length} size="small" sx={{ ml: 1 }} />
                    </AccordionSummary>
                    <AccordionDetails>
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                        {catTemplates.map(t => (
                          <Card key={t.id} sx={{ p: 2, cursor: 'pointer', '&:hover': { bgcolor: '#e8eaf6' } }} onClick={() => sendTemplate(t)}>
                            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                              <Typography variant="body2" sx={{ fontWeight: 'medium', color: '#000' }}>{t.title}</Typography>
                              <IconButton size="small"><SendIcon fontSize="small" /></IconButton>
                            </Box>
                            <Typography variant="caption" sx={{ color: '#666', display: 'block', mt: 0.5, maxWidth: '90%', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                              {t.message}
                            </Typography>
                          </Card>
                        ))}
                      </Box>
                    </AccordionDetails>
                  </Accordion>
                )
              })}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setTemplateDialogOpen(false)}>Закрыть</Button>
        </DialogActions>
      </Dialog>
    </Card>
  )
}
