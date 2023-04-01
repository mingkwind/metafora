组件安装：

# k8s

MAC docker destop设置中 Enable Kubernetes，自动进行k8s安装

验证安装：

```ruby
kubectl cluster-info
kubectl get nodes
kubectl describe node
```

## 安装 Kubernetes Dashboard

```csharp
# 部署 Kubernetes Dashboard
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.6.1/aio/deploy/recommended.yaml
```

创建用户并获取token

## 创建admin-user

复制如下配置，保存文件为`dashboard-adminuser.yaml`

```dts
apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-user
  namespace: kubernetes-dashboard
```

然后执行

```coq
kubectl apply -f dashboard-adminuser.yaml
```

绑定cluster-admin授权

复制如下配置，保存文件为`dashboard-clusteradmin.yaml`

```nestedtext
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: admin-user
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: admin-user
  namespace: kubernetes-dashboard
```

然后执行

```coq
kubectl apply -f dashboard-clusteradmin.yaml
```

## 获取登录token

开启代理并且设置代理端口为8001

```routeros
kubectl proxy --port=8001
```

打开新的命令窗口，执行

检查用户信息是否存在

```1c
curl 'http://127.0.0.1:8001/api/v1/namespaces/kubernetes-dashboard/serviceaccounts/admin-user'
```

获取token不带参数

```1c
curl 'http://127.0.0.1:8001/api/v1/namespaces/kubernetes-dashboard/serviceaccounts/admin-user/token' -H "Content-Type:application/json" -X POST -d '{}'
```

获取token带参数

```apache
curl 'http://127.0.0.1:8001/api/v1/namespaces/kubernetes-dashboard/serviceaccounts/admin-user/token' -H "Content-Type:application/json" -X POST -d '{"kind":"TokenRequest","apiVersion":"authentication.k8s.io/v1","metadata":{"name":"admin-user","namespace":"kubernetes-dashboard"},"spec":{"audiences":["https://kubernetes.default.svc.cluster.local"],"expirationSeconds":7600}}'
```

![image-20221025181622908](rabbitMQ%E5%AE%89%E8%A3%85/image-20221025181622908.png)

## 登录k8s Dashboard

复制上一步返回的token信息，浏览器访问如下地址，填入token即可登录

```awk
http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/#/login
```

![image.png](rabbitMQ%E5%AE%89%E8%A3%85/bVc1fmo.png)



# RabbitMQ

```bash
docker pull rabbitmq:management
docker run --name rabbitmq -d -p 15672:15672 -p 5672:5672 rabbitmq:management
```

- --name指定了容器名称
- -d 指定容器以后台守护进程方式运行
- -p指定容器内部端口号与宿主机之间的映射，rabbitMq默认要使用15672为其web端界面访问时端口，5672为数据通信端口

```
docker logs rabbitmq
```

可以看出默认创建了guest用户，密码是guest

进入docker容器

```bash
sudo docker exec -it a8ecac45cf0a /bin/bash
rabbitmq-plugins enable rabbitmq_management
```

此时可以通过网页查看RabbitMQ管理端

使用guest，guest账号登陆

添加一个新账户

①进入到rabbitMq容器内部

```ruby
sudo docker exec -it 79171582b670 /bin/bash
```

②添加用户，用户名为root，密码为123456

```ruby
rabbitmqctl add_user root 123456 
```

③赋予root用户所有权限

```javascript
rabbitmqctl set_permissions -p / root ".*" ".*" ".*"
```

④赋予root用户administrator角色

```ruby
rabbitmqctl set_user_tags root administrator
```

⑤查看所有用户即可看到root用户已经添加成功

```ini
rabbitmqctl list_users
```

执行`exit`命令，从容器内部退出即可。这时我们使用root账户登录web界面也是可以的。到此，rabbitMq的安装就结束了，接下里就实际代码开发。

在web管理界面添加exchange

```
exchange name=apiServers type=fanout
exchange name=dataServers type=fanout
```

