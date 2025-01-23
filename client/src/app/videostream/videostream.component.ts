import { Component, OnInit, Input } from '@angular/core';
import {MediaStreams} from '../peer';

@Component({
  selector: 'app-videostream',
  imports: [],
  templateUrl: './videostream.component.html',
  styleUrl: './videostream.component.scss'
})
export class VideostreamComponent implements OnInit{
  @Input() streamMetaData! : MediaStreams
  public width: number = 0
  public height : number = 0
  ngOnInit(): void {
  }

  constructor() {
  }

  onVideoClick() {
    console.log("clicked on: ", this.streamMetaData.stream.id)
  }
  onMetadata(ev: Event) {
    let video = ev.target as HTMLVideoElement
    this.height = video.videoHeight
    this.width = video.videoWidth
  }
}
