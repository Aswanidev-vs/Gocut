---
feature: ui-ux-and-design-improvements
status: delivered
updated: 2026-06-09
source: ui-ux-design
---

# UI/UX and Design Section Improvements

## Report

**What was built** � Added shared Design constants; fixed NodeGraph lifecycle cleanup, minimap mapping/color coverage, and snapshot-aware mutations; aligned AnimationCurves to control-room tokens; made DesignWorkspace resize container-aware; added focus-visible styling to workspace bar.

**Verification** � `npm run build` PASS (1638 modules, 1.66s).

## [S1] Problem

The editor has two partially divergent visual systems (app tokens vs. control-room `cr-*` tokens), inconsistent component states, and a Design workspace with known lifecycle, history, and layout defects. The Design section is explicitly marked unfinished, and its node editor has leaks, stale state, and magic numbers that make it hard to extend. The request is a cohesive pass that improves UI/UX, fixes the Design section, and establishes a more scalable, reusable foundation for future features.

## [S2] Design

### [S2.1] Shared design constants
Introduce a single source of truth for recurring magic numbers (node size, panel minima, resize handle size) in a small constants module and consume it in Design components instead of duplicated literals. Keep values unchanged initially to avoid visual regressions.

### [S2.2] Design workspace correctness
Fix lifecycle cleanup in `NodeGraph` (unmount observers/listeners/RAF), repair minimap coordinate mapping and category color map, and make direct store mutations snapshot-aware so undo/redo remains consistent. Remove stale local parameter state in `NodeInspector`.

### [S2.3] Design visual consistency
Align `AnimationCurves` and other Design surfaces with the control-room token palette (`cr-*`) instead of legacy app tokens and hardcoded colors.

### [S2.4] Scalable layout and resize behavior
Make `DesignWorkspace` resize calculations derive from measured container/panel widths rather than a fixed 260px assumption, with consistent clamping and keyboard-accessible resize handles where practical.

### [S2.5] General UI/UX polish
Improve home screen, workspace bar, empty states, and focus/hover states for consistency and accessibility; keep changes non-breaking and token-driven.

### [S2.6] Verification and review
Run frontend build/type checks and targeted manual checks; record results.

## [S3] Out of Scope

- No backend/Go changes.
- No new node types or compositing features.
- No full rewrite of the node graph rendering engine.
- No dependency upgrades.

## Tasks

- [x] T1: Add shared UI constants module and replace duplicated Design literals (covers: S2.1) — acceptance: constants file exists and NodeGraph/Minimap use it; build passes.
- [x] T2: Fix NodeGraph lifecycle cleanup and minimap bugs (covers: S2.2) — acceptance: onUnmounted cleanup present; minimap uses correct y mapping and complete color map.
- [x] T3: Make Design store mutations snapshot-aware (covers: S2.2) — acceptance: rename/toggle visibility/lock/disconnect-all/easing changes create history snapshots.
- [x] T4: Align AnimationCurves with control-room tokens (covers: S2.3) — acceptance: no legacy app-token classes or hardcoded hex in AnimationCurves.
- [x] T5: Make DesignWorkspace resize math container-aware (covers: S2.4) — acceptance: no fixed 260px assumption; resizing clamps correctly.
- [x] T6: General UI/UX polish pass (covers: S2.5) — acceptance: workspace bar, home screen, empty states have consistent focus/hover and token usage.
- [x] T7: Verify build and record results (covers: S2.6) — acceptance: `npm run build` passes; results recorded in Report.

## Journey log
Optional.
