package gstreamer

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0
#include "gstreamer_wrapper.h"
*/
import "C"
import (
	"errors"
	"log"
	"pion-webrtc/internal/dto"
	"time"
	"unsafe"
)

//export onBusMessage
func onBusMessage(msgType *C.char, msg *C.char, id C.int) {
	log.Printf("BusMessage(%d): %s(%s)", id, C.GoString(msgType), C.GoString(msg))
}

//export onNewFrame
func onNewFrame(frame unsafe.Pointer, size C.int, duration C.int, _ C.int) {
	frameBytes := C.GoBytes(frame, size) //nolint:nlreturn

	pipeline.ch <- dto.VideoFrame{
		Frame:    frameBytes,
		Duration: time.Duration(duration),
	}
}

type GstVideo struct {
	ch          chan dto.VideoFrame
	strPipeline string
	pipe        unsafe.Pointer
}

var pipeline *GstVideo //nolint:gochecknoglobals
var errPipeline = errors.New("error creating pipeline")

func NewGstVideo(pipelineStr string, ch chan dto.VideoFrame) *GstVideo {
	C.gstreamer_init()

	pipeline = &GstVideo{
		ch:          ch,
		strPipeline: pipelineStr,
		pipe:        nil,
	}

	return pipeline
}

func (gvid *GstVideo) Initialize() error {
	pipelineCString := C.CString(gvid.strPipeline)

	defer C.free(unsafe.Pointer(pipelineCString)) //nolint:nlreturn

	gvid.pipe = C.gstreamer_prepare_pipelines(pipelineCString, 1)
	if gvid.pipe == nil {
		return errPipeline
	}

	return nil
}

func (gvid *GstVideo) Run() {
	go func() {
		C.gstreamer_start_main_loop(gvid.pipe)
	}()
	C.gstreamer_start_pipeline(gvid.pipe)
}

func (gvid *GstVideo) Stop() {
	C.gstreamer_stop_pipeline(gvid.pipe)
	C.gstreamer_stop_main_loop(gvid.pipe)
}

func (gvid *GstVideo) Dispose() {
	C.gstreamer_deinit()
}
