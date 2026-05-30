import { describe, it, expect } from 'vitest'
import { normalize, findLibraryMatch, GYM_ONLY_EQUIPMENT, type LibraryExercise } from './exerciseMatch'

const ex = (overrides: Partial<LibraryExercise> & { name: string }): LibraryExercise => ({
  id: 1,
  muscle_group: 'chest',
  equipment: 'barbell',
  gif_url: '/api/v1/exercises/image/001',
  aliases: [],
  ...overrides,
})

describe('normalize', () => {
  it('lowercases and splits on spaces', () => {
    expect(normalize('Bench Press')).toEqual(['bench', 'press'])
  })

  it('strips punctuation', () => {
    expect(normalize('Push-Up')).toEqual(['push', 'up'])
  })

  it('returns empty array for blank string', () => {
    expect(normalize('')).toEqual([])
  })
})

describe('findLibraryMatch', () => {
  it('returns null for empty library', () => {
    expect(findLibraryMatch('Bench Press', ['chest'], [])).toBeNull()
  })

  it('returns null for empty exercise name', () => {
    const library = [ex({ name: 'Barbell Bench Press' })]
    expect(findLibraryMatch('', ['chest'], library)).toBeNull()
  })

  it('matches on exact name', () => {
    const library = [ex({ name: 'Barbell Bench Press' })]
    const result = findLibraryMatch('Barbell Bench Press', [], library)
    expect(result?.name).toBe('Barbell Bench Press')
  })

  it('matches via alias', () => {
    const library = [ex({ name: 'Barbell Bench Press', aliases: ['bench press', 'flat bench'] })]
    const result = findLibraryMatch('bench press', [], library)
    expect(result?.name).toBe('Barbell Bench Press')
  })

  it('returns null when overlap is below 0.5 threshold', () => {
    const library = [ex({ name: 'Lateral Raise' })]
    // "Bicep Curl" shares zero words with "Lateral Raise" — overlap = 0 < 0.5
    const result = findLibraryMatch('Bicep Curl', [], library)
    expect(result).toBeNull()
  })

  it('skips exercises with no gif_url', () => {
    const library = [ex({ name: 'Barbell Bench Press', gif_url: null })]
    const result = findLibraryMatch('Barbell Bench Press', [], library)
    expect(result).toBeNull()
  })

  it('muscle group boost breaks a tie in favour of correct group', () => {
    const lateral = ex({ id: 1, name: 'Lateral Raise', muscle_group: 'shoulders', aliases: ['raise'] })
    const calf    = ex({ id: 2, name: 'Calf Raise',    muscle_group: 'lower legs', aliases: ['calf raises', 'standing calf raise'] })
    const library = [lateral, calf]

    // "Standing Calf Raise" — both share "raise", but calf has "calf" too and muscle group matches
    const result = findLibraryMatch('Standing Calf Raise', ['Lower Legs'], library)
    expect(result?.name).toBe('Calf Raise')
  })

  it('returns the best match from multiple candidates', () => {
    // "Chest Press" has no aliases — only partial name overlap with "Barbell Bench Press"
    const weak   = ex({ id: 1, name: 'Chest Press', muscle_group: 'chest', aliases: [] })
    const strong = ex({ id: 2, name: 'Barbell Bench Press', muscle_group: 'chest', aliases: ['bench press', 'flat bench'] })
    const result = findLibraryMatch('Barbell Bench Press', ['chest'], [weak, strong])
    expect(result?.name).toBe('Barbell Bench Press')
  })
})

describe('findLibraryMatch — equipment filtering', () => {
  it('excludes barbell exercises when environment is home', () => {
    const library = [ex({ name: 'Barbell Bench Press', equipment: 'barbell', muscle_group: 'chest' })]
    expect(findLibraryMatch('Barbell Bench Press', ['chest'], library, 'home')).toBeNull()
  })

  it('includes barbell exercises when environment is gym', () => {
    const library = [ex({ name: 'Barbell Bench Press', equipment: 'barbell', muscle_group: 'chest' })]
    expect(findLibraryMatch('Barbell Bench Press', ['chest'], library, 'gym')?.name).toBe('Barbell Bench Press')
  })

  it('includes bodyweight exercises at home', () => {
    const library = [ex({ name: 'Push-Up', equipment: 'none', muscle_group: 'chest' })]
    expect(findLibraryMatch('Push-Up', ['chest'], library, 'home')?.name).toBe('Push-Up')
  })

  it('excludes barbell exercises when environment is outdoor', () => {
    const library = [ex({ name: 'Barbell Squat', equipment: 'barbell', muscle_group: 'legs' })]
    expect(findLibraryMatch('Barbell Squat', ['legs'], library, 'outdoor')).toBeNull()
  })

  it('defaults to gym behavior when environment is not provided', () => {
    const library = [ex({ name: 'Barbell Bench Press', equipment: 'barbell', muscle_group: 'chest' })]
    expect(findLibraryMatch('Barbell Bench Press', ['chest'], library)?.name).toBe('Barbell Bench Press')
  })

  it('GYM_ONLY_EQUIPMENT contains the expected gym-specific equipment', () => {
    expect(GYM_ONLY_EQUIPMENT.has('barbell')).toBe(true)
    expect(GYM_ONLY_EQUIPMENT.has('cable machine')).toBe(true)
    expect(GYM_ONLY_EQUIPMENT.has('none')).toBe(false)
    expect(GYM_ONLY_EQUIPMENT.has('dumbbells')).toBe(false)
  })
})
