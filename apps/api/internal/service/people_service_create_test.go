package service

import (
	"testing"
)

func TestCreateClientWithoutUser(t *testing.T) {
	// TODO: Implement test for creating a client without system user
}

func TestCreateClientWithUserBySystem(t *testing.T) {
	// TODO: Implement test for creating a client with user (role common) by system user
}

func TestCreateEmployeeWithUser(t *testing.T) {
	// TODO: Implement test for creating an employee with user (role system)
}

func TestCreateOutsourcedEmployeeWithUser(t *testing.T) {
	// TODO: Implement test for creating an outsourced employee with user (role system)
}

func TestBlockSystemCreatingEmployeeOrOutsourced(t *testing.T) {
	// TODO: Implement test to block system user from creating employee or outsourced_employee
}

func TestBlockSystemCreatingUserForSupplier(t *testing.T) {
	// TODO: Implement test to block system user from creating user for supplier
}
