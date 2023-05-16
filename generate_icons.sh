#!/bin/bash

icon_name="icon.png"
output_path="icon.iconset"
cd "$(dirname "${0}")" || exit
rm -rf "$(output_path)" || true
mkdir -p "${output_path}"

for size in 16 32 64 128 256 512 1024; do
  sips -z $size $size ${icon_name} --out "${output_path}/icon_${size}x${size}.png"
  sips -z $size $size ${icon_name} --out "${output_path}/icon_${size}x${size}@2x.png"
done

iconutil -c icns "${output_path}"
mv icon.icns Galvani.app/Contents/Resources/
