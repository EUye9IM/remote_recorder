/*************************************************
 * @abstract 以面向过程的方式进行设计
 * @brief    
 ************************************************/

let ws;
let cameraStream;
let screenStream;
// let peerConnection;
let streamType;
let userType;

let peerConnections = {}

let id2content = {};

function guuid() {
    return crypto.randomUUID()
}

const isSec = window.location.protocol == "https:";
const host = window.location.host;
let url;
if (isSec) {
    url = "wss://" + host + "/api/ws"
}
else {
    url = "ws://" + host + "/api/ws"
}

const servers = {
    iceServers: [
        {
            urls: 'stun:stun.l.google.com:19302'
        }
    ],
}

const displayMediaOptions = {
    video: {
        frameRate: { ideal: 15 }
    }
}

const mediaStreamConstrains = {
    video: {
        width: { min: 640, ideal: 1920, max: 1920 },
        height: { min: 480, ideal: 1080, max: 1080 },
    },
    audio: true
}

const getCookie = (name) => {
    return document.cookie.split('; ').reduce((r, v) => {
        const parts = v.split('=')
        return parts[0] === name ? decodeURIComponent(parts[1]) : r
    }, '')
}

const initWebSocket = (url) => {
    ws = new WebSocket(url)

    ws.onmessage = handleMessage
    ws.onopen = event => {
        console.log('Websocket open.')
        // 发送token信息
        ws.send(JSON.stringify({
            'action': "token",
            'data': getCookie("token"),
        }))
        console.log('token send.')
    }
    ws.onerror = event => {
        console.log('Websocket error!')
    }
    ws.onclose = event => {
        console.log('Websocket closed.')
        alert('与服务端失去连接，请直接关闭标签页')
        // window.location = '/login'
    }
}

// 等待websocket进入连接状态
async function waitForSocketConnection(socket, callback) {
    setTimeout(
        () => {
            if (socket.readyState === 1) {
                console.log("Connection is made.")
                if (callback != null) {
                    callback();
                }
            } else {
                console.log("wait for connection...")
                waitForSocketConnection(socket, callback);
            }
        }, 5); // wait 5 milisecond for the connection...
}

const handleMessage = async event => {
    const message = JSON.parse(event.data)
    // console.log('message from ' + message.from)
    // if (message.from === userType) {
    //     return
    // }
    console.log(message)
    console.log(`recieve ${message.action}.`)
    

    if (message.action === 'event') {
        // 事件处理
        handleEvent(message.data)
    }

    const uuid = message.uuid
    const data = message.data

    if (message.action === 'offer') {
        createAnswer(uuid, data)
    }

    if (message.action === 'answer') {
        addAnswer(uuid, data)
    }

    if (message.action === 'candidate') {
        AddIceCandidate(uuid, data)
    }

    if (message.action === 'uuid') {
        // 设计问题
        const uuid = data
        // 有监考端需要查看
        await createPeerConnection(uuid)
        // peerConnections[uuid] = peerConnections[serveruuid]
        await streamAddTrack(uuid)
        await negotiation(uuid)
    }
}

const AddIceCandidate = (uuid, candidate) => {
    if (peerConnections[uuid]) {
        peerConnections[uuid].addIceCandidate(candidate)
    }
}

// 这些都是监考端的事情，客户端无需考虑
const handleEvent = async (data) => {
    console.log(data)
    if (data.event === 'MemberJoined') {
        console.log('member joined.')
        handleMemberJoined(data.no, data.name, data.level)
    }
    if (data.event === 'MemberLeft') {
        handleMemberLeft(data.no, data.name, data.level)
    }

    if (data.event === 'SendStreamId') {
        // 获取stream id
        id2content = data.streamid
    }
}

// 完成 sdp 交换过程，必须在 addtrack 后调用
async function negotiation(uuid) {
    // let peerConnection = peerConnections[uuid]
    try {
        let offer = await peerConnections[uuid].createOffer()
        await peerConnections[uuid].setLocalDescription(offer)

        // 发送 offer 信息
        ws.send(JSON.stringify({
            action: 'offer',
            'data': offer,
            'uuid': uuid
        }))
        console.log('offer send.')
    } catch (err) {
        console.error('negotiation error: ', err)
    }
}


async function createPeerConnection(uuid) {
    peerConnections[uuid] = new RTCPeerConnection(servers)

    peerConnections[uuid].ontrack = event => {
        console.log("track event", event)
        if (id2content.camera === event.streams[0].id) {
            console.log('camera stream track')
            event.streams[0].getTracks().forEach(track => {
                cameraStream.addTrack(track)
                // console.log('track: ' + track.kind)
            })
            document.getElementById('cameraStream').srcObject = cameraStream

        } else if (id2content.screen === event.streams[0].id) {
            console.log('screen stream track')
            event.streams[0].getTracks().forEach(track => {
                screenStream.addTrack(track)
                // console.log('track: ' + track.kind)
            })
            document.getElementById('screenStream').srcObject = screenStream
        }

        // event.streams[0].getTracks().forEach(track => {
        //     screenStream.addTrack(track)
        //     console.log('track: ' + track.kind)
        // })
    }

    peerConnections[uuid].onicecandidate = async event => {
        if (event.candidate) {
            // 发送 candidate 信息
            ws.send(JSON.stringify({
                'action': 'candidate',
                'data': event.candidate,
                'uuid': uuid
            }))
            console.log('candidate send.')
        }
    }
}

const getStream = async () => {
    if (!cameraStream) {
        await getCameraStream()
    }

    if (!screenStream) {
        await getScreenStream()
    }
}

// 为一个peerConnection添加流
async function streamAddTrack(uuid) {
    // let peerConnection = peerConnections[uuid]
    // 添加音视频流
    cameraStream.getTracks().forEach(track => {
        peerConnections[uuid].addTrack(track, cameraStream)
        console.log('add track: ', track.kind)
    })

    screenStream.getTracks().forEach(track => {
        peerConnections[uuid].addTrack(track, screenStream)
        console.log(track.kind)
    })
}

async function getCameraStream() {
    try {
        cameraStream = await navigator.mediaDevices.getUserMedia(mediaStreamConstrains)
        document.getElementById('cameraStream').srcObject = cameraStream
        document.getElementById('cameraStream').play()

        // 获取之后进行监测
        cameraStream.oninactive = async () => {
            console.log('camera inactive')
            getCameraStream()
        }

        id2content['camera'] = cameraStream.id

    } catch (err) {
        console.log('CameraStream error: ' + err)

        if (err.name == "NotFoundError" || err.name == "DevicesNotFoundError") {
            //required track is missing 
            alert('请检查设备是否存在问题！！！')
        } else if (err.name == "NotReadableError" || err.name == "TrackStartError") {
            //webcam or mic are already in use 
            alert('请解除设备占用问题！！！')
        } else if (err.name == "OverconstrainedError" || err.name == "ConstraintNotSatisfiedError") {
            //constraints can not be satisfied by avb. devices 
            alert('你的设备无法满足最低录像限制！！！')
        } else if (err.name == "NotAllowedError" || err.name == "PermissionDeniedError") {
            //permission denied in browser 
            alert('请授予权限！！！')
        } else if (err.name == "TypeError") {
            //empty constraints object 
            alert('请联系开发人员！！！')
        } else if (err.name == 'AbortError') {
            alert("Starting videoinput failed，请检查后重试")
        }
        else {
            //other errors 
            alert('你几乎不可能遇见这种错误！！！')
        }

        getCameraStream()
    }
}


// 只有本地连接服务器的流需要
async function getScreenStream() {
    try {
        let displaySurface;
        do {
            screenStream = await navigator.mediaDevices.getDisplayMedia(displayMediaOptions)
            // 必须保证共享全屏
            displaySurface = await screenStream.getVideoTracks()[0].getSettings().displaySurface
            if (displaySurface !== 'monitor') {
                alert("你必须选择全屏共享！！！")
            }
        } while (displaySurface !== 'monitor')

        document.getElementById('screenStream').srcObject = screenStream
        document.getElementById('screenStream').play()

        // 获取之后进行监测
        screenStream.oninactive = async () => {
            console.log('screen inactive')
            await getScreenStream()
        }

        /// TODO: 如何将此处提取分离
        // screenStream.getTracks().forEach(track => {
        //     peerConnection.addTrack(track, screenStream)
        //     console.log(track.kind)
        // })

        id2content['screen'] = screenStream.id

    } catch (err) {
        console.error("ScreenStream error: " + err)
        if (err.name == "NotFoundError" || err.name == "DevicesNotFoundError") {
            //required track is missing 
            alert('没有可用于捕获的屏幕视频源！！！')
        } else if (err.name == "NotReadableError" || err.name == "TrackStartError") {
            //webcam or mic are already in use 
            alert('请解除设备占用问题！！！')
        } else if (err.name == "NotAllowedError" || err.name == "PermissionDeniedError") {
            //permission denied in browser 
            alert('请授予权限！！！')
        } else if (err.name == "OverconstrainedError") {
            alert("流兼容错误！！！")
        } else if (err.name == "AbortError") {
            alert("出现错误或故障！！！")
        } else if (err.name == "TypeError") {
            //empty constraints object 
            alert('约束错误，请联系开发人员！！！')
        } else if (err.name == 'InvalidStateError') {
            alert('部分浏览器错误，需要用户触发共享屏幕')
        }
        else {
            //other errors 
            alert('你几乎不可能遇见这种错误！！！')
        }
        getScreenStream()
    }

}

const createAnswer = async (uuid, offer) => {
    // let peerConnection = peerConnections[uuid]
    await createPeerConnection(uuid)
    await peerConnections[uuid].setRemoteDescription(offer)

    let answer = await peerConnections[uuid].createAnswer()
    await peerConnections[uuid].setLocalDescription(answer)

    const json = JSON.stringify({
        'action': 'answer',
        'data': answer,
        'uuid': uuid
    })
    ws.send(json)
    console.log('answer send.')
}

const addAnswer = async (uuid, answer) => {
    // let peerConnection = peerConnections[uuid]
    // 当前如果没有远程连接，则开始建立连接
    if (!peerConnections[uuid].currentRemoteDescription) {
        try {
            await peerConnections[uuid].setRemoteDescription(answer)
            console.log('set remote description finish')
        } catch (err) {
            console.error('setRemoteDescription error: ' + err)
        }
    }
}


let leaveChannel = async () => {
    // 关闭连接
    // peerConnections.forEach(peerConnection => {
    //     peerConnection.close()
    // })
    for (const uuid in peerConnections) {
        peerConnections[uuid].close()
    }
    
    console.log("RTCPeerConnection closed!")
}

window.addEventListener('beforeunload', leaveChannel)
