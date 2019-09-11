package controller

import (
	"github.com/wtfjoke/pod-replica/pkg/controller/podreplica"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, podreplica.Add)
}
