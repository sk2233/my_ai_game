package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	// 玩家移动速度（像素/帧）
	playerSpeed = 5.5
	// 玩家碰撞盒尺寸
	playerCollisionWidth  = 60.0
	playerCollisionHeight = 346.0
	// 重力加速度（像素/帧²）
	gravity = 0.6
	// 跳跃初始速度（像素/帧）
	jumpSpeed = -18.0
	// 飞行速度（像素/帧）
	flySpeed = 15.0
	// 飞行持续时间（帧数）
	flyDurationFrames = 300
)

// Player 玩家结构体
type Player struct {
	X                 float64              // X 坐标（原点在底部中心）
	Y                 float64              // Y 坐标（原点在底部中心）
	VelocityY         float64              // 垂直速度
	IsOnGround        bool                 // 是否在地面上
	wasSpaceDown      bool                 // 上一帧是否按下了空格键
	wasOnGround       bool                 // 上一帧是否在地面上
	FacingLeft        bool                 // 是否面向左边
	Animation         *AnimationController // 动画控制器
	jumpSound         *audio.Player        // 跳跃音效播放器
	dieSound          *audio.Player        // 死亡音效播放器
	IsDead            bool                 // 是否死亡
	hasPlayedDieSound bool                 // 是否已播放死亡音效
	IsFlying          bool                 // 是否处于飞行状态
	flyFrameCount     int                  // 飞行帧计数器
}

// NewPlayer 创建新玩家
// x: 初始 X 坐标
// y: 初始 Y 坐标
// audioManager: 音频管理器，用于加载音效
func NewPlayer(x, y float64, audioManager *AudioManager) *Player {
	player := &Player{
		X:           x,
		Y:           y,
		Animation:   NewAnimationController(),
		FacingLeft:  false,
		wasOnGround: true,
	}

	// 从音频管理器加载跳跃音效
	player.jumpSound = audioManager.LoadJumpSound()
	// 从音频管理器加载死亡音效
	player.dieSound = audioManager.LoadDieSound()

	return player
}

// Update 更新玩家状态（处理移动和重力）
// obstacles: 障碍物列表，用于碰撞检测
// mapWidth: 地图总宽度，用于限制玩家移动范围
// cameraX: 相机 X 坐标，用于检测玩家是否移出屏幕
func (p *Player) Update(obstacles []*Obstacle, mapWidth float64, cameraX float64) {
	// 检查玩家是否死亡（碰撞盒完全移出屏幕）
	if !p.IsDead {
		p.checkDeath(cameraX)
	}

	// 如果玩家已死亡，只更新动画，不再处理其他操作
	if p.IsDead {
		// 继续更新动画，直到死亡动画播放完毕
		p.Animation.Update()
		return
	}

	// 处理飞行状态
	if p.IsFlying {
		p.updateFlyingState(mapWidth)
		// 更新动画状态（飞行状态）
		p.updateAnimationState(false)
		// 更新动画帧
		p.Animation.Update()
		return
	}

	isMoving := false

	// 处理左右移动（移动前检查碰撞和地图边界）
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		// 尝试向左移动
		newX := p.X - playerSpeed
		// 检查是否超出地图左边界（玩家碰撞盒的左边界不能小于0）
		minX := playerCollisionWidth / 2.0
		if newX >= minX && !p.wouldCollideHorizontal(newX, obstacles) {
			p.X = newX
			p.FacingLeft = true
		}
		isMoving = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		// 尝试向右移动
		newX := p.X + playerSpeed
		// 检查是否超出地图右边界（玩家碰撞盒的右边界不能大于地图宽度）
		maxX := mapWidth - playerCollisionWidth/2.0
		if newX <= maxX && !p.wouldCollideHorizontal(newX, obstacles) {
			p.X = newX
			p.FacingLeft = false
		}
		isMoving = true
	}

	// 处理跳跃（只有在地面上才能跳跃，且只在按键按下时触发一次）
	spacePressed := ebiten.IsKeyPressed(ebiten.KeySpace)
	if p.IsOnGround && spacePressed && !p.wasSpaceDown {
		p.VelocityY = jumpSpeed
		p.IsOnGround = false
		// 播放跳跃音效
		if p.jumpSound != nil {
			// 重置到开头并播放
			p.jumpSound.Rewind()
			p.jumpSound.Play()
		}
	}
	p.wasSpaceDown = spacePressed

	// 应用重力
	p.VelocityY += gravity

	// 更新 Y 坐标（向上方向不检查碰撞，允许穿越）
	p.Y += p.VelocityY

	// 检查与障碍物的碰撞（只检查向下和左右，不检查向上）
	p.checkCollisionWithObstacles(obstacles)

	// 更新动画状态（根据玩家状态切换）
	p.updateAnimationState(isMoving)

	// 更新动画帧
	p.Animation.Update()
}

// updateFlyingState 更新飞行状态
func (p *Player) updateFlyingState(mapWidth float64) {
	// 增加飞行帧计数器
	p.flyFrameCount++

	// 检查飞行帧数是否超过设定值
	if p.flyFrameCount >= flyDurationFrames {
		// 飞行结束，转换为 jump_loop 状态
		p.IsFlying = false
		p.flyFrameCount = 0
		p.Animation.SetState(StateJumpLoop)
		// 恢复重力影响
		return
	}

	// 飞行状态下每帧向右移动 10 像素
	newX := p.X + flySpeed
	// 检查是否超出地图右边界
	maxX := mapWidth - playerCollisionWidth/2.0
	if newX <= maxX {
		p.X = newX
	}
	// 飞行状态下不受重力影响，VelocityY 保持不变
	// 飞行状态下无视任何碰撞，不检查碰撞
}

// updateAnimationState 根据玩家状态更新动画状态
func (p *Player) updateAnimationState(isMoving bool) {
	currentState := p.Animation.GetState()

	// 如果玩家已死亡，不再切换动画状态（保持死亡动画）
	if p.IsDead {
		// 死亡动画播放完毕后会定格在最后一帧（非循环动画会自动停在最后一帧）
		return
	}

	// 如果玩家处于飞行状态，保持飞行动画
	if p.IsFlying {
		if currentState != StateFly {
			p.Animation.SetState(StateFly)
		}
		return
	}

	// 检测从地面到空中的过渡
	if p.wasOnGround && !p.IsOnGround {
		// 从地面到空中，转到JumpBefore（过渡动画）
		if currentState != StateJumpBefore {
			p.Animation.SetState(StateJumpBefore)
		}
	}

	// 检测从空中到地面的过渡
	if !p.wasOnGround && p.IsOnGround {
		// 从空中到地面，转到JumpEnd（过渡动画）
		if currentState != StateJumpEnd {
			p.Animation.SetState(StateJumpEnd)
		}
	}

	// 根据当前状态和玩家状态决定动画切换
	switch currentState {
	case StateJumpBefore:
		// 起跳过渡动画播放完毕后，转到JumpLoop（循环播放）
		if p.Animation.IsFinished() {
			p.Animation.SetState(StateJumpLoop)
		}
	case StateJumpLoop:
		// 跳跃循环，在空中时一直循环播放
		// 如果落地，会在上面的检测中切换到JumpEnd
	case StateJumpEnd:
		// 落地过渡动画播放完毕后，根据是否移动转到Idle或Move
		if p.Animation.IsFinished() {
			if isMoving {
				p.Animation.SetState(StateMove)
			} else {
				p.Animation.SetState(StateIdle)
			}
		}
	case StateIdle, StateMove:
		// 地面状态，根据是否移动切换（都是循环播放）
		if p.IsOnGround {
			if isMoving {
				if currentState != StateMove {
					p.Animation.SetState(StateMove)
				}
			} else {
				if currentState != StateIdle {
					p.Animation.SetState(StateIdle)
				}
			}
		}
		// 如果不在空中，上面的检测会处理从地面到空中的过渡
	}

	p.wasOnGround = p.IsOnGround
}

// wouldCollideHorizontal 检查水平移动是否会碰撞
// 怪物和道具不阻挡水平移动，允许玩家移动到碰撞位置以触发相应逻辑
func (p *Player) wouldCollideHorizontal(newX float64, obstacles []*Obstacle) bool {
	// 临时保存原位置
	oldX := p.X
	p.X = newX

	// 使用 CheckCollision 检查是否会与障碍物碰撞
	for _, obstacle := range obstacles {
		// 怪物和道具不阻挡水平移动
		if obstacle.Type == ObstacleTypeMonster || obstacle.Type == ObstacleTypeTool {
			continue
		}

		if CheckCollision(p, obstacle) {
			// 恢复原位置
			p.X = oldX
			return true
		}
	}

	// 恢复原位置
	p.X = oldX
	return false
}

// handleDeath 处理玩家死亡逻辑（提取公共方法）
func (p *Player) handleDeath() {
	if !p.IsDead {
		// 玩家刚死亡，切换到死亡动画状态
		p.IsDead = true
		p.Animation.SetState(StateDie)
	}
	// 播放死亡音效（只播放一次）
	if !p.hasPlayedDieSound && p.dieSound != nil {
		p.dieSound.Rewind()
		p.dieSound.Play()
		p.hasPlayedDieSound = true
	}
}

// checkDeath 检查玩家是否死亡（碰撞盒完全移出屏幕）
func (p *Player) checkDeath(cameraX float64) {
	// 获取玩家碰撞盒边界
	left, right, top, bottom := p.GetCollisionBox()

	// 计算碰撞盒在屏幕上的位置（相对于相机）
	screenLeft := left - cameraX
	screenRight := right - cameraX
	screenTop := top
	screenBottom := bottom

	// 检查碰撞盒是否完全移出屏幕
	// 完全移出屏幕的条件：右边界在屏幕左边，或左边界在屏幕右边，或下边界在屏幕上边，或上边界在屏幕下边
	if screenRight < 0 || screenLeft > float64(windowWidth) || screenBottom < 0 || screenTop > float64(windowHeight) {
		p.handleDeath()
	}
}

// checkCollisionWithObstacles 检查玩家与障碍物的碰撞
// 只检查向下和左右方向的碰撞，不检查向上方向（允许向上穿越）
// 怪物：触碰到怪物立即死亡
func (p *Player) checkCollisionWithObstacles(obstacles []*Obstacle) {
	p.IsOnGround = false

	// 遍历所有障碍物检查碰撞
	for _, obstacle := range obstacles {
		// 使用 CheckCollision 检查是否发生碰撞
		if !CheckCollision(p, obstacle) {
			continue
		}

		// 根据障碍物类型处理
		switch obstacle.Type {
		case ObstacleTypeMonster:
			// 如果是怪物，触碰到立即死亡
			p.handleDeath()
			// 触碰到怪物后不再检查其他障碍物
			return
		case ObstacleTypeTool:
			// 如果是道具，跳过（由 Game.Update 处理移除）
			continue
		}

		// 普通障碍物：检查向下方向的碰撞
		_, _, obstacleTop, _ := obstacle.GetCollisionBox()

		// 只检查向下方向的碰撞（玩家正在下落）
		if p.VelocityY >= 0 && p.Y > obstacleTop {
			// 玩家站在障碍物上
			p.Y = obstacleTop
			p.VelocityY = 0
			p.IsOnGround = true
		}
	}
}

// GetCollisionBox 获取碰撞盒边界
// 返回：左边界, 右边界, 上边界, 下边界
func (p *Player) GetCollisionBox() (left, right, top, bottom float64) {
	halfWidth := playerCollisionWidth / 2.0
	left = p.X - halfWidth
	right = p.X + halfWidth
	top = p.Y - playerCollisionHeight
	bottom = p.Y
	return
}

// GetPosition 获取玩家位置
func (p *Player) GetPosition() (x, y float64) {
	return p.X, p.Y
}

// SetPosition 设置玩家位置
func (p *Player) SetPosition(x, y float64) {
	p.X = x
	p.Y = y
}

// Draw 绘制玩家动画
// screen: 绘制目标
// cameraX: 相机 X 坐标（用于计算屏幕坐标）
func (p *Player) Draw(screen *ebiten.Image, cameraX float64) {
	frame := p.Animation.GetCurrentFrame()
	if frame == nil {
		return
	}

	frameWidth, frameHeight := p.Animation.GetFrameSize()

	// 缩放比例
	scale := 0.5

	// 计算缩放后的尺寸
	scaledWidth := float64(frameWidth) * scale
	scaledHeight := float64(frameHeight) * scale

	// 获取当前动画的原点Y偏移
	originOffsetY := p.Animation.GetCurrentOriginOffsetY()

	// 计算绘制位置（缩放后图像的左上角位置）
	// 以帧动画中间最下方为原点与玩家原地对齐
	// 玩家原点在底部中心，所以：
	// - X: 玩家X - 缩放后帧宽度/2
	// - Y: 玩家Y - 缩放后帧高度 + 原点Y偏移（偏移是相对于帧底部的，需要缩放）
	screenX := p.X - scaledWidth/2.0 - cameraX
	screenY := p.Y - scaledHeight + originOffsetY*scale

	// 创建绘制选项
	op := &ebiten.DrawImageOptions{}

	// 如果面向左边，先翻转（以图像原点，即左上角(0,0)为轴）
	if p.FacingLeft {
		// 图像原点在(0,0)，水平翻转会绕(0,0)翻转
		op.GeoM.Scale(-1, 1)
		// 翻转后图像在负X区域，需要向右移动frameWidth来补偿
		op.GeoM.Translate(float64(frameWidth), 0)
	}

	// 然后移动到图片中心，缩放（以中心为原点）
	op.GeoM.Translate(-float64(frameWidth)/2.0, -float64(frameHeight)/2.0)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(float64(frameWidth)/2.0*scale, float64(frameHeight)/2.0*scale)

	// 移动到绘制位置
	op.GeoM.Translate(screenX, screenY)

	// 绘制当前帧
	screen.DrawImage(frame, op)
}
