#!/bin/bash
# Cretaed by Yevgeniy Gonvharov, https://lab.sys-adm.in

# Envs
# ---------------------------------------------------\
PATH=$PATH:/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin
SCRIPT_PATH=$(cd `dirname "${BASH_SOURCE[0]}"` && pwd)

BUILD_PATH="$SCRIPT_PATH/builds"
BINARY_NAME="kompressor"

# Build
cd $SCRIPT_PATH; cd ..

# Functions
# Help information
usage() {

    echo -e "" "\nParameters:\n"
    echo -e "-b - Build binary file"
    exit 1

}

timestamp() {
    echo `date +%d-%m-%Y_%H-%M-%S`
}

backupBinary() {
    if [[ -f "$BUILD_PATH/$BINARY_NAME" ]]; then
        bkp_name="${BINARY_NAME}-$(timestamp)"
        tar -zcvf $bkp_name.tar.gz $BUILD_PATH/$BINARY_NAME
        [ -d $BUILD_PATH/prev/ ] || mkdir -p $BUILD_PATH/prev/
        mv $bkp_name.tar.gz $BUILD_PATH/prev/
    fi
}

buildBLD() {

    echo "Building BLD release.. to $BUILD_PATH"
    backupBinary

    if [[ ! -d $SCRIPT_PATH/builds ]]; then
        mkdir $SCRIPT_PATH/builds
    fi

    env GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -gcflags '-trimpath'  -o $BUILD_PATH
}

if [[ -z "$1" ]]; then
    usage;
fi

# Checks arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -b|--build) BUILD=1; ;;
        -h|--help) usage ;; 
    esac
    shift
done

if [[ "$BUILD" -eq "1" ]]; then
    buildBLD; echo "Binary saved to: $SCRIPT_PATH/builds"; echo "Done!"
fi
