const start = () => {
    streamType = 'local'
	initWebSocket(url)

    waitForSocketConnection(ws, createOffer)

    // createOffer()
}

start()