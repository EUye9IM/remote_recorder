# 现有ws包整理（服务端视角）
考生端会在第一次连接，即开启本地流之后，上传streamid，为方便，在此处附上一个uuid，表明考生端将会建立一个`考生-服务端`的peerConnection连接，我们记此uuid为serveruuid，方便后续表述。下面以S/T代表学生和教师两端。

| Action    | 结构                                                                   | 发起方       | 服务端处理方式（处理/转发）                               |
| --------- | ---------------------------------------------------------------------- | ------------ | --------------------------------------------------------- |
| token     | `{action: "token", data: ...}`                                         | 客户S/T      | 处理                                                      |
| event     | `{action: "event", data:{event: event_type, ...}}`                     | 服务端/T     | 处理、发送                                                |
| streamid  | `{action: "streamid", data:{screen:..., camera:...}, uuid:serveruuid}` | 客户S        | 处理，记录serveruuid，随后将会收到连接offer               |
| offer     | `{action: "offer", data:... }`                                         | 客户S        | 处理/转发，处理uuid=serveruuid的情况，其余转发给对应的T端 |
| candidate | `{action: "candidate", data: ...}`                                     | 客户S/T/服务 | 处理/转发，与上同理，serveruuid外转发对端                 |
| answer    | `{action: "answer", data: ...}`                                        | 客户T/服务   | 处理转发同上-                                             |
| candidate | `{action: "candidate", data: ...}`                                     | 服务         | -                                                         |
