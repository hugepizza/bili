# 已废弃 不保证正常使用

# bili
bilibili 哔哩哔哩 视频下载工具

## 说明
b站视频的格式有m4s切片、flv、和mp4三种<br>
m4s切片下载后音视频是分离的两个文件 参数-M自动合并 需要`ffmpeg`支持(合并功能windows没测试 不保证能用)<br>
b站视频的清晰度有大会员专属的1080p+等<br>
默认并只支持非会员身份的最高清晰的的下载<br>
暂时不支持多p下载

## 命令行参数 
```
Usage:
  b-get [flags]

Flags:
      --av int        视频av号
  -h, --help          help for b-get
  -M, --merge         是否自动合并
      --path string   保存路径 (default "/home/wanlei/Videos")
```

## 禁止用于商业用途
