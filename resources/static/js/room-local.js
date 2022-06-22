const start = () => {
    streamType = 'local'
    userType = 'student'
	initWebSocket(url)
    createPeerConnection()
}

start()