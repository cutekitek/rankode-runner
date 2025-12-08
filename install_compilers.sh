#!/bin/sh
set -e

if [ -z "$LANGUAGES" ]; then
    echo "No LANGUAGES environment variable set. Exiting."
    exit 0
fi

echo "--- Preparing System ---"
echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories
apk update
apk upgrade

apk add --no-cache curl bash build-base

APK_PACKAGES=""

target_langs=$(echo $LANGUAGES | tr ',' ' ' | tr '[:upper:]' '[:lower:]')

for lang in $target_langs; do
    echo "--- Installing support for: $lang ---"
    case "$lang" in
        c)
            APK_PACKAGES="$APK_PACKAGES gcc musl-dev"
            ;;
        c++|cpp)
            APK_PACKAGES="$APK_PACKAGES g++ musl-dev"
            ;;
        go|golang)
            APK_PACKAGES="$APK_PACKAGES go"
            ;;
        java|java21)
            APK_PACKAGES="$APK_PACKAGES openjdk21"
            ;;
        python|py|python3)
            APK_PACKAGES="$APK_PACKAGES python3"
            ;;
        perl)
            APK_PACKAGES="$APK_PACKAGES perl"
            ;;
        c#|csharp|dotnet)
            APK_PACKAGES="$APK_PACKAGES dotnet-sdk icu-libs"
            ;;
        rust|rustlang)
            export RUSTUP_HOME=/usr/local/rustup
            export CARGO_HOME=/usr/local/cargo
            
            curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --profile minimal --no-modify-path
            
            echo "Copying Rust binaries to /usr/bin..."
            cp $CARGO_HOME/bin/rustc /usr/bin/rustc
            cp $CARGO_HOME/bin/cargo /usr/bin/cargo
            
            chmod +x /usr/bin/rustc /usr/bin/cargo
            chmod -R a+w $RUSTUP_HOME $CARGO_HOME
            
            echo "Rust installed and copied to /usr/bin."
            ;;
        javascript|js|bun)
            APK_PACKAGES="$APK_PACKAGES gcompat libstdc++"
            curl -fsSL https://bun.sh/install | bash
            mv /root/.bun/bin/bun /usr/bin/bun
            rm -rf /root/.bun
            echo "Bun installed to /usr/bin."
            ;;
        *)
            echo "Warning: Language '$lang' skipped (unknown)."
            ;;
    esac
done

if [ -n "$APK_PACKAGES" ]; then
    echo "Running APK install for: $APK_PACKAGES"
    apk add --no-cache $APK_PACKAGES
fi

echo "Cleaning up..."
apk del build-base curl
rm -rf /var/cache/apk/* /tmp/*

echo "Installation complete."