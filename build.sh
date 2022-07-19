#!/bin/bash

function usage() {
    echo -e "Usage: $0 -p [linux|darwin] -a [amd64|arm] -l [path_of_fconf]"
}

# check command-line
[ $# -eq 0 ] && usage && exit 1

# get parameters
while getopts p:a:l: flag; do
    case "${flag}" in
        p) plat=${OPTARG};;
        a) arch=${OPTARG};;
        l) conf=${OPTARG};;
    esac
done

[ "$plat" != "linux" ] && [ "$plat" != "darwin" ] && echo "invalid plat" && usage && exit 1
[ "$arch" != "amd64" ] && [ "$arch" != "arm" ] && echo "invalid arch" && usage && exit 1

# read config file
if [ "$conf" == "" ]; then
    echo "use default const vals"
    flags=""
else
    echo "use custom const vals"
    flags=""

    asset_name=`jq -r .smart_assets_name $conf`
    asset_symbol=`jq -r .smart_assets_symbol $conf`
    asset_supply=`jq -r .smart_assets_supply $conf`
    asset_precision=`jq -r .smart_assets_precision $conf`
    asset_owner=`jq -r .smart_assets_owner $conf`

    [ "$asset_name" == null ]      || [ "$asset_name" == "" ]      || flags+="-X 'github.com/SealSC/SealABC/config.SmartAssetsName=${asset_name}'"
    [ "$asset_symbol" == null ]    || [ "$asset_symbol" == "" ]    || flags+="-X 'github.com/SealSC/SealABC/config.SmartAssetsSymbol=${asset_symbol}'"
    [ "$asset_supply" == null ]    || [ "$asset_supply" == "" ]    || flags+="-X 'github.com/SealSC/SealABC/config.SmartAssetsSupply=${asset_supply}'"
    [ "$asset_precision" == null ] || [ "$asset_precision" == "" ] || flags+="-X 'github.com/SealSC/SealABC/config.SmartAssetsPrecision=${asset_precision}'"
    [ "$asset_owner" == null ]     || [ "$asset_owner" == "" ]     || flags+="-X 'github.com/SealSC/SealABC/config.SmartAssetsOwner=${asset_owner}'"
fi


# build
echo "building for plat=${plat} arch=${arch}"
echo "ldflags: $flags"
GOOS=${plat} GOARCH=${arch} go build -ldflags "${flags}" SealABC.go

echo "finished"
