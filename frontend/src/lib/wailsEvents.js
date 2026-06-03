export function onWailsEvent(event, callback) {
  if (typeof window !== 'undefined' && window.runtime) {
    window.runtime.EventsOn(event, callback)
  }
}

export function offWailsEvent(event, callback) {
  if (typeof window !== 'undefined' && window.runtime) {
    window.runtime.EventsOff(event, callback)
  }
}
