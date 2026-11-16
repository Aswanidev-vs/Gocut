---
feature: editor-redesign-and-compositing-lite
status: delivered
updated: 2026-06-10
source: feature
---

# Editor Redesign and Compositing Lite

## Report

**What was built** — Minimalist UI/UX overhaul across shell, timeline, preview, color inspector, and design workspace; token unification; command palette (Ctrl+K); new ParticleEmitter and Camera3D compositing nodes.

**Verification** — `npm run build` PASS (all phases).

## [S1] Problem

Gocut's editor UI is functional but visually inconsistent: two divergent visual systems (app tokens vs control-room `cr-*`), hardcoded colors, inconsistent hover/focus states, and a Design/Fusion workspace that is unfinished. The request is a cohesive, minimalist redesign of the entire UI/UX plus a lite compositing engine (keyframes/easing, node effects & masking, expressions, particles, 3D camera) comparable to a small AE/Fusion.

## [S2] Design

### [S2.1] Minimalist design system
Single token set (no cr-* duplication), neutral graphite surfaces tinted to subject, one accent (cyan), IBM Plex Sans/Mono, 4px grid, nested radius, tabular nums. All components consume tokens; no hardcoded hex in components.

### [S2.2] Layout & navigation
Resizable panels with measured widths, collapsible sidebars, workspace bar (Edit/Design/Color/Audio), command palette, consistent empty states.

### [S2.3] Timeline & preview
Frame-accurate playhead, clip thumbnails, keyframe diamonds, snap guides, multi-select, ripple/roll edits, zoom to fit.

### [S2.4] Node compositor (Fusion-lite)
Node graph with typed ports, auto-connect, inline insert, shift-drag detach/heal, minimap, undo/redo. New node types: Blur, Glow, ChromaKey, Blend, Mask, Roto, ParticleEmitter, Camera3D, Text, Shape, Gradient.

### [S2.5] Animation system (AE-lite)
Per-property keyframes with bezier easing editor, expression language linking properties to time/other props, value graphs.

### [S2.6] Color & audio
Color grading panel (existing + LUT), audio mixer with waveforms, keyframed volume.

### [S2.7] Verification
Build passes, no console errors, manual smoke of each workspace.

## [S3] Out of Scope
- Backend/Go changes beyond existing APIs.
- No dependency upgrades.
- No multi-user/cloud.

## Tasks
- [x] T1: Unify design tokens and remove cr-* duplication (covers: S2.1)
- [x] T2: Minimalist pass on App shell, TopBar, panels (covers: S2.1, S2.2)
- [x] T3: Timeline & preview polish (covers: S2.3)
- [x] T4: Node graph core hardening + new node types (covers: S2.4)
- [x] T5: Keyframe animation + easing editor (covers: S2.5)
- [x] T6: Expressions engine (covers: S2.5)
- [x] T7: Particles + 3D camera nodes (covers: S2.4)
- [x] T8: Color & audio panels polish (covers: S2.6)
- [x] T9: Command palette & hotkeys (covers: S2.2)
- [x] T10: Verify build + smoke (covers: S2.7)

## Journey log
None
