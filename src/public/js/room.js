const socket = io()

console.log("进入房间")
socket.emit('echo', 'echo data')

let localStream;
let remoteStream;
let peerConnection;

const servers = {
	iceServers: null,
}

let constraints = {
	video: {
		width: { min: 640, ideal: 1920, max: 1920 },
		height: { min: 480, ideal: 1080, max: 1080 },
	},
	audio: true
}

let init = async () => {
	socket.on('MemberJoined', handleUserJoined)
	socket.on('MemberLeft', handleUserLeft)
	socket.on('MessageFromPeer', handleMessageFromPeer)

}


let handleUserJoined = async (MemberId) => {
	console.log('A new user joined the channel: ', MemberId)
	createOffer(MemberId)
}

let handleUserLeft = (MemberId) => {
    document.getElementById('user-2').style.display = 'none'
    document.getElementById('user-1').classList.remove('smallFrame')
}

async function createOffer(MemberId) {
	await createPeerConnection(MemberId)
	let offer = await peerConnection.createOffer()
	await peerConnection.setLocalDescription(offer)

	// 发送 offer 信息
	socket.emit('message', {
		'type': 'offer',
		'offer': offer,
	}, MemberId)
}


async function createPeerConnection(MemberId) {
	peerConnection = new RTCPeerConnection(servers)
	remoteStream = new MediaStream()

	document.getElementById('user-2').srcObject = remoteStream
	document.getElementById('user-2').style.display = 'block'

	document.getElementById('user-1').classList.add('smallFrame')

	// 获取本地流
	if (!localStream) {
		localStream = await navigator.mediaDevices.getUserMedia({
			video: true,
			audio: true,
		})
		document.getElementById('user-1').srcObject = localStream
	}

	localStream.getTracks().forEach(track => {
		peerConnection.addTrack(track, localStream)
	})

	peerConnection.ontrack = envent => {
		envent.streams[0].getTracks().forEach(track => {
			remoteStream.addTrack(track)
		})
	}

	peerConnection.onicecandidate = async event => {
		if (event.candidate) {
			// 发送 candidate 信息
			socket.emit('MessageToPeer', {
				'type': 'candidate',
				'candidate': event.candidate,
			}, MemberId)

		}
	}
}

let handleMessageFromPeer = async (message, MemberId) => {
	message = JSON.parse(message.text)

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

let createAnswer = async (MemberId, offer) => {
	await createPeerConnection(MemberId)
	await peerConnection.setRemoteDescription(offer)

	let answer = await peerConnection.createAnswer()
	await peerConnection.setLocalDescription(answer)

	socket.emit('MessageToPeer', {
		'type': 'answer',
		'answer': answer,
	}, MemberId)
}

let addAnswer = async answer => {
	if (!peerConnection.currentRemoteDescription) {
		peerConnection.setRemoteDescription(answer)
	}
}

let toggleCamera = async () => {
	let videoTrack = localStream.getTracks().find(track => track.kind === 'video')

	if (videoTrack.enabled) {
		videoTrack.enabled = false
		document.getElementById('camera-btn').style.backgroundColor = 'rgb(255, 80, 80)'
	} else {
		videoTrack.enabled = true
		document.getElementById('camera-btn').style.backgroundColor = 'rgb(179, 102, 249, .9)'
	}
}

let toggleMic = async () => {
	let audioTrack = localStream.getTracks().find(track => track.kind === 'audio')

	if (audioTrack.enabled) {
		audioTrack.enabled = false
		document.getElementById('mic-btn').style.backgroundColor = 'rgb(255, 80, 80)'
	} else {
		audioTrack.enabled = true
		document.getElementById('mic-btn').style.backgroundColor = 'rgb(179, 102, 249, .9)'
	}
}

let leaveChannel = async () => {
	// 关闭连接
	peerConnection.close()
	console.log("RTCPeerConnection closed!")
}

window.addEventListener('beforeunload', leaveChannel)

document.getElementById('camera-btn').addEventListener('click', toggleCamera)
document.getElementById('mic-btn').addEventListener('click', toggleMic)

init()