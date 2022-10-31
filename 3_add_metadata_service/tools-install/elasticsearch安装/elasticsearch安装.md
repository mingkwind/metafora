```
docker pull elasticsearch:7.17.2
# 单节点部署

```



编辑config/elasticsearch.yml

```
cluster.name: "docker-cluster"
network.host: 0.0.0.0
xpack.security.enabled: true
xpack.license.self_generated.type: basic
xpack.security.transport.ssl.enabled: true
```

运行容器

```
docker run --name es -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "xpack.security.enabled=true" -e "xpack.license.self_generated.type=basic" -e "xpack.security.transport.ssl.enabled=true"  -dit elasticsearch:7.17.2 
```

设置密码

```
#进入es 名录目录 
cd /usr/share/elasticsearch/bin/
#执行命令，交互式设置密码（注意保存好全部密码）
./elasticsearch-setup-passwords interactive
# 密码设置为123456
```

网页登陆

```
http://localhost:9200/
elastic 123456
```

