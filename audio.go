package main

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const (
	// 音频采样率
	audioSampleRate = 44100
	// 背景音乐音量
	bgmVolume = 0.4
	// 跳跃音效音量
	soundVolume = 1
)

// AudioManager 音频管理器
type AudioManager struct {
	context   *audio.Context // 音频上下文
	bgmPlayer *audio.Player  // 背景音乐播放器
}

// NewAudioManager 创建音频管理器
func NewAudioManager() *AudioManager {
	manager := &AudioManager{
		context: audio.NewContext(audioSampleRate),
	}

	// 加载并播放背景音乐
	manager.loadBGM()

	return manager
}

// GetContext 获取音频上下文
func (am *AudioManager) GetContext() *audio.Context {
	return am.context
}

// loadBGM 加载并播放背景音乐
func (am *AudioManager) loadBGM() {
	// 打开背景音乐文件
	f, err := os.Open("res/audio/bgm.mp3")
	if err != nil {
		log.Printf("警告: 无法加载背景音乐: %v", err)
		return
	}

	// 读取整个文件到内存
	data, err := io.ReadAll(f)
	f.Close() // 立即关闭文件
	if err != nil {
		log.Printf("警告: 无法读取背景音乐文件: %v", err)
		return
	}

	// 从内存中的数据创建 Reader
	reader := bytes.NewReader(data)

	// 解码 MP3 文件
	stream, err := mp3.DecodeWithoutResampling(reader)
	if err != nil {
		log.Printf("警告: 无法解码背景音乐: %v", err)
		return
	}

	// 创建循环播放器（使用 InfiniteLoop 实现循环）
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	player, err := am.context.NewPlayer(loop)
	if err != nil {
		log.Printf("警告: 无法创建背景音乐播放器: %v", err)
		return
	}

	am.bgmPlayer = player
	player.SetVolume(bgmVolume) // 设置音量（0.0 到 1.0）
	player.Play()               // 开始播放
}

// SetBGMVolume 设置背景音乐音量
func (am *AudioManager) SetBGMVolume(volume float64) {
	if am.bgmPlayer != nil {
		am.bgmPlayer.SetVolume(volume)
	}
}

// PauseBGM 暂停背景音乐
func (am *AudioManager) PauseBGM() {
	if am.bgmPlayer != nil && am.bgmPlayer.IsPlaying() {
		am.bgmPlayer.Pause()
	}
}

// ResumeBGM 恢复背景音乐
func (am *AudioManager) ResumeBGM() {
	if am.bgmPlayer != nil && !am.bgmPlayer.IsPlaying() {
		am.bgmPlayer.Play()
	}
}

// LoadJumpSound 加载跳跃音效
// 返回音频播放器，如果加载失败返回 nil
func (am *AudioManager) LoadJumpSound() *audio.Player {
	// 打开跳跃音效文件
	f, err := os.Open("res/audio/jump.wav")
	if err != nil {
		// 如果文件不存在，只记录警告，不中断游戏
		return nil
	}

	// 读取整个文件到内存
	data, err := io.ReadAll(f)
	f.Close() // 立即关闭文件
	if err != nil {
		return nil
	}

	// 从内存中的数据创建 Reader
	reader := bytes.NewReader(data)

	// 解码 WAV 文件
	stream, err := wav.DecodeWithoutResampling(reader)
	if err != nil {
		return nil
	}

	// 创建播放器
	player, err := am.context.NewPlayer(stream)
	if err != nil {
		return nil
	}

	player.SetVolume(soundVolume) // 设置音量（0.0 到 1.0）
	return player
}

// LoadDieSound 加载死亡音效
// 返回音频播放器，如果加载失败返回 nil
func (am *AudioManager) LoadDieSound() *audio.Player {
	// 打开死亡音效文件
	f, err := os.Open("res/audio/die.mp3")
	if err != nil {
		// 如果文件不存在，只记录警告，不中断游戏
		return nil
	}

	// 读取整个文件到内存
	data, err := io.ReadAll(f)
	f.Close() // 立即关闭文件
	if err != nil {
		return nil
	}

	// 从内存中的数据创建 Reader
	reader := bytes.NewReader(data)

	// 解码 MP3 文件
	stream, err := mp3.DecodeWithoutResampling(reader)
	if err != nil {
		return nil
	}

	// 创建播放器
	player, err := am.context.NewPlayer(stream)
	if err != nil {
		return nil
	}

	player.SetVolume(soundVolume) // 设置音量（0.0 到 1.0）
	return player
}
