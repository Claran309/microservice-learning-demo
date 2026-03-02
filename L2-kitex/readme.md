```
1. 客户端（curl/浏览器）
   ↓ HTTP POST /api/v1/register
   
2. API网关（Hertz）
   ↓ 接收HTTP请求
   ↓ 解析JSON
   ↓ 调用init中初始化的userClient
   
3. RPC客户端（Kitex生成）
   ↓ 序列化请求
   ↓ 发送到127.0.0.1:8888
   
4. 网络传输
   ↓ TCP连接
   
5. RPC服务端（Kitex生成，监听8888端口）
   ↓ 接收请求
   ↓ 反序列化
   ↓ 调用handler.Register()
   
6. 业务逻辑（handler.go）
   ↓ 执行验证
   ↓ 返回响应
   
7. 反向流程
   RPC服务端 → 序列化响应 → 网络 → RPC客户端 → API网关 → HTTP响应
```

