
# deeplx-pro

使用 DeepL Pro的翻译服务

# 直接使用，畅行无阻。

## 如何使用这个镜像

### 拉取镜像

你可以使用以下命令从Docker Hub拉取这个镜像：

```bash
docker pull xiaoxiaofeihh/deeplx-pro:latest
```

### 运行镜像

你可以使用以下命令运行这个镜像：

```bash
docker run -d -p 9000:9000 -e COOKIE_VALUE=dl_session=你的dl_session xiaoxiaofeihh/deeplx-pro:latest
```

然后使用http://localhost:9000/translate
就可以愉快使用了

## 构建你自己的镜像

如果你想要构建你自己的`deeplx-pro`镜像，你可以在仓库的根目录运行以下命令：

```bash
docker build -t deeplx-pro .
```

## 问题和反馈

如果你在使用这个镜像时遇到了问题，或者你有任何反馈或建议，欢迎通过GitHub Issues向我们反馈。
