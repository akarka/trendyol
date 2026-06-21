// A4 şablonlu sticker kağıdı için etiket yerleşim geometrisi.
// settings tablosunda `label_layout` anahtarı altında JSON olarak saklanır.

export interface LabelLayout {
  pageWidthMm: number
  pageHeightMm: number
  columns: number
  rows: number
  labelWidthMm: number
  labelHeightMm: number
  marginTopMm: number
  marginLeftMm: number
  gapXMm: number
  gapYMm: number
  paddingMm: number // hücre içi güvenlik boşluğu (kesim kenarından)
}

// Varsayılan: A4'ü tam dolduran 3×8 = 24 etiket (70 × 37.125 mm), boşluksuz.
export const DEFAULT_LAYOUT: LabelLayout = {
  pageWidthMm: 210,
  pageHeightMm: 297,
  columns: 3,
  rows: 8,
  labelWidthMm: 70,
  labelHeightMm: 37.125,
  marginTopMm: 0,
  marginLeftMm: 0,
  gapXMm: 0,
  gapYMm: 0,
  paddingMm: 2,
}

export const LABEL_LAYOUT_KEY = 'label_layout'

export function parseLayout(settings: Record<string, string> | undefined): LabelLayout {
  const raw = settings?.[LABEL_LAYOUT_KEY]
  if (!raw) return DEFAULT_LAYOUT
  try {
    return { ...DEFAULT_LAYOUT, ...(JSON.parse(raw) as Partial<LabelLayout>) }
  } catch {
    return DEFAULT_LAYOUT
  }
}

export function cellCount(l: LabelLayout): number {
  return Math.max(0, Math.floor(l.columns)) * Math.max(0, Math.floor(l.rows))
}

export interface CellPos {
  xMm: number
  yMm: number
}

// 0-indeksli hücrenin sayfadaki sol-üst konumu (mm).
export function cellPosition(l: LabelLayout, index: number): CellPos {
  const col = index % l.columns
  const row = Math.floor(index / l.columns)
  return {
    xMm: l.marginLeftMm + col * (l.labelWidthMm + l.gapXMm),
    yMm: l.marginTopMm + row * (l.labelHeightMm + l.gapYMm),
  }
}
