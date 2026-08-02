# Gocut Development Timeline

## Completed

### Phase 1: Merge Systems & Fix Bugs
- Unified `DesignWorkspace.vue` replaces inline system
- Fixed template insertion double-node bug
- Fixed Chroma Key Demo missing connections
- Wired `DesignOnboarding.vue` and `SimpleEffectsPanel.vue`
- Added template search filter

### Phase 2: Undo/Redo & Save/Load
- `designHistoryStore.js` — snapshot-based undo/redo
- `designStore.serialize()`/`deserialize()` for persistence
- Design graph saved/loaded with project
- Go `Project` struct updated with `DesignGraph` field

### Phase 3: Right-Click Context Menus
- `ContextMenu.vue` — reusable with auto-positioning
- Canvas, node, and connection context menus

### Phase 4: Keyboard Shortcuts
- `useDesignHotkeys.js` composable
- Ctrl+Z/Y, Ctrl+C/V/X, Ctrl+D, Ctrl+A, Delete
- Ctrl+0 zoom fit, Ctrl++/- zoom, Escape deselect
- S snap toggle, Space play/pause, Arrows step

### Phase 5: Multi-Selection
- Ctrl+click additive selection
- Ctrl+click canvas for marquee/lasso
- Multi-node drag, selection highlights

### Phase 6: Node Search Palette
- `NodeSearchPalette.vue` — Shift+Space fuzzy search

### Phase 7: Inspector Improvements
- Proper easing interpolation (bounce, elastic, easeIn/Out)
- Keyframe list with expand/collapse, easing dropdown, delete

### Phase 8: Visual Polish
- `Minimap.vue` — SVG minimap with color-coded nodes

### Phase 9: Node Grouping & Bookmarks
- Ctrl+G/Ctrl+Shift+G group/ungroup
- Ctrl+1-9 save bookmarks, Ctrl+Shift+1-9 restore

### Phase 10: FFmpeg Motion Tracking Backend
- `internal/tracking/` package (analyzer, stabilizer, models)
- `AnalyzeMotion()` — FFmpeg vidstabdetect/mestimate
- `ApplyStabilization()` — FFmpeg vidstabtransform
- Tracking UI with progress tracking

### Phase 11: Graph Evaluation Engine
- `graphEvaluator.js` — topological sort + evaluateGraph()
- `nodeEvaluators.js` — evaluators for 16 node types
- `CompositingCanvas.vue` — Canvas2D compositing preview
- View mode toggle: Graph / Preview / Split

### Phase 12: Extended Node Types
- 15 new types: solidColor, noise, crop, cornerPin, channelSplit/Merge, levels, invert, temperature, directionalBlur, mask, timeShift, expression
- Total: 31 node types

---

## Future Phases

### Phase 13: Tracking → Keyframes Integration
- Convert `TrackingData` → keyframes on node params
- `trackingIntegration.js` — apply tracking to transform nodes
- Auto-create keyframes from tracked points
- UI: "Apply to Transform" button in tracking results
- Files to create:
  - `frontend/src/engine/trackingIntegration.js`
  - `frontend/src/components/tracking/TrackingResults.vue`
- Files to modify:
  - `frontend/src/stores/designStore.js` — add `applyTrackingKeyframes()`
  - `frontend/src/stores/trackingStore.js` — add conversion logic

### Phase 14: Export Pipeline
- Frame-by-frame rendering via WebGL/Canvas → FFmpeg encoding
- Frontend renders each frame → `canvas.toBlob('image/png')`
- Send frames to backend via `DesignRenderFrame(jobID, frameIndex, imageData)`
- Backend writes frames to temp dir → FFmpeg encodes to MP4
- Progress events during export
- Files to create:
  - `internal/render/designRender.go`
  - `frontend/src/engine/exportPipeline.js`
- Files to modify:
  - `internal/render/queue.go` — add `buildDesignRenderArgs()` path
  - `internal/app/app.go` — add `DesignRenderFrame()` method
  - `frontend/src/lib/wails.js` — add new method exports

### Phase 15: WebGL Renderer
- Upgrade Canvas2D → WebGL2 for better performance
- GLSL shaders per node type (already defined in `shaderSources.js`)
- FBO ping-pong for multi-pass effects
- Texture pool for memory management
- Fallback to Canvas2D if WebGL2 unavailable
- Files to create:
  - `frontend/src/engine/webglRenderer.js`
  - `frontend/src/engine/shaderCompiler.js`
- Files to modify:
  - `frontend/src/components/design/CompositingCanvas.vue` — use WebGL
  - `frontend/src/engine/nodeEvaluators.js` — return WebGL textures

### Phase 16: Expression Engine
- Math expression parser for `expression` node
- Cross-node parameter linking (e.g., `node2.x = node1.x * 2`)
- Time-based expressions (e.g., `sin(time * 3) * 50`)
- Variable system (node params, time, frame, random)
- Files to create:
  - `frontend/src/engine/expressionParser.js`
- Files to modify:
  - `frontend/src/engine/nodeEvaluators.js` — evaluate expressions
  - `frontend/src/stores/designStore.js` — expression evaluation in getParamValue()
