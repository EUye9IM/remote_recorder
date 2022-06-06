const start = async () => {
    streamType = 'local'
	initWebSocket(url)
    createOffer()
}

start()