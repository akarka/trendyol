import { useEffect, useRef } from 'react'

// Dış sarmalayıcının ekran ebadı SABİT (mm büyüdükçe/küçüldükçe div büyümez/küçülmez);
// içerideki hücre kutusu bu sabit alana en-boy oranı korunarak ("contain") sığdırılır —
// yani gösterilen gerçek bir ölçek değil, 1 hücre ile baskı alanı arasındaki ORANdır.
const DISPLAY_W = 280
const DISPLAY_H = 180

// Dış kutu, sayfadan türetilen hücre ebadının (labelWidthMm × labelHeightMm) oranını yansıtır.
// İçindeki kutu tarayıcının yerleşik resize tutamacıyla sürüklenebilir; boyutu, hücre
// kenarından baskı alanına bırakılan boşluğu (paddingMm) belirler.
export function LabelPaddingBox({
  labelWidthMm,
  labelHeightMm,
  paddingMm,
  onPaddingChange,
  previewUrl,
}: {
  labelWidthMm: number
  labelHeightMm: number
  paddingMm: number
  onPaddingChange: (mm: number) => void
  previewUrl: string
}) {
  const innerRef = useRef<HTMLDivElement>(null)
  const scale = Math.min(DISPLAY_W / labelWidthMm, DISPLAY_H / labelHeightMm)
  const outerWpx = labelWidthMm * scale
  const outerHpx = labelHeightMm * scale

  useEffect(() => {
    const el = innerRef.current
    if (!el) return
    const obs = new ResizeObserver(() => {
      const wMm = el.offsetWidth / scale
      const hMm = el.offsetHeight / scale
      const padW = (labelWidthMm - wMm) / 2
      const padH = (labelHeightMm - hMm) / 2
      const next = Math.max(0, Math.round(Math.min(padW, padH) * 10) / 10)
      onPaddingChange(next)
    })
    obs.observe(el)
    return () => obs.disconnect()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [labelWidthMm, labelHeightMm, scale])

  const innerW = Math.max(10, outerWpx - paddingMm * 2 * scale)
  const innerH = Math.max(10, outerHpx - paddingMm * 2 * scale)

  return (
    <div
      className="mx-auto flex items-center justify-center"
      style={{ width: DISPLAY_W, height: DISPLAY_H }}
    >
      <div
        className="relative flex items-center justify-center overflow-hidden border border-gray-400 bg-white"
        style={{ width: outerWpx, height: outerHpx }}
      >
        <div
          ref={innerRef}
          className="flex resize items-center justify-center overflow-hidden border-2 border-dashed border-gray-400 bg-gray-50"
          style={{
            width: innerW,
            height: innerH,
            maxWidth: outerWpx,
            maxHeight: outerHpx,
            minWidth: 10,
            minHeight: 10,
          }}
        >
          {previewUrl ? (
            <img
              src={previewUrl}
              alt="etiket önizleme"
              className="max-h-full max-w-full object-contain"
            />
          ) : (
            <span className="px-1 text-center text-[10px] text-gray-400">
              Ürün seçince önizleme görünür
            </span>
          )}
        </div>
      </div>
    </div>
  )
}
