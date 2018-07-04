package gst

/*
#cgo pkg-config: gstreamer-1.0 gstreamer-app-1.0

#include "gst.h"

*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/pions/webrtc"
)

func init() {
	go C.gstreamer_send_start_mainloop()
}

// Pipeline is a wrapper for a GStreamer Pipeline
type Pipeline struct {
	Pipeline *C.GstElement
	in       chan<- webrtc.RTCSample
	id       int
}

var pipelines = make(map[int]*Pipeline)
var pipelinesLock sync.Mutex

// CreatePipeline creates a GStreamer Pipeline
func CreatePipeline(codec webrtc.TrackType, in chan<- webrtc.RTCSample) *Pipeline {
	pipelineStr := "appsink name=appsink"
	switch codec {
	case webrtc.VP8:
		pipelineStr = "videotestsrc ! vp8enc ! " + pipelineStr
	case webrtc.VP9:
		pipelineStr = "videotestsrc ! vp9enc ! " + pipelineStr
	case webrtc.Opus:
		pipelineStr = "audiotestsrc ! opusenc ! " + pipelineStr
	default:
		panic("Unhandled codec " + codec.String())
	}

	pipelineStrUnsafe := C.CString(pipelineStr)
	defer C.free(unsafe.Pointer(pipelineStrUnsafe))

	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	pipeline := &Pipeline{
		Pipeline: C.gstreamer_send_create_pipeline(pipelineStrUnsafe),
		in:       in,
		id:       len(pipelines),
	}

	pipelines[pipeline.id] = pipeline
	return pipeline
}

// Start starts the GStreamer Pipeline
func (p *Pipeline) Start() {
	C.gstreamer_send_start_pipeline(p.Pipeline, C.int(p.id))
}

// Stop stops the GStreamer Pipeline
func (p *Pipeline) Stop() {
	C.gstreamer_send_stop_pipeline(p.Pipeline)
}

//export goHandlePipelineBuffer
func goHandlePipelineBuffer(buffer unsafe.Pointer, bufferLen C.int, samples C.int, pipelineId C.int) {
	pipelinesLock.Lock()
	defer pipelinesLock.Unlock()

	if pipeline, ok := pipelines[int(pipelineId)]; ok {
		pipeline.in <- webrtc.RTCSample{C.GoBytes(buffer, bufferLen), uint32(samples)}
	} else {
		fmt.Printf("discarding buffer, no pipeline with id %d", int(pipelineId))
	}
	C.free(buffer)
}