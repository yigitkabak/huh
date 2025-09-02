# HUH Universal Image Converter & Viewer

A powerful command-line tool and web application for converting, viewing, and managing images with support for a custom HUH format that includes metadata and compression.

## Features

- **Multi-format Support**: Convert between PNG, JPEG, GIF, and the custom HUH format
- **Terminal Image Viewer**: View images directly in your terminal with high-quality rendering
- **Web Interface**: Camera capture and gallery management through a modern web interface
- **Metadata Storage**: Embed and preserve metadata in the HUH format
- **Compression**: Built-in compression for efficient storage
- **Cross-platform**: Works on Linux, macOS, and Windows

## Installation

### Prerequisites

- Go 1.19 or later
- Terminal with true color support (for image viewing)

### Build from Source

1. Clone the repository:
```bash
git clone <repository-url>
cd huh-converter
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the binary:
```bash
go build -o huh main.go
```

4. Install system-wide (optional):
```bash
sudo mv huh /usr/local/bin/
```

### Automated Installation (Linux)

Use the provided installation script:

```bash
chmod +x install.sh
./install.sh
```

The script automatically detects your Linux distribution and installs Go if needed.

## Usage

### Command Line Interface

#### Convert Images

Convert between different image formats:

```bash
# Convert PNG to HUH format
huh convert image.png image.huh

# Convert HUH back to JPEG
huh convert image.huh image.jpg

# Convert between standard formats
huh convert image.png image.gif
```

#### View Images

Display images in your terminal:

```bash
# View a standard image file
huh view photo.jpg

# View a HUH file with metadata
huh view capture.huh
```

#### Web Server

Start the web interface for camera capture and gallery:

```bash
huh serve
```

Access the web interface at `http://localhost:8080`

#### Help

Display usage information:

```bash
huh help
```

### Web Interface Features

The web server provides:

- **Camera Capture**: Take photos directly from your webcam
- **Automatic HUH Conversion**: Captured images are automatically saved in HUH format
- **File Upload**: Upload existing HUH files to the gallery
- **Gallery View**: Browse and view all stored images
- **Metadata Display**: View embedded metadata for HUH files

## HUH Format Specification

The HUH format is a custom image format designed for efficient storage and metadata preservation.

### File Structure

```
Header:
- Magic Number: "HUH!" (4 bytes)
- Version: 2 (1 byte)
- Metadata Length: uint32 (4 bytes)
- Metadata: JSON string (variable length)
- Width: uint32 (4 bytes)
- Height: uint32 (4 bytes)

Image Data:
- Compressed RGB pixel data using DEFLATE compression
```

### Metadata

HUH files can store arbitrary metadata as JSON, including:
- Author information
- Creation date
- Source application
- Custom tags and descriptions

## API Reference

### Web API Endpoints

#### POST /api/upload
Upload image data from camera capture.

**Request Body:**
```json
{
  "image": "data:image/png;base64,iVBORw0KGgoAAAANS...",
  "author": "User Name"
}
```

**Response:**
```json
{
  "success": true,
  "filename": "capture-1234567890.huh"
}
```

#### POST /api/upload-file
Upload HUH file directly.

**Request:** Multipart form data with `huhfile` field

**Response:**
```json
{
  "success": true
}
```

#### GET /api/images
List all available HUH files.

**Response:**
```json
["capture-1.huh", "capture-2.huh", "photo.huh"]
```

#### GET /view/{filename}
Serve HUH file as PNG image.

**Response:** PNG image data

## Dependencies

### Go Modules

- `github.com/eliukblau/pixterm/pkg/ansimage` - Terminal image rendering
- `github.com/fatih/color` - Colored terminal output
- `golang.org/x/term` - Terminal control

### System Requirements

- Terminal with 256-color or true color support
- Camera access (for web interface capture)
- Modern web browser with getUserMedia support

## File Storage

- Uploaded and captured images are stored in the `uploads/` directory
- Files are automatically organized and secured
- Only HUH format files are accepted for upload

## Security Features

- File extension validation
- Filename sanitization
- Upload size limits (10MB)
- Path traversal protection

## Performance

- DEFLATE compression for efficient storage
- Progress indicators for large operations
- Optimized pixel processing
- Responsive web interface

## Platform Support

### Supported Linux Distributions

The installation script supports:
- Debian/Ubuntu (apt)
- RHEL/CentOS (yum)
- Fedora (dnf)
- Arch Linux (pacman)
- openSUSE (zypper)
- Alpine Linux (apk)
- Solus (eopkg)
- Pardus (pisi)
- Void Linux (xbps)
- Gentoo (emerge)
- NixOS (nix)
- Clear Linux (swupd)

### Other Platforms

Manual Go installation required for:
- macOS
- Windows
- FreeBSD
- Other Unix-like systems

## Development

### Building

```bash
go build -ldflags "-s -w" -o huh main.go
```

### Testing

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the GPL License. See the [LICENSE](LICENSE) file for details.

## Troubleshooting

### Camera Access Issues

If camera access fails in the web interface:
- Check browser permissions
- Ensure HTTPS is used for production
- Verify camera is not in use by another application

### Terminal Display Issues

For optimal image viewing:
- Use a terminal with true color support
- Ensure terminal size is adequate
- Check terminal color settings

### Installation Issues

If Go installation fails:
- Install Go manually from https://golang.org/dl/
- Check your PATH environment variable
- Verify Go version compatibility

## Support

For issues, questions, or contributions, please refer to the project repository or documentation.
