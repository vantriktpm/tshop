import { ref, type Ref } from 'vue'

const loading = ref(false) as Ref<boolean>

export function useLoading() {
  const setLoading = (value: boolean) => {
    loading.value = value
  }

  const withLoading = async <T>(fn: () => Promise<T>): Promise<T> => {
    loading.value = true
    try {
      return await fn()
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    setLoading,
    withLoading,
  }
}
