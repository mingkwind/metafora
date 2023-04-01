package version

import (
	"api-server/internal/models"
	"fmt"
	"net/http"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// 判断是不是GET请求
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// 获取所有版本信息
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	versions, err := models.GetAllVersions(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 返回所有版本信息
	w.WriteHeader(http.StatusOK)
	for v := range versions {
		w.Write([]byte(fmt.Sprintf("%s_%d: %s\n", v.Name, v.Version, v.Hash)))
	}
}
