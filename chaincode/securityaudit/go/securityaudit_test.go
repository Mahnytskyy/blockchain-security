// Файл: chaincode/securityaudit/go/securityaudit_test.go
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

func (s *MockStub) DelState(key string) error {
	args := s.Called(key)
	return args.Error(0)
}

func (s *MockStub) GetTxID() string {
	args := s.Called()
	return args.String(0)
}

func (s *MockStub) SetEvent(name string, payload []byte) error {
	args := s.Called(name, payload)
	return args.Error(0)
}

func (s *MockStub) GetStateByRange(startKey string, endKey string) (shim.StateQueryIteratorInterface, error) {
	args := s.Called(startKey, endKey)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Error(1)
}

// MockQueryIterator імітує StateQueryIteratorInterface
type MockQueryIterator struct {
	mock.Mock
	shim.StateQueryIteratorInterface
	Results []KVPair
	Index   int
}

type KVPair struct {
	Key   string
	Value []byte
}

func (i *MockQueryIterator) HasNext() bool {
	return i.Index < len(i.Results)
}

func (i *MockQueryIterator) Next() (*shim.KV, error) {
	if i.Index < len(i.Results) {
		result := i.Results[i.Index]
		i.Index++
		return &shim.KV{
			Key:   result.Key,
			Value: result.Value,
		}, nil
	}
	return nil, nil
}

func (i *MockQueryIterator) Close() error {
	return nil
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

// Тестування RecordEvent
func TestRecordEvent(t *testing.T) {
	// Ініціалізація мок-об'єктів
	mockStub := new(MockStub)
	mockContext := new(MockContext)
	mockContext.On("GetStub").Return(mockStub)

	// Підготовка даних для тесту
	eventType := "access_check"
	actor := "user123"
	resource := "resource123"
	action := "check_access"
	result := "granted"
	metadata := `{"source":"api","timestamp":"2024-05-01T12:00:00Z"}`
	
	// Очікуємо виклики методів
	mockStub.On("GetTxID").Return("tx123")
	mockStub.On("PutState", mock.Anything, mock.Anything).Return(nil)
	mockStub.On("SetEvent", "SecurityAuditEvent", mock.Anything).Return(nil)

	// Створення об'єкту смарт-контракту і виклик методу
	contract := new(SmartContract)
	err := contract.RecordEvent(mockContext, eventType, actor, resource, action, result, metadata)
	
	// Перевірка результатів
	assert.Nil(t, err)
	mockStub.AssertExpectations(t)
	mockContext.AssertExpectations(t)
	
	// Перевірка, що PutState був викликаний з коректними даними
	call := mockStub.Calls[1] // Другий виклик - це PutState
	actualKey := call.Arguments[0].(string)
	actualValue := call.Arguments[1].([]byte)
	
	// Перевірка, що ключ починається з префіксу "event:"
	assert.Contains(t, actualKey, "event:")
	
	// Десеріалізація події для перевірки полів
	var event SecurityEvent
	err = json.Unmarshal(actualValue, &event)
	assert.Nil(t, err)
	
	// Перевірка полів події
	assert.NotEmpty(t, event.ID)
	assert.Equal(t, eventType, event.Type)
	assert.Equal(t, actor, event.Actor)
	assert.Equal(t, resource, event.Resource)
	assert.Equal(t, action, event.Action)
	assert.Equal(t, result, event.Result)
	assert.NotZero(t, event.Timestamp)
	
	// Перевірка метаданих
	var expectedMetadata map[string]string
	err = json.Unmarshal([]byte(metadata), &expectedMetadata)
	assert.Nil(t, err)
	assert.Equal(t, expectedMetadata, event.Metadata)
}

// Тестування QueryEvents
func TestQueryEvents(t *testing.T) {
	// Ініціалізація мок-об'єктів
	mockStub := new(MockStub)
	mockContext := new(MockContext)
	mockContext.On("GetStub").Return(mockStub)

	// Підготовка тестових даних для відповіді від GetStateByRange
	events := []SecurityEvent{
		{
			ID:        "event1",
			Type:      "access_check",
			Timestamp: time.Now().Unix() - 100,
			Actor:     "user123",
			Resource:  "resource1",
			Action:    "check_access",
			Result:    "granted",
			Metadata:  map[string]string{"source": "api"},
		},
		{
			ID:        "event2",
			Type:      "access_check",
			Timestamp: time.Now().Unix() - 50,
			Actor:     "user123",
			Resource:  "resource2",
			Action:    "check_access",
			Result:    "denied",
			Metadata:  map[string]string{"source": "api"},
		},
	}
	
	// Створення серіалізованих подій для мок-відповіді
	var mockResults []KVPair
	for i, event := range events {
		eventJSON, _ := json.Marshal(event)
		mockResults = append(mockResults, KVPair{
			Key:   fmt.Sprintf("event:%d", i),
			Value: eventJSON,
		})
	}
	
	// Створення мок-ітератора
	mockIterator := &MockQueryIterator{
		Results: mockResults,
	}
	
	// Налаштування поведінки мок-об'єкта
	mockStub.On("GetStateByRange", "event:", "event~").Return(mockIterator, nil)
	
	// Створення параметрів запиту
	queryString := `{"startTime": 0, "endTime": 9999999999, "limit": 10}`
	
	// Створення об'єкту смарт-контракту і виклик методу
	contract := new(SmartContract)
	result, err := contract.QueryEvents(mockContext, queryString)
	
	// Перевірка результатів
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	
	// Перевірка, що події повернуті у правильному порядку
	assert.Equal(t, events[0].ID, result[0].ID)
	assert.Equal(t, events[1].ID, result[1].ID)
	
	mockStub.AssertExpectations(t)
	mockContext.AssertExpectations(t)
}

// Тестування фільтрації за параметрами
func TestQueryEventsWithFilters(t *testing.T) {
	// Тестові випадки для різних комбінацій фільтрів
	testCases := []struct {
		name           string
		queryParams    string
		expectedEvents int
		eventType      string
		actor          string
	}{
		{
			name:           "Фільтр за типом події",
			queryParams:    `{"startTime": 0, "endTime": 9999999999, "eventType": "login", "limit": 10}`,
			expectedEvents: 1,
			eventType:      "login",
			actor:          "",
		},
		{
			name:           "Фільтр за користувачем",
			queryParams:    `{"startTime": 0, "endTime": 9999999999, "actor": "admin", "limit": 10}`,
			expectedEvents: 1,
			eventType:      "",
			actor:          "admin",
		},
		{
			name:           "Комбінований фільтр",
			queryParams:    `{"startTime": 0, "endTime": 9999999999, "eventType": "access_check", "actor": "user123", "limit": 10}`,
			expectedEvents: 1,
			eventType:      "access_check",
			actor:          "user123",
		},
	}
	
	// Продовження функції TestQueryEventsWithFilters
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Ініціалізація мок-об'єктів
            mockStub := new(MockStub)
            mockContext := new(MockContext)
            mockContext.On("GetStub").Return(mockStub)
            
            // Підготовка тестових даних - різні події для тестування фільтрів
            events := []SecurityEvent{
                {
                    ID:        "event1",
                    Type:      "access_check",
                    Timestamp: time.Now().Unix() - 100,
                    Actor:     "user123",
                    Resource:  "resource1",
                    Action:    "check_access",
                    Result:    "granted",
                    Metadata:  map[string]string{"source": "api"},
                },
                {
                    ID:        "event2",
                    Type:      "login",
                    Timestamp: time.Now().Unix() - 50,
                    Actor:     "admin",
                    Resource:  "system",
                    Action:    "login",
                    Result:    "success",
                    Metadata:  map[string]string{"source": "web"},
                },
            }
            
            // Створення серіалізованих подій для мок-відповіді
            var mockResults []KVPair
            for i, event := range events {
                eventJSON, _ := json.Marshal(event)
                mockResults = append(mockResults, KVPair{
                    Key:   fmt.Sprintf("event:%d", i),
                    Value: eventJSON,
                })
            }
            
            // Створення мок-ітератора
            mockIterator := &MockQueryIterator{
                Results: mockResults,
            }
            
            // Налаштування поведінки мок-об'єкта
            mockStub.On("GetStateByRange", "event:", "event~").Return(mockIterator, nil)
            
            // Виклик методу з параметрами фільтрації
            contract := new(SmartContract)
            result, err := contract.QueryEvents(mockContext, tc.queryParams)
            
            // Перевірка результатів
            assert.Nil(t, err)
            assert.Len(t, result, tc.expectedEvents)
            
            // Перевірка фільтрації за типом події
            if tc.eventType != "" {
                for _, event := range result {
                    assert.Equal(t, tc.eventType, event.Type)
                }
            }
            
            // Перевірка фільтрації за актором
            if tc.actor != "" {
                for _, event := range result {
                    assert.Equal(t, tc.actor, event.Actor)
                }
            }
            
            mockStub.AssertExpectations(t)
            mockContext.AssertExpectations(t)
        })
    }
}