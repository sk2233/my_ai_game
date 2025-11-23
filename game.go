package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	// 窗口尺寸
	windowWidth  = 1280
	windowHeight = 720
	// 每个地图单元的宽度（像素）
	mapItemWidth = 120.0
	// 相机移动速度（像素/帧）
	cameraSpeed = 5.0
)

// Game 实现 ebiten.Game 接口
type Game struct {
	MapItems  []*MapItem
	Obstacles []*Obstacle // 所有障碍物对象（包括 grass 和 obstacle）
	Player    *Player     // 玩家
	CameraX   float64     // 相机位置（用于滚屏）

	// 图片资源
	bgImage       *ebiten.Image
	grassImage    *ebiten.Image
	obstacleImage *ebiten.Image
	monsterImage  *ebiten.Image
	toolImage     *ebiten.Image

	// 音频资源
	audioManager  *AudioManager // 音频管理器
	hasStoppedBGM bool          // 是否已停止背景音乐
}

func NewGame(count int) *Game {
	game := &Game{
		MapItems:  GenMap(count),
		Obstacles: make([]*Obstacle, 0),
		CameraX:   0,
	}

	// 初始化音频管理器（会自动加载并播放背景音乐）
	game.audioManager = NewAudioManager()

	// 加载图片资源
	var err error
	game.bgImage, _, err = ebitenutil.NewImageFromFile("res/image/bg.png")
	if err != nil {
		log.Fatalf("加载背景图片失败: %v", err)
	}

	game.grassImage, _, err = ebitenutil.NewImageFromFile("res/image/grass.png")
	if err != nil {
		log.Fatalf("加载道路图片失败: %v", err)
	}

	game.obstacleImage, _, err = ebitenutil.NewImageFromFile("res/image/obstacle.png")
	if err != nil {
		log.Fatalf("加载障碍图片失败: %v", err)
	}

	game.monsterImage, _, err = ebitenutil.NewImageFromFile("res/image/most_pix.png")
	if err != nil {
		log.Fatalf("加载怪物图片失败: %v", err)
	}

	game.toolImage, _, err = ebitenutil.NewImageFromFile("res/image/tool.png")
	if err != nil {
		log.Fatalf("加载道具图片失败: %v", err)
	}

	// 根据 MapItems 创建 Obstacle 对象
	game.initObstacles()

	// 初始化玩家，位置在屏幕中心
	// 玩家原点在底部中心，所以 X 在屏幕中心，Y 在窗口底部
	playerX := float64(windowWidth) / 2.0
	playerY := float64(windowHeight) / 2.0
	game.Player = NewPlayer(playerX, playerY, game.audioManager)

	return game
}

// initObstacles 根据 MapItems 初始化所有障碍物对象
func (g *Game) initObstacles() {
	// 预先计算所有图片尺寸，避免在循环中重复计算
	grassBounds := g.grassImage.Bounds()
	grassWidth := float64(grassBounds.Dx())
	grassHeight := float64(grassBounds.Dy())

	obstacleBounds := g.obstacleImage.Bounds()
	obstacleWidth := float64(obstacleBounds.Dx())
	obstacleHeight := float64(obstacleBounds.Dy())

	monsterBounds := g.monsterImage.Bounds()
	monsterHeight := float64(monsterBounds.Dy())

	toolBounds := g.toolImage.Bounds()
	toolWidth := float64(toolBounds.Dx())
	toolHeight := 120.0 // 道具高度固定为 120

	// 道路块在地图最下面的位置
	grassY := float64(windowHeight) - grassHeight

	// 预先分配容量，减少内存重新分配
	estimatedCount := len(g.MapItems) * 2 // 估算：每个 MapItem 平均 2 个障碍物（道路 + 其他）
	g.Obstacles = make([]*Obstacle, 0, estimatedCount)

	// 遍历所有 MapItem，创建对应的 Obstacle 对象
	for _, item := range g.MapItems {
		// 计算道路块的绝对坐标
		grassX := float64(item.Index) * grassWidth

		// 如果有道路，创建 grass Obstacle
		if item.HasRoad {
			grass := NewObstacle(grassX, grassY, grassX, grassY, grassWidth, grassHeight, g.grassImage, ObstacleTypeGrass)
			g.Obstacles = append(g.Obstacles, grass)

			// 如果有障碍，创建 obstacle Obstacle
			if item.HasObstacle {
				obstacleY := grassY - obstacleHeight
				obstacle := NewObstacle(grassX, obstacleY, grassX, obstacleY, obstacleWidth, obstacleHeight, g.obstacleImage, ObstacleTypeObstacle)
				g.Obstacles = append(g.Obstacles, obstacle)
			}

			// 如果有怪物，创建 monster Obstacle
			if item.HasMonster {
				// 怪物放在道路块上面，怪物的碰撞盒与绘制相比略小
				monsterDrawY := grassY - monsterHeight
				monsterCollisionX := grassX + 25
				monsterCollisionY := monsterDrawY + 12
				monster := NewObstacle(grassX, monsterDrawY, monsterCollisionX, monsterCollisionY, 70, 145, g.monsterImage, ObstacleTypeMonster)
				g.Obstacles = append(g.Obstacles, monster)
			}

			// 如果有道具，创建 tool Obstacle
			if item.HasTool {
				toolY := 120.0 // 道具 Y 坐标固定为 120
				tool := NewObstacle(grassX, toolY, grassX, toolY, toolWidth, toolHeight, g.toolImage, ObstacleTypeTool)
				g.Obstacles = append(g.Obstacles, tool)
			}
		}
	}
}

// Update 每帧更新游戏逻辑
func (g *Game) Update() error {
	// 更新玩家状态（传入障碍物列表和地图宽度用于碰撞检测和边界限制，以及相机位置用于死亡检测）
	if g.Player != nil {
		mapWidth := float64(len(g.MapItems)) * mapItemWidth
		g.Player.Update(g.Obstacles, mapWidth, g.CameraX)

		// 检查玩家是否死亡
		if g.Player.IsDead {
			// 玩家死亡后，停止背景音乐（只停止一次）
			if !g.hasStoppedBGM {
				g.audioManager.PauseBGM()
				g.hasStoppedBGM = true
			}
			// 玩家死亡后，相机不再移动
			return nil
		}

		// 检查玩家与道具的碰撞，移除被触碰的道具
		g.removeTouchedTools()
	}

	// 更新相机位置，自动向右移动（只有在玩家未死亡时才移动）
	g.updateCamera()

	return nil
}

// removeTouchedTools 移除玩家触碰到的道具，并触发飞行状态
func (g *Game) removeTouchedTools() {
	if g.Player == nil {
		return
	}

	// 从后往前遍历，避免删除时索引错乱
	for i := len(g.Obstacles) - 1; i >= 0; i-- {
		obstacle := g.Obstacles[i]
		// 如果是道具且与玩家发生碰撞
		if obstacle.Type == ObstacleTypeTool && CheckCollision(g.Player, obstacle) {
			// 触发飞行状态
			if !g.Player.IsFlying {
				g.Player.IsFlying = true
				g.Player.Y = 240
				g.Player.X = g.CameraX + float64(windowWidth)/2.0
				g.Player.Animation.SetState(StateFly)
			}
			g.Player.flyFrameCount = 0
			// 从切片中移除该元素
			g.Obstacles = append(g.Obstacles[:i], g.Obstacles[i+1:]...)
		}
	}
}

// updateCamera 更新相机位置，自动向右移动
// 相机每帧向右移动，速度根据玩家飞行状态调整
// 范围：0 ～ 生成地图块数量 * 120 - 屏幕宽度
// 如果玩家死亡，相机停止移动
func (g *Game) updateCamera() {
	// 如果玩家死亡，相机停止移动
	if g.Player != nil && g.Player.IsDead {
		return
	}

	// 计算相机的最大移动范围
	// 地图总宽度 = 地图块数量 * 120
	// 最大相机位置 = 地图总宽度 - 屏幕宽度
	maxCameraX := float64(len(g.MapItems))*mapItemWidth - float64(windowWidth)
	if maxCameraX < 0 {
		maxCameraX = 0
	}

	// 根据玩家飞行状态调整相机移动速度
	var currentSpeed float64
	if g.Player != nil && g.Player.IsFlying {
		currentSpeed = 15.0 // 飞行状态下每帧 15 像素（与玩家飞行速度同步）
	} else {
		currentSpeed = cameraSpeed // 正常状态下每帧 5 像素
	}

	// 如果相机还未到达边界，继续向右移动
	if g.CameraX < maxCameraX {
		g.CameraX += currentSpeed
		// 确保不超过边界
		if g.CameraX > maxCameraX {
			g.CameraX = maxCameraX
		}
	}
	// 如果已经到达边界，相机停止移动（保持在 maxCameraX）
}

// Draw 每帧绘制游戏画面
func (g *Game) Draw(screen *ebiten.Image) {
	// 绘制背景（无限滚动）
	g.drawBackground(screen)

	// 绘制道路和障碍
	g.drawMap(screen)

	// 绘制玩家碰撞盒（半透明绿色）
	g.drawPlayer(screen)

	// 在左上角显示帧率
	fps := fmt.Sprintf("FPS: %.0f", ebiten.ActualFPS())
	ebitenutil.DebugPrintAt(screen, fps, 10, 10)
}

// drawBackground 绘制背景图片（上下铺满，左右无限生成）
func (g *Game) drawBackground(screen *ebiten.Image) {
	bgBounds := g.bgImage.Bounds()
	bgWidth := float64(bgBounds.Dx())

	// 计算需要绘制的背景图片数量（左右各多绘制一张以确保无缝滚动）
	startX := int(g.CameraX/bgWidth) - 1
	endX := int((g.CameraX+float64(windowWidth))/bgWidth) + 1

	// 复用 DrawImageOptions 对象，减少内存分配
	op := &ebiten.DrawImageOptions{}

	for i := startX; i <= endX; i++ {
		x := float64(i)*bgWidth - g.CameraX
		op.GeoM.Reset()
		op.GeoM.Translate(x, 0)
		screen.DrawImage(g.bgImage, op)
	}
}

// drawMap 绘制地图（道路和障碍）
func (g *Game) drawMap(screen *ebiten.Image) {
	// 遍历所有障碍物，调用其 Draw 方法
	for _, obstacle := range g.Obstacles {
		obstacle.Draw(screen, g.CameraX)
	}
}

// drawPlayer 绘制玩家
func (g *Game) drawPlayer(screen *ebiten.Image) {
	if g.Player != nil {
		g.Player.Draw(screen, g.CameraX)
	}
}

// Layout 返回游戏逻辑尺寸
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}
