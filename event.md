# event 约定
## 会议房间的 event
目前考虑的端对端的连接，只有两台机器

| event | 服务端 | 客户端 | 说明 | 数据 |
| ---- | ----   | ----  | ---- | ---- |
| `MemberJoined` | emit  | on | 成员加入房间 | MemberId |
| `MemberLeft`   | emit  | on | 成员离开房间 | MemberId |
| `MessageToPeer`   | on      | emit | 端对端传递数据，配合MessageFromPeer传递，触发MessageFromPeer | json, memberid |
| `MessageFromPeer` | emit    | on | 端对端传递数据 | 同上 |


## api
url：`/api/sio`