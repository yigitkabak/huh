package main

import (
	"bufio"
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
const (
	HUH_MAGIC   = "HUH!"
	HUH_VERSION = 2
	UPLOADS_DIR = "uploads"
)

type Metadata map[string]string

var (
	imageToHuhFunc func(image.Image, Metadata, string) error
	huhToImageFunc func(string) (image.Image, Metadata, error)
)

func printFancyHeader() {
	fcolor.Cyan(LOGO)
	fmt.Println("\n   Universal Image Converter & Viewer v2 (Enhanced) \n")
}

func printInfo(message string) {
	fcolor.Blue("INFO: %s\n", message)
}

func printSuccess(message string) {
	fcolor.Green("SUCCESS: %s\n", message)
}

func printError(message string) {
	fcolor.Red("ERROR: %s\n", message)
}

func printProgress(progress float32) {
	const barWidth = 50
	filled := int(progress * float32(barWidth))
	empty := barWidth - filled
	bar := fcolor.GreenString(strings.Repeat("█", filled)) + strings.Repeat(" ", empty)
	fmt.Printf("\r[%s] %.1f%%", bar, progress*100)
	if progress >= 1.0 {
		fmt.Println()
	}
}

func imageToHuh(img image.Image, metadata Metadata, huhPath string) error {
	outFile, err := os.Create(huhPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if _, err := outFile.WriteString(HUH_MAGIC); err != nil {
		return err
	}
	if err := binary.Write(outFile, binary.LittleEndian, uint8(HUH_VERSION)); err != nil {
		return err
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	if err := binary.Write(outFile, binary.LittleEndian, uint32(len(metadataJSON))); err != nil {
		return err
	}
	if _, err := outFile.Write(metadataJSON); err != nil {
		return err
	}

	bounds := img.Bounds()
	width, height := uint32(bounds.Max.X), uint32(bounds.Max.Y)
	if err := binary.Write(outFile, binary.LittleEndian, width); err != nil {
		return err
	}
	if err := binary.Write(outFile, binary.LittleEndian, height); err != nil {
		return err
	}

	compressor, err := flate.NewWriter(outFile, flate.BestCompression)
	if err != nil {
		return err
	}
	defer compressor.Close()

	totalPixels := int(width * height)
	pixelCount := 0
	for y := 0; y < int(height); y++ {
		for x := 0; x < int(width); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixelData := []byte{byte(r >> 8), byte(g >> 8), byte(b >> 8)}
			if _, err := compressor.Write(pixelData); err != nil {
				return err
			}
			pixelCount++
			if totalPixels > 100 && pixelCount%(totalPixels/100+1) == 0 {
				printProgress(float32(pixelCount) / float32(totalPixels))
			}
		}
	}
	if totalPixels > 100 {
		printProgress(1.0)
	}
	return nil
}

func huhToImage(huhPath string) (image.Image, Metadata, error) {
	file, err := os.Open(huhPath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	magic := make([]byte, 4)
	if _, err := io.ReadFull(file, magic); err != nil || string(magic) != HUH_MAGIC {
		return nil, nil, errors.New("invalid HUH file: bad magic number")
	}

	var version uint8
	if err := binary.Read(file, binary.LittleEndian, &version); err != nil || version != HUH_VERSION {
		return nil, nil, fmt.Errorf("unsupported HUH version: %d", version)
	}

	var metaLen uint32
	if err := binary.Read(file, binary.LittleEndian, &metaLen); err != nil {
		return nil, nil, err
	}

	metaJSON := make([]byte, metaLen)
	if _, err := io.ReadFull(file, metaJSON); err != nil {
		return nil, nil, err
	}
	var metadata Metadata
	if err := json.Unmarshal(metaJSON, &metadata); err != nil {
		return nil, nil, err
	}

	var width, height uint32
	if err := binary.Read(file, binary.LittleEndian, &width); err != nil {
		return nil, nil, err
	}
	if err := binary.Read(file, binary.LittleEndian, &height); err != nil {
		return nil, nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	decompressor := flate.NewReader(file)
	defer decompressor.Close()

	totalPixels := int(width * height)
	pixelBuffer := make([]byte, totalPixels*3)
	if _, err := io.ReadFull(decompressor, pixelBuffer); err != nil {
		return nil, nil, fmt.Errorf("failed to decompress pixel data: %w", err)
	}

	for i := 0; i < totalPixels; i++ {
		x, y := i%int(width), i/int(width)
		offset := i * 3
		img.Set(x, y, color.RGBA{R: pixelBuffer[offset], G: pixelBuffer[offset+1], B: pixelBuffer[offset+2], A: 255})
		if totalPixels > 100 && i%(totalPixels/100+1) == 0 {
			printProgress(float32(i) / float32(totalPixels))
		}
	}
	if totalPixels > 100 {
		printProgress(1.0)
	}

	return img, metadata, nil
}

func convertImage(inputPath, outputPath string) error {
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
	return err
}

func viewImage(path string) error {
	var img image.Image
	var err error
	var meta Metadata

	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".huh" {
		printInfo("Decoding HUH file...")
		img, meta, err = huhToImageFunc(path)
		if err != nil {
			return err
		}
		printInfo("Displaying HUHv2 Image. Metadata:")
		for k, v := range meta {
			fmt.Printf("  - %s: %s\n", k, v)
		}
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
	if err := png.Encode(buf, img); err != nil {
		return err
	}
	ansImg, err := ansimage.NewScaledFromReader(
		bytes.NewReader(buf.Bytes()), w, h, color.Transparent, ansimage.ScaleModeFit, 0,
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

const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>HUH Camera & Gallery</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-slate-100 text-slate-800">
    <div class="container mx-auto p-4">
        
        <header class="text-center mb-6">
            <h1 class="text-3xl font-bold text-slate-900">HUH Camera & Gallery</h1>
            <p class="text-slate-600 mt-1">Capture images and save them in the .huh format.</p>
        </header>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            
            <div class="bg-white p-4 rounded-lg shadow-md">
                <h2 class="text-xl font-semibold mb-3 border-b pb-2">Camera</h2>
                <video id="video" class="w-full h-auto bg-slate-200 rounded-md" autoplay playsinline></video>
                <canvas id="canvas" class="hidden"></canvas>
                <div class="mt-3 text-center">
                    <button id="snap" class="bg-blue-600 text-white font-bold py-2 px-4 rounded-lg hover:bg-blue-700 transition-colors duration-200 disabled:bg-slate-400">
                        Fotoğraf Çek
                    </button>
                </div>
                <div id="status" class="mt-3 text-center font-medium h-5 text-sm"></div>
            </div>

            <div class="bg-white p-4 rounded-lg shadow-md flex flex-col">
                <h2 class="text-xl font-semibold mb-3 border-b pb-2">Gallery</h2>
                
                <div class="mb-4 border-b pb-4">
                    <h3 class="text-lg font-medium mb-2">HUH Dosyası Yükle</h3>
                    <form id="uploadForm" class="flex items-center gap-3">
                        <input type="file" id="fileInput" name="huhfile" accept=".huh" required class="block w-full text-sm text-slate-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-blue-50 file:text-blue-700 hover:file:bg-blue-100"/>
                        <button type="submit" id="uploadButton" class="bg-green-600 text-white font-bold py-2 px-4 rounded-lg hover:bg-green-700 transition-colors duration-200 disabled:bg-slate-400">Yükle</button>
                    </form>
                    <div id="uploadStatus" class="mt-2 text-center font-medium h-5 text-sm"></div>
                </div>

                <div id="gallery" class="grid grid-cols-2 sm:grid-cols-3 gap-3 overflow-y-auto flex-grow pr-2">
                    <p id="gallery-placeholder" class="col-span-full text-slate-500">Loading...</p>
                </div>
            </div>

        </div>
    </div>

    <script>
        const video = document.getElementById('video');
        const canvas = document.getElementById('canvas');
        const snap = document.getElementById('snap');
        const statusDiv = document.getElementById('status');
        const gallery = document.getElementById('gallery');
        const galleryPlaceholder = document.getElementById('gallery-placeholder');
        const context = canvas.getContext('2d');
        
        const uploadForm = document.getElementById('uploadForm');
        const fileInput = document.getElementById('fileInput');
        const uploadButton = document.getElementById('uploadButton');
        const uploadStatus = document.getElementById('uploadStatus');


        navigator.mediaDevices.getUserMedia({ video: true, audio: false })
            .then(stream => {
                video.srcObject = stream;
                video.play();
            })
            .catch(err => {
                console.error("Camera access error:", err);
                statusDiv.innerHTML = '<span class="text-red-600">Kamera erişimi reddedildi.</span>';
                snap.disabled = true;
            });

        snap.addEventListener("click", () => {
            statusDiv.innerHTML = '<span class="text-orange-500">İşleniyor...</span>';
            snap.disabled = true;

            canvas.width = video.videoWidth;
            canvas.height = video.videoHeight;
            context.drawImage(video, 0, 0, canvas.width, canvas.height);
            
            const dataURL = canvas.toDataURL('image/png');
            
            fetch('/api/upload', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ image: dataURL, author: 'WebApp User' })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    statusDiv.innerHTML = '<span class="text-green-600">Başarılı: ' + data.filename + '</span>';
                    addImageToGallery(data.filename, true);
                } else {
                    statusDiv.innerHTML = '<span class="text-red-600">Hata: ' + data.error + '</span>';
                }
            })
            .catch(err => {
                statusDiv.innerHTML = '<span class="text-red-600">Sunucu bağlantı hatası.</span>';
                console.error("Upload error:", err);
            })
            .finally(() => {
                snap.disabled = false;
            });
        });
        
        uploadForm.addEventListener('submit', (event) => {
            event.preventDefault();
            
            if (fileInput.files.length === 0) {
                uploadStatus.innerHTML = '<span class="text-red-600">Lütfen bir .huh dosyası seçin.</span>';
                return;
            }

            uploadStatus.innerHTML = '<span class="text-orange-500">Yükleniyor...</span>';
            uploadButton.disabled = true;

            const formData = new FormData();
            formData.append('huhfile', fileInput.files[0]);

            fetch('/api/upload-file', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    uploadStatus.innerHTML = '<span class="text-green-600">Dosya başarıyla yüklendi.</span>';
                    uploadForm.reset();
                    loadGallery(); 
                } else {
                    uploadStatus.innerHTML = '<span class="text-red-600">Hata: ' + data.error + '</span>';
                }
            })
            .catch(err => {
                uploadStatus.innerHTML = '<span class="text-red-600">Sunucu bağlantı hatası.</span>';
                console.error("File upload error:", err);
            })
            .finally(() => {
                uploadButton.disabled = false;
            });
        });

        const addImageToGallery = (filename, prepend = false) => {
            if (galleryPlaceholder) {
                galleryPlaceholder.style.display = 'none';
            }

            const container = document.createElement('div');
            container.className = 'relative group';

            const img = document.createElement('img');
            img.src = '/view/' + filename;
            img.alt = filename;
            img.className = 'w-full h-auto object-cover rounded-md shadow-sm aspect-square';

            const caption = document.createElement('div');
            caption.className = 'absolute bottom-0 left-0 right-0 bg-black bg-opacity-50 text-white text-xs text-center p-1 rounded-b-md truncate';
            caption.textContent = filename;

            container.appendChild(img);
            container.appendChild(caption);

            if (prepend) {
                gallery.prepend(container);
            } else {
                gallery.appendChild(container);
            }
        };

        const loadGallery = async () => {
            try {
                const response = await fetch('/api/images');
                const images = await response.json();
                
                gallery.innerHTML = ''; 

                if (images && images.length > 0) {
                    images.forEach(filename => addImageToGallery(filename));
                } else {
                    gallery.innerHTML = '<p id="gallery-placeholder" class="col-span-full text-slate-500">Galeride hiç resim yok.</p>';
                }
            } catch (error) {
                console.error('Failed to load gallery:', error);
                gallery.innerHTML = '<p id="gallery-placeholder" class="col-span-full text-slate-500">Galeri yüklenemedi.</p>';
            }
        };

        document.addEventListener('DOMContentLoaded', loadGallery);
    </script>
</body>
</html>
`

type UploadRequest struct {
	Image  string `json:"image"`
	Author string `json:"author"`
}

func ensureUploadsDir() {
	if _, err := os.Stat(UPLOADS_DIR); os.IsNotExist(err) {
		os.Mkdir(UPLOADS_DIR, 0755)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	b64data := req.Image[strings.IndexByte(req.Image, ',')+1:]
	imgReader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64data))
	img, _, err := image.Decode(imgReader)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid image data"})
		return
	}

	filename := fmt.Sprintf("capture-%d.huh", time.Now().UnixNano())
	outputPath := filepath.Join(UPLOADS_DIR, filename)

	metadata := Metadata{
		"author":        req.Author,
		"creation_date": time.Now().Format(time.RFC3339),
		"source":        "WebApp Camera API",
	}

	err = imageToHuhFunc(img, metadata, outputPath)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Error saving HUH file: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to save HUH file"})
		return
	}

	log.Printf("Successfully saved %s", outputPath)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "filename": filename})
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "File size exceeds limit"})
		return
	}

	file, handler, err := r.FormFile("huhfile")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid file upload request"})
		return
	}
	defer file.Close()

	// Security: Sanitize filename and check extension
	sanitizedFilename := filepath.Base(handler.Filename)
	if strings.ToLower(filepath.Ext(sanitizedFilename)) != ".huh" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Only .huh files are allowed"})
		return
	}

	dstPath := filepath.Join(UPLOADS_DIR, sanitizedFilename)
	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("Error creating destination file: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Could not save file on server"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error copying uploaded file: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to copy file data"})
		return
	}

	log.Printf("Successfully uploaded and saved %s", dstPath)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleViewImage(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/view/")
	if filename == "" {
		http.Error(w, "Filename not provided", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(UPLOADS_DIR, filename)
	img, _, err := huhToImageFunc(filePath)
	if err != nil {
		log.Printf("Failed to decode HUH file %s: %v", filename, err)
		http.Error(w, "Could not process image file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, img); err != nil {
		log.Printf("Failed to encode image to PNG for %s: %v", filename, err)
		http.Error(w, "Could not serve image", http.StatusInternalServerError)
	}
}

func handleListImages(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(UPLOADS_DIR)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]string{})
			return
		}
		http.Error(w, "Could not read image directory", http.StatusInternalServerError)
		return
	}

	var huhFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".huh") {
			huhFiles = append(huhFiles, file.Name())
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(huhFiles)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(huhFiles)
}

func startServer() {
	ensureUploadsDir()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, indexHTML)
	})
	http.HandleFunc("/api/upload", handleUpload)
	http.HandleFunc("/api/upload-file", handleFileUpload) // YENİ EKLENDİ
	http.HandleFunc("/view/", handleViewImage)
	http.HandleFunc("/api/images", handleListImages)

	port := "8080"
	printInfo(fmt.Sprintf("Starting web server on http://localhost:%s", port))
	printSuccess("Navigate to this address in your browser to use the camera capture & gallery.")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func printUsage() {
	printFancyHeader()
	fmt.Println("Usage:")
	fmt.Println("  huh convert <input_file> <output_file>  - Convert between image formats and HUH")
	fmt.Println("  huh view <file>                        - View an image or HUH file in the terminal")
	fmt.Println("  huh serve                              - Start the web API server for camera capture and gallery")
	fmt.Println("  huh help                               - Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  huh convert image.png image.huh")
	fmt.Println("  huh convert image.huh image.jpg")
	fmt.Println("  huh view image.huh")
	fmt.Println("  huh serve")
}

func main() {
	imageToHuhFunc = imageToHuh
	huhToImageFunc = huhToImage

	args := os.Args
	if len(args) < 2 {
		printUsage()
		return
	}

	command := args[1]
	var err error

	imageToHuhServer := func(img image.Image, metadata Metadata, huhPath string) error {
		oldStdout := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w
		defer func() {
			w.Close()
			os.Stdout = oldStdout
		}()
		return imageToHuh(img, metadata, huhPath)
	}
	huhToImageServer := func(huhPath string) (image.Image, Metadata, error) {
		oldStdout := os.Stdout
		_, w, _ := os.Pipe()
		os.Stdout = w
		defer func() {
			w.Close()
			os.Stdout = oldStdout
		}()
		return huhToImage(huhPath)
	}

	switch command {
	case "convert":
		if len(args) != 4 {
			printError("Invalid number of arguments for convert command")
			printUsage()
			return
		}
		inputPath, outputPath := args[2], args[3]
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			printError(fmt.Sprintf("Input file does not exist: %s", inputPath))
			return
		}

		inputExt := strings.ToLower(filepath.Ext(inputPath))
		outputExt := strings.ToLower(filepath.Ext(outputPath))

		printInfo(fmt.Sprintf("Converting %s to %s", inputPath, outputPath))

		if inputExt == ".huh" && outputExt != ".huh" {
			var img image.Image
			img, _, err = huhToImageFunc(inputPath)
			if err == nil {
				var outFile *os.File
				outFile, err = os.Create(outputPath)
				if err == nil {
					defer outFile.Close()
					switch outputExt {
					case ".png":
						err = png.Encode(outFile, img)
					case ".jpg", ".jpeg":
						err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90})
					case ".gif":
						err = gif.Encode(outFile, img, &gif.Options{NumColors: 256})
					default:
						err = fmt.Errorf("unsupported output format: %s", outputExt)
					}
				}
			}
		} else if inputExt != ".huh" && outputExt == ".huh" {
			var file *os.File
			file, err = os.Open(inputPath)
			if err == nil {
				defer file.Close()
				var img image.Image
				img, _, err = image.Decode(file)
				if err == nil {
					metadata := Metadata{"source_file": filepath.Base(inputPath)}
					err = imageToHuhFunc(img, metadata, outputPath)
				}
			}
		} else if inputExt != ".huh" && outputExt != ".huh" {
			err = convertImage(inputPath, outputPath)
		} else {
			err = errors.New("cannot convert from HUH to HUH")
		}

		if err == nil {
			printSuccess(fmt.Sprintf("Successfully converted %s to %s", inputPath, outputPath))
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
		printInfo(fmt.Sprintf("Viewing: %s", filePath))
		err = viewImage(filePath)

	case "serve":
		huhToImageFunc = huhToImageServer
		imageToHuhFunc = imageToHuhServer
		startServer()

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
