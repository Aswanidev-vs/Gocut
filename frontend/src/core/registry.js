// Minimal registry factory for Gocut UI extensibility.
const registries = new Map()

export function createRegistry(name) {
  const items = new Map()
  registries.set(name, items)
  return {
    name,
    register(item) {
      if (!item || !item.id) throw new Error(`Registry ${name}: item requires id`)
      items.set(item.id, item)
      return item.id
    },
    get(id) { return items.get(id) },
    list() { return [...items.values()] },
    has(id) { return items.has(id) },
  }
}

export const panelsRegistry = createRegistry('panels')
export const commandsRegistry = createRegistry('commands')
export const inspectorSectionsRegistry = createRegistry('inspectorSections')
export const trackTypesRegistry = createRegistry('trackTypes')
export const workspacesRegistry = createRegistry('workspaces')
export const nodeTypesRegistry = createRegistry('nodeTypes')
export const libraryTabsRegistry = createRegistry('libraryTabs')
export const statusBarItemsRegistry = createRegistry('statusBarItems')
export const settingsPagesRegistry = createRegistry('settingsPages')
export const exportPresetsRegistry = createRegistry('exportPresets')