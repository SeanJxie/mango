package trace

import (
	"slices"
)

type BVH struct {
	leftChild  Shape
	rightChild Shape
	bbox       *Aabb
}

func NewBVH(objects []Shape, start, end int) *BVH {
	if len(objects) == 0 {
		return nil
	}

	bbox := &Aabb{}
	for i := start; i < end; i++ {
		bbox = NewAabbFromUnion(bbox, objects[i].GetBoundingBox())
	}

	axis := bbox.GetLongestAxis()

	var compareFunction func(Shape, Shape) int
	switch axis {
	case 0:
		compareFunction = BoxCompareX
	case 1:
		compareFunction = BoxCompareY
	default:
		compareFunction = BoxCompareZ
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

		left = NewBVH(objects, start, mid)
		right = NewBVH(objects, mid, end)
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

func BoxCompareX(a, b Shape) int {
	return BoxCompare(a, b, 0)
}

func BoxCompareY(a, b Shape) int {
	return BoxCompare(a, b, 1)
}

func BoxCompareZ(a, b Shape) int {
	return BoxCompare(a, b, 2)
}
