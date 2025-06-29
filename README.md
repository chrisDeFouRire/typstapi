# Typst API

A simple web API that processes Typst documents and returns PDFs. The API accepts file uploads and JSON data, processes them using the Typst CLI, and returns a PDF (with optional gzip compression). It also supports merging additional PDFs before and after the Typst-generated PDF.

## Example

Check out the `example` directory for a complete working example of how to use this API, including:

- A sample Typst document (`main.typ`)
- Example JSON data
- A sample logo image (`splash192.png`)
- Instructions for testing the API

## Prerequisites

For local development:

- Go 1.23 or later
- Typst CLI installed and available in your PATH

For Docker deployment:

- Docker

## Installation

### Local Build

```bash
git clone <repository-url>
cd typstapi
go build
```

### Docker Build

```bash
docker build -t typstapi .
```

## Running the Server

### Local Run

```bash
# Set custom port (optional, defaults to 8080)
export PORT=3000
# Run the server
./typstapi
```

### Docker Run

```bash
# Run with default port 8080
docker run -p 8080:8080 typstapi

# Run with custom port
docker run -e PORT=3000 -p 3000:3000 typstapi
```

### A pre-built Docker image is available on Docker Hub
 
See the Docker Hub page [sslboard/typstapi](https://hub.docker.com/r/sslboard/typstapi) for more details.

```bash
# Pull the image
docker pull sslboard/typstapi:v1.0.0

# Run the container
docker run -p 8080:8080 sslboard/typstapi:v1.0.0
```

## API Usage

### POST /typst/:filename

Process a Typst document and return the compiled PDF. Optionally merge additional PDFs before and after the Typst-generated PDF.

**Request:**

- Method: POST
- Content-Type: multipart/form-data
- Path Parameter: `:filename` - The name of the main Typst file to process

**Form Fields:**

- Files: Upload your Typst files and any assets (images, etc.)
- `data`: (Optional) JSON string that will be saved as `data.json` in the processing directory
- Pre-PDFs: (Optional) Files with names starting with `pre_` and ending with `.pdf` will be merged before the Typst output
- Post-PDFs: (Optional) Files with names starting with `post_` and ending with `.pdf` will be merged after the Typst output

**Response:**

- Content-Type: application/pdf
- Content-Encoding: gzip (only if client supports it via Accept-Encoding header)
- The compiled PDF file (compressed with gzip only if client supports it)

**Example using curl:**

```bash
# Standard request (PDF will be compressed if browser supports gzip)
curl -X POST http://localhost:8080/typst/main.typ \
  -F "main.typ=@/path/to/main.typ" \
  -F "logo.jpeg=@/path/to/logo.jpeg" \
  -F 'data={"hello": "world"}' \
  --output output.pdf

# Request with pre and post PDFs
curl -X POST http://localhost:8080/typst/main.typ \
  -F "main.typ=@/path/to/main.typ" \
  -F "logo.jpeg=@/path/to/logo.jpeg" \
  -F 'data={"hello": "world"}' \
  -F "pre_1.pdf=@/path/to/cover.pdf" \
  -F "pre_2.pdf=@/path/to/toc.pdf" \
  -F "post_1.pdf=@/path/to/appendix.pdf" \
  --output output.pdf

# Explicitly request no compression by not sending Accept-Encoding header
curl -X POST http://localhost:8080/typst/main.typ \
  -H "Accept-Encoding: identity" \
  -F "main.typ=@/path/to/main.typ" \
  -F "logo.jpeg=@/path/to/logo.jpeg" \
  -F 'data={"hello": "world"}' \
  --output output.pdf
```

## PDF Merging

The API supports merging additional PDFs before and after the Typst-generated PDF:

1. Pre-PDFs:
   - Upload files with names starting with `pre_` and ending with `.pdf`
   - Files are merged in alphabetical order
   - Example: `pre_1.pdf`, `pre_2.pdf`, etc.

2. Post-PDFs:
   - Upload files with names starting with `post_` and ending with `.pdf`
   - Files are merged in alphabetical order
   - Example: `post_1.pdf`, `post_2.pdf`, etc.

The final PDF will be assembled in this order:

1. All pre-PDFs in alphabetical order
2. The Typst-generated PDF
3. All post-PDFs in alphabetical order

## Error Handling

The API returns appropriate HTTP status codes:

- 400: Bad Request (invalid input)
- 405: Method Not Allowed (non-POST requests)
- 500: Internal Server Error (processing failures)

When a Typst compilation error occurs, the API returns:

- The HTTP 500 status code
- An error message that includes the original error
- The complete stderr output from the Typst CLI to help with debugging

## Security Notes

- The API creates temporary directories for processing files
- All temporary files are automatically cleaned up after processing
- File size is limited to 32MB per request

## Author

Created by Chris Hartwig for [SSLBoard.com](https://sslboard.com).
