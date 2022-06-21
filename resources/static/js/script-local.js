/********************************
 * 设置按钮的事件等配置
 */

// 拦截直接输入url的非法访问
if (document.referrer === '') {
    alert('非法访问页面，请先登录！！！')
    window.location = '/login'
}

const userId = sessionStorage.getItem('userId')
if (userId === null) {
    window.location = '/login'
}

// $('#userId').attr('placeholder', userId)


// 下面处理按钮事件
let btnSystem = document.getElementById('btnSystem')
let switchCamera = document.getElementById('switchCamera')
let switchCameraAudio = document.getElementById('switchCameraAudio')
let switchScreen = document.getElementById('switchScreen')
let switchScreenAudio = document.getElementById('switchScreenAudio')

btnSystem.onclick = async () => {
    btnSystem.disabled = true
    
    if (mediaStreamConstrains.video) {
        switchCamera.checked = true
        switchCamera.disabled = false
    }

    if (mediaStreamConstrains.audio) {
        switchCameraAudio.checked = true
        switchCameraAudio.disabled = false
    }

    if (displayMediaOptions.video) {
        switchScreen.checked = true
        switchScreen.disabled = false
    }

    if (displayMediaOptions.audio) {
        switchScreen.checked = true
        switchScreenAudio.disabled = false
    }

    waitForSocketConnection(ws, async () => {
        await getStream(streamType)
        await ws.send(JSON.stringify({
            'action': 'streamid',
            'data': id2content,
            'from': userType
        }))
        await negotiation()
    })

}

async function toggleChange(toggle, tracks, msg)
{
    // console.log(tracks)
    if (toggle.checked) {
        if (tracks.length) {
            tracks.forEach(track => {
                track.enabled = true
            });
        } else {
            alert('系统不要求开启' + msg)
        }
    } else {
        tracks.forEach(track => {
            track.enabled = false
        });
    }
}

switchCamera.onchange = async () => {
    await toggleChange(switchCamera, cameraStream.getVideoTracks(), '摄像头')
}

switchCameraAudio.onchange = async () => {
    await toggleChange(switchCameraAudio, cameraStream.getAudioTracks(), '麦克风音频')
}

switchScreen.onchange = async () => {
    await toggleChange(switchScreen, screenStream.getVideoTracks(), '共享屏幕')
}

switchScreenAudio.onchange = async () => {
    await toggleChange(switchScreenAudio, screenStream.getAudioTracks(), '电脑音频')
}

