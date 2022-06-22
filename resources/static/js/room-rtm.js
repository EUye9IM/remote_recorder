// 监考系统的界面显示配置

// 获取到当前在会的所有成员信息，将其显示
const getAndUpdateMembers = async () => {
    let members = await getMembers()
    // 更新参会者人数
    updateMemberTotal(members)
    // 显示参会者成员信息
    for (let i = 0; members.length > i; i++){
        addMemberToDom(members[i].no, members[i].name)
    }
}

// 添加用户信息到左侧用户栏
let addMemberToDom = async (MemberId, name) => {
    let membersWrapper = document.getElementById('member__list')
    let memberItem = `<div class="member__wrapper" id="member__${MemberId}__wrapper">
                        <span class="green__icon"></span>
                        <p class="member_name">${name}</p>
                    </div>`
    membersWrapper.insertAdjacentHTML('beforeend', memberItem)
}

// 显示参会者人数的信息
let updateMemberTotal = async (MemberCount) => {
    let total = document.getElementById('members__count')
    total.innerText = MemberCount
}

// 成员加入房间
let handleMemberJoined = async (MemberId, name) => {
    console.log('A new member has joined the room:', MemberId)
    addMemberToDom(MemberId)

    let members = await getMembers()
    updateMemberTotal(members.length)

    // 信息栏显示通知
    addBotMessageToDom(`欢迎 ${MemberId} ${name} 加入房间`)
}
 
let handleMemberLeft = async (MemberId, name) => {
    removeMemberFromDom(MemberId)
    // 成员数减 1
    MemberCount = Number($("strong").html()) - 1
    
    updateMemberTotal(MemberCount)
}

// 成员离开，刷新左侧列表
let removeMemberFromDom = async (MemberId) => {
    let memberWrapper = document.getElementById(`member__${MemberId}__wrapper`)
    let name = memberWrapper.getElementsByClassName('member_name')[0].textContent
    addBotMessageToDom(`${name} 离开.`)
    
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


// 发送信息，目前还需要完善
let sendMessage = async (e) => {
    e.preventDefault()

    let message = e.target.message.value
    channel.sendMessage({text:JSON.stringify({'type':'chat', 'message':message, 'displayName':displayName})})
    addMessageToDom(displayName, message)
    e.target.reset()
}

let addMessageToDom = (name, message) => {
    let messagesWrapper = document.getElementById('messages')

    let newMessage = `<div class="message__wrapper">
                        <div class="message__body">
                            <strong class="message__author">${name}</strong>
                            <p class="message__text">${message}</p>
                        </div>
                    </div>`

    messagesWrapper.insertAdjacentHTML('beforeend', newMessage)

    let lastMessage = document.querySelector('#messages .message__wrapper:last-child')
    if(lastMessage){
        lastMessage.scrollIntoView()
    }
}


// 添加Bot信息（即通知信息）
let addBotMessageToDom = (botMessage) => {
    let messagesWrapper = document.getElementById('messages')

    let newMessage = `<div class="message__wrapper">
                        <div class="message__body__bot">
                            <strong class="message__author__bot">🤖 Mumble Bot</strong>
                            <p class="message__text__bot">${botMessage}</p>
                        </div>
                    </div>`

    messagesWrapper.insertAdjacentHTML('beforeend', newMessage)

    let lastMessage = document.querySelector('#messages .message__wrapper:last-child')
    if(lastMessage){
        lastMessage.scrollIntoView()
    }
}

// 获取到当前在会的所有成员信息，需要来自服务端
const getMembers = async () => {
    // post请求获取信息
    await $.post(
        '/api/getmembers',
        data => {
            // 获取成功
            if (data.res === 0) {
                console.log(data.msg)
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

window.addEventListener('beforeunload', leaveChannel)
let messageForm = document.getElementById('message__form')
messageForm.addEventListener('submit', sendMessage)