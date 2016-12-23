package graphgo

// Query on top of a Graph instance
// Aims to have "functional" style
type Query struct {
	Graph  *Graph
	result map[string]*Node
	Cache  map[string](map[string]interface{})

	Queries map[string]*Query
}

// NewEmptyQuery instanciates
func NewEmptyQuery() *Query {
	return &Query{
		Graph:  nil,
		result: map[string]*Node{},
		Cache:  map[string](map[string]interface{}){},

		Queries: map[string]*Query{},
	}
}

// NewQuery instanciates
func NewQuery(g *Graph, starts ...string) *Query {
	result := map[string]*Node{}

	for _, start := range starts {
		node, err := g.GetNode(start)
		if err != nil {
			continue
		}
		// At this point. we found the start node, and add it to the result graph
		result[node.Key] = node
	}

	return &Query{
		Graph:  g,
		result: result,
		Cache:  map[string](map[string]interface{}){},

		Queries: map[string]*Query{},
	}

}

// IsDeep checks if this is a nested query
func (q *Query) IsDeep() bool {
	if q.Queries == nil {
		return false
	}
	if len(q.Queries) == 0 {
		return false
	}
	return true
}

// IsDoubleDeep returns true if depth >= 2
func (q *Query) IsDoubleDeep() bool {
	if !q.IsDeep() {
		return false
	}

	for _, nestedQuery := range q.Queries {
		return nestedQuery.IsDeep()
	}

	return false
}

// Out returns outgoing nodes to this graph
func (q *Query) Out(label string) *Query {

	// Deep Calls
	if q.IsDeep() {
		for _, nestedQuery := range q.Queries {
			nestedQuery.Out(label)
		}
		return q
	}

	newResult := map[string]*Node{}

	// Loop over all the nodes in the current result
	for _, node := range q.result {

		// Loop over all relationships for this node
		for edgeKey, edgeLabel := range node.Out {

			// Only keep the ones with given label
			if edgeLabel == label {

				edge, err := q.Graph.GetEdge(edgeKey)
				if err != nil {
					continue
				}

				endNode, err := q.Graph.GetNode(edge.End)
				if err != nil {
					continue
				}

				newResult[endNode.Key] = endNode

			}

		}

	}

	q.result = newResult

	return q
}

// In returns outgoing nodes to this graph
func (q *Query) In(label string) *Query {

	// Deep Calls
	if q.IsDeep() {
		for _, nestedQuery := range q.Queries {
			nestedQuery.In(label)
		}
		return q
	}

	newResult := map[string]*Node{}

	// Loop over all the nodes in the current result
	for _, node := range q.result {

		// Loop over all relationships for this node
		for edgeKey, edgeLabel := range node.In {

			// Only keep the ones with given label
			if edgeLabel == label {

				edge, err := q.Graph.GetEdge(edgeKey)
				if err != nil {
					continue
				}

				startNode, err := q.Graph.GetNode(edge.Start)
				if err != nil {
					continue
				}

				newResult[startNode.Key] = startNode

			}

		}

	}

	q.result = newResult

	return q
}

// FilterNodes based on a predicate on their properties
func (q *Query) FilterNodes(predicate func(map[string]interface{}) bool) *Query {

	// Deep Calls
	if q.IsDeep() {
		for _, nestedQuery := range q.Queries {
			nestedQuery.FilterNodes(predicate)
		}
		return q
	}

	newResult := map[string]*Node{}

	// Loop over all the nodes in the current result
	for nodeKey, node := range q.result {

		if predicate(node.Props) {
			newResult[nodeKey] = node
		}

	}

	q.result = newResult
	return q

}

// Flatten function
// Get an iterable of all the keys, per node
func (q *Query) Save(keys ...string) *Query {

	// Deep Calls
	if q.IsDeep() {
		for _, nestedQuery := range q.Queries {
			nestedQuery.Save(keys...)
		}
		return q
	}

	// Loop over every node in the result
	for nodeKey, node := range q.result {
		_, exists := q.Cache[nodeKey]
		if !exists {
			q.Cache[nodeKey] = map[string]interface{}{}
		}

		// Loop over every key we care about
		for _, key := range keys {
			value, err := node.Get(key)
			if err != nil {
				continue
			}
			q.Cache[nodeKey][key] = value
		}
	}
	return q
}

// // Output just returns the result
// func (q *Query) Output() map[string]*Node {
// 	copy := map[string]*Node{}
//
// 	for nodeKey, node := range q.result {
// 		copy[nodeKey] = node.Copy()
// 	}
//
// 	return copy
// }

// DeepenQuery creates a new DeepQuery, from every node of a given Query
func (q *Query) Deepen() *Query {
	// Deep Calls
	if q.IsDeep() {
		for _, nestedQuery := range q.Queries {
			nestedQuery.Deepen()
		}
		return q
	}

	queries := map[string]*Query{}
	for _, r := range q.result {

		// Use the node key as a query key
		queries[r.Key] = NewQuery(q.Graph, r.Key)

	}

	q.Queries = queries
	return q
}

// Flatten flattens a query to the lower level
func (q *Query) DeepSave(name string) *Query {

	// Nothing to flatten
	if !q.IsDeep() {
		return q
	}

	// If it's actually too deep, we keep going
	if q.IsDoubleDeep() {
		for _, nestedQuery := range q.Queries {
			nestedQuery.Flatten()
		}
		return q
	}

	// Otherwise, this is the level before the lowest
	// We can flatten the cache
	for nodeKey, nestedQuery := range q.Queries {
		_, exists := q.Cache[nodeKey]
		if !exists {
			q.Cache[nodeKey] = map[string]interface{}{}
		}

		q.Cache[nodeKey][name] = nestedQuery.Cache
	}

	return q

}

// Flatten flattens a query to the lower level
func (q *Query) Flatten() *Query {

	// Nothing to flatten
	if !q.IsDeep() {
		return q
	}

	// If it's actually too deep, we keep going
	if q.IsDoubleDeep() {
		for _, nestedQuery := range q.Queries {
			nestedQuery.Flatten()
		}
		return q
	}

	q.Queries = map[string]*Query{}
	return q

}

// DeepFilter
func (q *Query) DeepFilter(keepQuery func(*Query) bool) *Query {

	// Nothing to flatten
	if !q.IsDeep() {
		return q
	}

	// If it's actually too deep, we keep going
	if q.IsDoubleDeep() {
		for _, nestedQuery := range q.Queries {
			nestedQuery.DeepFilter(keepQuery)
		}
		return q
	}

	// Otherwise, this is the level before the lowest
	// We can flatten the cache

	nodesToDiscard := []string{}
	for nodeKey, nestedQuery := range q.Queries {

		// if we need to filter this
		if !keepQuery(nestedQuery) {
			nodesToDiscard = append(nodesToDiscard, nodeKey)
		}

	}

	// Delete the nodes that have been filtered
	for _, nodeKey := range nodesToDiscard {
		delete(q.result, nodeKey)
		delete(q.Queries, nodeKey)
	}

	return q
}
