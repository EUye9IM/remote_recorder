/** Peer */

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
let URL

class Peer {
    ws;                     // websocket
    peerConnection;         // RtcPeerConnction
    cameraStream;
    screenStream;
    streamType;             // local and remote

    constructor(url, __streamType) {
        this.streamType = __streamType
        // this.init(url)
        URL = url
    }

    init(url) {
        this.initWebSocket(url)
        console.log(3)
        // this.createOffer('user')

        // 需要绑定事件
    }

    initWebSocket = url => {
        this.ws = new WebSocket(url)

        this.ws.onmessage = this.handleMessage
        this.ws.onopen = event => {
            console.log('Websocket open.')
        }
        this.ws.onerror = event => {
            console.log('Websocket error!')
        }
        this.ws.onclose = event => {
            console.log('Websocket closed.')
        }
        console.log(2)
    }

    handleMessage = event => {
        const message = JSON.parse(event.data)
        console.log('Recieve ' + message.type)

        if (message.type === 'event') {
            // 事件处理

        }

        if (message.type === 'offer') {
            this.createAnswer(message.offer)
        }

        if (message.type === 'answer') {
            this.addAnswer(message.answer)
        }

        if (message.type === 'candidate') {
            if (this.peerConnection) {
                this.peerConnection.addIceCandidate(message.candidate)
            }
        }

    }

    async createOffer() {
        console.log(4)
        await this.createPeerConnection()
        let offer = await this.peerConnection.createOffer()
        await this.peerConnection.setLocalDescription(offer)
        
        this.initWebSocket(URL)
        this.ws.send(JSON.stringify({
            'type': 'offer',
            'data': offer
        }), () => console.log('offer send.'))
    }

    async addAnswer(answer) {
        if (!this.peerConnection.currentRemoteDescription) {
            this.peerConnection.setRemoteDescription(answer)
        }
    }

    async createAnswer(offer) {
        await this.createPeerConnection()
        await this.peerConnection.setRemoteDescription(offer)
    
        let answer = await this.peerConnection.createAnswer()
        await this.peerConnection.setLocalDescription(answer)
    
        this.ws.send(JSON.stringify({
            'type': 'answer',
            'data': answer
        }), () => console.log('answer send.'))
    }

    async createPeerConnection() {
        this.peerConnection = new RTCPeerConnection(servers)
        if (this.streamType === 'local') {
            this.getLocalStream()
        }

        else if (this.streamType === 'remote') {
            this.cameraStream = new MediaStream()
            this.screenStream = new MediaStream()
        }

        // track 事件需要其他设置
        this.peerConnection.ontrack = envent => {
            envent.streams[0].getTracks().forEach(track => {
                // remoteStream.addTrack(track)
            })
        }
    
        this.peerConnection.onicecandidate = async event => {
            if (event.candidate) {
                // 发送 candidate 信息
                ws.send(JSON.stringify({
                    'type': 'candidate',
                    'candidate': event.candidate
                }), () => console.log('candidate send!'))
            }
        }

    }

    async getLocalStream() {
        // 获取本地流
        if (!this.cameraStream) {
            this.cameraStream = await navigator.mediaDevices.getUserMedia({
                video: true,
                audio: true,
            })
            // document.getElementById('cameraStream').srcObject = this.cameraStream
        }


        if (!this.screenStream) {
            try {
                let displaySurface;
                do {
                    this.screenStream = await navigator.mediaDevices.getDisplayMedia(displayMediaOptions)
                    // 必须保证共享全屏
                    displaySurface = screenStream.getVideoTracks()[0].getSettings().displaySurface
                    console.log('displaySurface: ' + displaySurface)
                    if (displaySurface !== 'monitor') {
                        alert("你必须选择全屏共享！！！")
                    }
                } while (displaySurface !== 'monitor')

            } catch (err) {
                console.error("ScreenStream error: " + err)
            }
            // document.getElementById('screenStream').srcObject = this.screenStream
        }

        // 添加音视频流
        this.cameraStream.getTracks().forEach(track => {
            this.peerConnection.addTrack(track, this.cameraStream)
        })

        this.screenStream.getTracks().forEach(track => {
            this.peerConnection.addTrack(track, this.screenStream)
        })
    }


}
