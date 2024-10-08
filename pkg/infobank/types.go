package infobank

import (
	"project.com/pkg/elevator"
)

type ElevatorInfo struct {
	Id                  string
	Requests            [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	Lights              [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	State               elevator.State
}
