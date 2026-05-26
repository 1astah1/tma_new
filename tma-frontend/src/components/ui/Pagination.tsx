interface Props {
  page: number
  total: number
  limit: number
  onPageChange: (page: number) => void
}

export function Pagination({ page, total, limit, onPageChange }: Props) {
  const totalPages = Math.ceil(total / limit)
  if (totalPages <= 1) return null

  return (
    <div className="flex items-center justify-center gap-2 mt-4">
      <button
        onClick={() => onPageChange(page - 1)}
        disabled={page <= 1}
        className="px-3 py-1.5 rounded-lg bg-[var(--tg-secondary)] text-sm disabled:opacity-30 hover:bg-[var(--tg-card)]"
      >
        ← Назад
      </button>
      <span className="text-sm text-[var(--tg-hint)]">
        {page} / {totalPages}
      </span>
      <button
        onClick={() => onPageChange(page + 1)}
        disabled={page >= totalPages}
        className="px-3 py-1.5 rounded-lg bg-[var(--tg-secondary)] text-sm disabled:opacity-30 hover:bg-[var(--tg-card)]"
      >
        Далее →
      </button>
    </div>
  )
}
