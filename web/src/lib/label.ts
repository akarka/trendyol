// Etiketi output.txt düzeninde monokrom bir bitmap (canvas) olarak üretir ve yazdırır.
// Aynı bitmap thermal'a geçişte ESC/POS raster için yeniden kullanılabilir.

import { cellPosition, LabelLayout } from './labelLayout'

export interface LabelLineItem {
  quantity: number
  name: string
  price: number
}

const SEP = '========================================'
const SUB = '----------------------------------------'

function pad(n: number): string {
  return String(n).padStart(2, '0')
}

export function buildLabelLines(opts: {
  orderNumber: string
  customer: string
  status?: string
  items: LabelLineItem[]
}): string[] {
  const now = new Date()
  const ts = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(
    now.getHours(),
  )}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`

  let total = 0
  const itemLines = opts.items.map((l) => {
    total += l.price * l.quantity
    return `${l.quantity} x ${l.name} (${l.price.toFixed(2)} TL)`
  })

  return [
    SEP,
    `Tarih      : ${ts}`,
    `Sipariş No : ${opts.orderNumber}`,
    `Durum      : ${opts.status ?? 'Yeni Sipariş'}`,
    `Müşteri    : ${opts.customer || 'Bilinmiyor'}`,
    SUB,
    'Ürünler:',
    ...itemLines,
    SUB,
    `Toplam     : ${total.toFixed(2)} TL`,
    SEP,
  ]
}

export function renderLabelCanvas(lines: string[], scale = 2): HTMLCanvasElement {
  const fontSize = 28
  const lineH = Math.round(fontSize * 1.35)
  const margin = 32
  const font = `${fontSize}px "Courier New", "Consolas", monospace`

  const measureCtx = document.createElement('canvas').getContext('2d')!
  measureCtx.font = font
  let maxW = 0
  for (const l of lines) maxW = Math.max(maxW, measureCtx.measureText(l).width)

  const w = Math.ceil(maxW + margin * 2)
  const h = lineH * lines.length + margin * 2

  const canvas = document.createElement('canvas')
  canvas.width = w * scale
  canvas.height = h * scale
  const ctx = canvas.getContext('2d')!
  ctx.scale(scale, scale)
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, w, h)
  ctx.fillStyle = '#000000'
  ctx.font = font
  ctx.textBaseline = 'top'

  let y = margin
  for (const l of lines) {
    ctx.fillText(l, margin, y)
    y += lineH
  }
  return canvas
}

export function labelDataUrl(lines: string[]): string {
  return renderLabelCanvas(lines).toDataURL('image/png')
}

// printLabel, etiketi A4 sayfasında verilen yerleşime göre seçilen hücreye (cellIndex) bastırır.
// Sadece o hücre dolar; şablonlu sticker kağıdında diğer etiketler boş/yeniden kullanılabilir kalır.
export function printLabel(lines: string[], layout: LabelLayout, cellIndex: number): void {
  const dataUrl = labelDataUrl(lines)
  const pos = cellPosition(layout, cellIndex)
  const pad = layout.paddingMm

  const win = window.open('', '_blank')
  if (!win) {
    alert('Yazdırma penceresi açılamadı. Popup engelini kaldırın.')
    return
  }
  win.document.write(
    `<!doctype html><html><head><meta charset="utf-8"><title>Etiket</title><style>` +
      `@page{size:${layout.pageWidthMm}mm ${layout.pageHeightMm}mm;margin:0}` +
      `html,body{margin:0;padding:0}` +
      `.sheet{position:relative;width:${layout.pageWidthMm}mm;height:${layout.pageHeightMm}mm}` +
      `.label{position:absolute;left:${pos.xMm}mm;top:${pos.yMm}mm;` +
      `width:${layout.labelWidthMm}mm;height:${layout.labelHeightMm}mm;` +
      `box-sizing:border-box;padding:${pad}mm;overflow:hidden;` +
      `display:flex;align-items:center;justify-content:center}` +
      `.label img{max-width:100%;max-height:100%}` +
      `</style></head>` +
      `<body><div class="sheet"><div class="label">` +
      `<img src="${dataUrl}" onload="window.focus();window.print();setTimeout(function(){window.close()},300)">` +
      `</div></div></body></html>`,
  )
  win.document.close()
}
