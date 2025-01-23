import { Injectable } from '@angular/core';
import {Observable, Subject} from 'rxjs';
import {webSocket, WebSocketSubject} from 'rxjs/webSocket';

export class WsclientService {

  private socket: WebSocketSubject<any>;
  constructor(url: string) {
    this.socket = webSocket(url)
  }

  send(msg : any) {
    this.socket.next(msg)
  }

  getMessages(): Observable<any> {

    return this.socket.asObservable();
  }
  closeConnection() {
    this.socket.complete()
  }
}
