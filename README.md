# 阿里云盘 for Server(cli)

> cli专用版本阿里云盘

## 功能介绍

​	本工具主要实现通过cli调用阿里云盘来实现相关功能。

| 功能           | action      |
| -------------- | ----------- |
| 从本地上传文件 | localUpload |
| 从网盘下载文件 | download    |
| 启动服务端     | server      |

## 功能必要参数

### 从本地上传文件

| 参数名   | 参数说明               |
| -------- | ---------------------- |
| filePath | 需要上传文件的绝对路径 |

```bash
# 例如
./ALIYUN -action localUplaod -filePath Your file path
```



### 从网盘下载文件

​	无需参数

```bash
# 例如
./ALIYUN -action download
```

​	cli页面会从更目录开始让您选择需要下载的文件,如果是文件夹则会转到文件夹目录继续上步操作，知道选择的是文件，输入下载到本地的目录，开始下载

### 服务端

#### 介绍

​	本客户端带有服务端的功能，可以在本地启动一个web服务，实现更快的上传

### 命令行调用参数

| 参数名 | 参数说明    | 是否必要           |
| :----: | ----------- | ------------------ |
|  port  | web服务端口 | 否,默认端口是13142 |

要启动服务，则需要执行以下命令

```bash
./ALIYUN -action server [-port your port]
```

### 接口

| 接口地址      | 接口说明                       |
| ------------- | ------------------------------ |
| /getUpload    | 创建上传文件请求、得到上传链接 |
| /complete     | 上传完成时调用此接口           |
| /directUpload | 直接上传文件到网盘             |

#### /getUpload

##### Method : POST

##### 请求参数

| 参数     | 说明                |
| -------- | ------------------- |
| fileName | 文件名              |
| fileSize | 文件大小(单位:Byte) |

##### 返回

```json
[
  "Upload Url",
  "Upload Id",
  "File Id"
]
```

#### /complete

##### Method : POST

##### 请求参数

| 参数     | 说明                          |
| -------- | ----------------------------- |
| fileid   | /getUpload中获取到的file id   |
| uploadid | /getUpload中获取到的upload id |

##### 返回

```json
{
  "result" : 阿里云盘接口返回
}
```

#### /directUpload

##### Mehod : POST

#####  请求参数

| 参数 | 参数说明 |
| ---- | -------- |
| file | 文件     |

使用form-data传

