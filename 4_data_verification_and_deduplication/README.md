# 访问接口和数据接口分离

其中有两个访问节点两个数据节点，使用rabbitmq做消息队列，用nginx做负载均衡

## 安装步骤：

采用docker compose安装

```cmd
cd depolyments
docker-compose up
```

访问方式：

```
curl --upload-file ./test.txt http://localhost:8888/objects/1.txt
curl http://localhost:8888/objects/1.txt
curl http://localhost:8888/locate/1.txt
```

