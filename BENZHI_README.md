# GeoIndex

GeoIndex 是通用地理空间索引服务：空间数据（点、矩形、多边形）写入后按网格
建立空间索引，范围查询与最近邻查询走索引加速；瓦片按网格批量生成给前端；
支持索引全量重建与空间变更预写日志，并提供浏览器地图瓦片浏览页面。

## 构建

```bash
./build_benzhi_docker.sh geoindex linux/amd64
./build_benzhi_docker.sh geoindex linux/arm64
```

## 运行

```bash
go run ./cmd/geoindex -addr :8080
```

启动后打开 http://localhost:8080 即可看到地图瓦片浏览页面。

## 接口

- `GET /healthz` 健康检查
- `POST /api/v1/points` 写入空间点
- `POST /api/v1/points/delete` 删除空间点
- `POST /api/v1/polygons` 更新多边形边界
- `GET /api/v1/query/range` 范围查询
- `GET /api/v1/query/nearest` 最近邻查询
- `GET /api/v1/tiles/{z}/{x}/{y}` 瓦片内容
- `POST /api/v1/tiles/invalidate` 清空瓦片缓存
- `POST /api/v1/rebuild` 全量重建索引
- `GET /api/v1/stats` 服务统计

## 容器内验证

```bash
go build ./...
go test ./...
go vet ./...
go run ./cmd/geoindex -addr :8080
```
