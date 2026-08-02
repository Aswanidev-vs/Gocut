// Graph evaluation engine — topological sort + frame-by-frame evaluation

/**
 * Topological sort of the node graph starting from output nodes.
 * Returns nodes in evaluation order (sources first, output last).
 */
export function topologicalSort(nodes, connections) {
  const adj = new Map()    // nodeId → [dependent nodeIds]
  const inDeg = new Map()  // nodeId → in-degree

  for (const node of nodes) {
    inDeg.set(node.id, 0)
    adj.set(node.id, [])
  }

  for (const conn of connections) {
    if (!adj.has(conn.fromNode)) continue
    adj.get(conn.fromNode).push(conn.toNode)
    inDeg.set(conn.toNode, (inDeg.get(conn.toNode) || 0) + 1)
  }

  // BFS from nodes with in-degree 0 (sources)
  const queue = []
  for (const [id, deg] of inDeg) {
    if (deg === 0) queue.push(id)
  }

  const sorted = []
  while (queue.length > 0) {
    const id = queue.shift()
    sorted.push(id)
    for (const dep of adj.get(id) || []) {
      const newDeg = inDeg.get(dep) - 1
      inDeg.set(dep, newDeg)
      if (newDeg === 0) queue.push(dep)
    }
  }

  // Detect cycles
  if (sorted.length !== nodes.length) {
    const missing = nodes.filter(n => !sorted.includes(n.id)).map(n => n.label || n.type)
    throw new Error(`Cycle detected in node graph involving: ${missing.join(', ')}`)
  }

  return sorted
}

/**
 * Gather input textures for a node from its connections.
 * Returns { portName: texture } map.
 */
export function gatherInputs(node, connections, nodeOutputs, nodeType) {
  const inputs = {}
  if (!nodeType) return inputs

  for (const conn of connections) {
    if (conn.toNode === node.id) {
      const fromOutput = nodeOutputs.get(conn.fromNode)
      if (fromOutput) {
        inputs[conn.toPort] = fromOutput
      }
    }
  }

  return inputs
}

/**
 * Evaluate the entire graph at a given time.
 * Returns Map<nodeId, texture> with the final output.
 */
export async function evaluateGraph(nodes, connections, time, resolution, evaluators) {
  const sorted = topologicalSort(nodes, connections)
  const nodeOutputs = new Map()

  for (const nodeId of sorted) {
    const node = nodes.find(n => n.id === nodeId)
    if (!node || node.visible === false) continue

    const evaluator = evaluators[node.type]
    if (!evaluator) continue

    // Get node type definition for port info
    const { getNodeType } = await import('../stores/designStore.js')
    const nodeType = getNodeType(node.type)

    // Gather input textures
    const inputs = gatherInputs(node, connections, nodeOutputs, nodeType)

    // Evaluate this node
    try {
      const output = await evaluator(node, { width: resolution.width, height: resolution.height, time }, inputs)
      nodeOutputs.set(nodeId, output)
    } catch (e) {
      console.error(`Error evaluating node ${node.label || node.type}:`, e)
    }
  }

  return nodeOutputs
}

/**
 * Find the output node in the graph.
 */
export function findOutputNode(nodes) {
  return nodes.find(n => n.type === 'output')
}
