package main

import (
	"math/rand"
	"time"
)

type MapItem struct {
	Index       int  // 从左往右数下标为几
	HasRoad     bool // 该位置是否有道路
	HasObstacle bool // 该道路是否有障碍
	HasMonster  bool // 该道路是否有怪物
	HasTool     bool // 该道路上是否有道具
}

// GenMap 生成地图
// count: 生成的地图列数
// 规则：
//   - 每一列可能有道路，也可能没有道路
//   - 只有有道路的情况下才能有障碍
//   - 障碍不能连续出现
//   - 怪物不能连续出现
//   - 怪物不会出现在连续道路段的边缘（道路段的开始和结束位置）
//   - 最多连续 2 个没有道路
func GenMap(count int) []*MapItem {
	if count <= 0 {
		return nil
	}

	// 初始化随机数种子
	random := rand.New(rand.NewSource(time.Now().Unix()))

	result := make([]*MapItem, 0, count)
	noRoadCount := 0         // 当前连续没有道路的数量
	prevHasObstacle := false // 上一个位置是否有障碍
	prevHasMonster := false  // 上一个位置是否有怪物

	for i := 0; i < count; i++ {
		item := &MapItem{
			Index:       i,
			HasRoad:     false,
			HasObstacle: false,
		}

		// 决定是否有道路
		// 前 10 块地图必须有道路，防止角色生成后掉下去
		if i < 10 {
			item.HasRoad = true
			noRoadCount = 0
		} else if noRoadCount >= 2 {
			// 如果已经连续 2 个没有道路，必须生成道路
			item.HasRoad = true
			noRoadCount = 0
		} else {
			// 80% 概率有道路，20% 概率没有道路
			item.HasRoad = random.Float32() < 0.8
			if item.HasRoad {
				noRoadCount = 0
			} else {
				noRoadCount++
			}
		}

		// 如果有道路，决定是否有障碍
		// 障碍不能连续出现
		if item.HasRoad && !prevHasObstacle {
			// 10% 概率有障碍
			item.HasObstacle = random.Float32() < 0.1
			prevHasObstacle = item.HasObstacle
		} else {
			prevHasObstacle = false
		}

		// 如果有道路且没有障碍物，决定是否有怪物
		// 怪物不能连续出现
		if item.HasRoad && !item.HasObstacle && !prevHasMonster {
			// 5% 概率有怪物
			item.HasMonster = random.Float32() < 0.05
			prevHasMonster = item.HasMonster
		} else {
			prevHasMonster = false
		}

		// 道具生成概率固定为 3%
		item.HasTool = random.Float32() < 0.03
		result = append(result, item)
	}

	// 移除连续道路段边缘的怪物
	// 遍历所有位置，如果当前位置是道路段的开始或结束，则移除怪物
	for i := 0; i < count; i++ {
		item := result[i]
		if !item.HasRoad {
			continue
		}

		// 检查是否是道路段的开始（前一个位置没有道路）
		isRoadStart := i == 0 || !result[i-1].HasRoad
		// 检查是否是道路段的结束（下一个位置没有道路）
		isRoadEnd := i == count-1 || !result[i+1].HasRoad

		// 如果是道路段的边缘，移除怪物
		if isRoadStart || isRoadEnd {
			item.HasMonster = false
		}
	}

	return result
}
