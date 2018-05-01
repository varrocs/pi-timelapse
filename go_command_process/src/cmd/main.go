package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	PREVIEW_START   = iota
	PREVIEW_STARTED = iota
	PREVIEW_STOP    = iota
	PREVIEW_STOPPED = iota
	RECORDING_START = iota
	RECORDING_STOP  = iota
)

const (
	default_interval      = 60
	filename_template     = "images/image_%s.png"
	filename_last_capture = "images/lastcapture.png"
)

type state struct {
	Recording         bool
	Streaming         bool
	RecordingInterval int
}

type recordingParams struct {
	running  bool
	interval int
}

var previewChan chan int
var previewProcessManagerChan chan int
var previewProcessCallbackChan chan int
var statusRequest chan int
var statusResponse chan state
var recordingChan chan recordingParams
var photoTakingChan chan int
var uploadChan chan string
var copyChan chan string

func write_url_for(w http.ResponseWriter, item string) {
	fmt.Fprintf(w, "<a href=\"%v\">%v</a><br>", item, item)
}

func redirect_to_status(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 302)
}

func startRecording(w http.ResponseWriter, r *http.Request) {
	previewChan <- RECORDING_START
	redirect_to_status(w, r)
}

func stopRecording(w http.ResponseWriter, r *http.Request) {
	previewChan <- RECORDING_STOP
	redirect_to_status(w, r)
}

func startPreview(w http.ResponseWriter, r *http.Request) {
	previewChan <- PREVIEW_START
	redirect_to_status(w, r)
}

func stopPreview(w http.ResponseWriter, r *http.Request) {
	previewChan <- PREVIEW_STOP
	redirect_to_status(w, r)
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	statusRequest <- 1
	s := <-statusResponse
	msg, err := json.Marshal(s)
	if err != nil {
		fmt.Fprintf(w, "Failed to marshal state object")
	} else {
		fmt.Fprintf(w, "%s", msg)
	}
}

func lastCapture(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filename_last_capture)
}

func menu(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	statusRequest <- 1
	s := <-statusResponse
	msg, err := json.Marshal(s)
	if err != nil {
		fmt.Fprintf(w, "<span>--Failed to marshal state object--</span><br>")
	} else {
		fmt.Fprintf(w, "<span>%s</span><br>", msg)
	}

	write_url_for(w, "/status")
	write_url_for(w, "/start_preview")
	write_url_for(w, "/stop_preview")
	write_url_for(w, "/start_recording")
	write_url_for(w, "/stop_recording")
	write_url_for(w, "/last_capture")
	fmt.Fprint(w, "<form><input type=\"text\" name=\"interval\"><input type=\"submit\"></form><br>")
}

func TakePicture(filename string) {
}

func setupChannels() {
	previewChan = make(chan int, 5)
	previewProcessManagerChan = make(chan int, 5)
	previewProcessCallbackChan = make(chan int, 5)
	statusRequest = make(chan int, 5)
	statusResponse = make(chan state, 5)
	recordingChan = make(chan recordingParams, 5)
	photoTakingChan = make(chan int, 5)
	uploadChan = make(chan string, 5)
	copyChan = make(chan string, 5)
}

func onPreviewMessage(message int, s *state) {
	switch message {
	case PREVIEW_START:
		if s.Streaming {
			fmt.Printf("StateLoop: Was streaming, skip\n")
		} else {
			previewProcessManagerChan <- PREVIEW_START
		}
	case PREVIEW_STOP:
		if !s.Streaming {
			fmt.Printf("StateLoop: Streaming was stopped, skip\n")
		} else {
			previewProcessManagerChan <- PREVIEW_STOP
		}
	case RECORDING_START:
		if s.Recording {
			fmt.Printf("StateLoop: Was recording, skip\n")
		} else {
			recordingChan <- recordingParams{running: true, interval: default_interval}
			s.Recording = true
		}
	case RECORDING_STOP:
		if !s.Recording {
			fmt.Printf("StateLoop: Was recording stopped, skip\n")
		} else {
			recordingChan <- recordingParams{running: false, interval: default_interval}
			s.Recording = false
		}
	}
}

func stateLoop() {
	fmt.Printf("StateLoop: START\n")
	s := state{
		Recording:         false,
		Streaming:         false,
		RecordingInterval: 3,
	}
	for {
		select {
		case i := <-previewChan:
			fmt.Printf("StateLoop: Got a message: %v\n", i)
			onPreviewMessage(i, &s)

		case j := <-previewProcessCallbackChan:
			switch j {
			case PREVIEW_STOPPED:
				fmt.Printf("StateLoop: Streaming stopped\n")
				s.Streaming = false
			case PREVIEW_STARTED:
				fmt.Printf("StateLoop: Streaming started\n")
				s.Streaming = true
			}

		case <-statusRequest:
			statusResponse <- s
		}
	}
}

func previewProcessHandlerLoop(c chan int, out chan int) {
	var cmd *exec.Cmd
	fmt.Printf("PreviewProcess: START \n")
	processEndedChannel := make(chan error)
	shouldRunning := false
	running := false
	for {

		select {
		case x := <-processEndedChannel:
			fmt.Printf("PreviewProcess: process ended: %v\n", x)
			running = false
			out <- PREVIEW_STOPPED
		case startSignal := <-c:
			fmt.Printf("PreviewProcess: got command: %v\n", startSignal)
			if startSignal == PREVIEW_START {
				shouldRunning = true
			} else if startSignal == PREVIEW_STOP {
				shouldRunning = false
			}

		}
		fmt.Printf("PreviewProcess: running: %v  should running: %v\n", running, shouldRunning)

		if shouldRunning && !running {
			fmt.Printf("PreviewProcess: starting process\n")
			cmd = exec.Command("bash", "start_streaming.sh")
			if err := cmd.Start(); err != nil {
				fmt.Printf("PreviewProcess: failed to start process: %v\n", err)
				running = false
				continue
			}

			running = true
			out <- PREVIEW_STARTED

			go func() {
				exitCode := cmd.Wait()
				processEndedChannel <- exitCode
			}()
		} else if !shouldRunning && running {
			fmt.Printf("PreviewProcess: killing process\n")
			if err := cmd.Process.Kill(); err != nil {
				fmt.Printf("PreviewProcess: failed to kill process: %v", err)
			}
		}
	}
}

func recordingLoop(commandChan chan recordingParams, photoTakingChan chan int) {
	fmt.Printf("RecordingLoop: START \n")

	var recordingP recordingParams
	var timerChan <-chan time.Time = nil

	for {
		select {
		case recordingP = <-commandChan:
			fmt.Printf("RecordingLoop: Got new parameters\n")
			if timerChan == nil && recordingP.running {
				fmt.Printf("RecordingLoop: Starting new timer\n")
				timerChan = time.After(time.Duration(recordingP.interval) * time.Second)
				photoTakingChan <- 1
			}
		case <-timerChan:
			fmt.Printf("RecordingLoop: Timer ticked\n")
			if recordingP.running {
				fmt.Printf("RecordingLoop: Rescheduling timer\n")
				timerChan = time.After(time.Duration(recordingP.interval) * time.Second)
				photoTakingChan <- 1
			} else {
				fmt.Printf("RecordingLoop: Stopping timer\n")
				timerChan = nil
			}
		}
	}
}

func generateFileName() string {
	now := time.Now()
	t := now.Format("2006-01-02-15_04_05")
	filename := fmt.Sprintf(filename_template, t)
	return filename
}

func photoTakingLoop(photoTakingChan chan int, copyChan chan string) {
	fmt.Printf("PhotoTakingLoop: START \n")
	for {
		<-photoTakingChan
		filename := generateFileName()
		fmt.Printf("PhotoTakingLoop: Taking a photo %s \n", filename)
		cmd := exec.Command("bash", "take_picture.sh", filename)
		if err := cmd.Start(); err == nil {
			exit := cmd.Wait()
			if exit == nil {
				fmt.Printf("PhotoTakingLoop: Photo taken %s \n", filename)
				copyChan <- filename
			} else {
				fmt.Printf("PhotoTakingLoop: Faild to take photo %s \n", filename)
			}
		} else {
			fmt.Printf("PhotoTakingLoop: Faild to start photo taking %s - Error: %v\n", filename, err)
		}

	}
}

func copyLoop(copyChan chan string, uploadChan chan string) {
	fmt.Printf("CopyLoop: START \n")
	for {
		filename := <-copyChan

		fmt.Printf("CopyLoop: Copying photo %s to %s \n", filename, filename_last_capture)
		cmd := exec.Command("cp", filename, filename_last_capture)
		if err := cmd.Start(); err != nil {
			fmt.Printf("PhotoTakingLoop: Failed to copy last captured image to %s, error: %v \n", filename_last_capture, err)
		} else {
			uploadChan <- filename
		}
	}
}

func uploadingLoop(uploadChan chan string) {
	fmt.Printf("UploadingLoop: START \n")
	for {
		filename := <-uploadChan
		success := false
		retryCount := 5
		for retryCount > 0 && !success {
			fmt.Printf("UploadingLoop: Uploading a photo %v, attempt #%v\n", filename, (6 - retryCount))
			cmd := exec.Command("bash", "upload_file.sh", filename)
			if err := cmd.Start(); err == nil {
				exit := cmd.Wait()
				if exit == nil {
					fmt.Printf("UploadingLoop: Photo uploaded: %s\n", filename)
					success = true
					if err = os.Remove(filename); err != nil {
						fmt.Printf("UploadingLoop: Failed to delete photo: %s. It stays there\n", filename)
					}
				} else {
					fmt.Printf("UploadingLoop: Faild to upload photo %s, exit code: %v \n", filename, exit)
				}
			} else {
				fmt.Printf("UploadingLoop: Faild to start uploading %s - Error: %v\n", filename, err)
			}
			retryCount -= 1
		}
	}
}

func main() {
	setupChannels()
	go stateLoop()
	go previewProcessHandlerLoop(previewProcessManagerChan, previewProcessCallbackChan)
	go recordingLoop(recordingChan, photoTakingChan)
	go photoTakingLoop(photoTakingChan, copyChan)
	go copyLoop(copyChan, uploadChan)
	go uploadingLoop(uploadChan)

	http.HandleFunc("/start_recording", startRecording)
	http.HandleFunc("/stop_recording", stopRecording)
	http.HandleFunc("/start_preview", startPreview)
	http.HandleFunc("/stop_preview", stopPreview)
	http.HandleFunc("/status", getStatus)
	http.HandleFunc("/last_capture", lastCapture)
	http.HandleFunc("/", menu)
	http.ListenAndServe(":8080", nil)
}
