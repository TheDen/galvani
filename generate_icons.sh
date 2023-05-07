#!/bin/bash

icon_name="icon.png"
output_path="icon.iconset"
cd "$(dirname "${0}")" || exit
rm -rf "output_path/*" || true

for size in 16 32 64 128 256 512 1024; do
  #  half="$((size / 2))"
  #  echo $half
  #  convert icon.png -resize x$size $output_path/icon_${size}x${size}.png
  #  convert icon.png -resize x$size $output_path/icon_${half}x${half}@2x.png
  sips -z $size $size ${icon_name} --out "${output_path}/icon_${size}x${size}.png"
  sips -z $size $size ${icon_name} --out "${output_path}/icon_${size}x${size}@2x.png"
done

iconutil -c icns "${output_path}"
mv icon.icns Galvani.app/Contents/Resources/
