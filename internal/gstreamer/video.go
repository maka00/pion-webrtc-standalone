package gstreamer

import (
	"fmt"
	"log"
	"pion-webrtc/internal/dto"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
)

type Video struct {
	mainLoop    *glib.MainLoop
	ch          chan dto.VideoFrame
	strPipeline string
}

func NewVideo(pipeline string, ch chan dto.VideoFrame) *Video {
	gst.Init(nil)

	return &Video{
		mainLoop:    glib.NewMainLoop(glib.MainContextDefault(), false),
		ch:          ch,
		strPipeline: pipeline,
	}
}

func (vid *Video) Run() {
	pipeline, err := gst.NewPipelineFromString(vid.strPipeline)
	if err != nil {
		log.Fatalf("Error creating pipeline: %v", err)
	}

	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() { //nolint:exhaustive
		case gst.MessageEOS: // When end-of-stream is received flush the pipeling and stop the main loop
			err := pipeline.BlockSetState(gst.StateNull)
			if err != nil {
				return false
			}

			vid.mainLoop.Quit()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			log.Printf("ERROR: %s", err.Error())

			if debug := err.DebugString(); debug != "" {
				fmt.Printf("DEBUG: %s", debug)
			}

			vid.mainLoop.Quit()
		default:
			// All messages implement a Stringer. However, this is
			// typically an expensive thing to do and should be avoided.
			fmt.Println(msg)
		}

		return true
	})

	elem, err := pipeline.GetElementByName("sink")
	if err != nil {
		log.Fatalf("no sink element")
	}

	sink := app.SinkFromElement(elem)
	sink.SetCallbacks(&app.SinkCallbacks{ //nolint:exhaustruct
		NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
			sample := sink.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}
			buffer := sample.GetBuffer()
			defer buffer.Unmap()

			memory := buffer.Map(gst.MapRead).Bytes()
			vid.OnNewFrame(memory, buffer.Duration())

			return gst.FlowOK
		},
	})

	if err := pipeline.SetState(gst.StatePlaying); err != nil {
		log.Fatalf("Error starting pipeline: %s", err.Error())
	}

	vid.mainLoop.Run()
}

func (vid *Video) OnNewFrame(frame []byte, time gst.ClockTime) {
	vid.ch <- dto.VideoFrame{Frame: frame, Duration: *time.AsDuration()}
}
