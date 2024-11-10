# go 八股项目——小微书（仿小红书）

## 分层结构
![img.png](imgs/img.png)
+ service: 代表的是领域服务（domain service），
  代表一个业务的完整的处理过程。
+ repository：按照 DDD 的说法，是代表领域对象的
  存储，这里你直观理解为存储数据的抽象。
+ dao: 代表的是数据库操作。
+ 还需要一个 domain，代表领域对象。

## 登录校验
