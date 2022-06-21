// 监考系统的界面显示配置

// 获取到当前在会的所有成员信息，需要来自服务端
const getMembers = async () => {
    let members = [
        {
            no: '1234567',
            name: '张三',
            stu_enable: 0
        }
    ]
    return members
}

const getUserAttributesByKeys = async (MemberId) => {
    const user = {
        no: '1234567',
        name: '张三',
        stu_enable: 0
    }
    return user 
}

// 获取到当前在会的所有成员信息，将其显示
const getAndUpdateMembers = async () => {
    let members = await getMembers()
    // 更新参会者人数
    updateMemberTotal(members)
    for (let i = 0; members.length > i; i++){
        addMemberToDom(members[i])
    }
}

// 成员加入房间
let handleMemberJoined = async (MemberId) => {
    console.log('A new member has joined the room:', MemberId)
    addMemberToDom(MemberId)

    let members = await getMembers()
    updateMemberTotal(members)

    let {name} = await getUserAttributesByKeys(MemberId)
    
    // 信息栏显示通知
    addBotMessageToDom(`欢迎 ${name} 加入房间`)
}

let addMemberToDom = async (MemberId) => {
    let {name} = await getUserAttributesByKeys(MemberId)

    let membersWrapper = document.getElementById('member__list')
    let memberItem = `<div class="member__wrapper" id="member__${MemberId}__wrapper">
                        <span class="green__icon"></span>
                        <p class="member_name">${name}</p>
                    </div>`

    membersWrapper.insertAdjacentHTML('beforeend', memberItem)
}

let updateMemberTotal = async (members) => {
    let total = document.getElementById('members__count')
    total.innerText = members.length
}
 
let handleMemberLeft = async (MemberId) => {
    removeMemberFromDom(MemberId)

    let members = await getMembers()
    updateMemberTotal(members)
}

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

let leaveChannel = async () => {
    await channel.leave()
    await rtmClient.logout()
}

window.addEventListener('beforeunload', leaveChannel)
let messageForm = document.getElementById('message__form')
messageForm.addEventListener('submit', sendMessage)