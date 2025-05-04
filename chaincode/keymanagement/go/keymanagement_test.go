// Файл: chaincode/keymanagement/go/keymanagement_test.go
package main

import (
    "encoding/json"
    "testing"
    "time"

    "github.com/hyperledger/fabric-chaincode-go/shim"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockStub імітує ChainCodeStubInterface
type MockStub struct {
    mock.Mock
    shim.ChaincodeStubInterface
}

func (s *MockStub) GetState(key string) ([]byte, error) {
    args := s.Called(key)
    return args.Get(0).([]byte), args.Error(1)
}

func (s *MockStub) PutState(key string, value []byte) error {
    args := s.Called(key, value)
    return args.Error(0)
}

func (s *MockStub) GetTxID() string {
    args := s.Called()
    return args.String(0)
}

// MockContext імітує TransactionContextInterface
type MockContext struct {
    mock.Mock
    contractapi.TransactionContextInterface
}

func (c *MockContext) GetStub() shim.ChaincodeStubInterface {
    args := c.Called()
    return args.Get(0).(shim.ChaincodeStubInterface)
}

// Тестування GenerateKey
func TestGenerateKey(t *testing.T) {
    // Ініціалізація мок-об'єктів
    mockStub := new(MockStub)
    mockContext := new(MockContext)
    mockContext.On("GetStub").Return(mockStub)

    // Підготовка даних для тесту
    id := "key123"
    keyType := "symmetric"
    algorithm := "AES"
    ownerIDs := `["user1", "user2"]`
    expirationDays := 365
    
    // Очікуємо виклики методів
    mockStub.On("GetTxID").Return("tx123")
    mockStub.On("PutState", mock.Anything, mock.Anything).Return(nil)

    // Створення об'єкту смарт-контракту і виклик методу
    contract := new(SmartContract)
    err := contract.GenerateKey(mockContext, id, keyType, algorithm, ownerIDs, expirationDays)
    
    // Перевірка результатів
    assert.Nil(t, err)
    mockStub.AssertExpectations(t)
    mockContext.AssertExpectations(t)
    
    // Перевірка, що PutState був викликаний з коректними даними
    call := mockStub.Calls[1] // Другий виклик - це PutState
    actualKey := call.Arguments[0].(string)
    actualValue := call.Arguments[1].([]byte)
    
    // Перевірка, що ключ починається з префіксу "cryptokey:"
    assert.Contains(t, actualKey, "cryptokey:")
    
    // Десеріалізація ключа для перевірки полів
    var cryptoKey CryptoKey
    err = json.Unmarshal(actualValue, &cryptoKey)
    assert.Nil(t, err)
    
    // Перевірка полів ключа
    assert.Contains(t, cryptoKey.ID, id)
    assert.Equal(t, keyType, cryptoKey.Type)
    assert.Equal(t, algorithm, cryptoKey.Algorithm)
    assert.Equal(t, "active", cryptoKey.Status)
    
    // Перевірка власників
    var expectedOwners []string
    err = json.Unmarshal([]byte(ownerIDs), &expectedOwners)
    assert.Nil(t, err)
    assert.Equal(t, expectedOwners, cryptoKey.OwnerIDs)
    
    // Перевірка часових міток
    now := time.Now().Unix()
    assert.LessOrEqual(t, cryptoKey.CreatedAt, now)
    assert.LessOrEqual(t, cryptoKey.ActivatedAt, now)
    assert.Greater(t, cryptoKey.ExpiresAt, now)
    assert.Equal(t, int64(0), cryptoKey.RevokedAt)
}

// Тестування GrantKeyAccess
func TestGrantKeyAccess(t *testing.T) {
    // Ініціалізація мок-об'єктів
    mockStub := new(MockStub)
    mockContext := new(MockContext)
    mockContext.On("GetStub").Return(mockStub)

    // Підготовка даних для тесту
    keyID := "key123"
    keyJSON := []byte(`{"id":"key123","type":"symmetric","algorithm":"AES","status":"active","ownerIds":["user1"],"createdAt":1620000000,"activatedAt":1620000000,"expiresAt":1651536000,"revokedAt":0}`)
    userID := "user2"
    accessType := "encrypt-only"
    grantedBy := "user1"
    expirationDays := 30
    
    // Очікуємо виклики методів
    mockStub.On("GetState", "cryptokey:"+keyID).Return(keyJSON, nil)
    mockStub.On("PutState", mock.Anything, mock.Anything).Return(nil)

    // Створення об'єкту смарт-контракту і виклик методу
    contract := new(SmartContract)
    err := contract.GrantKeyAccess(mockContext, keyID, userID, accessType, grantedBy, expirationDays)
    
    // Перевірка результатів
    assert.Nil(t, err)
    mockStub.AssertExpectations(t)
    mockContext.AssertExpectations(t)
    
    // Перевірка, що PutState був викликаний з коректними даними
    call := mockStub.Calls[1] // Другий виклик - це PutState
    actualKey := call.Arguments[0].(string)
    actualValue := call.Arguments[1].([]byte)
    
    // Перевірка, що ключ має правильний формат
    expectedKey := "keyaccess:key123-user2"
    assert.Equal(t, expectedKey, actualKey)
    
    // Десеріалізація доступу до ключа для перевірки полів
    var keyAccess KeyAccess
    err = json.Unmarshal(actualValue, &keyAccess)
    assert.Nil(t, err)
    
    // Перевірка полів доступу
    assert.Equal(t, keyID, keyAccess.KeyID)
    assert.Equal(t, userID, keyAccess.UserID)
    assert.Equal(t, accessType, keyAccess.AccessType)
    assert.Equal(t, grantedBy, keyAccess.GrantedBy)
    
    // Перевірка часових міток
    now := time.Now().Unix()
    assert.LessOrEqual(t, keyAccess.GrantedAt, now)
    assert.Greater(t, keyAccess.ExpiresAt, now)
}

// Тестування RevokeKeyAccess
func TestRevokeKeyAccess(t *testing.T) {
    // Ініціалізація мок-об'єктів
    mockStub := new(MockStub)
    mockContext := new(MockContext)
    mockContext.On("GetStub").Return(mockStub)

    // Підготовка даних для тесту
    keyID := "key123"
    userID := "user2"
    accessKey := "keyaccess:key123-user2"
    accessJSON := []byte(`{"keyId":"key123","userId":"user2","accessType":"encrypt-only","grantedAt":1620000000,"expiresAt":1651536000,"grantedBy":"user1"}`)
    
    // Очікуємо виклики методів
    mockStub.On("GetState", accessKey).Return(accessJSON, nil)
    mockStub.On("DelState", accessKey).Return(nil)

    // Створення об'єкту смарт-контракту і виклик методу
    contract := new(SmartContract)
    err := contract.RevokeKeyAccess(mockContext, keyID, userID)
    
    // Перевірка результатів
    assert.Nil(t, err)
    mockStub.AssertExpectations(t)
    mockContext.AssertExpectations(t)
}

// Тестування RotateKey
func TestRotateKey(t *testing.T) {
    // Ініціалізація мок-об'єктів
    mockStub := new(MockStub)
    mockContext := new(MockContext)
    mockContext.On("GetStub").Return(mockStub)

    // Підготовка даних для тесту
    keyID := "key123"
    oldKeyJSON := []byte(`{"id":"key123","type":"symmetric","algorithm":"AES","status":"active","ownerIds":["user1"],"createdAt":1620000000,"activatedAt":1620000000,"expiresAt":1651536000,"revokedAt":0}`)
    
    // Очікуємо виклики методів
    mockStub.On("GetState", "cryptokey:"+keyID).Return(oldKeyJSON, nil)
    mockStub.On("GetTxID").Return("tx456")
    mockStub.On("PutState", mock.Anything, mock.Anything).Return(nil).Times(2) // Один раз для старого ключа, один для нового

    // Створення об'єкту смарт-контракту і виклик методу
    contract := new(SmartContract)
    newKeyID, err := contract.RotateKey(mockContext, keyID)
    
    // Перевірка результатів
    assert.Nil(t, err)
    assert.NotEmpty(t, newKeyID)
    assert.NotEqual(t, keyID, newKeyID)
    mockStub.AssertExpectations(t)
    mockContext.AssertExpectations(t)
    
    // Перевірка, що старий ключ був оновлений
    call1 := mockStub.Calls[1] // Другий виклик - це PutState для старого ключа
    oldKeyValue := call1.Arguments[1].([]byte)
    
    var updatedOldKey CryptoKey
    err = json.Unmarshal(oldKeyValue, &updatedOldKey)
    assert.Nil(t, err)
    assert.Equal(t, "rotated", updatedOldKey.Status)
    
    // Перевірка, що новий ключ був створений
    call2 := mockStub.Calls[2] // Третій виклик - це PutState для нового ключа
    newKeyValue := call2.Arguments[1].([]byte)
    
    var newKey CryptoKey
    err = json.Unmarshal(newKeyValue, &newKey)
    assert.Nil(t, err)
    assert.Contains(t, newKey.ID, keyID) // Новий ключ повинен містити ID старого
    assert.Equal(t, "active", newKey.Status)
    assert.Equal(t, updatedOldKey.Type, newKey.Type)
    assert.Equal(t, updatedOldKey.Algorithm, newKey.Algorithm)
    assert.Equal(t, updatedOldKey.OwnerIDs, newKey.OwnerIDs)
}