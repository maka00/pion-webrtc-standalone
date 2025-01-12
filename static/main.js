class Peer {
    constructor() {
        const configuration = {};
        this.peer = new RTCPeerConnection(configuration);
        this.peer.addEventListener('icecandidate', (event) => this.iceCandidate(event));
        this.peer.addEventListener('track', (evt) => {
            console.log("**got track: " + evt.track.kind)
            const elem = document.createElement(evt.track.kind)
            elem.srcObject = evt.streams[0]
            elem.style.width = '100%'
            elem.autoplay = true
            elem.controls = true
            elem.muted = true
            document.getElementById('video').appendChild(elem)
        } )
        this.peer.onicecandidateerror = (event) => { console.error("onicecandidateerror", event) }
        this.peer.ondatachannel = (event) => {
            event.channel.onmessage = (event) => {
                console.log("onmessage", event)
                var enc = new TextDecoder("utf-8")
                let data = enc.decode(event.data)
                console.log("datachannel: ", enc.decode(event.data))
            }
            console.log("ondatachannel", event)
        }
        this.onIceCandidate = function() {console.warn("onIceCandidate not overwritten!")}
        this.onSessionDescription = function() {console.warn("onSessionDescription not overwritten!")}
    }

    iceCandidate(event) {
        if (event.candidate == null) {
            return
        }
        this.onIceCandidate(event.candidate)
    }

    async answer() {
        let answer = await this.peer.createAnswer()
        await this.peer.setLocalDescription(answer)
        this.onSessionDescription(answer)
    }

    async setSessionDescription(description) {
        await this.peer.setRemoteDescription(new RTCSessionDescription(description))
        if (description.type === "offer") {
            await this.answer()
        }
    }

    async addIceCandidate(candidate) {
        let line = JSON.parse(candidate)
        await this.peer.addIceCandidate(line)
    }
}

class Signaler {
    constructor() {
        this.ws = new WebSocket(location.origin.replace(/^http/, 'ws') + '/ws');
        this.ws.onopen = function (openEvent) {
            console.log("websocket connection opened");
        }
        this.ws.onmessage = (messageEvent) => {this.onMessage(messageEvent)}
        this.ws.onclose = function (closeEvent) {
            console.log("websocket connection closed");
        }
        this.onSessionDescription = function() {console.warn("onSessionDescription not overwritten!")}
        this.onIceCandidate = function() {console.warn("onIceCandidate not overwritten!")}
    }
    onMessage(messageEvent) {
        console.log(messageEvent.data);
        let msg = JSON.parse(messageEvent.data)
        if (msg.type === "sessionDescription") {
            console.log("got a sessionDescription")
            let payload = JSON.parse(msg.data)
            this.onSessionDescription(payload)
        } else if (msg.type === "iceCandidate") {
            console.log("got an iceCandidate")
            let payload = msg.data
            this.onIceCandidate(payload)
        }
    };
    sendMsg(type, data) {
        const message = {
            type: type,
            data: data
        }
        const messageAsString = JSON.stringify(message.data)
        const coverMessage = {
            type: type,
            data: messageAsString
        }
        this.ws.send(JSON.stringify(coverMessage));
    }
}
const signaler = new Signaler();
const peer = new Peer();
peer.onIceCandidate = function(candidate) {
    signaler.sendMsg("iceCandidate", candidate)
}
peer.onSessionDescription = function(description) {
    signaler.sendMsg("sessionDescription", description)
}

signaler.onSessionDescription =  function(description) {
    peer.setSessionDescription(description)
}
signaler.onIceCandidate = function(candidate) {
    peer.addIceCandidate(candidate)
}


