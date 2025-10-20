.PHONY: mock
mock:
	@mockgen -source=.\internal\service\careful\system\user.go -package=svcmocks -destination=.\internal\service\careful\mocks\user.mock.go
	@mockgen -source=.\internal\service\careful\tools\dict.go -package=svcmocks -destination=.\internal\service\careful\mocks\dict.mock.go
	@mockgen -source=.\internal\repository\repository\careful\tools\dict.go -package=repomocks -destination=.\internal\repository\repository\careful\mocks\dict.mock.go
	@mockgen -source=.\internal\service\careful\tools\dict_type.go -package=svcmocks -destination=.\internal\service\careful\mocks\dict_type.mock.go
	@mockgen -source=.\internal\repository\repository\careful\tools\dict_type.go -package=repomocks -destination=.\internal\repository\repository\careful\mocks\dict_type.mock.go
	@go mod tidy



