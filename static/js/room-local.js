const start = () => {
    streamType = 'local'
    userType = 'student'
	initWebSocket(url)
    createPeerConnection()

    // waitForSocketConnection(ws, () => {
    //     createOffer()
    // })

}

start()