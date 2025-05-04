package main

import (
    "fmt"
    "encoding/json"
    "time"
    
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract представляє смарт-контракт управління ключами
type SmartContract struct {
    contractapi.Contract
}

// CryptoKey структура криптографічного ключа
type CryptoKey struct {
    ID          string   `json:"id"`
    Type        string   `json:"type"` // symmetric, asymmetric
    Algorithm   string   `json:"algorithm"` // AES, RSA, ECDSA
    Status      string   `json:"status"` // active, rotated, revoked
    OwnerIDs    []string `json:"ownerIds"`
    CreatedAt   int64    `json:"createdAt"`
    ActivatedAt int64    `json:"activatedAt"`
    ExpiresAt   int64    `json:"expiresAt"`
    RevokedAt   int64    `json:"revokedAt"`
    Metadata    string   `json:"metadata"` // шифровані метадані
}

// KeyAccess структура доступу до ключа
type KeyAccess struct {
    KeyID      string   `json:"keyId"`
    UserID     string   `json:"userId"`
    AccessType string   `json:"accessType"` // full, encrypt-only, decrypt-only
    GrantedAt  int64    `json:"grantedAt"`
    ExpiresAt  int64    `json:"expiresAt"`
    GrantedBy  string   `json:"grantedBy"`
}

// Префікси для ключів у world state
const (
    keyPrefix    = "cryptokey:"
    accessPrefix = "keyaccess:"
)

// InitLedger ініціалізує стан смарт-контракту
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    fmt.Println("Контракт управління ключами ініціалізовано")
    return nil
}

// GenerateKey створює новий криптографічний ключ
func (s *SmartContract) GenerateKey(ctx contractapi.TransactionContextInterface, id string, keyType string, algorithm string, ownerIDs string, expirationDays int) error {
    // Створюємо унікальний ідентифікатор ключа
    txID := ctx.GetStub().GetTxID()
    keyID := fmt.Sprintf("%s-%s", id, txID[:8])
    
    // Парсимо власників з JSON рядка
    var ownersList []string
    err := json.Unmarshal([]byte(ownerIDs), &ownersList)
    if err != nil {
        return fmt.Errorf("помилка при розборі власників: %v", err)
    }
    
    // Встановлюємо час дії
    now := time.Now().Unix()
    expiresAt := now + int64(expirationDays*24*60*60)
    
    // Створюємо запис ключа
    key := CryptoKey{
        ID:          keyID,
        Type:        keyType,
        Algorithm:   algorithm,
        Status:      "active",
        OwnerIDs:    ownersList,
        CreatedAt:   now,
        ActivatedAt: now,
        ExpiresAt:   expiresAt,
        RevokedAt:   0,
        Metadata:    "", // В реальному коді тут були б шифровані метадані
    }
    
    // Серіалізуємо ключ
    keyJSON, err := json.Marshal(key)
    if err != nil {
        return err
    }
    
    // Зберігаємо в state database
    return ctx.GetStub().PutState(keyPrefix+keyID, keyJSON)
}

// GrantKeyAccess надає доступ до ключа певному користувачу
func (s *SmartContract) GrantKeyAccess(ctx contractapi.TransactionContextInterface, keyID string, userID string, accessType string, grantedBy string, expirationDays int) error {
    // Отримання даних ключа
    keyJSON, err := ctx.GetStub().GetState(keyPrefix + keyID)
    if err != nil {
        return fmt.Errorf("помилка читання ключа: %v", err)
    }
    if keyJSON == nil {
        return fmt.Errorf("ключ %s не існує", keyID)
    }
    
    var key CryptoKey
    err = json.Unmarshal(keyJSON, &key)
    if err != nil {
        return fmt.Errorf("помилка десеріалізації даних ключа: %v", err)
    }
    
    // Перевірка статусу ключа
    if key.Status != "active" {
        return fmt.Errorf("ключ %s не активний", keyID)
    }
    
    // Встановлюємо час дії
    now := time.Now().Unix()
    expiresAt := now + int64(expirationDays*24*60*60)
    
    // Створюємо запис доступу
    access := KeyAccess{
        KeyID:      keyID,
        UserID:     userID,
        AccessType: accessType,
        GrantedAt:  now,
        ExpiresAt:  expiresAt,
        GrantedBy:  grantedBy,
    }
    
    // Серіалізуємо доступ
    accessJSON, err := json.Marshal(access)
    if err != nil {
        return err
    }
    
    // Створюємо складений ключ для доступу
    accessKey := fmt.Sprintf("%s%s-%s", accessPrefix, keyID, userID)
    
    // Зберігаємо в state database
    return ctx.GetStub().PutState(accessKey, accessJSON)
}

func main() {
    chaincode, err := contractapi.NewChaincode(new(SmartContract))
    if err != nil {
        fmt.Printf("Помилка створення чейнкоду: %s", err.Error())
        return
    }
    
    if err := chaincode.Start(); err != nil {
        fmt.Printf("Помилка запуску чейнкоду: %s", err.Error())
    }
}