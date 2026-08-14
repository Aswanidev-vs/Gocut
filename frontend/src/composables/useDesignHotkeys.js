import { onMounted, onUnmounted } from 'vue'
import { useDesignStore, NODE_TYPES } from '../stores/designStore'
import { useDesignHistoryStore } from '../stores/designHistoryStore'
import { useUiStore } from '../stores/uiStore'

function getComboString(e) {
  const combo = []
  if (e.ctrlKey || e.metaKey) combo.push('ctrl')
  if (e.shiftKey) combo.push('shift')
  if (e.altKey) combo.push('alt')

  if (e.key !== 'Control' && e.key !== 'Shift' && e.key !== 'Alt' && e.key !== 'Meta') {
    combo.push(e.key.toLowerCase())
  }

  return combo.join('+')
}

export function useDesignHotkeys() {
  const handleKeyDown = (e) => {
    const ui = useUiStore()
    if (ui.activeWorkspace !== 'design') return

    const tag = e.target.tagName.toLowerCase()
    if (tag === 'input' || tag === 'textarea' || e.target.isContentEditable) return

    const combo = getComboString(e)
    if (!combo) return

    const designStore = useDesignStore()
    const designHistory = useDesignHistoryStore()

    switch (combo) {
      // Undo / Redo
      case 'ctrl+z':
        e.preventDefault()
        designHistory.undo()
        break
      case 'ctrl+y':
      case 'ctrl+shift+z':
        e.preventDefault()
        designHistory.redo()
        break

      // Copy / Paste / Cut
      case 'ctrl+c':
        e.preventDefault()
        copySelectedNode(designStore)
        break
      case 'ctrl+v':
        e.preventDefault()
        pasteNode(designStore)
        break
      case 'ctrl+x':
        e.preventDefault()
        cutSelectedNode(designStore)
        break

      // Select / Duplicate / Delete
      case 'ctrl+a':
        e.preventDefault()
        if (designStore.nodes.length > 0) {
          designStore.selectedNodeId = designStore.nodes[0].id
        }
        break
      case 'ctrl+d':
        e.preventDefault()
        designStore.duplicateSelectedNode()
        break
      case 'delete':
      case 'backspace':
        e.preventDefault()
        designStore.removeSelectedNode()
        break

      // DaVinci Resolve Fusion Viewers: 1 = Viewer 1 (Left), 2 = Viewer 2 (Right)
      case '1':
        if (designStore.selectedNodeId) {
          e.preventDefault()
          designStore.setViewer1(designStore.selectedNodeId)
        }
        break
      case '2':
        if (designStore.selectedNodeId) {
          e.preventDefault()
          designStore.setViewer2(designStore.selectedNodeId)
        }
        break

      // Zoom
      case 'ctrl+0':
        e.preventDefault()
        designStore.zoomFit()
        break
      case 'ctrl+=':
      case 'ctrl++':
        e.preventDefault()
        designStore.zoomIn()
        break
      case 'ctrl+-':
        e.preventDefault()
        designStore.zoomOut()
        break
      case 'ctrl+1':
        e.preventDefault()
        designStore.zoom = 1
        designStore.panX = 0
        designStore.panY = 0
        break

      // Deselect
      case 'escape':
        e.preventDefault()
        designStore.selectedNodeId = null
        designStore.selectedConnectionId = null
        break

      // Toggle snap
      case 's':
        e.preventDefault()
        designStore.snapEnabled = !designStore.snapEnabled
        useUiStore().addToast(`Snap ${designStore.snapEnabled ? 'on' : 'off'}`, 'info', 1200)
        break

      // Play/Pause
      case ' ':
        e.preventDefault()
        // Dispatch to DesignWorkspace play toggle
        window.dispatchEvent(new CustomEvent('design:togglePlay'))
        break

      // Step playhead
      case 'arrowleft':
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('design:stepPlayhead', { detail: -1 }))
        break
      case 'arrowright':
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('design:stepPlayhead', { detail: 1 }))
        break

      // Node search palette
      case 'shift+ ':
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('design:openSearch'))
        break

      // Group / Ungroup
      case 'ctrl+g':
        e.preventDefault()
        if (designStore.selectedNodeIds.size > 1) {
          designStore.groupNodes([...designStore.selectedNodeIds])
        }
        break
      case 'ctrl+shift+g':
        e.preventDefault()
        // Find group containing selected node and ungroup it
        if (designStore.selectedNodeId) {
          const group = designStore.groups.find(g => g.nodeIds.includes(designStore.selectedNodeId))
          if (group) designStore.ungroupNode(group.id)
        }
        break

      // Bookmarks (Ctrl+1-9 save, Ctrl+Shift+1-9 load)
      default: {
        const bookmarkMatch = combo.match(/^ctrl(\+shift)?\+([1-9])$/)
        if (bookmarkMatch) {
          e.preventDefault()
          const slot = parseInt(bookmarkMatch[2])
          if (bookmarkMatch[1]) {
            designStore.loadBookmark(slot)
          } else {
            designStore.saveBookmark(slot)
          }
        }
        break
      }
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyDown)
  })
}

// ============ Clipboard helpers ============

let clipBoard = null

function copySelectedNode(designStore) {
  if (!designStore.selectedNodeId) return
  const node = designStore.nodes.find(n => n.id === designStore.selectedNodeId)
  if (!node) return
  clipBoard = JSON.parse(JSON.stringify(node))
  useUiStore().addToast('Copied', 'info', 800)
}

function cutSelectedNode(designStore) {
  if (!designStore.selectedNodeId) return
  copySelectedNode(designStore)
  designStore.removeSelectedNode()
}

function pasteNode(designStore) {
  if (!clipBoard) return
  const nodeData = JSON.parse(JSON.stringify(clipBoard))
  nodeData.id = undefined
  nodeData.x = (nodeData.x || 0) + 40
  nodeData.y = (nodeData.y || 0) + 40
  if (NODE_TYPES[nodeData.type]) {
    designStore.addNode(nodeData.type, nodeData)
    useUiStore().addToast('Pasted', 'info', 800)
  }
}
