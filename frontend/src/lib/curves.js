export function computeSpline(points) {
  const n = points.length - 1
  if (n === 0) return () => points[0].y
  if (n === 1) {
    return (x) => {
      const p0 = points[0]
      const p1 = points[1]
      return p0.y + (p1.y - p0.y) * ((x - p0.x) / (p1.x - p0.x))
    }
  }

  const a = points.map(p => p.y)
  const h = []
  for (let i = 0; i < n; i++) h.push(points[i + 1].x - points[i].x)

  const alpha = [0]
  for (let i = 1; i < n; i++) {
    alpha.push((3 / h[i]) * (a[i + 1] - a[i]) - (3 / h[i - 1]) * (a[i] - a[i - 1]))
  }

  const c = new Array(n + 1).fill(0)
  const l = new Array(n + 1).fill(1)
  const mu = new Array(n + 1).fill(0)
  const z = new Array(n + 1).fill(0)

  for (let i = 1; i < n; i++) {
    l[i] = 2 * (points[i + 1].x - points[i - 1].x) - h[i - 1] * mu[i - 1]
    mu[i] = h[i] / l[i]
    z[i] = (alpha[i] - h[i - 1] * z[i - 1]) / l[i]
  }

  const b = new Array(n).fill(0)
  const d = new Array(n).fill(0)

  for (let j = n - 1; j >= 0; j--) {
    c[j] = z[j] - mu[j] * c[j + 1]
    b[j] = (a[j + 1] - a[j]) / h[j] - h[j] * (c[j + 1] + 2 * c[j]) / 3
    d[j] = (c[j + 1] - c[j]) / (3 * h[j])
  }

  return (x) => {
    if (x <= points[0].x) return a[0]
    if (x >= points[n].x) return a[n]
    let j = 0
    for (let i = 0; i < n; i++) {
      if (x >= points[i].x && x <= points[i + 1].x) {
        j = i
        break
      }
    }
    const dx = x - points[j].x
    return a[j] + b[j] * dx + c[j] * dx * dx + d[j] * dx * dx * dx
  }
}

export function parseCurvesFilter(curvesStr) {
  const result = { master: [], red: [], green: [], blue: [] }
  if (!curvesStr) return result

  const parts = curvesStr.split(':')
  for (const part of parts) {
    const [chName, pathStr] = part.split('=')
    if (!chName || !pathStr) continue
    const cleanPath = pathStr.replace(/'/g, '')
    const pts = cleanPath.split(' ').filter(Boolean).map(str => {
      const [x, y] = str.split('/')
      return { x: parseFloat(x) * 100, y: parseFloat(y) * 100 }
    })
    if (result[chName]) result[chName] = pts
  }
  return result
}

function getValuesFromPoints(points) {
  if (!points || points.length === 0) {
    // Identity mapping
    const vals = []
    for (let i = 0; i < 256; i++) vals.push(i / 255)
    return vals
  }
  const spline = computeSpline(points)
  const vals = []
  for (let i = 0; i < 256; i++) {
    const x = (i / 255) * 100
    const y = spline(x)
    vals.push(Math.max(0, Math.min(100, y)) / 100)
  }
  return vals
}

export function getCombinedTables(curvesStr) {
  const parsed = parseCurvesFilter(curvesStr)
  
  const masterVals = getValuesFromPoints(parsed.master)
  const redVals = getValuesFromPoints(parsed.red)
  const greenVals = getValuesFromPoints(parsed.green)
  const blueVals = getValuesFromPoints(parsed.blue)

  const finalRed = []
  const finalGreen = []
  const finalBlue = []

  for (let i = 0; i < 256; i++) {
    // Apply channel curve, then master curve
    const rIdx = Math.round(redVals[i] * 255)
    finalRed.push(masterVals[Math.max(0, Math.min(255, rIdx))].toFixed(4))

    const gIdx = Math.round(greenVals[i] * 255)
    finalGreen.push(masterVals[Math.max(0, Math.min(255, gIdx))].toFixed(4))

    const bIdx = Math.round(blueVals[i] * 255)
    finalBlue.push(masterVals[Math.max(0, Math.min(255, bIdx))].toFixed(4))
  }

  return {
    r: finalRed.join(' '),
    g: finalGreen.join(' '),
    b: finalBlue.join(' ')
  }
}
