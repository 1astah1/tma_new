import { DataProvider, fetchUtils } from 'react-admin'

const apiUrl = '/api/v1/admin'
const httpClient = (url: string, options: fetchUtils.Options = {}) => {
  const token = localStorage.getItem('token')
  const headers = new Headers(options.headers as HeadersInit)
  if (token) headers.set('Authorization', `Bearer ${token}`)
  options.headers = headers
  return fetchUtils.fetchJson(url, options)
}

export const dataProvider: DataProvider = {
  getList: async (resource, params) => {
    const { page = 1, perPage = 20 } = params.pagination || {}
    const { field, order } = params.sort || {}
    const q = new URLSearchParams()
    q.set('page', String(page))
    q.set('limit', String(perPage))
    if (field) q.set('sort', field)
    if (order) q.set('order', order)
    const filter = params.filter || {}
    Object.entries(filter).forEach(([k, v]) => {
      if (v !== undefined && v !== '') q.set(k, String(v))
    })
    const { json } = await httpClient(`${apiUrl}/${resource}?${q}`)
    return { data: json.data, total: json.meta?.total || json.data?.length || 0 }
  },

  getOne: async (resource, params) => {
    const { json } = await httpClient(`${apiUrl}/${resource}/${params.id}`)
    const order = json.order || json
    const history = json.history || []
    const account = json.account || null
    const key = json.key || null
    const assigned_admin = json.assigned_admin || null
    return { data: { ...order, history, account, key, assigned_admin } }
  },

  create: async (resource, params) => {
    const { json } = await httpClient(`${apiUrl}/${resource}`, {
      method: 'POST',
      body: JSON.stringify(params.data),
    })
    return { data: { ...params.data, id: json.id } as any }
  },

  update: async (resource, params) => {
    const { json } = await httpClient(`${apiUrl}/${resource}/${params.id}`, {
      method: 'PUT',
      body: JSON.stringify(params.data),
    })
    return { data: { ...json, id: params.id } }
  },

  delete: async (resource, params) => {
    await httpClient(`${apiUrl}/${resource}/${params.id}`, { method: 'DELETE' })
    return { data: { id: params.id } as any }
  },

  getMany: async (resource, params) => {
    const promises = params.ids.map((id) =>
      httpClient(`${apiUrl}/${resource}/${id}`).then((r) => r.json)
    )
    const data = await Promise.all(promises)
    return { data }
  },

  getManyReference: async (resource, params) => {
    return dataProvider.getList(resource, {
      pagination: params.pagination,
      sort: params.sort,
      filter: { ...params.filter, [params.target]: params.id },
    })
  },

  updateMany: async (resource, params) => {
    const promises = params.ids.map((id) =>
      httpClient(`${apiUrl}/${resource}/${id}`, {
        method: 'PUT',
        body: JSON.stringify(params.data),
      })
    )
    await Promise.all(promises)
    return { data: params.ids }
  },

  deleteMany: async (resource, params) => {
    const promises = params.ids.map((id) =>
      httpClient(`${apiUrl}/${resource}/${id}`, { method: 'DELETE' })
    )
    await Promise.all(promises)
    return { data: params.ids }
  },
}
