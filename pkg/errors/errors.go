package errors

import (
	"errors"
	"fmt"
)

// Common error types for the Unit of Work pattern
var (
	// Transaction errors
	ErrTransactionNotStarted     = errors.New("transaction has not been started")
	ErrTransactionAlreadyOpen    = errors.New("transaction is already open")
	ErrTransactionCommitFailed   = errors.New("failed to commit transaction")
	ErrTransactionRollbackFailed = errors.New("failed to rollback transaction")

	// Entity errors
	ErrEntityNotFound   = errors.New("entity not found")
	ErrEntityExists     = errors.New("entity already exists")
	ErrInvalidEntity    = errors.New("invalid entity")
	ErrEntityValidation = errors.New("entity validation failed")

	// Repository errors
	ErrRepositoryNotFound    = errors.New("repository not found")
	ErrInvalidRepositoryType = errors.New("invalid repository type")
	ErrRepositoryOperation   = errors.New("repository operation failed")

	// Database errors
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseTimeout    = errors.New("database operation timeout")
	ErrDatabaseConstraint = errors.New("database constraint violation")
	ErrDatabaseDeadlock   = errors.New("database deadlock detected")

	// Query errors
	ErrInvalidQuery       = errors.New("invalid query")
	ErrQueryExecution     = errors.New("query execution failed")
	ErrInvalidQueryParams = errors.New("invalid query parameters")
)

// UnitOfWorkError wraps errors with context information
// Provides structured error handling for debugging and monitoring
type UnitOfWorkError struct {
	Op     string    // Operation that failed
	Entity string    // Entity type involved
	Err    error     // Underlying error
	Code   ErrorCode // Error classification
}

// ErrorCode categorizes errors for better handling
type ErrorCode int

const (
	CodeUnknown ErrorCode = iota
	CodeValidation
	CodeNotFound
	CodeExists
	CodeConstraint
	CodeTransaction
	CodeConnection
	CodeTimeout
	CodeDeadlock
)

// Error implements the error interface
func (e *UnitOfWorkError) Error() string {
	if e.Entity != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Entity, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap returns the underlying error for error unwrapping
func (e *UnitOfWorkError) Unwrap() error {
	return e.Err
}

// Is implements error matching for errors.Is()
func (e *UnitOfWorkError) Is(target error) bool {
	if t, ok := target.(*UnitOfWorkError); ok {
		return e.Code == t.Code
	}
	return errors.Is(e.Err, target)
}

// NewUnitOfWorkError creates a new structured error
func NewUnitOfWorkError(op, entity string, err error, code ErrorCode) *UnitOfWorkError {
	return &UnitOfWorkError{
		Op:     op,
		Entity: entity,
		Err:    err,
		Code:   code,
	}
}

// Wrap wraps an error with operation context
func Wrap(err error, op string) error {
	if err == nil {
		return nil
	}
	return &UnitOfWorkError{
		Op:   op,
		Err:  err,
		Code: CodeUnknown,
	}
}

// WrapWithEntity wraps an error with operation and entity context
func WrapWithEntity(err error, op, entity string) error {
	if err == nil {
		return nil
	}
	return &UnitOfWorkError{
		Op:     op,
		Entity: entity,
		Err:    err,
		Code:   CodeUnknown,
	}
}

// IsNotFound checks if the error is a "not found" error
func IsNotFound(err error) bool {
	var uowErr *UnitOfWorkError
	if errors.As(err, &uowErr) {
		return uowErr.Code == CodeNotFound
	}
	return errors.Is(err, ErrEntityNotFound)
}

// IsValidation checks if the error is a validation error
func IsValidation(err error) bool {
	var uowErr *UnitOfWorkError
	if errors.As(err, &uowErr) {
		return uowErr.Code == CodeValidation
	}
	return errors.Is(err, ErrEntityValidation)
}

// IsConstraint checks if the error is a constraint violation
func IsConstraint(err error) bool {
	var uowErr *UnitOfWorkError
	if errors.As(err, &uowErr) {
		return uowErr.Code == CodeConstraint
	}
	return errors.Is(err, ErrDatabaseConstraint)
}

// IsTransaction checks if the error is transaction-related
func IsTransaction(err error) bool {
	var uowErr *UnitOfWorkError
	if errors.As(err, &uowErr) {
		return uowErr.Code == CodeTransaction
	}
	return errors.Is(err, ErrTransactionNotStarted) ||
		errors.Is(err, ErrTransactionAlreadyOpen) ||
		errors.Is(err, ErrTransactionCommitFailed) ||
		errors.Is(err, ErrTransactionRollbackFailed)
}

// IsConnection checks if the error is connection-related
func IsConnection(err error) bool {
	var uowErr *UnitOfWorkError
	if errors.As(err, &uowErr) {
		return uowErr.Code == CodeConnection
	}
	return errors.Is(err, ErrDatabaseConnection)
}

// IsTimeout checks if the error is timeout-related
func IsTimeout(err error) bool {
	var uowErr *UnitOfWorkError
	if errors.As(err, &uowErr) {
		return uowErr.Code == CodeTimeout
	}
	return errors.Is(err, ErrDatabaseTimeout)
}

// IsDeadlock checks if the error is deadlock-related
func IsDeadlock(err error) bool {
	var uowErr *UnitOfWorkError
	if errors.As(err, &uowErr) {
		return uowErr.Code == CodeDeadlock
	}
	return errors.Is(err, ErrDatabaseDeadlock)
}
