package models

type Metadata struct { //元数据结构体 包含名称、版本、size、哈希值 ,在es中的索引名为name_version
	Name    string `json:"name"`
	Version int    `json:"version"`
	Size    int64  `json:"size"`
	Hash    string `json:"hash"`
}

const mapping = `{
	"mappings": {
		"properties": {
			"name": {
				"type": "keyword",
				"index": "true"
			},
			"version": {
				"type": "integer"
			},
			"size": {
				"type": "integer"
			},
			"hash": {
				"type": "keyword",
				"index": "true"
			}
		}
	}
}`
