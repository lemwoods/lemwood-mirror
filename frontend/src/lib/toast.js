import { reactive } from 'vue'

const state = reactive({ items: [] })

let uid = 0
const timers = new Map()

const dismiss = id => {
  const timer = timers.get(id)
  if (timer) {
    clearTimeout(timer)
    timers.delete(id)
  }
  const index = state.items.findIndex(item => item.id === id)
  if (index !== -1) state.items.splice(index, 1)
}

const push = (type, message, duration = 1500) => {
  const id = ++uid
  state.items.push({ id, type, message })
  timers.set(
    id,
    setTimeout(() => {
      dismiss(id)
    }, duration)
  )
  return id
}

export const toast = {
  items: state.items,
  success: message => push('success', message),
  error: message => push('error', message),
  dismiss
}
