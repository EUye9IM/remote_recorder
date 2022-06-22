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

```