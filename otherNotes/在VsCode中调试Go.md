# 在VsCode中调试Go程序

主要设置一下`launch.json`配置文件：

```json
{
    // 使用 IntelliSense 了解相关属性。 
    // 悬停以查看现有属性的描述。
    // 欲了解更多信息，请访问: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${fileDirname}",	// 这是当前选择的文件
            "showLog": false,
            "console": "integratedTerminal"	// 支持终端输入
        }
    ]
}
```



