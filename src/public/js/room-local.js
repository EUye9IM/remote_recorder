const isSec = window.location.protocol == "https:";
const host = window.location.host;
let url;
if (isSec)
    url = "wss://" + host + "/api/ws"
else
    url = "ws://" + host + "/api/ws"

const start = () => {
    const localPeer = new Peer(url, 'local')
    // localPeer.initWebSocket(url)
    localPeer.createOffer()
    document.getElementById('cameraStream').srcObject = localPeer.cameraStream
    document.getElementById('screenStream').srcObject = localPeer.screenStream
}

start()