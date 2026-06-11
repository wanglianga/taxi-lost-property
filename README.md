# 出租车遗失物认领服务

## 项目简介

本系统为出租车公司提供遗失物认领全流程管理服务，涵盖乘客报失、订单智能匹配、司机上交登记、物品入库管理、客服核验认领、归还记录追踪等完整业务闭环。

## 原始需求

> 出租车公司需要遗失物认领服务，Go 接口处理乘客报失、订单匹配、司机上交、物品入库、认领核验和归还记录。业务对象包括乘车时间、上下车地点、支付流水、车牌、司机、物品描述、照片、站点柜号、核验材料和寄回方式。乘客不知道车牌时，服务通过支付时间和路线匹配可能车辆；司机发现物品后登记上交；客服核验物品特征后安排到店领取或快递寄回。服务要区分无匹配车辆、司机未发现、物品已上交、多人认领和贵重物品需实名核验。

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite（嵌入式，开箱即用）
- **容器化**: Docker + Docker Compose

## 项目结构

```
wl-299/
├── main.go                      # 程序入口
├── internal/
│   ├── model/
│   │   └── models.go           # 数据模型定义
│   ├── repository/
│   │   └── db.go               # 数据库初始化和种子数据
│   ├── service/
│   │   └── service.go          # 业务逻辑层
│   └── handler/
│       └── handler.go          # HTTP处理器层
├── Dockerfile
├── docker-compose.yml
├── .dockerignore
├── go.mod
├── go.sum
└── README.md
```

## 数据模型

| 模型 | 说明 |
|------|------|
| LostReport | 乘客遗失物报失记录 |
| TaxiOrder | 出租车订单信息 |
| DriverSubmission | 司机物品上交记录 |
| ItemInventory | 物品入库库存 |
| Station | 认领站点信息 |
| ClaimRecord | 认领申请与归还记录 |

## 业务状态流转

### 报失单状态 (LostReportStatus)
- `pending_match` - 待匹配订单
- `no_match` - 无匹配车辆
- `driver_not_found` - 已匹配车辆，待司机反馈
- `submitted` - 司机已上交物品
- `multiple_claims` - 多人认领冲突
- `verified` - 认领核验通过
- `returned` - 物品已归还

### 认领单状态 (ClaimStatus)
- `pending` - 待核验
- `verified` - 核验通过
- `rejected` - 核验拒绝
- `returned` - 已归还

## API 文档

### 基础路径: `/api/v1`

所有请求和响应均使用 JSON 格式。

### 通用响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code: 0` 表示成功，非 0 表示失败
- `message` 为错误描述或 "success"

---

### 1. 乘客报失

**POST** `/api/v1/reports`

请求体：
```json
{
  "passenger_name": "张三",
  "passenger_phone": "13800138000",
  "passenger_id_card": "110101199001011234",
  "ride_time": "2026-06-10T14:30:00Z",
  "boarding_point": "北京首都机场T3",
  "alighting_point": "朝阳区国贸中心",
  "payment_no": "PAY20260610001",
  "amount": 98.50,
  "plate_no": "",
  "driver_name": "",
  "item_description": "黑色双肩包，内有笔记本电脑一台",
  "item_category": "电子产品",
  "item_value": 8000,
  "is_valuable": true,
  "photos": "url1,url2"
}
```

**注意**：贵重物品 (`is_valuable=true`) 必须提供身份证号。

---

### 2. 查询报失记录列表

**GET** `/api/v1/reports?status=pending_match&page=1&size=10`

查询参数：
- `status` (可选): 按状态筛选
- `page`: 页码，默认 1
- `size`: 每页数量，默认 10，最大 100

---

### 3. 查询单条报失记录

**GET** `/api/v1/reports/:id`

---

### 4. 匹配可能的订单（智能匹配）

**POST** `/api/v1/reports/:id/match`

系统根据乘车时间、支付流水号、上下车地点、金额、车牌号等多维度匹配可能的出租车订单，并返回匹配度评分。

---

### 5. 确认匹配订单

**POST** `/api/v1/reports/:id/confirm-order`

请求体：
```json
{
  "order_id": 1
}
```

---

### 6. 司机上交物品

**POST** `/api/v1/submissions`

请求体：
```json
{
  "report_id": 1,
  "order_id": 1,
  "driver_id": 1001,
  "driver_name": "张师傅",
  "plate_no": "京B·12345",
  "item_description": "黑色双肩包，内有笔记本电脑一台",
  "item_category": "电子产品",
  "item_value": 8000,
  "is_valuable": true,
  "photos": "url1,url2",
  "found_location": "后排座椅下方",
  "found_time": "2026-06-10T16:00:00Z",
  "remark": "物品完好"
}
```

---

### 7. 查询司机上交记录列表

**GET** `/api/v1/submissions?page=1&size=10`

---

### 8. 物品入库

**POST** `/api/v1/inventory`

请求体：
```json
{
  "submission_id": 1,
  "report_id": 1,
  "station_id": 1,
  "cabinet_no": "A-01-03",
  "stored_by": "客服小王"
}
```

---

### 9. 查询库存列表

**GET** `/api/v1/inventory?status=in_stock&page=1&size=10`

---

### 10. 查询所有站点

**GET** `/api/v1/stations`

---

### 11. 乘客提交认领申请

**POST** `/api/v1/claims`

请求体（到店领取）：
```json
{
  "report_id": 1,
  "inventory_id": 1,
  "passenger_name": "张三",
  "passenger_phone": "13800138000",
  "passenger_id_card": "110101199001011234",
  "verify_materials": "购买凭证照片",
  "verify_description": "电脑序列号为XXX，包内有一张名片写有本人姓名",
  "return_method": "pickup",
  "pickup_station_id": 1
}
```

请求体（快递寄回）：
```json
{
  "report_id": 1,
  "inventory_id": 1,
  "passenger_name": "张三",
  "passenger_phone": "13800138000",
  "passenger_id_card": "110101199001011234",
  "verify_materials": "购买凭证照片",
  "verify_description": "电脑序列号为XXX",
  "return_method": "express",
  "express_company": "顺丰速运",
  "express_no": "",
  "receiver_name": "张三",
  "receiver_phone": "13800138000",
  "receiver_address": "北京市朝阳区建国路88号"
}
```

**注意**：贵重物品认领必须提供身份证号。

---

### 12. 查询认领申请列表

**GET** `/api/v1/claims?status=pending&report_id=1&page=1&size=10`

---

### 13. 客服核验认领申请

**POST** `/api/v1/claims/:id/verify`

请求体（通过）：
```json
{
  "verified_by": "客服小李",
  "pass": true
}
```

请求体（拒绝）：
```json
{
  "verified_by": "客服小李",
  "pass": false,
  "reject_reason": "物品特征描述不符"
}
```

---

### 14. 确认物品归还

**POST** `/api/v1/claims/:id/return`

请求体：
```json
{
  "operator": "客服小王",
  "returned_at": "2026-06-11T10:00:00Z"
}
```

---

### 15. 查询订单列表

**GET** `/api/v1/orders?page=1&size=10`

---

### 16. 健康检查

**GET** `/health`

响应：
```json
{
  "service": "taxi-lost-property",
  "status": "ok"
}
```

## 启动方式

### 前置要求

- **方式一（推荐）**: Docker 20.10+、Docker Compose v2+
- **方式二**: Go 1.21+

### Docker 一键启动（推荐）

#### 1. 启动服务

```bash
docker compose up --build
```

后台运行：

```bash
docker compose up --build -d
```

#### 2. 查看服务状态

```bash
docker compose ps
```

#### 3. 查看日志

```bash
docker compose logs -f
```

#### 4. 停止并清理服务

```bash
docker compose down
```

访问地址：http://localhost:8080

健康检查：http://localhost:8080/health

---

### 本地源码启动

#### 1. 安装依赖

```bash
go mod download
```

#### 2. 启动服务

```bash
go run main.go
```

访问地址：http://localhost:8080

健康检查：http://localhost:8080/health

## 注意事项

1. 数据库使用 SQLite，数据文件 `taxi_lost_property.db` 会自动创建在当前目录
2. Docker 模式下，数据通过 volume 挂载到 `./data` 目录，容器重建数据不丢失
3. 系统内置 3 个示例站点和 3 条示例订单数据用于测试
4. 贵重物品报失和认领强制要求身份证号实名核验
5. 当同一物品出现多人认领时，报失单状态会标记为 `multiple_claims`，需客服人工处理
