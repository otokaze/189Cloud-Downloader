# 189Cloud-Downloader
一个189云盘的下载器。（支持分享链接下载、支持Windows、Linux、macOS）Based Go.

## 使用说明
```
NAME:
   189Cloud-Downloader - 一个189云盘的下载器。（支持分享链接）

USAGE:
   189Cloud-Downloader [global options] command [command options] [arguments...]

COMMANDS:
   login     登陆189账号
   logout    退出登陆
   exit      退出程序
   share     读取分享链接
   cd        切换至目录
   pwd       查看当前路径
   get       下载这个目录(递归)|文件
   ls        遍历目录（精简）
   ll        遍历目录（详细）
   userinfo  查看当前登录的用户信息
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

### 登陆
```
NAME:
   189Cloud-Downloader login - 登陆189账号

USAGE:
   189Cloud-Downloader login [command options] <username> <password>

OPTIONS:
   --cookie value  cookie, 取 COOKIE_LOGIN_USER 字段就行
   --help, -h      show help (default: false)
```

#### Example
```
./189Cloud-Downloader login ${USERNME} ${PASSWORD}
```
or
```
./189Cloud-Downloader
> login  ${USERNME} ${PASSWORD}
```

### 读取分享链接
*USAGE 中的：“?”，指可以忽略的参数*
```
NAME:
   189Cloud-Downloader share - 读取分享链接

USAGE:
   189Cloud-Downloader share [command options] <link> <key>?

OPTIONS:
   --help, -h  show help (default: false)
```
#### Example
```
./189Cloud-Downloader share https://cloud.189.cn/t/xxxxxx
```
or
```
./189Cloud-Downloader
> share https://cloud.189.cn/t/xxxxxx
```

### 切换目录
*cd 命令后面必须是 fileId（可以通过 ll 命令得到 fileId）*
```
NAME:
   189Cloud-Downloader cd - 切换目录

USAGE:
   189Cloud-Downloader cd [command options] <fileId>

OPTIONS:
   --help, -h  show help (default: false)
```
#### Example
```
./189Cloud-Downloader share https://cloud.189.cn/t/xxxxxx
个人收集电影...> ll
[D]2150137850933107     0.00B   2020-04-30 21:50:58     100部纯英文系列电影
[D]3152737831639376     0.00B   2020-04-30 21:53:47     2016.信号 signal.16集全
[D]7142737850912074     0.00B   2020-04-30 21:39:20     3年A班
[D]7142737850912085     0.00B   2020-04-30 21:39:20     EVA 新世纪福音战士
> cd 2150137850933107
100部纯英文...>
```
切换到个人空间
```
100部纯英文...> cd ~
全部文件> ll
[D]0    0.00B   2021-01-07 22:08:48     同步盘
[D]-12  0.00B   2021-01-07 22:08:48     我的图片
[D]-14  0.00B   2021-01-07 22:08:48     我的音乐
[D]-13  0.00B   2021-01-07 22:08:48     我的视频
[D]-15  0.00B   2021-01-07 22:08:48     我的文档
[D]-16  0.00B   2021-01-07 22:08:48     我的应用
```
切换回刚才的分享目录
```
全部文件> cd share
个人收集电影...> ll
[D]2150137850933107     0.00B   2020-04-30 21:50:58     100部纯英文系列电影
[D]3152737831639376     0.00B   2020-04-30 21:53:47     2016.信号 signal.16集全
[D]7142737850912074     0.00B   2020-04-30 21:39:20     3年A班
[D]7142737850912085     0.00B   2020-04-30 21:39:20     EVA 新世纪福音战士
```

### 查看当前路径
#### Example
```
./189Cloud-Downloader share https://cloud.189.cn/t/xxxxxx
个人收集电影...> pwd
/个人收集电影大合集
```

### 遍历目录（精简）
*USAGE 中的：“?”，指可以忽略的参数*  
*ls 可以遍历指定 fileId 的目录（可以通过 ll 命令得到 fileId）*  
```
NAME:
   189Cloud-Downloader ls - 遍历目录（精简）

USAGE:
   189Cloud-Downloader ls [command options] <fileId>?

OPTIONS:
   --pn value     页码 (default: 1)
   --ps value     页长 (default: 60)
   --order value  排序，ASC：顺排 DESC：倒排 (default: "ASC")
   --help, -h     show help (default: false)
```
#### Example
```
./189Cloud-Downloader share https://cloud.189.cn/t/xxxxxx
个人收集电影...> ls
100部纯英文系列电影     2016.信号 signal.16集全 3年A班  EVA 新世纪福音战士
```
遍历个人空间
```
个人收集电影...> ls ~
同步盘  我的图片        我的音乐        我的视频        我的文档        我的应用
```

### 遍历目录（详细）
*USAGE 中的：“?”，指可以忽略的参数*  
*ll 可以遍历指定 fileId 的目录*  
```
NAME:
   189Cloud-Downloader ll - 遍历目录（详细）

USAGE:
   189Cloud-Downloader ll [command options] <fileId>?

OPTIONS:
   --pn value     页码 (default: 1)
   --ps value     页长 (default: 60)
   --order value  排序，ASC：顺排 DESC：倒排 (default: "ASC")
   --help, -h     show help (default: false)
```
#### Example
```
./189Cloud-Downloader share https://cloud.189.cn/t/xxxxxx
个人收集电影...> ll
[D]2150137850933107     0.00B   2020-04-30 21:50:58     100部纯英文系列电影
[D]3152737831639376     0.00B   2020-04-30 21:53:47     2016.信号 signal.16集全
[D]7142737850912074     0.00B   2020-04-30 21:39:20     3年A班
[D]7142737850912085     0.00B   2020-04-30 21:39:20     EVA 新世纪福音战士
```
遍历个人空间
```
个人收集电影...> ll ~
[D]0    0.00B   2021-01-07 22:08:02     同步盘
[D]-12  0.00B   2021-01-07 22:08:02     我的图片
[D]-14  0.00B   2021-01-07 22:08:02     我的音乐
[D]-13  0.00B   2021-01-07 22:08:02     我的视频
[D]-15  0.00B   2021-01-07 22:08:02     我的文档
[D]-16  0.00B   2021-01-07 22:08:02     我的应用
```

### 下载这个目录(递归)|文件
- *USAGE 中的：“?”，指可以忽略的参数*  
- *当不指定\<topath\>的时候默认下载到系统临时目录*  
```
NAME:
   189Cloud-Downloader get - 下载这个目录(递归)|文件

USAGE:
   189Cloud-Downloader get [command options] <fileId> or ./ <topath>?

OPTIONS:
   --concurrency value, -c value  并发数 (default: 10)
   --tmp value                    工作路径 (default: /tmp)
   --help, -h                     show help (default: false)
```
#### Example
下载指定 fileId 的文件|目录，并且指定保存目录为 /Users/otokaze/Downloads
```
./189Cloud-Downloader share https://cloud.189.cn/t/xxxxxx
个人收集电影...> ll
[D]2150137850933107     0.00B   2020-04-30 21:50:58     100部纯英文系列电影
[D]3152737831639376     0.00B   2020-04-30 21:53:47     2016.信号 signal.16集全
[D]7142737850912074     0.00B   2020-04-30 21:39:20     3年A班
[D]7142737850912085     0.00B   2020-04-30 21:39:20     EVA 新世纪福音战士
个人收集电影...> get 7142737850912085 /Users/otokaze/Downloads
```
下载当前目录
```
个人收集电影...> cd 7142737850912085
EVA 新世纪福...> get ./ /Users/otokaze/Downloads
```
取消下载
```
^C (CTRL+C)
```

### 查看当前登录的用户信息
#### Example
```
./189Cloud-Downloader login ${USERNME} ${PASSWORD}
全部文件> userinfo
UserId: 756719517
UserAccount: otokaze
已用容量: 22.88GB
可用容量: 7.12GB
总容量: 30.00GB
```

### 退出登陆
```
> logout
```

### 退出程序
```
> exit
```

### Donate
![weixin](https://raw.githubusercontent.com/otokaze/189Cloud-Downloader/master/donate_weixin.png)
![alipay](https://raw.githubusercontent.com/otokaze/189Cloud-Downloader/master/donate_alipay.png)

### 感谢
本项目部分参考了以下项目实现免验证登陆
- https://github.com/tickstep/cloudpan189-api
