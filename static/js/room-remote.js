const start = async () => {
    streamType = 'remote'
    userType = 'teacher'
	initWebSocket(url)
    await getStream(streamType)
    // createPeerConnection()
    // createOffer()
}

start()