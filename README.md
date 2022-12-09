## 1.0.0

### 验证器改造

1. 自定义msg【 &check; 】
2. strconv类型转换错误屏蔽【 &check; 】
3. 验证器采用断路模式，一个字段有问题，立即返回。【 &check; 】
4. 时间time_format 支持 unixmilli、unixmicro, 以及时间错误封装【 &check; 】
5. json类型支持default: `json:a,default=1`【 &check; 】
6. 正则 【 &check; 】
7. required 验证器有语义偏差重新实现 【 &check; 】
8. 扩展验证函数： 【 &check; 】
   1. idcard: 大陆身份证验证
   2. mobile: 大陆手机号验证
   3. regex: 正则验证器
9. 废弃部分不常用验证器 【 &check; 】
10. 重新实现 required_if 验证器 【 &check; 】

## 1.1.0

1. 迁移 ginx core 核心功能；
2. utils 工具函数封装；
3. jsonp callback验证处理
4. 常用middleware内置；
5. ctx rsp扩展服务内置；
