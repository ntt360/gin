## 版本关系
| ginx 版本 | gin 官方版本 |
|:-------:|:--------:|
| 1.0.0+  |  1.8.1   |
| 1.9.0+  | 1.9.1 |

`ginx` 尽量与 `gin` 兼容，但我们扩展了`gin`功能，以及进行了二次深度重构。

## 1.0.0

1. [验证器重构](https://github.com/ntt360/gin/blob/master/docs/validator.md)
2. 内置gzip模块；

## 1.1.0 ~ 1.8.4

1. 迁移 ginx core 核心功能；
2. 内置 utils 工具函数封装；
3. jsonp callback 函数名限制；
4. 常用middleware内置；
5. ctx rsp扩展服务内置；

## 1.9.0

适配 `gin` 官方版本 `1.9.1`，无其它变化。
