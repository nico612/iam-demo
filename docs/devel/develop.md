


## 编译检查并发问题

`go build -race cmd/iam-authz-server/authzserver.go`, 然后在测试过程中观察程序日志，看有无并发问题出现。





## 数据库

mysql

redis

### mongodb 
mongodb 文档：https://www.mongodb.com/docs/manual/administration/install-community/

#### mac 安装
```shell
# 添加 mongodb 源
brew tap mongodb/brew

brew update

brew install mongodb-community@7.0
```
上面安装已经包含了：
- The `mongod` server
- The `mongos` sharded cluster query router
- The MongoDB Shell, `mongosh`

#### 运行
启动服务
```shell
brew services start mongodb-community@7.0
```
停止
```shell
brew services stop mongodb-community@7.0
```

验证 MongoDB 是否运行
```shell
# 查看服务列表
brew services list
```

查看mongoDB配置文件， 可根据官方文档说明查看，但是我的mac M2 芯片 安装后配置文件路径有点不一样，如果找不到，可以全局搜索, 一般在homebrew/etc目录下
```shell
# 超找配置文件
sudo find / -name mongod.conf 
# /usr/local/homebrew/etc/mongod.conf

cat /usr/local/homebrew/etc/mongod.conf

```
输出信息
```shell
systemLog:
  destination: file
  path: /usr/local/homebrew/var/log/mongodb/mongo.log
  logAppend: true
storage:
  dbPath: /usr/local/homebrew/var/mongodb
net:
  bindIp: 127.0.0.1, ::1
  ipv6: true
```
**MongoDB 默认使用 27017 端口。**

go 连接mongo 三方库：github.com/vinllen/mgo