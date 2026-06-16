import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import sv from './locales/sv.json'

const validLocales = ['en', 'sv'] as const
type Locale = typeof validLocales[number]
const raw = localStorage.getItem('app-locale')
const savedLocale: Locale = validLocales.includes(raw as Locale) ? (raw as Locale) : 'en'

export const i18n = createI18n({
  legacy: false,
  locale: savedLocale,
  fallbackLocale: 'en',
  messages: { en, sv },
})

export function setLocale(lang: Locale) {
  i18n.global.locale.value = lang
  localStorage.setItem('app-locale', lang)
}
