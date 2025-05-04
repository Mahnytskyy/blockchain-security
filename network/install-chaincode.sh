#!/bin/bash
# install-chaincode.sh - скрипт для встановлення та активації смарт-контрактів

# Вивід команд, які виконуються
set -x

# Зупинка при помилках
set -e

# Змінні середовища
export FABRIC_CFG_PATH=$PWD
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051

# Функція для пакування смарт-контракту
packageChaincode() {
  echo "Пакування смарт-контракту $1..."
  peer lifecycle chaincode package ${PWD}/../chaincode/$1.tar.gz --path ${PWD}/../chaincode/$1/go/ --lang golang --label $1_1.0
}

# Функція для встановлення смарт-контракту
installChaincode() {
  echo "Встановлення смарт-контракту $1 на Org1..."
  peer lifecycle chaincode install ${PWD}/../chaincode/$1.tar.gz
  
  export CORE_PEER_LOCALMSPID="Org2MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
  
  echo "Встановлення смарт-контракту $1 на Org2..."
  peer lifecycle chaincode install ${PWD}/../chaincode/$1.tar.gz
  
  export CORE_PEER_LOCALMSPID="Org3MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/crypto-config/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org3.example.com:11051
  
  echo "Встановлення смарт-контракту $1 на Org3..."
  peer lifecycle chaincode install ${PWD}/../chaincode/$1.tar.gz
}

# Функція для затвердження смарт-контракту
approveChaincode() {
  # Отримання ID пакету
  PACKAGE_ID=$(peer lifecycle chaincode queryinstalled | grep $1_1.0 | awk '{print $3}' | sed 's/,//')
  
  # Затвердження від Org1
  export CORE_PEER_LOCALMSPID="Org1MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
  
  echo "Затвердження смарт-контракту $1 від Org1..."
  peer lifecycle chaincode approveformyorg -o orderer1.orderer.example.com:7050 --channelID security-channel --name $1 --version 1.0 --package-id $PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer1.orderer.example.com/msp/tlscacerts/tlsca.orderer.example.com-cert.pem
  
  # Затвердження від Org2
  export CORE_PEER_LOCALMSPID="Org2MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/crypto-config/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
  
  echo "Затвердження смарт-контракту $1 від Org2..."
  peer lifecycle chaincode approveformyorg -o orderer1.orderer.example.com:7050 --channelID security-channel --name $1 --version 1.0 --package-id $PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer1.orderer.example.com/msp/tlscacerts/tlsca.orderer.example.com-cert.pem
  
  # Затвердження від Org3
  export CORE_PEER_LOCALMSPID="Org3MSP"
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/crypto-config/peerOrganizations/org3.example.com/users/Admin@org3.example.com/msp
  export CORE_PEER_ADDRESS=peer0.org3.example.com:11051
  
  echo "Затвердження смарт-контракту $1 від Org3..."
  peer lifecycle chaincode approveformyorg -o orderer1.orderer.example.com:7050 --channelID security-channel --name $1 --version 1.0 --package-id $PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer1.orderer.example.com/msp/tlscacerts/tlsca.orderer.example.com-cert.pem
}

# Функція для активації смарт-контракту
commitChaincode() {
  echo "Активація смарт-контракту $1..."
  peer lifecycle chaincode commit -o orderer1.orderer.example.com:7050 --channelID security-channel --name $1 --version 1.0 --sequence 1 --tls --cafile ${PWD}/crypto-config/ordererOrganizations/orderer.example.com/orderers/orderer1.orderer.example.com/msp/tlscacerts/tlsca.orderer.example.com-cert.pem --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles ${PWD}/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles ${PWD}/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --peerAddresses peer0.org3.example.com:11051 --tlsRootCertFiles ${PWD}/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt
}

# Пакування смарт-контрактів
packageChaincode "accesscontrol"
packageChaincode "securityaudit"
packageChaincode "keymanagement"

# Встановлення смарт-контрактів
installChaincode "accesscontrol"
installChaincode "securityaudit"
installChaincode "keymanagement"

# Затвердження та активація смарт-контрактів
approveChaincode "accesscontrol"
commitChaincode "accesscontrol"

approveChaincode "securityaudit"
commitChaincode "securityaudit"

approveChaincode "keymanagement"
commitChaincode "keymanagement"

echo "Смарт-контракти успішно встановлені та активовані!"
