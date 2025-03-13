# Typst API

A simple web API that processes Typst documents and returns PDFs. The API accepts file uploads and JSON data, processes them using the Typst CLI, and returns a gzip-compressed PDF.

## Example

Check out the `example` directory for a complete working example of how to use this API, including:
- A sample Typst document (`main.typ`)
- Example JSON data
- A sample logo image (`splash192.png`)
- Instructions for testing the API

## Prerequisites

For local development:
- Go 1.16 or later
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

## API Usage

### POST /typst/:filename

Process a Typst document and return the compiled PDF.

**Request:**
- Method: POST
- Content-Type: multipart/form-data
- Path Parameter: `:filename` - The name of the main Typst file to process

**Form Fields:**
- Files: Upload your Typst files and any assets (images, etc.)
- `data`: (Optional) JSON string that will be saved as `data.json` in the processing directory

**Response:**
- Content-Type: application/pdf
- Content-Encoding: gzip
- The compiled PDF file, gzip compressed

**Example using curl:**
```bash
curl -X POST http://localhost:8080/typst/main.typ \
  -F "main.typ=@/path/to/main.typ" \
  -F "logo.jpeg=@/path/to/logo.jpeg" \
  -F 'data={"hello": "world"}' \
  --output output.pdf
```

## Error Handling

The API returns appropriate HTTP status codes:
- 400: Bad Request (invalid input)
- 405: Method Not Allowed (non-POST requests)
- 500: Internal Server Error (processing failures)

## Security Notes

- The API creates temporary directories for processing files
- All temporary files are automatically cleaned up after processing
- File size is limited to 32MB per request
