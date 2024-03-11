#!/usr/bin/env bash

package=$1
output_dir=$2

if [[ -z "$package" || -z "$output_dir" ]]; then
  echo "usage: $0 <package-name> <output-directory>"
  exit 1
fi

if [[ ! -d "$output_dir" ]]; then
  echo "Error: Output directory does not exist."
  exit 1
fi

package_name=$(basename "$package")

platforms=("linux/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name="$output_dir/$package_name-$GOOS-$GOARCH"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
