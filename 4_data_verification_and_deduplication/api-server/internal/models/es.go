package models

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/olivere/elastic/v7"
)

var (
	elasticClient *elastic.Client
	ctx           = context.Background()
)

func init() {
	log.Println("init elasticsearch client")
	//从配置文件中读取es的地址
	urlOpt := elastic.SetURL(os.Getenv("ES_URL"))
	sniffOpt := elastic.SetSniff(false)
	basicAuthOpt := elastic.SetBasicAuth(os.Getenv("ES_USER"), os.Getenv("ES_PASS"))
	var err error
	elasticClient, err = elastic.NewClient(urlOpt, sniffOpt, basicAuthOpt)
	if err != nil {
		// Handle error
		panic(err)
	}
	log.Println("connect to es success")
	//ping
	info, code, err := elasticClient.Ping(os.Getenv("ES_URL")).Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	log.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
	esVersion, err := elasticClient.ElasticsearchVersion(os.Getenv("ES_URL"))
	if err != nil {
		// Handle error
		panic(err)
	}
	log.Printf("Elasticsearch version %s\n", esVersion)
	// 判断索引metadata是否存在
	exists, err := elasticClient.IndexExists("metadata").Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	if !exists {
		// 创建索引
		createIndex, err := elasticClient.CreateIndex("metadata").BodyString(mapping).Do(ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
			log.Println("create index metadata failed")
		}
	}
}

func AddVersion(name string, size int64, hash string) error {
	err := PutMetadata(&Metadata{
		Name: name,
		Size: size,
		Hash: hash,
	})
	return err
}

func PutMetadata(metadata *Metadata) error {
	// 查询 当前name是否存在
	// 如果存在,则获取最新的版本,并且version+1，采用乐观锁
	// 如果不存在,则version=1
	lastVersion, err := GetLatestVersion(metadata.Name)
	if err != nil {
		return err
	}
	if lastVersion.Name == "" {
		metadata.Version = 1
	} else {
		metadata.Version = lastVersion.Version + 1
	}
	// 如果hash相同,则不需要更新
	if lastVersion.Hash == metadata.Hash {
		return nil
	}
	log.Printf("put metadata: %#v\n", *metadata)
	// 保存到es
	_, err = elasticClient.Index().
		Index("metadata").
		Id(fmt.Sprintf("%s_%d", metadata.Name, metadata.Version)).
		BodyJson(metadata).
		Do(ctx)
	if err != nil {
		return err
	}
	return nil
}

func DelMetadata(name string, version int) error {
	//elastic7 删除
	id := fmt.Sprintf("%s_%d", name, version)
	_, err := elasticClient.Delete().
		Index("metadata").
		Id(id).
		Do(ctx)
	if err != nil {
		log.Printf("delete metadata %s_%d failed, err:%v\n", name, version, err)
		return err
	}
	log.Printf("delete metadata %s_%d success\n", name, version)
	return nil
}

func DelAllMetadata(name string) error {
	// 查询name的所有版本
	// 删除
	metadataChan, err := GetAllVersions(name)
	if err != nil {
		return err
	}
	for metadata := range metadataChan {
		err = DelMetadata(metadata.Name, metadata.Version)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetMetadata(name string, versionId int) (Metadata, error) {
	//如果versionId为0,则获取最新的版本
	if versionId == 0 {
		return GetLatestVersion(name)
	} else {
		// 查询name和version,使用id=name_version
		id := fmt.Sprintf("%s_%d", name, versionId)
		get, err := elasticClient.Get().
			Index("metadata").
			Id(id).
			Do(ctx)
		if err != nil {
			return Metadata{}, err
		}
		if get.Found {
			var metadata Metadata
			err = json.Unmarshal(get.Source, &metadata)
			if err != nil {
				return Metadata{}, err
			}
			return metadata, nil
		}
		return Metadata{}, nil
	}
}

// 获取最新的版本
func GetLatestVersion(name string) (Metadata, error) {
	// 查询name和version,使用id=name_version
	// 获取版本最大的
	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermQuery("name", name))
	searchResult, err := elasticClient.Search().
		Index("metadata").
		Query(query).
		Sort("version", false).
		From(0).Size(1).
		Pretty(true).
		Do(ctx)
	if err != nil {
		return Metadata{}, err
	}
	if searchResult.Hits.TotalHits.Value > 0 {
		var metadata Metadata
		err = json.Unmarshal(searchResult.Hits.Hits[0].Source, &metadata)
		if err != nil {
			return Metadata{}, err
		}
		return metadata, nil
	}
	return Metadata{}, nil
}

func GetAllVersions(name string) (chan Metadata, error) {
	// 使用SearchAfter
	// 循环查询
	query := elastic.NewBoolQuery()
	query = query.Must(elastic.NewTermQuery("name", name))
	searchResult, err := elasticClient.Search().
		Index("metadata").
		Query(query).
		Sort("version", false).
		Size(10).
		Pretty(true).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	var metadata Metadata
	var lastSort []interface{}
	ch := make(chan Metadata, 10)
	go func() {
		defer close(ch)
		for {
			if len(searchResult.Hits.Hits) > 0 {
				for _, item := range searchResult.Each(reflect.TypeOf(metadata)) {
					if t, ok := item.(Metadata); ok {
						ch <- t
					}
				}
				lastSort = searchResult.Hits.Hits[len(searchResult.Hits.Hits)-1].Sort
				searchResult, err = elasticClient.Search().
					Index("metadata").
					Query(query).
					Sort("version", false).
					SearchAfter(lastSort...).
					Size(10).
					Pretty(true).
					Do(ctx)
				if err != nil {
					log.Println(err)
					break
				}
			} else {
				break
			}
		}
	}()
	return ch, nil
}
