package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/media/control"
)

var debug = flag.Bool("debug", false, "enable debug logging")

func init() {
	flag.Parse()
	if !*debug {
		log.SetOutput(io.Discard)
	}
}

func main() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize COM and WinRT
	log.Printf("Initializing COM/WinRT...")
	if err := ole.RoInitialize(1); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize COM/WinRT: %v\n", err)
		os.Exit(1)
	}

	// Get session manager
	log.Printf("Getting global system media transport controls session manager...")
	op, err := control.GlobalSystemMediaTransportControlsSessionManagerRequestAsync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get global system media manager: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Awaiting global system media transport controls session manager async operation...")
	if err := awaitAsyncOperation(op, control.SignatureGlobalSystemMediaTransportControlsSessionManager); err != nil {
		fmt.Fprintf(os.Stderr, "failed to await async operation: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Getting results from async operation...")
	sessionManagerRes, err := op.GetResults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get results from async operation: %v\n", err)
		os.Exit(1)
	}

	if uintptr(sessionManagerRes) == 0 {
		fmt.Fprintf(os.Stderr, "no media session available\n")
		os.Exit(1)
	}

	sessionManager := (*control.GlobalSystemMediaTransportControlsSessionManager)(sessionManagerRes)

	// Access current session
	log.Printf("Getting current media session...")
	session, err := sessionManager.GetCurrentSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get current session: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Getting media properties asynchronously...")
	mediaPropsOp, err := session.TryGetMediaPropertiesAsync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get media properties async: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Awaiting media properties async operation...")
	if err := awaitAsyncOperation(mediaPropsOp, control.SignatureGlobalSystemMediaTransportControlsSessionMediaProperties); err != nil {
		fmt.Fprintf(os.Stderr, "failed to await media properties async operation: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Getting results from media properties async operation...")
	mediaPropsRes, err := mediaPropsOp.GetResults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get results from media properties async operation: %v\n", err)
		os.Exit(1)
	}

	mediaProps := (*control.GlobalSystemMediaTransportControlsSessionMediaProperties)(mediaPropsRes)
	log.Printf("Retrieving artist and title from media properties...")
	artist, err := mediaProps.GetArtist()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get artist from media properties: %v\n", err)
		os.Exit(1)
	}

	title, err := mediaProps.GetTitle()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get title from media properties: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Now playing: %s - %s\n", artist, title)

	// Control playback
	log.Printf("Getting playback info...")
	playbackInfo, err := session.GetPlaybackInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get playback info: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Getting playback status...")
	status, err := playbackInfo.GetPlaybackStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get playback status: %v\n", err)
		os.Exit(1)
	}

	var statusStr string
	switch status {
	case control.GlobalSystemMediaTransportControlsSessionPlaybackStatusClosed:
		statusStr = "Closed"
	case control.GlobalSystemMediaTransportControlsSessionPlaybackStatusOpened:
		statusStr = "Opened"
	case control.GlobalSystemMediaTransportControlsSessionPlaybackStatusChanging:
		statusStr = "Changing"
	case control.GlobalSystemMediaTransportControlsSessionPlaybackStatusStopped:
		statusStr = "Stopped"
	case control.GlobalSystemMediaTransportControlsSessionPlaybackStatusPlaying:
		statusStr = "Playing"
	case control.GlobalSystemMediaTransportControlsSessionPlaybackStatusPaused:
		statusStr = "Paused"
	default:
		statusStr = "Unknown"
	}

	fmt.Printf("Current playback status before toggle: %s\n", statusStr)

	log.Printf("Trying to toggle play/pause async...")
	toggleOp, err := session.TryTogglePlayPauseAsync()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get toggle play/pause async operation: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Awaiting toggle play/pause async operation...")
	if err := awaitAsyncOperation(toggleOp, winrt.SignatureBool); err != nil {
		fmt.Fprintf(os.Stderr, "failed to await toggle play/pause async operation: %v\n", err)
		os.Exit(1)
	}

	log.Printf("Play/pause toggled successfully.")
}

func awaitAsyncOperation(asyncOperation *foundation.IAsyncOperation, genericParamSignature string) error {
	var status foundation.AsyncStatus

	// We need to obtain the GUID of the AsyncOperationCompletedHandler, but its a generic delegate
	// so we also need the generic parameter type's signature:
	// AsyncOperationCompletedHandler<genericParamSignature>
	iid := winrt.ParameterizedInstanceGUID(foundation.GUIDAsyncOperationCompletedHandler, genericParamSignature)

	// Wait until the async operation completes.
	waitChan := make(chan struct{})
	handler := foundation.NewAsyncOperationCompletedHandler(ole.NewGUID(iid), func(instance *foundation.AsyncOperationCompletedHandler, asyncInfo *foundation.IAsyncOperation, asyncStatus foundation.AsyncStatus) {
		status = asyncStatus
		close(waitChan)
	})
	defer handler.Release()

	asyncOperation.SetCompleted(handler)

	// Wait until async operation has stopped, and finish.
	<-waitChan

	if status != foundation.AsyncStatusCompleted {
		return fmt.Errorf("async operation failed with status %d", status)
	}
	return nil
}
