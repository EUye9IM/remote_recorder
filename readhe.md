# 现有ws包整理（服务端视角）

|Action | 结构 | 发起方 |服务端处理方式（处理/转发） |
|---| ---|------|---------------|
| token | `{action: "token", data: ...}` | 客户 | 处理 |
| event | `{action: "event", data:{event: event_type, ...}}` | 客户 | 处理 |
| streamid | `{action: "streamid", data:{screen:..., camera:...}}` | 客户 | 处理 |
| offer | `{action: "offer", data:... }` | 客户 | 处理 |
| candidate | `{action: "candidate", data: ...}` | 客户 | 处理 |
| answer | `{action: "answer", data: ...}` | 服务 | - |
| candidate | `{action: "candidate", data: ...}` | 服务 | - |
