package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/eliukblau/pixterm/pkg/ansimage"
	fcolor "github.com/fatih/color"
	"golang.org/x/term"
)

const LOGO = `
  _   _ _   _ _   _   _____                           _
 | | | | | | | | | | /  __ \                         | |
 | |_| | | | | |_| | | /  \/ ___  _ ____   _____ _ __| |_ ___ _ __
 |  _  | | | |  _  | | |    / _ \| '_ \ \ / / _ \ '__| __/ _ \ '__|
 | | | | |_| | | | | | \__/\ (_) | | | \ V /  __/ |  | ||  __/ |
 \_| |_/\___/\_| |_|  \____/\___/|_|_|\_/ \___|_|   \__\___|_|
`

func printFancyHeader() {
	fcolor.Cyan(LOGO)
	fmt.Println("\n   🖼️  Universal Image Converter & Viewer  🖼️\n")
}

func printInfo(message string) {
	fcolor.Blue("ℹ️ %s\n", message)
}

func printSuccess(message string) {
	fcolor.Green("✅ %s\n", message)
}

func printError(message string) {
	fcolor.Red("❌ %s\n", message)
}

func printProgress(progress float32) {
	const barWidth = 50
	filled := int(progress * float32(barWidth))
	empty := barWidth - filled

	bar := fcolor.GreenString(strings.Repeat("█", filled)) + strings.Repeat(" ", empty)
	fmt.Printf("\r[%s] %.1f%%", bar, progress*100)
	if progress == 1.0 {
		fmt.Println()
	}
}

func imageToHuh(imagePath, huhPath string) error {
	printInfo(fmt.Sprintf("Converting %s to %s", imagePath, huhPath))

	file, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	outFile, err := os.Create(huhPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	bounds := img.Bounds()
	width, height := uint32(bounds.Max.X), uint32(bounds.Max.Y)
	totalPixels := int(width * height)

	if err := binary.Write(outFile, binary.LittleEndian, width); err != nil {
		return err
	}
	if err := binary.Write(outFile, binary.LittleEndian, height); err != nil {
		return err
	}

	pixelCount := 0
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixelData := []byte{byte(r >> 8), byte(g >> 8), byte(b >> 8)}
			if _, err := outFile.Write(pixelData); err != nil {
				return err
			}

			pixelCount++
			if pixelCount%(totalPixels/100+1) == 0 {
				printProgress(float32(pixelCount) / float32(totalPixels))
			}
		}
	}

	printProgress(1.0)
	printSuccess(fmt.Sprintf("Successfully converted %s to %s", imagePath, huhPath))
	return nil
}

func huhToImage(huhPath, imagePath string) error {
	printInfo(fmt.Sprintf("Converting %s to %s", huhPath, imagePath))

	file, err := os.Open(huhPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var width, height uint32

	if err := binary.Read(file, binary.LittleEndian, &width); err != nil {
		return errors.New("invalid HUH file: could not read width")
	}
	if err := binary.Read(file, binary.LittleEndian, &height); err != nil {
		return errors.New("invalid HUH file: could not read height")
	}

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	totalPixels := int(width * height)

	pixelBuffer := make([]byte, 3)
	for i := 0; i < totalPixels; i++ {
		if _, err := io.ReadFull(file, pixelBuffer); err != nil {
			return fmt.Errorf("invalid HUH file: unexpected EOF at pixel %d", i)
		}
		x, y := i%int(width), i/int(width)
		img.Set(x, y, color.RGBA{R: pixelBuffer[0], G: pixelBuffer[1], B: pixelBuffer[2], A: 255})

		if i%(totalPixels/100+1) == 0 {
			printProgress(float32(i) / float32(totalPixels))
		}
	}

	printProgress(1.0)

	outFile, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".png":
		err = png.Encode(outFile, img)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90})
	case ".gif":
		err = gif.Encode(outFile, img, &gif.Options{NumColors: 256})
	default:
		return fmt.Errorf("unsupported output format: %s", ext)
	}

	if err != nil {
		return err
	}

	printSuccess(fmt.Sprintf("Successfully converted %s to %s", huhPath, imagePath))
	return nil
}

func convertImage(inputPath, outputPath string) error {
	printInfo(fmt.Sprintf("Converting %s to %s", inputPath, outputPath))

	inputFile, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	img, _, err := image.Decode(inputFile)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".png":
		err = png.Encode(outputFile, img)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 90})
	case ".gif":
		err = gif.Encode(outputFile, img, &gif.Options{NumColors: 256})
	default:
		return fmt.Errorf("unsupported output format: %s", ext)
	}

	if err != nil {
		return err
	}

	printSuccess(fmt.Sprintf("Successfully converted %s to %s", inputPath, outputPath))
	return nil
}

func viewImage(path string) error {
	printInfo(fmt.Sprintf("Viewing image: %s", path))

	var img image.Image
	var err error

	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".huh" {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		var width, height uint32
		binary.Read(file, binary.LittleEndian, &width)
		binary.Read(file, binary.LittleEndian, &height)

		rgbaImg := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		pixelBuffer := make([]byte, 3)
		totalPixels := int(width * height)

		for i := 0; i < totalPixels; i++ {
			io.ReadFull(file, pixelBuffer)
			x, y := i%int(width), i/int(width)
			rgbaImg.Set(x, y, color.RGBA{R: pixelBuffer[0], G: pixelBuffer[1], B: pixelBuffer[2], A: 255})
		}
		img = rgbaImg
	} else {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		img, _, err = image.Decode(file)
		if err != nil {
			return err
		}
	}

	w, h, _ := term.GetSize(int(os.Stdout.Fd()))

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return err
	}

	ansImg, err := ansimage.NewScaledFromReader(
		bytes.NewReader(buf.Bytes()),
		w,
		h,
		color.Transparent,
		ansimage.ScaleModeFit,
		0,
	)
	if err != nil {
		return err
	}
	ansImg.Draw()

	fmt.Println("\nPress 'q' to exit viewer...")

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	reader := bufio.NewReader(os.Stdin)
	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return err
		}
		if char == 'q' || char == 'Q' || char == 3 { 
			break
		}
	}

	return nil
}

func printUsage() {
	printFancyHeader()
	fmt.Println("Usage:")
	fmt.Println("  huh convert <input_file> <output_file>  - Convert between image formats and HUH")
	fmt.Println("  huh view <file>                        - View an image or HUH file")
	fmt.Println("  huh help                               - Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  huh convert image.png image.huh         - Convert PNG to HUH")
	fmt.Println("  huh convert image.huh image.jpg         - Convert HUH to JPG")
	fmt.Println("  huh convert image.png image.jpg         - Convert PNG to JPG")
	fmt.Println("  huh view image.png                      - View a PNG image")
	fmt.Println("  huh view image.huh                      - View a HUH file")
}

func main() {
	args := os.Args

	if len(args) < 2 {
		printUsage()
		return
	}

	command := args[1]
	var err error

	switch command {
	case "convert":
		if len(args) != 4 {
			printError("Invalid number of arguments for convert command")
			printUsage()
			return
		}
		inputPath := args[2]
		outputPath := args[3]

		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			printError(fmt.Sprintf("Input file does not exist: %s", inputPath))
			return
		}

		inputExt := strings.ToLower(filepath.Ext(inputPath))
		outputExt := strings.ToLower(filepath.Ext(outputPath))

		if inputExt == ".huh" && outputExt != ".huh" {
			err = huhToImage(inputPath, outputPath)
		} else if inputExt != ".huh" && outputExt == ".huh" {
			err = imageToHuh(inputPath, outputPath)
		} else if inputExt != ".huh" && outputExt != ".huh" {
			err = convertImage(inputPath, outputPath)
		} else {
			printError("Cannot convert from HUH to HUH")
		}

	case "view":
		if len(args) != 3 {
			printError("Invalid number of arguments for view command")
			printUsage()
			return
		}
		filePath := args[2]

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			printError(fmt.Sprintf("File does not exist: %s", filePath))
			return
		}

		err = viewImage(filePath)

	case "help", "--help", "-h":
		printUsage()

	default:
		printError(fmt.Sprintf("Unknown command: %s", command))
		printUsage()
	}

	if err != nil {
		printError(fmt.Sprintf("An error occurred: %v", err))
		os.Exit(1)
	}
}

