# eru-stats

## 功能

- 抓取etcd中 core 和 agent 中的数据，统计信息
- 透传citadel api中的数据

## 配置 - 环境变量

- `AGENT_PREFIX` etcd中agent的prefix (默认是 /agent2)
- `CORE_PREFIX` etcd中core的prefix (默认是 /eru-core)
- `CITADEL_URL` citadel的地址 (默认是 http://citadel.ricebook.net)
- `CITADEL_AUTH_TOKEN` citadel的认证token (默认是 hello)
- `ETCD_ENDPOINTS` etcd地址 (默认是本机IP)

## 接口示例

### `/stats`

```json
{
  "Agent": {
    "Containers": 256,
    "Nodes": 30
  },
  "Core": {
    "Containers": 256,
    "Nodes": 30
  },
  "NodesMemcap": {
    "redis-master": {
      "c2-docker-26": {
        "diff": "-36666605568 bytes",
        "total": "31.16 GiB",
        "used": "28.17 GiB",
        "used_by_memcap": "-6417285120 bytes"
      },
      "c2-docker-29": {
        "diff": "-51581550592 bytes",
        "total": "31.16 GiB",
        "used": "22.98 GiB",
        "used_by_memcap": "-26910654464 bytes"
      }
    }
  }
}
```

- `Agent` agent中的数据
- `Core` core中的数据
  - `Containers` 容器数量
  - `Nodes` 节点数量
- `NodesMemcap` node的内存对比
  - `total` node总内存量
  - `used_by_memcap` 总内存 - 保留内存 - 容器使用内存(core中记录)
  - `used` node上所有容器统计使用内存 (citadel mysql统计)
  - `diff` `used_by_memcap` - `used`

### `/diff`

```json
{
  "container": {
    "agentLess": [],
    "agentMore": []
  },
  "nodes": {
    "agentLess": [],
    "agentMore": []
  }
}
```

- `container`
  - `agentLess` agent比core少的container记录
  - `agentMore` agent比core多的container记录
- `nodes`
  - `agentLess` agent比core少的nodes记录
  - `agentMore` agent比core多的nodes记录

### `/apps`

```json
{
  "abtest": {
    "Entrypoints": {
      "web": {
        "Count": 1,
        "Mem": 536870912
      }
    },
    "MemTotal": 536870912,
    "Mem": "512 MB",
    "CPUTotal": 0,
    "Count": 1
  },
  "akihabara": {
    "Entrypoints": {
      "prod-c2": {
        "Count": 2,
        "Mem": 8589934592
      }
    },
    "MemTotal": 8589934592,
    "Mem": "8192 MB",
    "CPUTotal": 0,
    "Count": 2
  }
}
```

- 首项为app的name
- `Entrypoints` 中记录app的所有Entrypoint
  - `Count` Entrypoint部署数量
  - `Mem` Entrypoint总内存消耗 (bytes)
- `MemTotal` | `Mem` app使用内存量 (bytes | Mb)
- `CPUTotal` app使用CPU资源
- `Count` app数量

### `/pods`

```json
{
  "data": [
    {
      "HostName": "c2-data-3",
      "PodName": "data",
      "Mem": "256567 MB | 250 GB"
    },
    {
      "HostName": "c2-data-5",
      "PodName": "data",
      "Mem": "256567 MB | 250 GB"
    }
  ],
  "elb": [
    {
      "HostName": "c2-docker-11",
      "PodName": "elb",
      "Mem": "623 MB | 0 GB"
    }
  ]
}
```

- `HostName` 主机名
- `PodName` pod名
- `Mem` node总内存