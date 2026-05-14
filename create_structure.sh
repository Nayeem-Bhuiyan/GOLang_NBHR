#!/bin/bash

# Create directories
mkdir -p cmd/api
mkdir -p config
mkdir -p internal/bootstrap
mkdir -p internal/constants
mkdir -p internal/domain/entity
mkdir -p internal/domain/errors
mkdir -p internal/dto
mkdir -p internal/middleware
mkdir -p internal/modules/auth
mkdir -p internal/modules/user
mkdir -p internal/modules/role
mkdir -p internal/modules/permission
mkdir -p internal/shared/response
mkdir -p internal/shared/pagination
mkdir -p internal/shared/filter
mkdir -p internal/shared/validator
mkdir -p internal/shared/crypto
mkdir -p internal/shared/jwt
mkdir -p internal/shared/logger
mkdir -p internal/router
mkdir -p migrations

# Create empty files
touch cmd/api/main.go
touch config/config.go
touch config/database.go
touch internal/bootstrap/app.go
touch internal/constants/constants.go
touch internal/domain/entity/user.go
touch internal/domain/entity/role.go
touch internal/domain/entity/permission.go
touch internal/domain/entity/token.go
touch internal/domain/errors/errors.go
touch internal/dto/auth_dto.go
touch internal/dto/user_dto.go
touch internal/dto/role_dto.go
touch internal/dto/permission_dto.go
touch internal/middleware/auth.go
touch internal/middleware/rbac.go
touch internal/middleware/cors.go
touch internal/middleware/logger.go
touch internal/middleware/request_id.go
touch internal/middleware/rate_limit.go
touch internal/middleware/timeout.go
touch internal/modules/auth/handler.go
touch internal/modules/auth/service.go
touch internal/modules/auth/repository.go
touch internal/modules/auth/routes.go
touch internal/modules/user/handler.go
touch internal/modules/user/service.go
touch internal/modules/user/repository.go
touch internal/modules/user/routes.go
touch internal/modules/role/handler.go
touch internal/modules/role/service.go
touch internal/modules/role/repository.go
touch internal/modules/role/routes.go
touch internal/modules/permission/handler.go
touch internal/modules/permission/service.go
touch internal/modules/permission/repository.go
touch internal/modules/permission/routes.go
touch internal/shared/response/response.go
touch internal/shared/pagination/pagination.go
touch internal/shared/filter/filter.go
touch internal/shared/validator/validator.go
touch internal/shared/crypto/crypto.go
touch internal/shared/jwt/jwt.go
touch internal/shared/logger/logger.go
touch internal/router/router.go
touch migrations/migrate.go
touch .env.example
touch go.mod
touch go.sum
touch README.md

echo "✅ NBHR Empty Project Structure Created Successfully!"
ls -R
