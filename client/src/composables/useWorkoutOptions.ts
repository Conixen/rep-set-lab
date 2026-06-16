import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

export function useWorkoutOptions() {
  const { t } = useI18n()

  const environmentOptions = computed(() => [
    { key: 'gym',     label: t('environments.gym') },
    { key: 'home',    label: t('environments.home') },
    { key: 'outdoor', label: t('environments.outdoor') },
  ])

  const muscleGroupOptions = computed(() => [
    { key: 'Chest',     label: t('ai.muscleGroups.chest') },
    { key: 'Back',      label: t('ai.muscleGroups.back') },
    { key: 'Legs',      label: t('ai.muscleGroups.legs') },
    { key: 'Shoulders', label: t('ai.muscleGroups.shoulders') },
    { key: 'Arms',      label: t('ai.muscleGroups.arms') },
    { key: 'Core',      label: t('ai.muscleGroups.core') },
    { key: 'Calves',    label: t('ai.muscleGroups.calves') },
    { key: 'Wrists',    label: t('ai.muscleGroups.wrists') },
  ])

  const levelOptions = computed(() => [
    { key: 'Beginner',     label: t('ai.experience.beginner') },
    { key: 'Intermediate', label: t('ai.experience.intermediate') },
    { key: 'Advanced',     label: t('ai.experience.advanced') },
  ])

  const goalOptions = computed(() => [
    { key: 'Muscle gain', label: t('ai.goal.muscleGain') },
    { key: 'Fat loss',    label: t('ai.goal.fatLoss') },
    { key: 'Strength',    label: t('ai.goal.strength') },
    { key: 'Endurance',   label: t('ai.goal.endurance') },
  ])

  return { environmentOptions, muscleGroupOptions, levelOptions, goalOptions }
}
