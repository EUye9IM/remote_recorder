# 补丁
## 新约定
为解决监考端的通信问题，现在开始考虑使用新约定，以求增加考生端压力

```
下面的uuid是为区分peerConnection而设置，由考生端生成，对端连接回复时只需要携带转发即可

--------------------
action: 'offer' 
data: {
    'data': offer值
    uuid: uuid
}

--------------------
action: 'answer'
data: {
    'data': answer值
    uuid: uuid
}

---------------------
action: 'candidate'
data: {
    'data': candidate值
    uuid: uuid
}

----------------------
action: 'uuid'
data: {
    'uuid': uuid
}

```

## 现有ws包整理（服务端视角）
考生端会在第一次连接，即开启本地流之后，上传streamid，为方便，在此处附上一个uuid，表明考生端将会建立一个`考生-服务端`的peerConnection连接，我们记此uuid为serveruuid，方便后续表述。下面以S/T代表学生和教师两端。

### 服务端收
| Action    | 结构                                                                    | 发起方  | 服务端处理方式（处理/转发）                        |
| --------- | ----------------------------------------------------------------------- | ------- | -------------------------------------------------- |
| token     | `{action: "token", data: ...}`                                          | 客户S/T | 处理                                               |
| event     | `{action: "event", data:{event: event_type, ...}}`                      | 客户T   | 处理、发送                                         |
| streamid  | `{action: "streamid", data:{screen:..., camera:...}, uuid:serveruuid }` | 客户S   | 处理，记录serveruuid，随后将会收到连接offer        |
| offer     | `{action: "offer", data:{data: ..., uuid: uuid} }`                      | 客户S   | 处理/转发，处理uuid=serveruuid的情况，其余转发对端 |
| candidate | `{action: "candidate", data:{data: ..., uuid: uuid} }`                  | 客户S/T | 处理/转发，处理uuid=serveruuid的情况，其余转发对端 |
| answer    | `{action: "answer",  data:{data: ..., uuid: uuid} }`                    | 客户T   | 转发                                               |

### 服务端发
| Action    | 结构                                                                       | 接受方     |
| --------- | -------------------------------------------------------------------------- | ---------- |
| answer    | `{action: "answer", data: ...}`                                            | S          |
| answer    | `{action: "answer", data:{data: ..., uuid: uuid}}`                         | S(转发自T) |
| event     | `{action: "event", data:{event: SendStreamId,streamid: ... , uuid: uuid}}` | T          |
| candidate | `{action: "candidate", data: ...}`                                         | S          |
| candidate | `{action: "candidate",data:{data: ..., uuid: uuid}}`                       | S(转发自T) |
| offer     | `{action: "offer", data:{data: ..., uuid: uuid} }`                         | 服务(转发) |


<!-- ### 监控查看时序

```
        S                        服务器                         T
                                    <-- event:GetMemberStream --|       
                                   |----event:SendStreamId----->         服务器编一个uuid -->




```