// Package config provides a configuration struct and methods for loading and saving it.
package speak

import (
	"fmt"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func Play(filePath string) error {

	tempPath := filePath + ".temp.mp3"
	// Replace the original file with the temp file if it exists
	if _, err := os.Stat(tempPath); err == nil {
		fmt.Println("Replacing original file with temp file")
		err := os.Remove(filePath)
		if err != nil {
			return err
		}
		err = os.Rename(tempPath, filePath)
		if err != nil {
			return err
		}
		err = os.Remove(tempPath)
		if err != nil {
			return err
		}
	}

	// Replace the original file with the temp file if it exists
	if _, err := os.Stat(tempPath); err == nil {
		os.Remove(filePath)
		err := os.Rename(tempPath, filePath)
		if err != nil {
			return err
		}
		err = os.Remove(tempPath)
		if err != nil {
			return err
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return err
	}
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done

	return nil
}
