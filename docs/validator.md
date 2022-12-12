# validator

1. [如何使用](https://github.com/ntt360/gin/blob/master/docs/validator.md#%E5%A6%82%E4%BD%95%E4%BD%BF%E7%94%A8)
2. [示例](https://github.com/ntt360/gin/blob/master/docs/validator.md#%E7%A4%BA%E4%BE%8B)
3. [支持的验证函数]()
4. [自定义错误消息]()
5. [valid.Error说明]()

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

|                                        -                                         | - |                                                   -                                                    |                                                      -                                                       | - |
|:--------------------------------------------------------------------------------:|:------:|:------------------------------------------------------------------------------------------------------:|:------------------------------------------------------------------------------------------------------------:|:-------------------------------------------------------------------------:|
| [required](https://github.com/ntt360/gin/blob/master/docs/validator.md#required) |[required_if](https://github.com/ntt360/gin/blob/master/docs/validator.md#required_if)|              [len](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Length)               |                 [min](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Minimum)                 | [max](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Maximum) |
|    [eq](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Equals)    |[ne](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Not_Equal)|             [lt](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Less_Than)              |                [lte](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Less_Than)                |[gt](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Less_Than)| 
|  [gte](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Less_Than)  |[eqfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Equals_Another_Field)| [eqcsfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Equals_Another_Field)  | [nefield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Does_Not_Equal_Another_Field)  |[gtfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Greater_Than_Another_Field) |
|[gtefield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Greater_Than_or_Equal_To_Another_Field)|[ltfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Less_Than_Another_Field)|[ltefield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Less_Than_or_Equal_To_Another_Field)|[necsfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Does_Not_Equal_Another_Field)|[gtcsfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Greater_Than_Another_Relative_Field)|
|[gtecsfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Greater_Than_or_Equal_To_Another_Relative_Field)|[ltcsfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Less_Than_Another_Relative_Field)|[ltecsfield](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Less_Than_or_Equal_To_Another_Relative_Field)|[fieldcontains](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Field_Contains_Another_Field)|fieldexcludes|
|[alpha](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Alpha_Only)|[alphanum](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Alphanumeric) |[alphaunicode](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Alpha_Unicode)|[alphanumunicode](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Alphanumeric_Unicode)|[boolean](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Alphanumeric_Unicode)|
|[numeric](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Numeric)|[number](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Number)|email|[url](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-URI_String)|[uri](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-URI_String)|
|[urn_rfc2141](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Urn_RFC_2141_String) |[file](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-File_path)|[base64](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Base64_String)|[base64url](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Base64URL_String)|[contains](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Contains)|
|[containsany](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Contains_Any)|[containsrune](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Contains_Rune)|[excludes](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Excludes)|[excludesall](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Excludes_All)|[excludesrune](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Excludes_Rune)|
|[startswith](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Starts_With)|[endswith](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Ends_With)|[startsnotwith](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Does_Not_Start_With)|[endsnotwith](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Does_Not_End_With)|md5|
|sha256|[ipv4](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Internet_Protocol_Address_IPv4)|[ipv6](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Internet_Protocol_Address_IPv6)|[ip](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Internet_Protocol_Address_IP)|[tcp4_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Transmission_Control_Protocol_Address_TCPv4)|
|[tcp6_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Transmission_Control_Protocol_Address_TCPv6)|[tcp_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Transmission_Control_Protocol_Address_TCP)|[udp4_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-User_Datagram_Protocol_Address_UDPv4)|[udp6_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-User_Datagram_Protocol_Address_UDPv6)|[udp_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-User_Datagram_Protocol_Address_UDP)|
|[ip4_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Internet_Protocol_Address_IPv4)|[ip6_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Internet_Protocol_Address_IPv6)|[ip_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Internet_Protocol_Address_IP)|[unix_addr](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Unix_domain_socket_end_point_Address)|[mac](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Media_Access_Control_Address_MAC)|
|[hostname](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Hostname_RFC_952)|[hostname_rfc1123](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Hostname_RFC_1123)|[unique](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Unique)|[oneof](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-One_Of)|[html](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-HTML_Tags)|
|[html_encoded](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-HTML_Encoded)|[url_encoded](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-URL_Encoded)|[json](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-JSON_String)|[jwt](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-JWT_String)|[hostname_port](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-HostPort)|
|[lowercase](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Lowercase_String)|[uppercase](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Uppercase_String)|[datetime](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-Datetime)|[timezone](https://pkg.go.dev/github.com/go-playground/validator/v10#hdr-TimeZone)||

### 扩展的验证函数

|  -     |    -     |    -     |
|:------------:|:-----------------:|:-------------------:|
| [regex](https://github.com/ntt360/gin/blob/master/docs/validator.md#regex) | [mobile](https://github.com/ntt360/gin/blob/master/docs/validator.md#mobile) | [idcard](https://github.com/ntt360/gin/blob/master/docs/validator.md#idcard) |


### required

校验验证的参数是否存在，这和官方的提供的`required`验证器有差别，我们调整了 `required` 验证器语义，仅会验证参数是否存在。

```go
type Params struct {
	Page int `form:"page" binding:"required"`
}
```

验证结果

```shell
curl --url 'http://xxxx/?' # 验证不通过

curl --url 'http://xxxx/?page=0' # 验证通过
curl --url 'http://xxxx/?page=1' # 验证通过
```

### required_if

有条件的必填验证器，注意这个验证实现也和官方不一致（官方验证器仅能验证值是否相等），我们可以实现一些常规的大于、小于、等于判断：

格式：

```
required_if={Field} {Cond} {Value}
```
`{Field}` 为选择对比的字段，`{Cond}` 为对比的条件，`{Value}` 为对比的值。目前 `Cond` 支持如下系列：

```shell
gt  # 大于，required_if=Num gt 1
gte # 大于等于，required_if=Num gte 1
eq  # 相等，required_if=Num eq 1
lt  # 小于，required_if=Num lt 1
lte # 小于等于，required_if=Num lte 1
ne  # 不等，required_if=Num ne 1
```

示例：

```go
type Params struct {
	// Num = 2 时，page 参数要求为：必填
	Page int `form:"page" binding:"required_if=Num eq 2"`
	Num int `form:"num" binding:"required"`
}
```

验证结果

```shell
curl 'http://xxxx/?num=1' # 验证通过
curl 'http://xxxx/?num=2&page=1' # 验证通过

curl 'http://xxxx:3000/?num=2' # 验证不通过
```

### regex

正则表达式验证，需要配合另一个 `tag` 字段：`pattern` 配合使用（原先的验证器设计中，引号、逗号、等号都有特殊含义，我们沿用这种规则，所以引入了新的tag机制来存储正则表达式）

```go
type Params struct {
	Name string `form:"name" pattern:"^\\w{8,20}$" binding:"required,regex" msg:"name参数不正确"`
}
```

如上的 `name` 参数需要满足正则规则：`^\\w{8,20}$`


### mobile

支持大陆三家运行商手机号验证，支持附带(+86)前缀。

### idcard

大陆的身份证号验证

## 自定义错误消息

很多时候我们都需要自定义错误消息，所以我们增加了一个：`msg` 标签来满足该需求，由于 `POST` `JSON` `body` 支持无线嵌套数据，所以 `msg` 消息也较为复杂，我们尽量简单描述如何使用它。

### 全局错误消息

如果你仅仅希望每个字段多个验证规则，均使用同一个自定义错误时，那么这种方式使用是比较简单的：

```go
type Params struct {
	Name string `form:"name" pattern:"^\\w{8,20}$" binding:"required,min=1,max=20,regex" msg:"name参数不正确"`
}
```

如上的 `name` 字段有4个验证规则函数，都返回同一个错误消息：`name参数不正确`

### 每个验证规则返回错误

```go
type Params struct {
	Name string `form:"name" pattern:"^\\w{8,20}$" binding:"required,min=1,max=3,regex" msg:"default='name参数不合法',required='name参数必填',min='name长度必须大于1',max='name长度不超过3个字符'"`
}
```

上面便是为多个验证规则提供不同的错误消息，当都不匹配时（regex验证不过），会使用一个 `default='xxx'` 默认错误消息。

验证结果：

```shell
curl 'http://xxxx/?'          # 返回错误：name参数必填
curl 'http://xxxx/?name='     # 返回错误：name长度必须大于1
curl 'http://xxxx/?name=1234' # 返回错误：name长度不超过3个字符
curl 'http://xxxx/>name=#$%^' # 返回错误：name参数不合法
```

`default` 默认错误消息，并非必须，如果不存在，则会使用内置的错误消息：`the param name not valid`

需要注意的是：针对每个验证函数的自定义错误的，一定需要使用单引号包含错误消息内容：required='错误消息内容'

```shell
# 正确定义方式
requierd='错误消息'

# 错误定义方式
required=错误消息
```

之所以需要单引号，主要是因为消息的拆分使用了正则表达式拆分多个错误内容，单引号是判断边界，当使用全局错误消息时没有此要求。

### 验证规则重复处理: > 符号

有时候很存在多个验证器名称重复，此时如果使用规则错误会有些问题，例如：

```go
type Params struct {
	Name [][]string `json:"name" binding:"min=1,dive,min=3,dive,min=10" msg:""`
}
```

如上的验证规则中，`min=1` 用于验证 `name` 数组本身长度必须大于等于1，`min=3` 用于验证数组中的第二层数组长度必须要大于3个字符。第三层要求具体元素长度比如不小于10。但此时我们发觉，验证器规则出现了3个`min`验证函数，那此时我们如果定义错误消息呢？

所以我们引入了 `>` 符号用于描述当前验证器路径层次关系。，所以此时的错误消息定义如下：

```shell
msg:"min='数组至少1个元素',>min='子数组不少于3元素',>>min='元素长度必须不少于10个字符'"
```

## valid.Error 说明

验证器我们定义了一个 `valid.Error` 错误类型，默认的错误消息有两种消息类型：1.具体参数类错误 2. 全局的错误；错误默认消息内容如下。

```shell
# 具体参数字段错误
the param xx not valid

# 全局错误，如请求参数格式不正确，但不涉及某个具体字段，如请求json不合法。
request data is not valid json

# 其它全局错误
...
```

我们隐藏了具体参数是因为哪个验证器规则导致的错误，但会在返回的`valid.Error`有对应的字段予以标识，你可以通过反解 `err`，或者使用`%+v`输出符号打印出更明细的错误内容用以记录错误日志：

```go
e := ctx.Valid(&rel)
if e != nil {
  var validErr *valid.Error
  if errors.As(e, &validErr) {
      fmt.Printf("%+v", e) // 输出错误详细内容
  }
    
  // 前端仅返回错误概述或自定义错误，避免系统实现暴露
  rsp.JSONErr(ctx, &rsp.Values{Msg: e.Error()})
  return
}
```


