# validator

我们对 `gin` 的官网验证器进行了重写，解决了如下一些问题：

1. 支持自定义错误消息；
2. 修复某个验证字段数据类型解析失败产生的系统不友好的提示错误；
3. `json` 和 `form` 同时支持 `time_format` 时间格式值；
4. `json` 支持 `default` 默认值，保持和 `form` 一致；
5. `required`、`required_if` 验证器重写实现；
6. 扩展了几个常用验证函数：
   1. 正则验证器：`regex`
   2. 大陆手机号验证器：`mobile`
   3. 大陆身份证验证器：`idcard`
7. 废弃了部分不常用的验证器

## 如何使用

我们扩展了 `gin.Context` ，提供了以 `ctx.ValidXXX()` 系列的验证函数，对标原`gin` 内置的 `ctx.ShouldBindXXX()` 系列验证函数。同样，**你仍然可以继续使用原先的验证函数，也可以是新的方式**。

| 扩展的函数名                   | 对标函数                       | 说明           |
|:-------------------------|:---------------------------|:-------------|
| ctx.Valid()              | ctx.ShouldBind()           | 默认form验证函数   |
| ctx.ValidJSON()          | ctx.ShouldBindJSON()       | json body 验证 |
| ctx.ValidQuery()         | ctx.ShouldBindQuery()      | query 验证     |
| ctx.ValidHeader()        | ctx.ShouldBindHeader()     | header参数绑定验证 |
| ctx.ValidWith()          | ctx.ShouldBindWith()       | -            |
| ctx.ValidBodyWith()      | ctx.ShouldBindBodyWith()   | -            |


**注意：**

我们并没有完全实现所有的原 `ctx.ShouldBindXXX` 其它相关验证器（不常用），另外验证类型目前不支持 `xml`、`protobuf`、`yaml` 等，仅支持如上的方式，但是已经足以满足我们的业务需要，如有需要，后续再考虑扩展。

## 示例

```go
package main

import (
   "errors"
   "fmt"
   "github.com/ntt360/gin"
   "github.com/ntt360/gin/valid"
)

type Params struct {
	Page int `form:"page,default=1" binding:"numeric,min=1" msg:"页码不正确"`
	Size int `form:"size,default=30" binding:"numeric,max=30" msg:"size 不正确"`
}

func Form(ctx *gin.Context) {
	var rel Params
	e := ctx.Valid(&rel)
	if e != nil {
		var validErr *valid.Error
		if errors.As(e, &validErr) {
			fmt.Printf("%+v", e)
		}
	}
	
	// TODO 业务处理
}	
```

使用方式和使用官方的验证器绑定一致，只需要替换原有的 `ctx.ShouldBindXXX` 系列函数为 `ctx.ValidXXX` 系列函数即可。同样你会看到，我们使用 `msg` 自定义tag来自定义错误消息（后续会详细说明）。

## 支持的验证函数

我们复用了原有的 `validator` 内置验证函数，所以实现了绝大部分验证函数，但有小部分不常用的验证函数尚未实现，已实现函数列表如下：

[required](https://github.com/ntt360/gin/blob/master/docs/validator.md#required)   
[required_if](https://github.com/ntt360/gin/blob/master/docs/validator.md#required_if)


### required

校验验证的参数是否存在，这和官方的提供的`required`验证器有差别，我们调整了 `required` 验证器语义，仅会验证参数是否存在。

```go
type Params struct {
	Page int `form:"page" binding:"required"`
}
```

验证结果

```shell
curl --url 'http://xxxx:3000/?' # 验证不通过

curl --url 'http://xxxx:3000/?page=0' # 验证通过
curl --url 'http://xxxx:3000/?page=1' # 验证通过
```

### required_if

有条件的必填验证器，注意这个验证实现也和官方不一致（官方验证器仅能验证值是否相等），我们可以实现一些常规的大于、小于、等于判断：

```go
type Params struct {
	// Num = 2 时，page 参数必填
	Page int `form:"page" binding:"required_if=Num eq 2"`
	Num int `form:"num" binding:"required"`
}
```

验证结果

```shell
curl 'http://xxxx:3000/?num=1' # 验证通过
curl 'http://xxxx:3000/?num=2&page=1' # 验证通过

curl 'http://xxxx:3000/?num=2' # 验证不通过
```

目前支持的规则有：

```shell
gt  # 大于，required_if=Num gt 1
gte # 大于等于，required_if=Num gte 1
eq  # 相等，required_if=Num eq 1
lt  # 小于，required_if=Num lt 1
lte # 小于等于，required_if=Num lte 1
ne  # 不等，required_if=Num ne 1
```
