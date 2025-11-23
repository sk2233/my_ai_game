package main

// CollisionBox 碰撞盒接口
type CollisionBox interface {
	GetCollisionBox() (left, right, top, bottom float64)
}

// CheckCollision 检查两个碰撞盒是否发生碰撞（矩形碰撞）
// 返回 true 表示发生碰撞
func CheckCollision(a, b CollisionBox) bool {
	aLeft, aRight, aTop, aBottom := a.GetCollisionBox()
	bLeft, bRight, bTop, bBottom := b.GetCollisionBox()

	// AABB 碰撞检测：检查两个矩形是否重叠
	// 如果两个矩形在 X 轴和 Y 轴上都重叠，则发生碰撞
	return aLeft < bRight && aRight > bLeft && aTop < bBottom && aBottom > bTop
}
