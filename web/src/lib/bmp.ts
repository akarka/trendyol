// Canvas'ı 1-bit monokrom Windows BMP'ye encode eder. 1 bit/piksel ESC/POS raster'a doğrudan eşlenir.

export function canvasToMonoBmp(canvas: HTMLCanvasElement, threshold = 128): Blob {
  const ctx = canvas.getContext('2d')!
  const { width, height } = canvas
  const { data } = ctx.getImageData(0, 0, width, height)

  const rowBytes = Math.floor((width + 31) / 32) * 4 // BMP satırları 4 byte hizalı
  const pixelDataSize = rowBytes * height
  const offset = 14 + 40 + 8 // file header + info header + 2 renkli palet
  const fileSize = offset + pixelDataSize

  const buf = new ArrayBuffer(fileSize)
  const view = new DataView(buf)
  let p = 0

  view.setUint8(p++, 0x42)
  view.setUint8(p++, 0x4d)
  view.setUint32(p, fileSize, true); p += 4
  view.setUint32(p, 0, true); p += 4
  view.setUint32(p, offset, true); p += 4

  view.setUint32(p, 40, true); p += 4
  view.setInt32(p, width, true); p += 4
  view.setInt32(p, height, true); p += 4 // pozitif → satırlar alttan üste
  view.setUint16(p, 1, true); p += 2
  view.setUint16(p, 1, true); p += 2 // bitcount = 1
  view.setUint32(p, 0, true); p += 4 // BI_RGB
  view.setUint32(p, pixelDataSize, true); p += 4
  view.setInt32(p, 0, true); p += 4
  view.setInt32(p, 0, true); p += 4
  view.setUint32(p, 2, true); p += 4
  view.setUint32(p, 2, true); p += 4

  view.setUint32(p, 0x00000000, true); p += 4 // palet[0] siyah
  view.setUint32(p, 0x00ffffff, true); p += 4 // palet[1] beyaz

  const bytes = new Uint8Array(buf)
  for (let y = 0; y < height; y++) {
    const srcRow = (height - 1 - y) * width
    const dstRow = offset + y * rowBytes
    for (let x = 0; x < width; x++) {
      const i = (srcRow + x) * 4
      const lum = (data[i] * 299 + data[i + 1] * 587 + data[i + 2] * 114) / 1000
      if (lum >= threshold) bytes[dstRow + (x >> 3)] |= 0x80 >> (x & 7) // açık → beyaz biti (1)
    }
  }

  return new Blob([buf], { type: 'image/bmp' })
}
