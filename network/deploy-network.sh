#!/bin/bash
# deploy-network.sh - скрипт для розгортання мережі Hyperledger Fabric

# Вивід команд, які виконуються
set -x

# Зупинка при помилках
set -e

# Очищення попередніх артефактів
echo "Очищення попередніх артефактів..."
rm -rf ./crypto-config/*
rm -rf ./channel-artifacts/*

# Генерація криптографічних матеріалів
echo "Генерація криптографічних матеріалів..."
cryptogen generate --config=./crypto-config.yaml

# Генерація блоку генезису та артефактів каналів
echo "Генерація артефактів мережі..."
export FABRIC_CFG_PATH=$PWD
configtxgen -profile OrdererGenesis -channelID system-channel -outputBlock ./channel-artifacts/genesis.block
configtxgen -profile CommonChannel -outputCreateChannelTx ./channel-artifacts/common-channel.tx -channelID common-channel
configtxgen -profile SecurityChannel -outputCreateChannelTx ./channel-artifacts/security-channel.tx -channelID security-channel

# Генерація артефактів для приєднання організацій до каналів
echo "Генерація артефактів для приєднання до каналів..."
configtxgen -profile CommonChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors_common.tx -channelID common-channel -asOrg Org1MSP
configtxgen -profile CommonChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors_common.tx -channelID common-channel -asOrg Org2MSP
configtxgen -profile CommonChannel -outputAnchorPeersUpdate ./channel-artifacts/Org3MSPanchors_common.tx -channelID common-channel -asOrg Org3MSP

configtxgen -profile SecurityChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors_security.tx -channelID security-channel -asOrg Org1MSP
configtxgen -profile SecurityChannel -outputAnchorPeersUpdate ./channel-artifacts/Org2MSPanchors_security.tx -channelID security-channel -asOrg Org2MSP
configtxgen -profile SecurityChannel -outputAnchorPeersUpdate ./channel-artifacts/Org3MSPanchors_security.tx -channelID security-channel -asOrg Org3MSP

echo "Криптографічні матеріали та артефакти каналів успішно згенеровані!"
echo "Для запуску мережі потрібно буде використати docker-compose"