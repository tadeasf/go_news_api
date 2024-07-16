#!/bin/bash

# Set GIN_MODE to release
export GIN_MODE=release

# Function to increment version
increment_version() {
    local version=$1
    local delimiter=.
    local array=($(echo "$version" | tr $delimiter '\n'))
    array[2]=$((array[2] + 1))
    echo $(
        local IFS=$delimiter
        echo "${array[*]}"
    )
}

# Read current version from build.log (create if not exists)
if [ ! -f build.log ]; then
    echo "----- version: 1.0.0 -----" >build.log
    current_version="1.0.0"
else
    current_version=$(grep -oP '(?<=version: )[0-9.]+' build.log | tail -1)
fi

# Increment version
new_version=$(increment_version $current_version)

echo "Building version $new_version"

# Tidy up dependencies
go mod tidy

# Build the application
go build -o ./dist/go_news_api -ldflags="-w -s -X main.Version=$new_version" .

# Log the build
echo "" >>build.log
echo "----- version: $new_version -----" >>build.log
echo "$(date): Built version $new_version" >>build.log

echo "Build completed. New version: $new_version"
