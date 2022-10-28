package main

import (
	"bytes"
	"fmt"
	"github.com/dhowden/tag"
	"github.com/jessevdk/go-flags"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var opts struct {
	Noop      bool   `short:"n" long:"noop" description:"Run, but do not rename. Useful for testing when combined with verbose"`
	Verbose   bool   `short:"v" long:"verbose" description:"Show verbose debug information"`
	Directory string `short:"d" long:"directory" description:"Base directory containing the audio files" required:"true"`
	Recursive bool   `short:"r" long:"recursive" description:"Recursively target subdirectories"`
}

func main() {
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		fmt.Println("Could not parse arguments")
		os.Exit(1)
	}

	err = Run(opts.Directory)
	if err != nil {
		fmt.Println("Could not read directory")
		os.Exit(1)
	}
}

func Run(directory string) error {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}
	for _, file := range files {
		filePath := filepath.Join(directory, file.Name())
		if file.IsDir() {
			if opts.Recursive {
				Run(filePath)
			}
			continue
		}

		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Warning: cannot file read %s\n", filePath)
		}
		reader := bytes.NewReader(fileBytes)
		metadata, err := tag.ReadFrom(reader)
		if err != nil {
			// Assume file is not an audio file
			continue
		}

		title := metadata.Title()

		if title == "" {
			if opts.Verbose {
				fmt.Printf("Could not rename file %s. No title\n", filePath)
			}
			continue
		}

		extension := filepath.Ext(filePath)
		title = escapeChars(title)
		title = title + extension
		newPath := filepath.Join(directory, title)

		if opts.Verbose {
			fmt.Printf("Renaming \"%s\" to \"%s\"\n", filePath, newPath)
		}

		if opts.Noop {
			continue
		}

		err = os.Rename(filePath, newPath)
		if err != nil {
			fmt.Printf("Warning: could not rename %s\n!", filePath)
			fmt.Println(err.Error())
		}
	}
	return nil
}

func escapeChars(filename string) string {
	switch runtime.GOOS {
	default:
		fallthrough
	case "windows":
		return escapeWindowsCharacters(filename)
	case "linux":
		return escapeLinuxChars(filename)
	case "darwin":
		return escapeWindowsCharacters(filename)
	}
}

func escapeWindowsCharacters(filename string) string {
	filename = strings.ReplaceAll(filename, "\\", "-")
	filename = strings.ReplaceAll(filename, "/", "-")
	filename = strings.ReplaceAll(filename, ":", " -")
	filename = strings.ReplaceAll(filename, "*", "-")
	filename = strings.ReplaceAll(filename, "?", "-")
	filename = strings.ReplaceAll(filename, "\"", "-")
	filename = strings.ReplaceAll(filename, "<", "")
	filename = strings.ReplaceAll(filename, ">", "")
	filename = strings.ReplaceAll(filename, "|", "-")

	return filename
}

func escapeLinuxChars(filename string) string {
	return strings.ReplaceAll(filename, "/", " ")
}

func escapeMacChars(filename string) string {
	filename = strings.ReplaceAll(filename, ":", " -")
	filename = strings.ReplaceAll(filename, "/", " ")

	return filename
}
