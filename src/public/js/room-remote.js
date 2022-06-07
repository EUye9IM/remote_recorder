const start = () => {
    streamType = 'remote'
    userType = 'teacher'
    console.log('teacher')
	initWebSocket(url)
    // createOffer()
}

start()