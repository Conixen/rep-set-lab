export interface LibraryExercise {
  id: number
  name: string
  muscle_group: string
  equipment: string
  gif_url: string | null
  aliases: string[]
}

// Equipment that only exists in a gym. Exercises using these are excluded
// from GIF matches when the user's environment is 'home' or 'outdoor'.
export const GYM_ONLY_EQUIPMENT = new Set([
  'barbell',
  'cable machine',
  'machine',
  'leg press machine',
  'dip bars',
])

export function normalize(s: string): string[] {
  return s.toLowerCase().replace(/[^a-z0-9 ]/g, ' ').split(' ').filter(Boolean)
}

export function findLibraryMatch(
  aiName: string,
  muscleHints: string[],
  library: LibraryExercise[],
  environment = 'gym'
): LibraryExercise | null {
  const aiWords = normalize(aiName)
  if (aiWords.length === 0) return null

  let best: { ex: LibraryExercise; score: number } | null = null

  for (const ex of library) {
    if (!ex.gif_url) continue
    if (environment !== 'gym' && GYM_ONLY_EQUIPMENT.has(ex.equipment)) continue

    const candidates = [ex.name, ...(ex.aliases ?? [])]
    let maxOverlap = 0

    for (const candidate of candidates) {
      const exWords = normalize(candidate)
      if (exWords.length === 0) continue
      const [shorter, longer] = aiWords.length <= exWords.length
        ? [aiWords, exWords]
        : [exWords, aiWords]
      const overlap = shorter.filter(w => longer.includes(w)).length / shorter.length
      if (overlap > maxOverlap) maxOverlap = overlap
    }

    if (maxOverlap < 0.5) continue

    const mgMatch = muscleHints.some(mg =>
      ex.muscle_group.toLowerCase().includes(mg.toLowerCase()) ||
      mg.toLowerCase().includes(ex.muscle_group.toLowerCase())
    )
    const score = maxOverlap + (mgMatch ? 0.3 : 0)
    if (!best || score > best.score) best = { ex, score }
  }

  return best?.ex ?? null
}
