import { cellCount, cellPosition, LabelLayout } from '../lib/labelLayout'

const DISPLAY_W = 240 // px

// A4 sayfasını ve etiket hücrelerini ölçekli gösterir. onSelect verilirse hücreler tıklanabilir.
export function SheetPreview({
  layout,
  selected,
  onSelect,
}: {
  layout: LabelLayout
  selected?: number
  onSelect?: (index: number) => void
}) {
  const scale = DISPLAY_W / layout.pageWidthMm
  const n = cellCount(layout)

  return (
    <div
      className="relative shrink-0 border border-gray-400 bg-white"
      style={{ width: layout.pageWidthMm * scale, height: layout.pageHeightMm * scale }}
    >
      {Array.from({ length: n }).map((_, i) => {
        const p = cellPosition(layout, i)
        const isSel = selected === i
        return (
          <button
            key={i}
            type="button"
            onClick={onSelect ? () => onSelect(i) : undefined}
            disabled={!onSelect}
            className={`absolute flex items-center justify-center border text-[9px] ${
              isSel
                ? 'border-gray-900 bg-gray-900 text-white'
                : 'border-gray-300 text-gray-400 hover:bg-gray-100'
            } ${onSelect ? 'cursor-pointer' : 'cursor-default'}`}
            style={{
              left: p.xMm * scale,
              top: p.yMm * scale,
              width: layout.labelWidthMm * scale,
              height: layout.labelHeightMm * scale,
            }}
          >
            {i + 1}
          </button>
        )
      })}
    </div>
  )
}
