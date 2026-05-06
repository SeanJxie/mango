package mango

import (
	"math"
	"sort"
)

const (
	bvhMaxLeafPrimitives = 4
	bvhBucketCount       = 12
	bvhTraversalStack    = 128
)

type BVH struct {
	nodes      []bvhNode
	primitives []Shape
	bbox       *Aabb
}

type bvhNode struct {
	bbox         Aabb
	left, right  int
	first, count int
	splitAxis    int
}

type bvhPrimitiveInfo struct {
	shape    Shape
	bbox     *Aabb
	centroid Vector3
}

type bvhBucket struct {
	count int
	bbox  *Aabb
}

func BuildBVH(objects []Shape, start, end int) *BVH {
	if start < 0 {
		start = 0
	}
	if end > len(objects) {
		end = len(objects)
	}
	if start >= end {
		return nil
	}

	primitiveInfos := make([]bvhPrimitiveInfo, end-start)
	for i := range primitiveInfos {
		shape := objects[start+i]
		bbox := shape.GetBoundingBox()
		primitiveInfos[i] = bvhPrimitiveInfo{
			shape:    shape,
			bbox:     bbox,
			centroid: bbox.Centroid(),
		}
	}

	bvh := &BVH{
		nodes:      make([]bvhNode, 0, 2*len(primitiveInfos)-1),
		primitives: make([]Shape, 0, len(primitiveInfos)),
	}
	root := bvh.buildRecursive(primitiveInfos)
	bvh.bbox = copyAabb(&bvh.nodes[root].bbox)

	return bvh
}

func (bvh *BVH) buildRecursive(primitives []bvhPrimitiveInfo) int {
	nodeIndex := len(bvh.nodes)
	bvh.nodes = append(bvh.nodes, bvhNode{})

	bounds := boundsOfPrimitives(primitives)
	if len(primitives) <= bvhMaxLeafPrimitives {
		bvh.nodes[nodeIndex] = bvh.makeLeaf(bounds, primitives)
		return nodeIndex
	}

	centroidBounds := boundsOfCentroids(primitives)
	axis := centroidBounds.GetLongestAxis()
	if centroidBounds.GetAxisInterval(axis).GetSize() <= Epsilon {
		bvh.nodes[nodeIndex] = bvh.makeLeaf(bounds, primitives)
		return nodeIndex
	}

	mid := partitionPrimitives(primitives, axis, bounds, centroidBounds)
	if mid <= 0 || mid >= len(primitives) {
		sortPrimitivesByAxis(primitives, axis)
		mid = len(primitives) / 2
	}

	left := bvh.buildRecursive(primitives[:mid])
	right := bvh.buildRecursive(primitives[mid:])
	bvh.nodes[nodeIndex] = bvhNode{
		bbox:      *bounds,
		left:      left,
		right:     right,
		splitAxis: axis,
	}

	return nodeIndex
}

func (bvh *BVH) makeLeaf(bounds *Aabb, primitives []bvhPrimitiveInfo) bvhNode {
	first := len(bvh.primitives)
	for _, primitive := range primitives {
		bvh.primitives = append(bvh.primitives, primitive.shape)
	}

	return bvhNode{
		bbox:  *bounds,
		first: first,
		count: len(primitives),
	}
}

func partitionPrimitives(primitives []bvhPrimitiveInfo, axis int, bounds, centroidBounds *Aabb) int {
	buckets := [bvhBucketCount]bvhBucket{}
	for _, primitive := range primitives {
		bucket := centroidBucketIndex(primitive.centroid, axis, centroidBounds)
		buckets[bucket].count++
		buckets[bucket].bbox = unionAabb(buckets[bucket].bbox, primitive.bbox)
	}

	bestSplit := -1
	bestCost := math.Inf(1)
	parentArea := bounds.SurfaceArea()
	if parentArea <= 0 {
		return 0
	}

	for split := 0; split < bvhBucketCount-1; split++ {
		leftBounds, rightBounds := (*Aabb)(nil), (*Aabb)(nil)
		leftCount, rightCount := 0, 0

		for i := 0; i <= split; i++ {
			leftBounds = unionAabb(leftBounds, buckets[i].bbox)
			leftCount += buckets[i].count
		}
		for i := split + 1; i < bvhBucketCount; i++ {
			rightBounds = unionAabb(rightBounds, buckets[i].bbox)
			rightCount += buckets[i].count
		}

		if leftCount == 0 || rightCount == 0 {
			continue
		}

		cost := 0.5 + (float64(leftCount)*leftBounds.SurfaceArea()+float64(rightCount)*rightBounds.SurfaceArea())/parentArea
		if cost < bestCost {
			bestCost = cost
			bestSplit = split
		}
	}

	if bestSplit < 0 {
		return 0
	}

	leafCost := float64(len(primitives))
	if len(primitives) <= bvhMaxLeafPrimitives && bestCost >= leafCost {
		return 0
	}

	mid := 0
	for i := range primitives {
		if centroidBucketIndex(primitives[i].centroid, axis, centroidBounds) <= bestSplit {
			primitives[i], primitives[mid] = primitives[mid], primitives[i]
			mid++
		}
	}

	return mid
}

func (bvh *BVH) Intersect(ray *Ray, tMin, tMax float64) (bool, *ShapeIntersection) {
	if bvh == nil || len(bvh.nodes) == 0 {
		return false, nil
	}

	invDir := inverseRayDirection(ray)
	stackStorage := [bvhTraversalStack]int{0}
	stack := stackStorage[:1]
	closest := tMax
	var closestIntersection *ShapeIntersection

	for len(stack) > 0 {
		nodeIndex := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := &bvh.nodes[nodeIndex]
		if !node.bbox.IntersectWithInv(ray, invDir, tMin, closest) {
			continue
		}

		if node.isLeaf() {
			for i := node.first; i < node.first+node.count; i++ {
				hit, intersection := bvh.primitives[i].Intersect(ray, tMin, closest)
				if hit {
					closest = intersection.T
					closestIntersection = intersection
				}
			}
			continue
		}

		nearChild, farChild := node.left, node.right
		if axisValue(ray.Direction, node.splitAxis) < 0 {
			nearChild, farChild = farChild, nearChild
		}
		stack = append(stack, farChild, nearChild)
	}

	return closestIntersection != nil, closestIntersection
}

func (bvh *BVH) IntersectBool(ray *Ray, tMin, tMax float64) bool {
	if bvh == nil || len(bvh.nodes) == 0 {
		return false
	}

	invDir := inverseRayDirection(ray)
	stackStorage := [bvhTraversalStack]int{0}
	stack := stackStorage[:1]

	for len(stack) > 0 {
		nodeIndex := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := &bvh.nodes[nodeIndex]
		if !node.bbox.IntersectWithInv(ray, invDir, tMin, tMax) {
			continue
		}

		if node.isLeaf() {
			for i := node.first; i < node.first+node.count; i++ {
				if bvh.primitives[i].IntersectBool(ray, tMin, tMax) {
					return true
				}
			}
			continue
		}

		nearChild, farChild := node.left, node.right
		if axisValue(ray.Direction, node.splitAxis) < 0 {
			nearChild, farChild = farChild, nearChild
		}
		stack = append(stack, farChild, nearChild)
	}

	return false
}

func (bvh *BVH) GetBoundingBox() *Aabb {
	return bvh.bbox
}

func (bvh *BVH) SurfaceArea() float64 {
	if bvh == nil || bvh.bbox == nil {
		return 0
	}
	return bvh.bbox.SurfaceArea()
}

func (bvh *BVH) Pdf() float64 {
	return 0
}

func (bvh *BVH) SamplePoint(sample Vector2) Vector3 {
	return Zero3
}

func (node *bvhNode) isLeaf() bool {
	return node.count > 0
}

func sortPrimitivesByAxis(primitives []bvhPrimitiveInfo, axis int) {
	sort.Slice(primitives, func(i, j int) bool {
		iCentroid := axisValue(primitives[i].centroid, axis)
		jCentroid := axisValue(primitives[j].centroid, axis)
		if iCentroid == jCentroid {
			return primitives[i].bbox.GetAxisInterval(axis).Min < primitives[j].bbox.GetAxisInterval(axis).Min
		}
		return iCentroid < jCentroid
	})
}

func centroidBucketIndex(centroid Vector3, axis int, centroidBounds *Aabb) int {
	axisInterval := centroidBounds.GetAxisInterval(axis)
	offset := (axisValue(centroid, axis) - axisInterval.Min) / axisInterval.GetSize()
	bucket := int(float64(bvhBucketCount) * Clamp(offset, 0, 0.999999))
	return min(bucket, bvhBucketCount-1)
}

func boundsOfPrimitives(primitives []bvhPrimitiveInfo) *Aabb {
	var bounds *Aabb
	for _, primitive := range primitives {
		bounds = unionAabb(bounds, primitive.bbox)
	}
	return bounds
}

func boundsOfCentroids(primitives []bvhPrimitiveInfo) *Aabb {
	first := primitives[0].centroid
	bounds := NewAabbFromExtrema(first, first)
	for i := 1; i < len(primitives); i++ {
		bounds = unionAabb(bounds, NewAabbFromExtrema(primitives[i].centroid, primitives[i].centroid))
	}
	return bounds
}

func unionAabb(a, b *Aabb) *Aabb {
	if a == nil {
		return copyAabb(b)
	}
	if b == nil {
		return copyAabb(a)
	}
	return NewAabbFromUnion(a, b)
}

func copyAabb(box *Aabb) *Aabb {
	if box == nil {
		return nil
	}
	copied := *box
	return &copied
}

func axisValue(v Vector3, axis int) float64 {
	switch axis {
	case 0:
		return v.X
	case 1:
		return v.Y
	default:
		return v.Z
	}
}

func inverseRayDirection(ray *Ray) Vector3 {
	return Vector3{
		X: 1 / ray.Direction.X,
		Y: 1 / ray.Direction.Y,
		Z: 1 / ray.Direction.Z,
	}
}
