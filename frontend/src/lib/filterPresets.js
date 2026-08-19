/**
 * Filter presets shared editor-wide. Values are the exact color grade
 * offsets also used by the inspector (brightness/contrast/saturation/hue/temp).
 */
export const FILTER_PRESETS = [
  { name: 'Natural',   grade: { brightness: 0,  contrast: 5,   saturation: 0,   hue: 0 } },
  { name: 'Cinema',    grade: { brightness: -5, contrast: 15,  saturation: -10, hue: 0 } },
  { name: 'Warm',      grade: { brightness: 5,  contrast: 0,   saturation: 10,  hue: 5,  temp: 10 } },
  { name: 'Cool',      grade: { brightness: 0,  contrast: 5,   saturation: -5,  hue: -5, temp: -10 } },
  { name: 'Vintage',   grade: { brightness: 5,  contrast: -5,  saturation: -25, hue: 10, temp: 15 } },
  { name: 'B&W',         grade: { saturation: -100 } },
  { name: 'Fade',        grade: { brightness: 15, contrast: -20, saturation: -30 } },
  { name: 'Vivid',       grade: { brightness: 5,  contrast: 20,  saturation: 30 } },
  { name: 'Matte',       grade: { brightness: 5,  contrast: -10, saturation: -15,  hue: -5 } },
  { name: 'Golden Hour', grade: { brightness: 10, contrast: 0,   saturation: 15,   hue: 8,  temp: 20 } },
  { name: 'Cyberpunk',   grade: { brightness: 0,  contrast: 20,  saturation: 25,   hue: -15, temp: -20 } },
  { name: 'Soft',        grade: { brightness: 8,  contrast: -10, saturation: -10 } },
]

/**
 * Approximate a color grade as CSS filter functions, good enough for
 * distinguishable preset thumbnails.
 *
 * @param {object} grade { brightness, contrast, saturation, hue, temp }
 * @returns {{ filter: string[], opacity: number }}
 */
export function cssFilterForGrade(grade = {}) {
  const brightness = 1 + (grade.brightness || 0) / 100
  const contrast = 1 + (grade.contrast || 0) / 100
  const saturation = 1 + (grade.saturation || 0) / 100
  const hue = grade.hue || 0
  const temp = grade.temp || 0

  const filters = []
  if (brightness !== 1) filters.push(`brightness(${brightness.toFixed(2)})`)
  if (contrast !== 1) filters.push(`contrast(${contrast.toFixed(2)})`)

  if (saturation <= 0.05) {
    filters.push('grayscale(1)')
  } else if (saturation !== 1) {
    filters.push(`saturate(${saturation.toFixed(2)})`)
  }

  if (hue !== 0) filters.push(`hue-rotate(${hue}deg)`)

  // Simple warm/cool tinting: warm shifts toward sepia/amber, cool toward blue.
  if (temp > 0) {
    filters.push(`sepia(${Math.min(0.6, temp / 50).toFixed(2)})`)
    filters.push('hue-rotate(-8deg)')
  } else if (temp < 0) {
    filters.push('hue-rotate(190deg)')
    filters.push(`saturate(${(1 + Math.min(0.35, Math.abs(temp) / 60)).toFixed(2)})`)
  }

  // Very bright "fade"-style looks read better slightly washed out.
  const opacity = brightness > 1.1 ? 0.94 : 1

  return { filter: filters.length ? filters : ['none'], opacity }
}
