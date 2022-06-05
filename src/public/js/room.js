const isSec = window.location.protocol == "https:";

let ws;
let cameraStream;
let screenStream;
// let remoteStream;
let peerConnection;

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

const InitWebSocket = url => {

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
		createAnswer(MemberId, message.offer)
	}

	if (message.type === 'answer') {
		addAnswer(message.answer)
	}

	if (message.type === 'candidate') {
		if (peerConnection) {
			peerConnection.addIceCandidate(message.candidate)
		}
	}

}
const init = async () => {
	const host = window.location.host;
	if (isSec)
		url = "wss://" + host + "/api/ws"
	else
		url = "ws://" + host + "/api/ws"

	InitWebSocket(url)

	createOffer('user')

	// 需要绑定事件
	// socket.on('MemberJoined', handleUserJoined)
	// socket.on('MemberLeft', handleUserLeft)
	// socket.on('MessageFromPeer', handleMessageFromPeer)
}


const handleUserJoined = async (MemberId) => {
	console.log('A new user joined the channel: ', MemberId)
	createOffer(MemberId)
}

const handleUserLeft = (MemberId) => {
	// document.getElementById('user-2').style.display = 'none'
	// document.getElementById('user-1').classList.remove('smallFrame')
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
}


async function createPeerConnection(MemberId) {
	peerConnection = new RTCPeerConnection(servers)

	// 获取本地流
	if (!cameraStream) {
		cameraStream = await navigator.mediaDevices.getUserMedia({
			video: true,
			audio: true,
		})
		document.getElementById('cameraStream').srcObject = cameraStream
	}

	const displayMediaOptions = {
		video: {
			frameRate: { ideal: 15 }
		}
	}
	if (!screenStream) {
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

		} catch (err) {
			console.error("ScreenStream error: " + err)
		}
		document.getElementById('screenStream').srcObject = screenStream
	}

	// 添加音视频流
	cameraStream.getTracks().forEach(track => {
		peerConnection.addTrack(track, cameraStream)
	})

	screenStream.getTracks().forEach(track => {
		peerConnection.addTrack(track, screenStream)
	})

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
				'candidate': event.candidate
			}))
		}
	}
}

const createAnswer = async (MemberId, offer) => {
	await createPeerConnection(MemberId)
	await peerConnection.setRemoteDescription(offer)

	let answer = await peerConnection.createAnswer()
	await peerConnection.setLocalDescription(answer)

	ws.send(JSON.stringify({
		'type': 'answer',
		'answer': answer
	}))
}

const addAnswer = async answer => {
	if (!peerConnection.currentRemoteDescription) {
		peerConnection.setRemoteDescription(answer)
	}
}

// const toggleCamera = async () => {
// 	let videoTrack = localStream.getTracks().find(track => track.kind === 'video')

// 	if (videoTrack.enabled) {
// 		videoTrack.enabled = false
// 		document.getElementById('camera-btn').style.backgroundColor = 'rgb(255, 80, 80)'
// 	} else {
// 		videoTrack.enabled = true
// 		document.getElementById('camera-btn').style.backgroundColor = 'rgb(179, 102, 249, .9)'
// 	}
// }

// const toggleMic = async () => {
// 	let audioTrack = localStream.getTracks().find(track => track.kind === 'audio')

// 	if (audioTrack.enabled) {
// 		audioTrack.enabled = false
// 		document.getElementById('mic-btn').style.backgroundColor = 'rgb(255, 80, 80)'
// 	} else {
// 		audioTrack.enabled = true
// 		document.getElementById('mic-btn').style.backgroundColor = 'rgb(179, 102, 249, .9)'
// 	}
// }

let leaveChannel = async () => {
	// 关闭连接
	peerConnection.close()
	console.log("RTCPeerConnection closed!")
}

window.addEventListener('beforeunload', leaveChannel)

// document.getElementById('camera-btn').addEventListener('click', toggleCamera)
// document.getElementById('mic-btn').addEventListener('click', toggleMic)

init()