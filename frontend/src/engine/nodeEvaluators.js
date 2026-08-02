// Per-node-type evaluation functions
// Each evaluator takes (node, ctx, inputs) and returns a texture or null

import { getNodeType, NODE_TYPES } from '../stores/designStore.js'

/**
 * Get the animated value of a parameter at a given time.
 */
function getParam(node, paramId, time, designStore) {
  return designStore.getParamValue(node.id, paramId, time)
}

/**
 * All node evaluators keyed by node type.
 */
export const nodeEvaluators = {
  // ============ SOURCES ============

  media(node, ctx, inputs, time, designStore) {
    // Return a reference to the media asset — the renderer will load the frame
    return {
      type: 'media',
      assetId: node.params.assetId,
      startTime: getParam(node, 'startTime', time, designStore),
      duration: getParam(node, 'duration', time, designStore),
    }
  },

  text(node, ctx, inputs, time, designStore) {
    return {
      type: 'text',
      text: node.params.text || 'Hello',
      fontSize: getParam(node, 'fontSize', time, designStore) || 48,
      color: node.params.color || '#FFFFFF',
      fontFamily: node.params.fontFamily || 'sans-serif',
      bold: node.params.bold,
      italic: node.params.italic,
    }
  },

  rectangle(node, ctx, inputs, time, designStore) {
    return {
      type: 'shape',
      shape: 'rectangle',
      width: getParam(node, 'width', time, designStore) || 200,
      height: getParam(node, 'height', time, designStore) || 200,
      fill: node.params.fill || '#00D4FF',
      cornerRadius: getParam(node, 'cornerRadius', time, designStore) || 0,
    }
  },

  ellipse(node, ctx, inputs, time, designStore) {
    return {
      type: 'shape',
      shape: 'ellipse',
      width: getParam(node, 'width', time, designStore) || 200,
      height: getParam(node, 'height', time, designStore) || 200,
      fill: node.params.fill || '#EC4899',
    }
  },

  polygon(node, ctx, inputs, time, designStore) {
    return {
      type: 'shape',
      shape: 'polygon',
      sides: getParam(node, 'sides', time, designStore) || 5,
      radius: getParam(node, 'radius', time, designStore) || 100,
      fill: node.params.fill || '#10B981',
    }
  },

  star(node, ctx, inputs, time, designStore) {
    return {
      type: 'shape',
      shape: 'star',
      numPoints: getParam(node, 'numPoints', time, designStore) || 5,
      innerRadius: getParam(node, 'innerRadius', time, designStore) || 40,
      outerRadius: getParam(node, 'outerRadius', time, designStore) || 100,
      fill: node.params.fill || '#F59E0B',
    }
  },

  gradient(node, ctx, inputs, time, designStore) {
    return {
      type: 'gradient',
      color1: node.params.color1 || '#00D4FF',
      color2: node.params.color2 || '#EC4899',
      angle: getParam(node, 'angle', time, designStore) || 90,
    }
  },

  // ============ TRANSFORM ============

  transform(node, ctx, inputs, time, designStore) {
    const input = inputs['in']
    return {
      type: 'transform',
      input,
      x: getParam(node, 'x', time, designStore) || 0,
      y: getParam(node, 'y', time, designStore) || 0,
      scaleX: getParam(node, 'scaleX', time, designStore) ?? 1,
      scaleY: getParam(node, 'scaleY', time, designStore) ?? 1,
      rotation: getParam(node, 'rotation', time, designStore) || 0,
      opacity: getParam(node, 'opacity', time, designStore) ?? 1,
    }
  },

  // ============ COMPOSITE ============

  merge(node, ctx, inputs, time, designStore) {
    return {
      type: 'merge',
      fg: inputs['fg'],
      bg: inputs['bg'],
      mode: node.params.mode || 'normal',
    }
  },

  // ============ EFFECTS ============

  blur(node, ctx, inputs, time, designStore) {
    return {
      type: 'blur',
      input: inputs['in'],
      radius: getParam(node, 'radius', time, designStore) || 5,
    }
  },

  glow(node, ctx, inputs, time, designStore) {
    return {
      type: 'glow',
      input: inputs['in'],
      intensity: getParam(node, 'intensity', time, designStore) || 1,
      color: node.params.color || '#00D4FF',
      radius: getParam(node, 'radius', time, designStore) || 10,
    }
  },

  shadow(node, ctx, inputs, time, designStore) {
    return {
      type: 'shadow',
      input: inputs['in'],
      color: node.params.color || '#000000',
      blur: getParam(node, 'blur', time, designStore) || 8,
      offsetX: getParam(node, 'offsetX', time, designStore) || 0,
      offsetY: getParam(node, 'offsetY', time, designStore) || 4,
    }
  },

  colorCorrect(node, ctx, inputs, time, designStore) {
    return {
      type: 'colorCorrect',
      input: inputs['in'],
      brightness: getParam(node, 'brightness', time, designStore) || 0,
      contrast: getParam(node, 'contrast', time, designStore) || 0,
      saturation: getParam(node, 'saturation', time, designStore) || 0,
      hue: getParam(node, 'hue', time, designStore) || 0,
    }
  },

  chromaKey(node, ctx, inputs, time, designStore) {
    return {
      type: 'chromaKey',
      input: inputs['in'],
      keyColor: node.params.keyColor || '#00FF00',
      similarity: getParam(node, 'similarity', time, designStore) || 0.4,
      smoothness: getParam(node, 'smoothness', time, designStore) || 0.1,
    }
  },

  // ============ MATH ============

  math(node, ctx, inputs, time, designStore) {
    return {
      type: 'math',
      a: inputs['a'],
      b: inputs['b'],
      operation: node.params.operation || 'add',
    }
  },

  // ============ OUTPUT ============

  output(node, ctx, inputs) {
    return {
      type: 'output',
      input: inputs['in'],
    }
  },

  // ============ EXTENDED NODE TYPES ============

  solidColor(node, ctx, inputs, time, designStore) {
    return {
      type: 'solid',
      fill: node.params.color || '#00D4FF',
      width: getParam(node, 'width', time, designStore) || 1920,
      height: getParam(node, 'height', time, designStore) || 1080,
    }
  },

  noise(node, ctx, inputs, time, designStore) {
    return {
      type: 'noise',
      noiseType: node.params.noiseType || 'perlin',
      scale: getParam(node, 'scale', time, designStore) || 50,
      octaves: getParam(node, 'octaves', time, designStore) || 4,
      seed: getParam(node, 'seed', time, designStore) || 0,
    }
  },

  crop(node, ctx, inputs, time, designStore) {
    return {
      type: 'crop',
      input: inputs['in'],
      x: getParam(node, 'x', time, designStore) || 0,
      y: getParam(node, 'y', time, designStore) || 0,
      width: getParam(node, 'width', time, designStore) || 1920,
      height: getParam(node, 'height', time, designStore) || 1080,
    }
  },

  cornerPin(node, ctx, inputs, time, designStore) {
    return {
      type: 'cornerPin',
      input: inputs['in'],
      tl: { x: getParam(node, 'tlX', time, designStore) || 0, y: getParam(node, 'tlY', time, designStore) || 0 },
      tr: { x: getParam(node, 'trX', time, designStore) || 1920, y: getParam(node, 'trY', time, designStore) || 0 },
      bl: { x: getParam(node, 'blX', time, designStore) || 0, y: getParam(node, 'blY', time, designStore) || 1080 },
      br: { x: getParam(node, 'brX', time, designStore) || 1920, y: getParam(node, 'brY', time, designStore) || 1080 },
    }
  },

  channelSplit(node, ctx, inputs) {
    return {
      type: 'channelSplit',
      input: inputs['in'],
    }
  },

  channelMerge(node, ctx, inputs) {
    return {
      type: 'channelMerge',
      r: inputs['r'],
      g: inputs['g'],
      b: inputs['b'],
      a: inputs['a'],
    }
  },

  levels(node, ctx, inputs, time, designStore) {
    return {
      type: 'levels',
      input: inputs['in'],
      inBlack: getParam(node, 'inBlack', time, designStore) || 0,
      inWhite: getParam(node, 'inWhite', time, designStore) || 255,
      gamma: getParam(node, 'gamma', time, designStore) || 1,
      outBlack: getParam(node, 'outBlack', time, designStore) || 0,
      outWhite: getParam(node, 'outWhite', time, designStore) || 255,
    }
  },

  invert(node, ctx, inputs) {
    return {
      type: 'invert',
      input: inputs['in'],
    }
  },

  temperature(node, ctx, inputs, time, designStore) {
    return {
      type: 'temperature',
      input: inputs['in'],
      temperature: getParam(node, 'temperature', time, designStore) || 0,
      tint: getParam(node, 'tint', time, designStore) || 0,
    }
  },

  directionalBlur(node, ctx, inputs, time, designStore) {
    return {
      type: 'directionalBlur',
      input: inputs['in'],
      angle: getParam(node, 'angle', time, designStore) || 0,
      distance: getParam(node, 'distance', time, designStore) || 10,
    }
  },

  mask(node, ctx, inputs, time, designStore) {
    return {
      type: 'mask',
      input: inputs['in'],
      maskType: node.params.maskType || 'rectangle',
      x: getParam(node, 'x', time, designStore) || 0,
      y: getParam(node, 'y', time, designStore) || 0,
      width: getParam(node, 'width', time, designStore) || 1920,
      height: getParam(node, 'height', time, designStore) || 1080,
      feather: getParam(node, 'feather', time, designStore) || 0,
    }
  },

  timeShift(node, ctx, inputs, time, designStore) {
    return {
      type: 'timeShift',
      input: inputs['in'],
      offset: getParam(node, 'offset', time, designStore) || 0,
    }
  },

  expression(node, ctx, inputs, time, designStore) {
    return {
      type: 'expression',
      a: inputs['a'],
      b: inputs['b'],
      c: inputs['c'],
      d: inputs['d'],
      expression: node.params.expression || 'a + b',
    }
  },
}
