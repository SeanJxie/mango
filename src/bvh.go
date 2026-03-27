package mango

import (
	"slices"
)

type BVH struct {
	leftChild  Shape
	rightChild Shape
	bbox       *Aabb
}

func BuildBVH(objects []Shape, start, end int) *BVH {
	if len(objects) == 0 {
		return nil
	}

	bbox := &Aabb{}
	for i := start; i < end; i++ {
		bbox = NewAabbFromUnion(bbox, objects[i].GetBoundingBox())
	}

	axis := bbox.GetLongestAxis()

	compareFunction := func(a Shape, b Shape) int {
		aAxisInterval := a.GetBoundingBox().GetAxisInterval(axis)
		bAxisInterval := b.GetBoundingBox().GetAxisInterval(axis)

		if aAxisInterval.Min < bAxisInterval.Max {
			return 1
		}
		return -1
	}

	objectSpan := end - start
	var left, right Shape

	switch objectSpan {
	case 1:
		left = objects[start]
		right = objects[start]
	case 2:
		left = objects[start]
		right = objects[start+1]
	default:
		slices.SortStableFunc(objects[start:end], compareFunction)

		mid := start + objectSpan/2

		left = BuildBVH(objects, start, mid)
		right = BuildBVH(objects, mid, end)
	}

	return &BVH{
		leftChild:  left,
		rightChild: right,
		bbox:       bbox,
	}
}

func (bvh *BVH) Intersect(ray *Ray, tMin, tMax float64) (bool, *ShapeIntersection) {
	if bvh == nil {
		return false, nil
	}

	if hit, _ := bvh.bbox.Intersect(ray, tMin, tMax); !hit {
		return false, nil
	}

	hitLeftChild, recLeft := bvh.leftChild.Intersect(ray, tMin, tMax)

	if hitLeftChild {
		tMax = recLeft.T
	}

	hitRightChild, recRight := bvh.rightChild.Intersect(ray, tMin, tMax)

	if hitRightChild {
		return true, recRight
	}

	if hitLeftChild {
		return true, recLeft
	}

	return false, nil
}

func (bvh *BVH) IntersectBool(ray *Ray, tMin, tMax float64) bool {
	if bvh == nil {
		return false
	}

	if hit := bvh.bbox.IntersectBool(ray, tMin, tMax); !hit {
		return false
	}

	// Only need to use non-bool version here for occlusion (objects block objects behind them).
	hitLeftChild, recLeft := bvh.leftChild.Intersect(ray, tMin, tMax)

	if hitLeftChild {
		tMax = recLeft.T
	}

	hitRightChild := bvh.rightChild.IntersectBool(ray, tMin, tMax)

	if hitRightChild {
		return true
	}

	if hitLeftChild {
		return true
	}

	return false
}

func (bvh *BVH) GetBoundingBox() *Aabb {
	return bvh.bbox
}

func BoxCompare(a, b Shape, axis int) int {
	aAxisInterval := a.GetBoundingBox().GetAxisInterval(axis)
	bAxisInterval := b.GetBoundingBox().GetAxisInterval(axis)

	if aAxisInterval.Min < bAxisInterval.Max {
		return 1
	}
	return -1
}

// Redundant.
func (bvg *BVH) SurfaceArea() float64 {
	return 0
}

// Redundant.
func (bvg *BVH) Pdf() float64 {
	return 0
}

// Redundant.
func (bvg *BVH) SamplePoint(sample Vector2) Vector3 {
	return Zero3
}

// TODO: Full PBRT-style BVH implementation
//
// type BvhNode struct {
// 	LeftChild, RightChild *BvhNode
// 	BoundingBox           *Aabb
// 	SplittingAxis         int
// 	FirstElementOffset    int
// 	NumElements           int
// }

// func NewBvhNodeLeaf(firstElementOffset, numElements int, boundingBox *Aabb) *BvhNode {
// 	return &BvhNode{
// 		LeftChild:          nil,
// 		RightChild:         nil,
// 		BoundingBox:        boundingBox,
// 		FirstElementOffset: firstElementOffset,
// 		NumElements:        numElements,
// 	}
// }

// // Non-leaf node.
// func NewBvhNodeInternal(splittingAxis int, leftChild, rightChild *BvhNode) *BvhNode {
// 	return &BvhNode{
// 		LeftChild:     leftChild,
// 		RightChild:    rightChild,
// 		BoundingBox:   NewAabbFromUnion(leftChild.BoundingBox, rightChild.BoundingBox),
// 		SplittingAxis: splittingAxis,
// 		NumElements:   0,
// 	}
// }

// type BvhElement struct {
// 	SliceIndex  int
// 	BoundingBox *Aabb
// }

// func ToBvhElements(shapes []Shape) []*BvhElement {
// 	ret := make([]*BvhElement, len(shapes))
// 	for i := 0; i < len(shapes); i++ {
// 		ret[i].SliceIndex = i
// 		ret[i].BoundingBox = shapes[i].GetBoundingBox()
// 	}
// 	return ret
// }

// func (elem *BvhElement) Centroid() Vector3 {
// 	min := Vector3{elem.BoundingBox.x.Min, elem.BoundingBox.y.Min, elem.BoundingBox.z.Min}
// 	max := Vector3{elem.BoundingBox.x.Max, elem.BoundingBox.y.Max, elem.BoundingBox.z.Max}
// 	return Add3(ScalarMultiply3(min, 0.5), ScalarMultiply3(max, 0.5))
// }

// type Bvh struct {
// 	elements []*BvhElement
// 	root     *BvhNode
// }

// func BuildBvhRecursive(bvhElements []*BvhElement, totalNodes, orderedElementsOffset int, orderedElements []Shape) *BvhNode {
// 	totalNodes++
// 	bbox := &Aabb{}
// 	for i := 0; i < len(bvhElements); i++ {
// 		bbox = NewAabbFromUnion(bbox, bvhElements[i].BoundingBox)
// 	}

// 	if len(bvhElements) == 1 {

// 	}

// 	return nil
// }
