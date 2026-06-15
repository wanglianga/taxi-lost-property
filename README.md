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

### 16. 模糊查询候选车辆（隐私保护）

**POST** `/api/v1/fuzzy/search-vehicles`

当乘客只记得大概乘车时间和区域时，使用此接口查询候选车辆。接口返回脱敏后的车辆信息，不直接暴露司机姓名、电话、完整车牌号等隐私信息。

请求体：
```json
{
  "approx_time": "2026-06-10T14:30:00Z",
  "time_window_min": 120,
  "area": "朝阳区",
  "route_keyword": "机场",
  "payment_no_part": "PAY2026",
  "fleet_company": "首汽",
  "amount_min": 50,
  "amount_max": 200
}
```

请求参数说明：
- `approx_time` (必填): 大概乘车时间
- `time_window_min`: 时间窗口（分钟），默认 120 分钟
- `area`: 区域关键词（如朝阳区、海淀区）
- `route_keyword`: 路线关键词（匹配上下车地点）
- `payment_no_part`: 支付流水号片段
- `fleet_company`: 车队公司名称
- `amount_min`: 最小金额
- `amount_max`: 最大金额

响应示例（脱敏后）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "order_id": 1,
        "plate_no_masked": "京B****5",
        "fleet_company": "首汽集团",
        "ride_time": "2026-06-10T10:30:00Z",
        "boarding_point": "北京首都机场T3",
        "alighting_point": "朝阳区国贸中心",
        "amount": 98.5,
        "payment_no_masked": "PAY*******01",
        "match_rate": 0.95
      }
    ],
    "total": 1
  }
}
```

**隐私保护说明**：
- 车牌号仅显示前3位和最后1位，中间用 `*` 代替
- 支付流水号仅显示前3位和最后2位，中间用 `*` 代替
- 不返回司机姓名、司机电话、司机ID等隐私信息
- 结果按匹配度从高到低排序

---

### 17. 获取贵重物品核验要求

**GET** `/api/v1/verify/requirements?inventory_id=1`

查询认领某物品时需要提供的核验材料。系统会自动识别物品类型（手机/钱包/证件/电脑），并返回对应的核验要求。

查询参数：
- `inventory_id` (必填): 物品库存ID

响应示例（手机）：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "item_type": "phone",
    "is_valuable": true,
    "requirements": [
      {
        "field": "wallpaper_desc",
        "description": "请描述手机屏保图片内容（如：蓝色海洋壁纸、家人照片等）",
        "required": true
      }
    ]
  }
}
```

---

### 18. 乘客提交认领申请（含贵重物品核验）

**POST** `/api/v1/claims`

对于手机、钱包、证件、电脑等贵重物品，需要根据物品类型提供对应的核验材料。

请求体（认领手机）：
```json
{
  "report_id": 1,
  "inventory_id": 1,
  "passenger_name": "张三",
  "passenger_phone": "13800138000",
  "passenger_id_card": "110101199001011234",
  "verify_materials": "购买凭证照片",
  "valuable_verify_materials": {
    "wallpaper_desc": "蓝色海洋壁纸，右下角有一张全家福照片"
  },
  "return_method": "pickup",
  "pickup_station_id": 1
}
```

请求体（认领钱包）：
```json
{
  "report_id": 1,
  "inventory_id": 2,
  "passenger_name": "李四",
  "passenger_phone": "13900139000",
  "passenger_id_card": "110101199002025678",
  "valuable_verify_materials": {
    "wallet_items": "内有身份证1张、建行银行卡2张、现金约500元、星巴克会员卡1张"
  },
  "return_method": "pickup",
  "pickup_station_id": 1
}
```

请求体（认领证件）：
```json
{
  "report_id": 1,
  "inventory_id": 3,
  "passenger_name": "王五",
  "passenger_phone": "13700137000",
  "passenger_id_card": "110101199003039012",
  "valuable_verify_materials": {
    "id_card_tail": "9012"
  },
  "return_method": "express",
  "express_company": "顺丰速运",
  "receiver_name": "王五",
  "receiver_phone": "13700137000",
  "receiver_address": "北京市海淀区中关村大街1号"
}
```

请求体（认领电脑）：
```json
{
  "report_id": 1,
  "inventory_id": 4,
  "passenger_name": "赵六",
  "passenger_phone": "13600136000",
  "passenger_id_card": "110101199004043456",
  "valuable_verify_materials": {
    "computer_info": "MacBook Pro 14寸银色，桌面背景是雪山，登录密码提示为生日"
  },
  "return_method": "pickup",
  "pickup_station_id": 1
}
```

**贵重物品核验规则**：
- **手机**: 必须提供屏保图片描述，至少5个字符
- **钱包**: 必须描述包内物品，至少10个字符
- **证件**: 必须提供身份证号后4位，长度必须为4位
- **电脑**: 必须提供设备特征描述，至少10个字符

---

### 19. 健康检查

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
