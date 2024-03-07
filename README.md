
# deeplx-pro

使用 DeepL Pro的翻译服务

## 如何使用这个镜像

### 拉取镜像

你可以使用以下命令从Docker Hub拉取这个镜像：

```bash
docker pull xiaoxiaofeihh/deeplx-pro:latest
```

### 运行镜像

你可以使用以下命令运行这个镜像：

```bash
docker run -d -p 9000:9000 xiaoxiaofeihh/deeplx-pro:latest
```

在这个命令中，`-d`选项让Docker在后台运行这个镜像，`-p`选项将容器的8080端口映射到主机的8080端口。

## 构建你自己的镜像

如果你想要构建你自己的`deeplx-pro`镜像，你可以在仓库的根目录运行以下命令：

```bash
docker build -t deeplx-pro .
```

## 问题和反馈

如果你在使用这个镜像时遇到了问题，或者你有任何反馈或建议，欢迎通过GitHub Issues向我们反馈。
