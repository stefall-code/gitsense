package louvain

import (
	"math"
)

// Graph 无向加权图
type Graph struct {
	Nodes []string                   // 节点列表
	Adj   map[string]map[string]float64 // 邻接表: node → neighbor → weight
}

// NewGraph 创建空图
func NewGraph() *Graph {
	return &Graph{
		Adj: make(map[string]map[string]float64),
	}
}

// AddNode 添加节点
func (g *Graph) AddNode(node string) {
	if _, ok := g.Adj[node]; !ok {
		g.Adj[node] = make(map[string]float64)
		g.Nodes = append(g.Nodes, node)
	}
}

// AddEdge 添加无向加权边
func (g *Graph) AddEdge(a, b string, weight float64) {
	g.AddNode(a)
	g.AddNode(b)
	g.Adj[a][b] += weight
	g.Adj[b][a] += weight
}

// Degree 节点度数（加权）
func (g *Graph) Degree(node string) float64 {
	sum := 0.0
	for _, w := range g.Adj[node] {
		sum += w
	}
	return sum
}

// TotalWeight 图的总权重（每条边只算一次）
func (g *Graph) TotalWeight() float64 {
	sum := 0.0
	seen := make(map[string]bool, len(g.Nodes))
	for _, n := range g.Nodes {
		for nb, w := range g.Adj[n] {
			key := n + "|" + nb
			if !seen[key] {
				sum += w
				seen[key] = true
			}
		}
	}
	return sum
}

// Community 社区检测结果
type Community struct {
	ID    int      // 社区 ID
	Nodes []string // 社区内节点
}

// Result Louvain 检测结果
type Result struct {
	Communities []Community           // 社区列表
	NodeMap     map[string]int        // node → community ID
	Modularity  float64               // 最终模块度
}

// SubGraph 提取子图，仅包含指定节点及其之间的边，保留原始边权重
func SubGraph(g *Graph, nodes []string) *Graph {
	sub := NewGraph()
	nodeSet := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		nodeSet[n] = true
	}
	for _, n := range nodes {
		sub.AddNode(n)
		for nb, w := range g.Adj[n] {
			if nodeSet[nb] {
				// 仅在 n < nb 时添加，避免重复加权重
				if n < nb {
					sub.AddEdge(n, nb, w)
				}
			}
		}
	}
	return sub
}

// Detect 执行 Louvain 社区检测
func Detect(g *Graph, minSize int) *Result {
	return DetectWithPenalty(g, minSize, false)
}

// DetectWithPenalty 执行 Louvain 社区检测，可选 hub topic 惩罚
// 当 hubPenalty=true 时，边权重除以 log(max(degree_a, degree_b) + 1)，
// 降低高度数 hub topic（如 python, golang, ai）的支配性
func DetectWithPenalty(g *Graph, minSize int, hubPenalty bool) *Result {
	if hubPenalty {
		penalized := applyHubPenalty(g)
		return detectInternal(penalized, minSize)
	}
	return detectInternal(g, minSize)
}

// applyHubPenalty 对图的每条边应用 hub 惩罚：
// effective_weight = weight / log(max(degree_a, degree_b) + 1)
func applyHubPenalty(g *Graph) *Graph {
	pg := NewGraph()
	// 预计算每个节点的度数（邻居数量，非加权度数）
	degree := make(map[string]int, len(g.Nodes))
	for _, n := range g.Nodes {
		degree[n] = len(g.Adj[n])
	}
	seen := make(map[string]bool, len(g.Nodes))
	for _, n := range g.Nodes {
		for nb, w := range g.Adj[n] {
			key := n + "|" + nb
			if seen[key] {
				continue
			}
			seen[key] = true
			maxDeg := degree[n]
			if degree[nb] > maxDeg {
				maxDeg = degree[nb]
			}
			penalizedW := w / math.Log(float64(maxDeg)+1)
			pg.AddEdge(n, nb, penalizedW)
		}
	}
	return pg
}

// detectInternal Louvain 核心算法
func detectInternal(g *Graph, minSize int) *Result {
	if len(g.Nodes) == 0 {
		return &Result{NodeMap: make(map[string]int)}
	}

	m := g.TotalWeight()
	if m == 0 {
		// 无边图，每个节点一个社区
		nodeMap := make(map[string]int, len(g.Nodes))
		comms := make([]Community, len(g.Nodes))
		for i, n := range g.Nodes {
			nodeMap[n] = i
			comms[i] = Community{ID: i, Nodes: []string{n}}
		}
		return &Result{Communities: comms, NodeMap: nodeMap, Modularity: 0}
	}

	// 初始化：每个节点一个社区
	node2comm := make(map[string]int, len(g.Nodes))
	commNodes := make(map[int][]string, len(g.Nodes))
	commDegree := make(map[int]float64, len(g.Nodes))

	for i, n := range g.Nodes {
		node2comm[n] = i
		commNodes[i] = []string{n}
		commDegree[i] = g.Degree(n)
	}

	// Phase 1: 局部移动
	improved := true
	for improved {
		improved = false
		for _, node := range g.Nodes {
			curComm := node2comm[node]
			ki := g.Degree(node)

			// 计算从当前社区移除的 delta
			kiCur := 0.0
			for nb, w := range g.Adj[node] {
				if node2comm[nb] == curComm {
					kiCur += w
				}
			}

			// 尝试移到邻居社区
			neighborComms := make(map[int]float64) // comm → sigma_in
			for nb, w := range g.Adj[node] {
				neighborComms[node2comm[nb]] += w
			}

			bestComm := curComm
			bestDelta := 0.0

			for comm, kiComm := range neighborComms {
				if comm == curComm {
					continue
				}
				// ΔQ = [ki_comm/m - commDegree[comm]*ki/(2m²)] - [ki_cur/m - commDegree[curComm]*ki/(2m²)]
				// 简化: ΔQ = (ki_comm - ki_cur)/m - ki*(commDegree[comm] - commDegree[curComm] + ki)/(2m²)
				deltaQ := (kiComm - kiCur)/m - ki*(commDegree[comm]-(commDegree[curComm]-ki))/(2*m*m)
				if deltaQ > bestDelta {
					bestDelta = deltaQ
					bestComm = comm
				}
			}

			if bestComm != curComm {
				// 移动节点
				node2comm[node] = bestComm

				// 更新社区节点列表
				newCurNodes := make([]string, 0, len(commNodes[curComm])-1)
				for _, n := range commNodes[curComm] {
					if n != node {
						newCurNodes = append(newCurNodes, n)
					}
				}
				commNodes[curComm] = newCurNodes
				commNodes[bestComm] = append(commNodes[bestComm], node)

				// 更新社区度数
				commDegree[curComm] -= ki
				commDegree[bestComm] += ki

				improved = true
			}
		}
	}

	// 过滤小社区
	commIDMap := make(map[int]int) // old comm id → new sequential id
	newID := 0
	for oldID, nodes := range commNodes {
		if len(nodes) >= minSize {
			commIDMap[oldID] = newID
			newID++
		}
	}

	// 构建结果
	result := &Result{
		NodeMap:    make(map[string]int, len(g.Nodes)),
		Modularity: computeModularity(g, node2comm, m),
	}

	// 分配社区
	commResult := make(map[int][]string)
	for _, n := range g.Nodes {
		oldComm := node2comm[n]
		if newComm, ok := commIDMap[oldComm]; ok {
			result.NodeMap[n] = newComm
			commResult[newComm] = append(commResult[newComm], n)
		} else {
			result.NodeMap[n] = -1 // 过滤掉的小社区
		}
	}

	for id := 0; id < newID; id++ {
		result.Communities = append(result.Communities, Community{
			ID:    id,
			Nodes: commResult[id],
		})
	}

	return result
}

// computeModularity 计算模块度 Q
func computeModularity(g *Graph, node2comm map[string]int, m float64) float64 {
	if m == 0 {
		return 0
	}

	Q := 0.0
	for _, n1 := range g.Nodes {
		for n2, w := range g.Adj[n1] {
			if node2comm[n1] == node2comm[n2] {
				Q += w - g.Degree(n1)*g.Degree(n2)/(2*m)
			}
		}
	}
	Q /= (2 * m)

	// 防止浮点误差
	if math.Abs(Q) < 1e-10 {
		Q = 0
	}
	return Q
}
