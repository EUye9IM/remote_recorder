/*************************************************
 * @abstract 以面向过程的方式进行设计
 * @brief    
 ************************************************/

let ws;
let cameraStream;
let screenStream;
let peerConnection;
let streamType;


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

const constraints = {
    video: {
        width: { min: 640, ideal: 1920, max: 1920 },
        height: { min: 480, ideal: 1080, max: 1080 },
    },
    audio: true
}

const displayMediaOptions = {
    video: {
        frameRate: { ideal: 15 }
    }
}

const initWebSocket = url => {
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

const handleMessage = event => {
    console.log(event.data)
    const message = JSON.parse(event.data)
    if (message.type === 'event') {
        // 事件处理

    }

    if (message.type === 'offer') {
        createAnswer(MemberId, message.data)
    }

    if (message.type === 'answer') {
        addAnswer(message.data)
    }

    if (message.type === 'candidate') {
        if (peerConnection) {
            peerConnection.addIceCandidate(message.data)
        }
    }
}

const handleUserJoined = async (MemberId) => {
    console.log('A new user joined the channel: ', MemberId)
    createOffer(MemberId)
}

async function createOffer(MemberId) {
    await createPeerConnection(MemberId)
    let offer = await peerConnection.createOffer()
    await peerConnection.setLocalDescription(offer)

    // 发送 offer 信息
    ws.send(JSON.stringify({
        'type': 'offer',
        'offer': offer
    }))
    console.log('offer send.')
}


async function createPeerConnection() {
    peerConnection = new RTCPeerConnection(servers)

    if (streamType === 'local') {
        await getLocalStream()
    }
    else if (streamType === 'remote') {
        cameraStream = new MediaStream()
        screenStream = new MediaStream()
    }

    peerConnection.ontrack = envent => {
        envent.streams[0].getTracks().forEach(track => {
            remoteStream.addTrack(track)
        })
    }

    peerConnection.onicecandidate = async event => {
        if (event.candidate) {
            // 发送 candidate 信息
            ws.send(JSON.stringify({
                'type': 'candidate',
                'data': event.candidate
            }))
            console.log('candidate send.')
        }
    }
}

const getLocalStream = async () => {
    // 获取本地流
    if (!cameraStream) {
        await getCameraStream()
    }

    if (!screenStream) {
        await getScreenStream()
    }
}

async function getCameraStream() {
    try {
        cameraStream = await navigator.mediaDevices.getUserMedia({
            video: true,
            audio: false,
        })
        document.getElementById('cameraStream').srcObject = cameraStream

        // 添加音视频流
        cameraStream.getTracks().forEach(track => {
            peerConnection.addTrack(track, cameraStream)
        })

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
        } else {
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
            displaySurface = screenStream.getVideoTracks()[0].getSettings().displaySurface
            if (displaySurface !== 'monitor') {
                alert("你必须选择全屏共享！！！")
            }
        } while (displaySurface !== 'monitor')

        screenStream.getTracks().forEach(track => {
            peerConnection.addTrack(track, screenStream)
        })
    } catch (err) {
        console.error("ScreenStream error: " + err)
    }
    document.getElementById('screenStream').srcObject = screenStream
}

const createAnswer = async (offer) => {
    await createPeerConnection()
    await peerConnection.setRemoteDescription(offer)

    let answer = await peerConnection.createAnswer()
    await peerConnection.setLocalDescription(answer)

    ws.send(JSON.stringify({
        'type': 'answer',
        'data': answer
    }))
    console.log('answer send.')
}

const addAnswer = async answer => {
    if (!peerConnection.currentRemoteDescription) {
        peerConnection.setRemoteDescription(answer)
    }
}


let leaveChannel = async () => {
    // 关闭连接
    peerConnection.close()
    console.log("RTCPeerConnection closed!")
}

window.addEventListener('beforeunload', leaveChannel)
