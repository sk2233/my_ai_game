package main

import "github.com/hajimehoshi/ebiten/v2"

// ObstacleType 障碍物类型枚举
type ObstacleType int

const (
	ObstacleTypeGrass    ObstacleType = iota // 道路（草地）
	ObstacleTypeObstacle                     // 障碍物
	ObstacleTypeMonster                      // 怪物
	ObstacleTypeTool                         // 道具
)

// Obstacle 障碍物类（用于 grass、obstacle、monster 和 tool）
type Obstacle struct {
	Dx, Dy        float64       // 绘制使用的 x y
	X, Y          float64       // 碰撞检查使用的 x y
	Width, Height float64       // 碰撞检查使用的 宽度与高度
	Image         *ebiten.Image // 图片资源
	Type          ObstacleType  // 障碍物类型
}

// NewObstacle 创建新障碍物
// dx, dy: 绘制坐标
// x, y: 碰撞盒坐标（左上角）
// width, height: 尺寸
// image: 图片资源
// obstacleType: 障碍物类型
func NewObstacle(dx, dy, x, y, width, height float64, image *ebiten.Image, obstacleType ObstacleType) *Obstacle {
	return &Obstacle{
		Dx:     dx,
		Dy:     dy,
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Image:  image,
		Type:   obstacleType,
	}
}

// GetCollisionBox 获取碰撞盒边界
// 返回：左边界, 右边界, 上边界, 下边界
func (o *Obstacle) GetCollisionBox() (left, right, top, bottom float64) {
	left = o.X
	right = o.X + o.Width
	top = o.Y
	bottom = o.Y + o.Height
	return
}

// GetPosition 获取障碍物位置
func (o *Obstacle) GetPosition() (x, y float64) {
	return o.X, o.Y
}

// SetPosition 设置障碍物位置
func (o *Obstacle) SetPosition(x, y float64) {
	o.X = x
	o.Y = y
}

// Draw 绘制障碍物
// screen: 绘制目标
// cameraX: 相机 X 坐标（用于计算屏幕坐标）
func (o *Obstacle) Draw(screen *ebiten.Image, cameraX float64) {
	if o.Image == nil {
		return
	}

	// 计算相对于相机的屏幕坐标
	screenX := o.Dx - cameraX
	screenY := o.Dy

	// 只绘制窗口内的内容（使用全局常量）
	if screenX+o.Width < 0 || screenX > float64(windowWidth) {
		return
	}
	if screenY+o.Height < 0 || screenY > float64(windowHeight) {
		return
	}

	// 绘制障碍物
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)
	screen.DrawImage(o.Image, op)
}
