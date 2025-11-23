package main

import (
	"image"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	// 游戏帧率（假设60fps）
	gameFPS = 60.0
)

// AnimationState 动画状态
type AnimationState int

const (
	StateIdle AnimationState = iota
	StateMove
	StateJumpBefore
	StateJumpLoop
	StateJumpEnd
	StateDie
	StateFly
)

// Animation 动画结构体
type Animation struct {
	Image         *ebiten.Image // 动画图片（精灵表）
	FrameCount    int           // 总帧数
	FrameWidth    int           // 每帧宽度
	FrameHeight   int           // 每帧高度
	Loop          bool          // 是否循环播放
	FPS           float64       // 动画播放速度（帧/秒）
	OriginOffsetY float64       // 动画原点Y偏移（相对于帧底部，正数向上偏移）
}

// NewAnimation 创建新动画
// imagePath: 图片路径
// frameCount: 帧数
// loop: 是否循环播放
// onComplete: 播放完成回调
// fps: 动画播放速度（帧/秒）
// originOffsetY: 动画原点Y偏移（相对于帧底部，正数向上偏移）
func NewAnimation(imagePath string, frameCount int, loop bool, fps float64, originOffsetY float64) *Animation {
	img, _, err := ebitenutil.NewImageFromFile(imagePath)
	if err != nil {
		log.Fatalf("加载动画图片失败 %s: %v", imagePath, err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 计算每帧宽度（水平均等拆分）
	frameWidth := width / frameCount

	return &Animation{
		Image:         img,
		FrameCount:    frameCount,
		FrameWidth:    frameWidth,
		FrameHeight:   height,
		Loop:          loop,
		FPS:           fps,
		OriginOffsetY: originOffsetY,
	}
}

// GetFrame 获取指定帧的图片
func (a *Animation) GetFrame(frameIndex int) *ebiten.Image {
	if frameIndex < 0 || frameIndex >= a.FrameCount {
		return nil
	}

	// 从精灵表中提取单帧
	frameRect := image.Rect(
		frameIndex*a.FrameWidth,
		0,
		(frameIndex+1)*a.FrameWidth,
		a.FrameHeight,
	)

	return a.Image.SubImage(frameRect).(*ebiten.Image)
}

// AnimationController 动画控制器
// 只负责更新当前动画的下一帧和判断是否动画结束
type AnimationController struct {
	currentState AnimationState
	currentFrame float64 // 当前帧（浮点数，用于平滑播放）
	animations   map[AnimationState]*Animation
}

// NewAnimationController 创建动画控制器
func NewAnimationController() *AnimationController {
	controller := &AnimationController{
		currentState: StateIdle,
		currentFrame: 0,
		animations:   make(map[AnimationState]*Animation),
	}

	// 加载所有动画（不设置回调，由Player控制状态切换）
	// 参数：图片路径, 帧数, 是否循环, 完成回调, 播放速度(FPS), 原点Y偏移
	controller.animations[StateIdle] = NewAnimation("res/image/idle.png", 39, true, 20.0, 22)
	controller.animations[StateMove] = NewAnimation("res/image/move.png", 26, true, 20.0, 45)
	controller.animations[StateJumpBefore] = NewAnimation("res/image/jump_before.png", 10, false, 27.0, 16)
	controller.animations[StateJumpLoop] = NewAnimation("res/image/jump_loop.png", 1, true, 1.0, 35)
	controller.animations[StateJumpEnd] = NewAnimation("res/image/jump_end.png", 7, false, 27.0, 13)
	controller.animations[StateDie] = NewAnimation("res/image/die.png", 30, false, 20.0, 18)
	controller.animations[StateFly] = NewAnimation("res/image/fly.png", 22, true, 20.0, 0.0)

	return controller
}

// SetState 设置动画状态
func (ac *AnimationController) SetState(state AnimationState) {
	if ac.currentState != state {
		ac.currentState = state
		ac.currentFrame = 0
	}
}

// GetState 获取当前动画状态
func (ac *AnimationController) GetState() AnimationState {
	return ac.currentState
}

// Update 更新动画帧（只更新当前动画的下一帧）
func (ac *AnimationController) Update() {
	anim := ac.animations[ac.currentState]
	if anim == nil {
		return
	}

	// 使用当前动画的FPS计算帧步进
	frameStep := anim.FPS / gameFPS
	ac.currentFrame += frameStep

	// 处理帧数溢出
	if ac.currentFrame >= float64(anim.FrameCount) {
		if anim.Loop {
			// 循环播放
			ac.currentFrame = ac.currentFrame - float64(anim.FrameCount)
		} else {
			// 非循环动画，保持在最后一帧
			ac.currentFrame = float64(anim.FrameCount) - 1
		}
	}
}

// IsFinished 判断当前动画是否播放完毕（仅对非循环动画有效）
func (ac *AnimationController) IsFinished() bool {
	anim := ac.animations[ac.currentState]
	if anim == nil || anim.Loop {
		return false
	}
	return ac.currentFrame >= float64(anim.FrameCount)-0.1 // 允许小的浮点误差
}

// GetCurrentFrame 获取当前帧图片
func (ac *AnimationController) GetCurrentFrame() *ebiten.Image {
	anim := ac.animations[ac.currentState]
	if anim == nil {
		return nil
	}

	frameIndex := int(ac.currentFrame)
	return anim.GetFrame(frameIndex)
}

// GetFrameSize 获取当前动画帧的尺寸
func (ac *AnimationController) GetFrameSize() (width, height int) {
	anim := ac.animations[ac.currentState]
	if anim == nil {
		return 0, 0
	}
	return anim.FrameWidth, anim.FrameHeight
}

// GetCurrentFPS 获取当前动画的播放速度（帧/秒）
func (ac *AnimationController) GetCurrentFPS() float64 {
	anim := ac.animations[ac.currentState]
	if anim == nil {
		return 0
	}
	return anim.FPS
}

// GetCurrentOriginOffsetY 获取当前动画的原点Y偏移
func (ac *AnimationController) GetCurrentOriginOffsetY() float64 {
	anim := ac.animations[ac.currentState]
	if anim == nil {
		return 0
	}
	return anim.OriginOffsetY
}

// SetAnimationFPS 设置指定动画的播放速度
func (ac *AnimationController) SetAnimationFPS(state AnimationState, fps float64) {
	anim := ac.animations[state]
	if anim != nil {
		anim.FPS = fps
	}
}

// SetAnimationOriginOffsetY 设置指定动画的原点Y偏移
func (ac *AnimationController) SetAnimationOriginOffsetY(state AnimationState, offsetY float64) {
	anim := ac.animations[state]
	if anim != nil {
		anim.OriginOffsetY = offsetY
	}
}
