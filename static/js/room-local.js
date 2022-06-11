const start = () => {
    streamType = 'local'
    userType = 'student'
	initWebSocket(url)

    waitForSocketConnection(ws, () => {
        createOffer()
    })

}

start()