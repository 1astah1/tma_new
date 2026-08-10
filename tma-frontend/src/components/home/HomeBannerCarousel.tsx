import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { HomeBanner } from '../../types/home'

const AUTOPLAY_MS = 4500

export function HomeBannerCarousel({ banners }: { banners: HomeBanner[] }) {
  const nav = useNavigate()
  const scrollRef = useRef<HTMLDivElement>(null)
  const [activeIndex, setActiveIndex] = useState(0)
  const pauseUntilRef = useRef(0)

  const scrollToIndex = useCallback((index: number) => {
    const el = scrollRef.current
    if (!el || !banners.length) return
    const next = ((index % banners.length) + banners.length) % banners.length
    el.scrollTo({ left: next * el.clientWidth, behavior: 'smooth' })
    setActiveIndex(next)
  }, [banners.length])

  useEffect(() => {
    if (banners.length <= 1) return
    const timer = window.setInterval(() => {
      if (Date.now() < pauseUntilRef.current) return
      scrollToIndex(activeIndex + 1)
    }, AUTOPLAY_MS)
    return () => window.clearInterval(timer)
  }, [activeIndex, banners.length, scrollToIndex])

  const onScroll = () => {
    const el = scrollRef.current
    if (!el || !el.clientWidth) return
    const index = Math.round(el.scrollLeft / el.clientWidth)
    if (index !== activeIndex) setActiveIndex(index)
  }

  const pauseAutoplay = () => {
    pauseUntilRef.current = Date.now() + AUTOPLAY_MS * 2
  }

  const openBanner = (banner: HomeBanner) => {
    if (!banner.link_url) return
    const link = banner.link_url.trim()
    if (link.startsWith('http://') || link.startsWith('https://')) {
      window.open(link, '_blank', 'noopener,noreferrer')
      return
    }
    nav(link.startsWith('/') ? link : `/${link}`)
  }

  if (!banners.length) return null

  return (
    <section className="mb-4" aria-label="Баннеры">
      <div className="relative">
        <div
          ref={scrollRef}
          onScroll={onScroll}
          onTouchStart={pauseAutoplay}
          onMouseDown={pauseAutoplay}
          className="flex snap-x snap-mandatory items-start overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
        >
          {banners.map((banner, index) => (
            <button
              key={banner.id}
              type="button"
              onClick={() => openBanner(banner)}
              className={`block w-full shrink-0 snap-center p-0 leading-[0] ${
                banner.link_url ? 'cursor-pointer' : 'cursor-default'
              }`}
            >
              <img
                src={banner.image_url}
                alt={banner.title || 'Баннер'}
                width={banner.width}
                height={banner.height}
                className="block h-auto w-full"
                draggable={false}
                decoding={index === 0 ? 'sync' : 'async'}
                fetchPriority={index === 0 ? 'high' : 'auto'}
              />
              {banner.title ? (
                <div className="bg-gradient-to-t from-black/70 to-transparent px-4 pb-3 pt-2 text-left leading-normal">
                  <span className="line-clamp-2 text-sm font-bold text-white">{banner.title}</span>
                </div>
              ) : null}
            </button>
          ))}
        </div>

        {banners.length > 1 ? (
          <div className="pointer-events-none absolute inset-x-0 bottom-2 z-10 flex justify-center gap-1.5 px-4">
            {banners.map((banner, index) => (
              <button
                key={banner.id}
                type="button"
                aria-label={`Баннер ${index + 1}`}
                onClick={() => {
                  pauseAutoplay()
                  scrollToIndex(index)
                }}
                className={`pointer-events-auto h-1.5 rounded-full shadow-sm transition-all ${
                  index === activeIndex ? 'w-5 bg-amber-400' : 'w-1.5 bg-white/60'
                }`}
              />
            ))}
          </div>
        ) : null}
      </div>
    </section>
  )
}
