package heartbeat

import (
	"math/rand"
)

// 获取n个随机节点，其中这些节点不能包括exclued中的节点
func ChooseRandomDataServers(n int, exclued map[int]string) (ds []string) {
	candidates := []string{}
	reverseExcluedMap := map[string]int{}
	for id, addr := range exclued {
		reverseExcluedMap[addr] = id
	}
	server := GetDataServers()
	for _, addr := range server {
		if _, ok := reverseExcluedMap[addr]; !ok {
			candidates = append(candidates, addr)
		}
	}
	if len(candidates) < n {
		return
	}
	// 从候选节点中随机选择不重复的n个节点
	// 生成len(candidates)的排列数
	perm := rand.Perm(len(candidates))
	for i := 0; i < n; i++ {
		ds = append(ds, candidates[perm[i]])
	}
	return
}
