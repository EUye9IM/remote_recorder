// 监考系统的界面显示配置

// 获取到当前在会的所有成员信息，将其显示
const getAndUpdateMembers = async () => {
    waitForSocketConnection(ws, async () => {
        let members = await getMembers()
        // 更新参会者人数
        updateMemberTotal(members.length)
        // 显示参会者成员信息
        for (let i = 0; members.length > i; i++) {
            addMemberToDom(members[i].no, members[i].name, members[i].stu_level)
        }
        // 将 student 添加到 select 选项中
        members.forEach( member => {
            if (member.stu_level == 0)
                $('#Select__Members').append($('<option>').val(`${member.no}`).text(`${member.no} ${member.name}`))
        })
    })
}

// 添加用户信息到左侧用户栏
let addMemberToDom = async (MemberId, name, level) => {
    let membersWrapper = document.getElementById('member__list')
    let memberItem = `<div class="member__wrapper" id="member__${MemberId}__wrapper">
                        <span class="green__icon"></span>
                        <p class="member_name">${MemberId} </p>
                        <p class="member_name">${name}</p>
                        <p class="member_name">level = ${level}</p>
                    </div>`
    membersWrapper.insertAdjacentHTML('beforeend', memberItem)
}

// 显示参会者人数的信息
let updateMemberTotal = async (MemberCount) => {
    let total = document.getElementById('members__count')
    total.innerText = MemberCount
}

// 成员加入房间
let handleMemberJoined = async (MemberId, name, level) => {
    console.log('A new member has joined the room:', MemberId)
    addMemberToDom(MemberId, name, level)

    let members = await getMembers()
    updateMemberTotal(members.length)

    // 信息栏显示通知
    addBotMessageToDom(`欢迎 ${MemberId} ${name} 加入房间`)

    // 加入option选项
    if (level == 0)
        $('#Select__Members').append($('<option>').val(`${MemberId}`).text(`${MemberId} ${name}`))
}
 
let handleMemberLeft = async (MemberId, name) => {
    removeMemberFromDom(MemberId)
    // 成员数减 1
    MemberCount = Number($("strong").html()) - 1
    
    updateMemberTotal(MemberCount)
    $('#Select__Members option').each(function() {
        if ( $(this).val() == `${MemberId}` ) {
            $(this).remove();
            // 移除一个即可
            return
        }
    });
}

// 成员离开，刷新左侧列表
let removeMemberFromDom = async (MemberId) => {
    let memberWrapper = document.getElementById(`member__${MemberId}__wrapper`)
    let name = memberWrapper.getElementsByClassName('member_name')[1].textContent
    addBotMessageToDom(`${MemberId} ${name} 离开.`)
    
    memberWrapper.remove()
}


// 处理信息交流，此处并无作用
let handleChannelMessage = async (messageData, MemberId) => {
    console.log('A new message was received')
    let data = JSON.parse(messageData.text)

    if(data.type === 'chat'){
        addMessageToDom(data.displayName, data.message)
    }

    if(data.type === 'user_left'){
        document.getElementById(`user-container-${data.uid}`).remove()

        if(userIdInDisplayFrame === `user-container-${uid}`){
            displayFrame.style.display = null
    
            for(let i = 0; videoFrames.length > i; i++){
                videoFrames[i].style.height = '300px'
                videoFrames[i].style.width = '300px'
            }
        }
    }
}


// 添加Bot信息（即通知信息）
let addBotMessageToDom = (botMessage) => {
    let messagesWrapper = document.getElementById('messages')

    let newMessage = `<div class="message__wrapper">
                        <div class="message__body__bot">
                            <strong class="message__author__bot">🤖 Webrtc Bot</strong>
                            <p class="message__text__bot">${botMessage}</p>
                        </div>
                    </div>`

    messagesWrapper.insertAdjacentHTML('beforeend', newMessage)

    let lastMessage = document.querySelector('#messages .message__wrapper:last-child')
    if(lastMessage){
        lastMessage.scrollIntoView()
    }
}

const joinStream = async () => {

    // 需要获取远程某位考生的信息
    MemberId = $("#Select__Members option:selected").val()
    if (MemberId == null) {
        return;
    }

    console.log(`get ${MemberId} stream`)
    await ws.send(JSON.stringify({
        'action': 'event',
        'data': {
            'event': 'GetMemberStream',
            'no': MemberId
        }
    }))

    // 接下来可以获取到远程的发来的streamid

    // cameraStream = await navigator.mediaDevices.getUserMedia(mediaStreamConstrains)
    // screenStream = await navigator.mediaDevices.getDisplayMedia(displayMediaOptions)
    // document.getElementById('cameraStream').srcObject = cameraStream
    // document.getElementById('screenStream').srcObject = screenStream
}

// 获取到当前在会的所有成员信息，需要来自服务端
const getMembers = async () => {
    let members;
    // post请求获取信息
    await $.post(
        '/api/getmembers',
        data => {
            // 获取成功
            if (data.res === 0) {
                console.log(data.msg)
                console.log(data.data)
                members = data.data
                return;
            }
            if (data.res === -1) {
                alert(data.msg)
                return;
            }
        }
    )
    return members
}


// 退出登录
$('#logout').click(async () => {
    $.post(
        '/api/logout',
        data => {
            if (data.res === 0) {
                console.log('退出登录')
                $(location).attr('href', '/login')
                return;
            }
            if (data.res === -1) {
                alert('退出登录失败')
                return;
            }
        }
    )
})

// 查看视频流开关
$('#join-stream-btn').click(async event => {
    await event.preventDefault()
    console.log("click join-btn")
    
    // 首先移除视频流原本的track
    cameraStream.getTracks().forEach( track => {
        console.log('remove track ', track)
        cameraStream.removeTrack(track)
    })
    screenStream.getTracks().forEach( track => {
        console.log('remove track ', track)
        screenStream.removeTrack(track)
    })
    for (key in peerConnections) {
        peerConnections[key].close()
    }
    
    // 清空peerConnection
    // peerConnections.clear()

    // await getStream()
    joinStream()
})

window.addEventListener('beforeunload', leaveChannel)
// let messageForm = document.getElementById('message__form')
// messageForm.addEventListener('submit', sendMessage)

const start = async () => {
    streamType = 'remote'
    userType = 'teacher'
	initWebSocket(url)

    // 存储两种流
    cameraStream = new MediaStream()
    screenStream = new MediaStream()
    
    // 显示当前的会议成员信息
    getAndUpdateMembers()
    addBotMessageToDom(`Welcome to the room! 👋`)
}

start()