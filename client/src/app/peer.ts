import {SignallerService} from './signaller.service';
import {Injectable} from '@angular/core';
import {Message} from './message';
import {Subject} from 'rxjs';

const icecandidate = "icecandidate"
const track = "track"


export class MediaStreams{
  constructor(public trackId: string,
              public trackKind: string,
              public stream : MediaStream,
) {
  }
}

@Injectable({
  providedIn: 'root'
})
export class Peer {
  peer : RTCPeerConnection
  private subject = new Subject()

  constructor(private signaller : SignallerService) {
    this.peer = new RTCPeerConnection()
    this.peer.addEventListener(icecandidate, (evt: RTCPeerConnectionIceEvent) => {
      this.onIceCandidateEvent(evt)
    })
    this.peer.onicecandidateerror = this.logError
    this.signaller.setOnRemoteDescription((desc: any) => {this.onDescription(desc)})
    this.signaller.setOnRemoteIcecandidate((desc: any) => {this.onIceCandidate(desc)})
    this.peer.ondatachannel = (ev: RTCDataChannelEvent) => {
      ev.channel.onmessage = (ev: MessageEvent) => {
        let enc = new TextDecoder('utf-8')
        //console.log(enc.decode(ev.data))
      }
    }
    this.peer.ontrack = (ev: RTCTrackEvent) => {
      console.log(`trackid=${ev.track.id} kind=${ev.track.kind} `)
      for (let stream of ev.streams) {
        console.log(`streamid=${stream.id}`)
      }
      let t =         {
        kind: ev.track.kind, payload: new MediaStreams(ev.track.id, ev.track.kind, ev.streams[0])}
      this.subject.next(t)
    }
  }


  getStats() {
    return this.peer.getStats(null)
  }

  listen(eventName: string, callback : (event: any) => void) {
    this.subject.asObservable().subscribe((nextObj : any) => {
      if (eventName === nextObj.kind) {
        callback(nextObj.payload)
      }
    })
  }
  private logError(evt: Event) {
    console.log("Error: ", evt)
  }

  private onIceCandidateEvent(evt: RTCPeerConnectionIceEvent) {
    if (evt.candidate == null) {
      return
    }
    this.signaller.sendLocalIceCandidate(evt.candidate)
  }

  async onDescription(desc: Message) {
    console.log("Remote description: ", desc)
    let payload = JSON.parse(desc.data)
    await this.peer.setRemoteDescription(new RTCSessionDescription(payload))
    let answer = await this.peer.createAnswer()
    await this.peer.setLocalDescription(answer)
    this.signaller.sendLocalDescription(answer)
  }

  async onIceCandidate(desc: Message) {
    console.log("Remote IceCandidate: ", desc)
    let payload = JSON.parse(desc.data)
    await this.peer.addIceCandidate(payload)
  }
}
