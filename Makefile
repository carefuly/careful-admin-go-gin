.PHONY: mock
mock:
	@mockgen -source=.\internal\service\careful\tools\dict.go -package=svcmocks -destination=.\internal\service\careful\mocks\dict.mock.go
	@mockgen -source=.\internal\service\careful\system\user.go -package=svcmocks -destination=.\internal\service\careful\mocks\user.mock.go
	@go mod tidy



