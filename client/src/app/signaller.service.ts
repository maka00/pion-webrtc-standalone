import { Injectable } from '@angular/core';
import {WsclientService} from './wsclient.service';
import {Message} from './message';

const icecandidate = "iceCandidate"
const description = "sessionDescription"
const track = "track"
@Injectable({
  providedIn: 'root'
})
export class SignallerService {

  private onRemoteDescriptionCallback :(desc:any) => void
  private onRemoteIceCandidateCallback :(desc:any) => void

  constructor(private ws : WsclientService) {
    this.onRemoteDescriptionCallback = (desc: any) => {console.log(`remoteDescription not set: ${desc}`)}
    this.onRemoteIceCandidateCallback = (desc: any) => {console.log(`remoteDescription not set: ${desc}`)}
    this.ws.getMessages().subscribe((msg: any) => {
      let message: Message = msg
      switch (message.type) {
        case icecandidate:
          this.onRemoteIceCandidate(msg)
          break
        case description:
          this.onRemoteDescription(msg)
          break;
        default:
          console.log("got unknown type: ", message.type)
      }
    })
  }

  setOnRemoteDescription(callback :(desc:any) => void) {
    this.onRemoteDescriptionCallback = callback
  }
  setOnRemoteIcecandidate(callback :(desc:any) => void) {
    this.onRemoteIceCandidateCallback = callback
  }

  onRemoteDescription(desc: any) {
    this.onRemoteDescriptionCallback(desc)
  }

  onRemoteIceCandidate(ice: any) {
    this.onRemoteIceCandidateCallback(ice)
  }

  sendLocalDescription(desc: any) {
    this.sendMessage(description, desc)
  }

  sendLocalIceCandidate(candidate: any) {
    this.sendMessage(icecandidate, candidate)
  }

  sendMessage(type: string, data: any) {
    console.log(`sending: ${type} with ${data}`)
    const message = new Message(type, data)
    const messageString = JSON.stringify(message.data)
    this.ws.send(new Message(type, messageString))
  }
}
