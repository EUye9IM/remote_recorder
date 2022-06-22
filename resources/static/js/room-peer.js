/*************************************************
 * @abstract 以面向过程的方式进行设计
 * @brief    
 ************************************************/

let ws;
let cameraStream;
let screenStream;
let peerConnection;
let streamType;
let userType;

let id2content  = {};

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

const initWebSocket = (url) => {
    ws = new WebSocket(url)

    ws.onmessage = handleMessage
    ws.onopen = event => {
        console.log('Websocket open.')
    }
    ws.onerror = event => {
        console.log('Websocket error!')
    }
    ws.onclose = event => {
        console.log('Websocket closed.')
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

const handleMessage = event => {
    const message = JSON.parse(event.data)
    // console.log('message from ' + message.from)
    if (message.from === userType) {
        return
    }
    console.log(message)
    console.log(`recieve ${message.action} from ${message.from}.`)

    if (message.action === 'event') {
        // 事件处理
        handleEvent(message.data)
    }

    if (message.action === 'streamid') {
        id2content = message.data
    }

    if (message.action === 'offer') {
        createAnswer(message.data)
    }
    

    if (message.action === 'answer') {
        addAnswer(message.data)
    }

    if (message.action === 'candidate') {
        if (peerConnection) {
            peerConnection.addIceCandidate(message.data)
        }
    }
}

const handleEvent = async (data) => {
    if (data.event === 'MemberJoined') {
        handleUserJoined(data.no, data.name)
    }
    if (data.event === 'MemberLeft') {
        handleUserLeft(data.no, data.name)
    }
}

// 完成 sdp 交换过程，必须在 addtrack 后调用
async function negotiation() {
    try {
        let offer = await peerConnection.createOffer()
        await peerConnection.setLocalDescription(offer)

        // 发送 offer 信息
        ws.send(JSON.stringify({
            action: 'offer',
            'data': offer,
            'from': userType
        }))
        console.log('offer send.')
    } catch (err) {
        console.error('negotiation error: ', err)
    }
}


async function createPeerConnection() {
    peerConnection = new RTCPeerConnection(servers)

    peerConnection.ontrack = event => {
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

    peerConnection.onicecandidate = async event => {
        if (event.candidate) {
            // 发送 candidate 信息
            ws.send(JSON.stringify({
                'action': 'candidate',
                'data': event.candidate,
                'from': userType
            }))
            console.log('candidate send.')
        }
    }
}

const getStream = async (__streamType) => {
    if (__streamType === 'local') {
        if (!cameraStream) {
            await getCameraStream()
        }

        if (!screenStream) {
            await getScreenStream()
        }
        // await getLocalStream()
    }
    if (__streamType === 'remote') {
        cameraStream = new MediaStream()
        screenStream = new MediaStream()
    }

}


async function getCameraStream() {
    try {
        cameraStream = await navigator.mediaDevices.getUserMedia(mediaStreamConstrains)
        document.getElementById('cameraStream').srcObject = cameraStream

        // 获取之后进行监测
        cameraStream.oninactive = async () => {
            console.log('camera inactive')
            getCameraStream()
        }

        // 添加音视频流
        cameraStream.getTracks().forEach(track => {
            peerConnection.addTrack(track, cameraStream)
            console.log('add track: ', track.kind)
        })

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

        // 获取之后进行监测
        screenStream.oninactive = async () => {
            console.log('screen inactive')
            await getScreenStream()
        }
        screenStream.getTracks().forEach(track => {
            peerConnection.addTrack(track, screenStream)
            console.log(track.kind)
        })

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

const createAnswer = async (offer) => {
    await createPeerConnection()
    await peerConnection.setRemoteDescription(offer)

    let answer = await peerConnection.createAnswer()
    await peerConnection.setLocalDescription(answer)

    const json = JSON.stringify({
        'action': 'answer',
        'data': answer,
        'from': userType
    })
    ws.send(json)
    console.log('answer send.')
}

const addAnswer = async answer => {
    // if (!peerConnection.currentRemoteDescription) {
    try {
        await peerConnection.setRemoteDescription(answer)
        console.log('set remote description finish')
    } catch (err) {
        console.error('setRemoteDescription error: ' + err)
    }
    // }
}


let leaveChannel = async () => {
    // 关闭连接
    peerConnection.close()
    console.log("RTCPeerConnection closed!")
    // 移除用户登录信息
    sessionStorage.removeItem("user")
}

window.addEventListener('beforeunload', leaveChannel)
// window.onunload = () => {
//     peerConnection.close()
//     console.log('broswer refresh')
// }