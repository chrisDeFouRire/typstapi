# Typst API Example

This directory contains example files to demonstrate how to use the Typst API.

## Files

- `main.typ`: A sample Typst document that uses data from a JSON file and includes an image
- `sample-data.json`: Example JSON data to be submitted with the API request
- `splash192.png`: A logo image used in the Typst document

## Testing the API

Once you have the API running, you can test it with:

```bash
# This will produce a regular PDF file without compression
curl -X POST http://localhost:8080/typst/main.typ \
  -F "main.typ=@main.typ" \
  -F "splash192.png=@splash192.png" \
  -F "data=$(cat sample-data.json)" \
  --output example.pdf
```

This will:

1. Upload the Typst document (`main.typ`)
2. Upload the logo image (`splash192.png`)
3. Submit the JSON data from `sample-data.json`
4. Save the resulting PDF as `example.pdf`

## Understanding the Example

The `main.typ` file demonstrates:

- Importing and using the JSON data from the API request
- Including an image file
- Creating a formatted document with headings, tables, and lists
- Conditionally displaying content based on the JSON data

The JSON data (`sample-data.json`) shows different data types that can be used in the Typst document. 
