# ShadowMixer — 企业级大模型隐私混淆与边界防御引擎

ShadowMixer 是一款专为 AI 时代设计的开源安全中间件。它通过核心的碎片化混淆（Fragmented Obfuscation）与多租户匿名聚合技术，将 OpenAI、Anthropic、Gemini 等大模型厂商彻底降级为受控的“匿名算力供应商”。在“零信任” AI 架构下，ShadowMixer 能够确保任何敏感意图、核心代码或私域知识在离开企业内网前，就已经被拆解为无法还原的语义碎片。

本项目不仅是开发者的利器，更是构建 **AI 安全关口（AI Security Gateway）** 的核心引擎。在多用户并发环境下，ShadowMixer 会产生强大的“群体掩护”效应，让追踪单个用户的商业意图在统计学上变得几乎不可能。

> 💡 **核心比喻：“消失在人海里的红烧肉”**
>
> 你想吃红烧肉，但不想让外界知道食谱。ShadowMixer 将食材切碎，混入全城成千上万人的食材订单中，随机分发给不同的厨师（LLM 厂商）。厨师们只看到无数人在买“糖、肉、酱油”，却无法拼凑出谁要吃红烧肉，更无法偷走你的独家秘方。

## ✨ 核心安全特性

1. **群体匿名效应 (Crowd Anonymity)**
   - **多租户混淆**：不同用户的任务碎片进入同一个全局调度池。在大模型厂商看来，这些请求序列是交织在一起的“语义流”，无法通过 IP 或 API Key 区分行为边界。
   - **网络规模增益**：用户越多，隐私越强。随着并发量增加，单个用户的特征会被淹没在海量的背景噪声中，彻底瓦解厂商的用户画像能力。

2. **拟人化外壳与算力节约 (Anthropomorphic Shell & Efficiency)**
   - **防风控伪装**：为干瘪的碎片穿上自然语言“外壳”，使其看起来像合法的、独立的咨询请求，规避厂商的输入完整性校验。
   - **零算力浪费 (Zero Compute Waste)**：ShadowMixer 专注于高效混淆，**绝不通过发送无效请求来浪费宝贵的算力资源**。每一分算力都用于真实的业务价值。

3. **本地状态机与分层路由 (Local State & Tiered Routing)**
   - **逻辑重组**：本地数据库实时维护任务状态，无需将上下文传回云端。
   - **隐私分级**：极高密级任务本地小模型处理，计算型任务云端混淆处理。

## ⚙️ 核心工作流

1. **分解与脱敏 (Decompose & Mask)**：将复杂指令拆解为 $N$ 个原子碎片，并在本地完成实体加密/占位。
2. **群体注入 (Shuffle & Inject)**：将所有用户的碎片混入高并发池，加入随机延迟（Jitter）和顺序打乱。
3. **算力路由 (Compute Routing)**：Worker 节点从池中捞取碎片，利用 Key Pooling 分布式请求上游厂商。
4. **智能聚合 (Reassemble)**：聚合器根据 TaskID 剥离外壳、滤除噪声、还原实体，将拼装好的结果交付用户。

## 🎯 适用场景

- **企业 AI 隐私防火墙**：解决企业员工违规使用 ChatGPT 泄露代码、商业计划书的合规痛点。
- **去中心化 AI 安全代理**：作为 Agent 的安全通信层，阻断云端对企业“思考链”的侦听。
- **数据资产脱敏中台**：医疗、金融领域在合规前提下利用公有云算力进行超大规模文档处理。
- **低成本隐私替代方案**：相比于联邦大模型（Federated LLM）极高的算力门槛与复杂的跨机构协同成本，ShadowMixer 提供了一种“零基础设施”的隐私保护路径。
- **公有云能力平替**：企业无需构建昂贵的本地 GPU 集群进行模型训练或微调，即可在保障私域数据安全的前提下，直接调动公有云顶尖模型的推理能力，极大地降低了 AI 合规的 TCO（总拥有成本）。

## 🗺️ 架构图

```mermaid
graph TD
    subgraph "Enterprise Secure Zone (Multi-User)"
        U1[User A] --> Gateway
        U2[User B] --> Gateway
        U3[User C] --> Gateway
        Gateway -->|Decompose| LocalEngine[Local NLP / State DB]
        LocalEngine -->|Shuffle & Jitter| FragmentPool[Global Anonymous Pool]
    end
     
    subgraph "Obfuscation & Distribution Layer"
        FragmentPool -->|Encapsulate| Worker1
        FragmentPool -->|Encapsulate| Worker2
        FragmentPool -->|Encapsulate| Worker3
    end
     
    subgraph "Public Cloud (Compute Providers)"
        Worker1 -->|Fragment| OpenAI
        Worker2 -->|Fragment| Gemini
        Worker3 -->|Fragment| Anthropic
    end
     
    OpenAI -->|Result| LocalEngine
    Gemini -->|Result| LocalEngine
     
    LocalEngine -->|Reassemble| Gateway
    Gateway -->|Final Response| U1
```

## 🚀 快速开始

### 1. 启动安全引擎

```bash
# 部署 ShadowMixer 多用户隐私集群
docker-compose up --build -d
```

### 2. 配置安全策略 (`config.yaml`)

```yaml
security:
  anonymization_level: "high"   # 开启多用户交叉混淆
  local_masking: true          # 开启本地实体脱敏

routing:
  api_pools: 
    - provider: "openai"
      keys: ["sk-1", "sk-2", "sk-3"]
    - provider: "anthropic"
      keys: ["sk-ant-1"]
```

### 3. API 调用 (OpenAI 兼容模式)

ShadowMixer 提供完全透明的接口封装，只需更改 Base URL 即可实现隐私加固：

```bash
curl -X POST http://localhost:8080/v1/secure/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "分析这份核心代码逻辑：[代码碎片...]"}]
  }'
```

## 📄 许可证

本项目采用 MIT License。

---

# ShadowMixer — Enterprise-Grade LLM Privacy Obfuscation & Defense Engine

ShadowMixer is an open-source security middleware designed for the AI era. Through its core Fragmented Obfuscation and Multi-Tenant Anonymous Aggregation technologies, it effectively downgrades major LLM providers like OpenAI, Anthropic, and Gemini into controlled "anonymous compute vendors." Under a "Zero Trust" AI architecture, ShadowMixer ensures that any sensitive intent, core code, or proprietary knowledge is decomposed into irreversible semantic fragments before leaving the corporate intranet.

This project is not just a tool for developers but the core engine for building an **AI Security Gateway**. In a multi-user concurrent environment, ShadowMixer generates a powerful "crowd cover" effect, making it statistically impossible to trace the commercial intent of a single user.

> 💡 **Core Analogy: "The Braised Pork in the Crowd"**
>
> You want to eat braised pork, but you don't want the outside world to know the recipe. ShadowMixer chops the ingredients and mixes them into the orders of thousands of people across the city, distributing them randomly to different chefs (LLM vendors). The chefs only see countless people buying "sugar, meat, soy sauce," but they cannot piece together who is eating braised pork, nor can they steal your exclusive recipe.

## ✨ Core Security Features

1. **Crowd Anonymity**
   - **Multi-Tenant Obfuscation**: Task fragments from different users enter the same global scheduling pool. To LLM vendors, these request sequences appear as interwoven "semantic streams," indistinguishable by IP or API Key boundaries.
   - **Network Scale Gain**: The more users, the stronger the privacy. As concurrency increases, a single user's characteristics are drowned out in massive background noise, completely dismantling the vendor's user profiling capabilities.

2. **Anthropomorphic Shell & Efficiency**
   - **Anti-Risk Control Camouflage**: Wraps dry fragments in a natural language "shell" to make them look like legitimate, independent inquiries, bypassing vendor input integrity checks.
   - **Zero Compute Waste**: ShadowMixer focuses on efficient obfuscation and **never wastes precious compute resources by sending invalid requests**. Every bit of compute is used for real business value.

3. **Local State & Tiered Routing**
   - **Logic Reassembly**: A local database maintains task state in real-time, eliminating the need to send context back to the cloud.
   - **Privacy Tiering**: Extremely sensitive tasks are handled by local small models, while computational tasks are processed via cloud obfuscation.

## ⚙️ Core Workflow

1. **Decompose & Mask**: Decomposes complex instructions into $N$ atomic fragments and performs entity encryption/masking locally.
2. **Shuffle & Inject**: Mixes all user fragments into a high-concurrency pool, adding random delays (Jitter) and shuffling the order.
3. **Compute Routing**: Worker nodes retrieve fragments from the pool and use Key Pooling to distribute requests to upstream vendors.
4. **Reassemble**: The aggregator strips shells, filters noise, restores entities based on TaskID, and delivers the assembled result to the user.

## 🎯 Use Cases

- **Enterprise AI Privacy Firewall**: Solves compliance issues where employees leak code or business plans while using ChatGPT.
- **Decentralized AI Security Agent**: Acts as a secure communication layer for Agents, blocking cloud eavesdropping on the enterprise "chain of thought."
- **Data Asset Desensitization Hub**: Enables healthcare and finance sectors to use public cloud compute for massive document processing under compliance.
- **Low-Cost Privacy Alternative**: Compared to the high compute threshold and complex cross-organization coordination of Federated LLMs, ShadowMixer offers a "zero infrastructure" privacy protection path.
- **Public Cloud Capability Replacement**: Enterprises can leverage top-tier public cloud model inference without building expensive local GPU clusters for training or fine-tuning, significantly reducing the TCO of AI compliance while ensuring data safety.

## 🗺️ Architecture Diagram

```mermaid
graph TD
    subgraph "Enterprise Secure Zone (Multi-User)"
        U1[User A] --> Gateway
        U2[User B] --> Gateway
        U3[User C] --> Gateway
        Gateway -->|Decompose| LocalEngine[Local NLP / State DB]
        LocalEngine -->|Shuffle & Jitter| FragmentPool[Global Anonymous Pool]
    end
     
    subgraph "Obfuscation & Distribution Layer"
        FragmentPool -->|Encapsulate| Worker1
        FragmentPool -->|Encapsulate| Worker2
        FragmentPool -->|Encapsulate| Worker3
    end
     
    subgraph "Public Cloud (Compute Providers)"
        Worker1 -->|Fragment| OpenAI
        Worker2 -->|Fragment| Gemini
        Worker3 -->|Fragment| Anthropic
    end
     
    OpenAI -->|Result| LocalEngine
    Gemini -->|Result| LocalEngine
     
    LocalEngine -->|Reassemble| Gateway
    Gateway -->|Final Response| U1
```

## 🚀 Quick Start

### 1. Start the Security Engine

```bash
# Deploy ShadowMixer Multi-User Privacy Cluster
docker-compose up --build -d
```

### 2. Configure Security Policy (`config.yaml`)

```yaml
security:
  anonymization_level: "high"   # Enable multi-user cross-obfuscation
  local_masking: true          # Enable local entity masking

routing:
  api_pools: 
    - provider: "openai"
      keys: ["sk-1", "sk-2", "sk-3"]
    - provider: "anthropic"
      keys: ["sk-ant-1"]
```

### 3. API Call (OpenAI Compatible Mode)

ShadowMixer provides a fully transparent interface wrapper. Just change the Base URL to achieve privacy hardening:

```bash
curl -X POST http://localhost:8080/v1/secure/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Analyze this core code logic: [Code Fragment...]"}]
  }'
```

## 📄 License

This project is licensed under the MIT License.
