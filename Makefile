.PHONY: test test-accesscontrol test-securityaudit test-keymanagement

# Запуск всіх тестів
test: test-accesscontrol test-securityaudit test-keymanagement

# Тестування smарт-контракту управління доступом
test-accesscontrol:
	cd chaincode/accesscontrol/go && go test -v

# Тестування smарт-контракту аудиту безпеки
test-securityaudit:
	cd chaincode/securityaudit/go && go test -v

# Тестування smарт-контракту управління ключами
test-keymanagement:
	cd chaincode/keymanagement/go && go test -v

# Очищення тимчасових файлів
clean:
	find . -name "*.test" -delete
	find . -name "*.out" -delete
