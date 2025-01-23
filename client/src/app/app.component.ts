import {Component, OnDestroy, OnInit} from '@angular/core';
import {WsclientService} from './wsclient.service';
import {MediaStreams, Peer} from './peer';
import {NgForOf, NgIf} from '@angular/common';
import {SignallerService} from './signaller.service';
import {VideostreamComponent} from './videostream/videostream.component';

@Component({
  selector: 'app-root',
  imports: [NgForOf, NgIf, VideostreamComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent implements OnInit, OnDestroy{
  mediaStreams : MediaStreams[] = []
  private peer: Peer
  private signaller: SignallerService
  constructor() {
    console.log(window.location.port)
    let websocketAddr = "ws://" + window.location.host + "/ws"
    //let websocketAddr = "ws://localhost:8080/ws"
    let ws : WsclientService= new WsclientService(websocketAddr)
    this.signaller = new SignallerService(ws)
    console.log(websocketAddr)
    this.peer = new Peer(this.signaller)
    this.peer.listen("video",(evt: any) => {
      let strm : MediaStreams = evt
      this.mediaStreams.push(strm)
    })
    this.peer.listen("audio",(evt: any) => {
      let strm : MediaStreams = evt
      this.mediaStreams.push(strm)
    })
  }

  ngOnDestroy(): void {
  }

  ngOnInit(): void {
  }

}
