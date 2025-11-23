package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// 设置窗口大小
	ebiten.SetWindowSize(windowWidth, windowHeight)

	// 设置窗口标题
	ebiten.SetWindowTitle("雪莉酱の大冒险")

	// 禁用窗口调整大小
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	// 创建游戏实例
	game := NewGame(512)

	// 运行游戏
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
