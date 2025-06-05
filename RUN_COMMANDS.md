# MongoDB Unit of Work System - Run Commands

## Quick Test Commands

### 1. Verify Architecture (No Database Required)
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go run test_architecture_simple.go
```

### 2. Run Layered Architecture Demo
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go run examples/layered_architecture_demo.go
```

### 3. Build All Packages
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go build ./...
```

### 4. Run Unit Tests
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go test ./pkg/... -v
```

### 5. Run Tests with Coverage
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go test -cover ./pkg/...
```

### 6. Run Performance Benchmarks
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go test -bench=. ./pkg/mongodb
```

### 7. Install Dependencies
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go mod download
go mod tidy
```

### 8. Format Code
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go fmt ./...
```

### 9. Check for Issues
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go vet ./...
```

### 10. Run Integration Tests (Requires MongoDB)
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
go test ./test/... -v
```

## Expected Outputs

### Architecture Verification
```
Architecture Verification
User validation working
Product validation working
Layered architecture verified successfully!
```

### Demo Output (Offline Mode)
```
MongoDB Layered Architecture Demo
=================================
Demonstrating: Service → Repository → Base Repository → Unit of Work → Database

Configuration: localhost:27017/layered_architecture_demo
Unit of Work factories created
Base repositories created
Specific repositories created
Service layer created

Testing MongoDB Connection:
MongoDB operation failed (expected if no MongoDB): failed to create unit of work

Demonstrating Layered Architecture (Offline Mode):
This demonstrates the architectural layers without requiring MongoDB:

The layered architecture is working correctly!
Connect to MongoDB to see full database operations.
```

### Unit Tests
```
=== RUN   TestConfig_Validate
--- PASS: TestConfig_Validate (0.00s)
=== RUN   TestConfig_ConnectionString  
--- PASS: TestConfig_ConnectionString (0.00s)
=== RUN   TestIdentifier_ToBSON
--- PASS: TestIdentifier_ToBSON (0.00s)
=== RUN   TestNewConfig
--- PASS: TestNewConfig (0.00s)
=== RUN   TestFactory_Create
--- PASS: TestFactory_Create (0.00s)
=== RUN   TestBaseEntity_Methods
--- PASS: TestBaseEntity_Methods (0.00s)
=== RUN   TestQueryParams_Validate
--- PASS: TestQueryParams_Validate (0.00s)
PASS
ok      github.com/arash-mosavi/mongo-unit-of-work-system/pkg/mongodb
```

## Development Commands

### Create New Service
```bash
# Template for new service implementation
mkdir pkg/services/new_service
touch pkg/services/new_service/service.go
```

### Add New Repository
```bash
# Template for new repository
touch pkg/mongodb/new_repository.go
```

### Add Tests
```bash
# Add test file
touch pkg/mongodb/new_feature_test.go
```

## Docker Commands (Optional)

### Start MongoDB Container
```bash
docker run --name mongo-uow -p 27017:27017 -d mongo:latest
```

### Stop MongoDB Container
```bash
docker stop mongo-uow
docker rm mongo-uow
```

## Production Deployment

### Build for Production
```bash
cd /media/arash/670edafe-2f0a-4a79-b60c-29b3cde041b61/projects/d-self-projects/Bug-Hunters-shiraz/mongo-unit-of-work-system
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go
```

### Cross-Platform Build
```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o main.exe ./cmd/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o main-mac ./cmd/main.go

# Linux ARM
GOOS=linux GOARCH=arm64 go build -o main-arm ./cmd/main.go
```
