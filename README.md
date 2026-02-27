# ShadowMixer — 面向大模型的开源 API 隐私网关

ShadowMixer 是一个用于拦截和代理发往大模型厂商（OpenAI/Anthropic 等）的请求的隐私网关。它通过身份剥离、队列抖动和多 Key 轮询混淆请求指纹，最大化降低用户被外部服务关联与画像的风险，同时保持对业务透明的响应，包括流式返回。

## 亮点特性
- 身份剥离（Stateless）
  - 丢弃来自客户端的真实身份标识（Headers/IP 等），以全新服务端请求调用上游 LLM
- 队列与抖动（Queue & Jitter）
  - 请求入队至 Redis，Worker 取出时强制加入 0.1s–2.0s 的随机延迟，打乱时间序列指纹
- 多 Key 轮询（Key Pooling）
  - 从配置的 Key 池中按轮询策略选择不同的上游 API Key，进一步模糊来源
- 透明返回（Streaming 支持）
  - 支持流式转发（SSE/Chunked），TTFT 更短，体验接近原生上游

## 架构与数据流
1. 客户端请求到达网关（Gin）
2. 网关读取 Body，忽略所有客户端身份标识；生成 TaskID 入队（Redis List）
3. Worker 从队列取任务，随机延迟（0.1–2.0s），按轮询选 Key 调用上游 LLM
4. Worker 将上游响应以流式块发布到 Redis Pub/Sub（response:TaskID）
5. 网关订阅对应频道，按客户端期望（JSON 或 event-stream）透明转发

## 安全与隐私
- 不保留客户端原始 Headers/IP
- 服务端重建请求并附加 API Key
- 时间抖动与 Key 轮询降低可关联性
- 默认不打印敏感数据；请遵循最小可见日志策略

## 路线图
- 更完整的 SSE/分块协议适配（OpenAI/Anthropic）
- 自适应限流与并发控制
- 可插拔的混淆策略（抖动分布、Key 选择算法）
- 指标与观测（Prometheus/Grafana）
- 简单鉴权（可选）与租户隔离

## 许可证
MIT License
