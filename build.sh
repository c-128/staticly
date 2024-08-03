#!/bin/bash

mkdir -p build
rm -rf build/*

for os in "linux" "windows" "darwin"; do
    for arch in "amd64" "arm" "arm64"; do
        output="build/staticly-$os-$arch"
        if [[ $os == "windows" ]]; then
            output="$output.exe"
        fi

        GOOS=$os GOARCH=$arch go build -o $output *.go
    done
done
