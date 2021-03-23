# simpleOss
最简化对接阿里，华为，Azure等的oss，配置分离，开箱即用
###1.go build 编译,-h命令行参数查看帮助
```golang
./main -h
Usage of 
  -c string
        General configuration file (default "config.ini")
  -d    When upload finished, delete local file
  -w string
        Upload work directory (default "./")
```
###2.跨平台自行交叉编译
```golang
GOOS=linux      GOARCH=arm      go build -o main-arm-linux        main.go
GOOS=linux      GOARCH=amd64    go build -o main-amd64-linux      main.go
GOOS=windows    GOARCH=amd64    go build -o main-amd64-windows    main.go
```