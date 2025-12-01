.PHONY: mock
mock:
	@mockgen -source=.\internal\service\careful\system\user.go -package=svcmocks -destination=.\internal\service\careful\mocks\user.mock.go
	@mockgen -source=.\internal\service\careful\system\dept.go -package=svcmocks -destination=.\internal\service\careful\mocks\dept.mock.go
	@mockgen -source=.\internal\service\careful\system\post.go -package=svcmocks -destination=.\internal\service\careful\mocks\post.mock.go
	@mockgen -source=.\internal\repository\repository\careful\system\dept.go -package=repomocks -destination=.\internal\repository\repository\careful\mocks\dept.mock.go
	@mockgen -source=.\internal\repository\repository\careful\system\post.go -package=repomocks -destination=.\internal\repository\repository\careful\mocks\post.mock.go
	@go mod tidy