#let data = json("data.json")

#set document(title: "Typst API Example", author: "Typst API")
#set page(margin: 2cm)
#set text(font: "New Computer Modern")

#align(center)[
  #block(text(weight: "bold", size: 24pt)[Typst API Example Document])
  #v(1cm)
  #image("splash192.png", width: 50%)
  #v(1cm)
]

= Introduction

This is a sample document that demonstrates how the Typst API works. It shows how to:

- Include external images
- Use data from a JSON file
- Format text and create headings

= Data from JSON

The following data was loaded from the `data.json` file that was submitted with the API request:

#table(
  columns: (auto, auto),
  inset: 10pt,
  align: horizon,
  [*Key*], [*Value*],
  ..data.keys().map(key => {
    (
      [#raw(key)],
      if type(data.at(key)) == "array" or type(data.at(key)) == "dictionary" {
        [#raw(repr(data.at(key)))]
      } else {
        [#raw(str(data.at(key)))]
      }
    )
  }).flatten()
)

= Dynamic Content

This document was generated on #datetime.today().display()

The content can be fully customized based on the data provided in the JSON payload.

#if "items" in data.keys() [
  == Items List
  
  #for item in data.at("items") [
    - #item
  ]
] 