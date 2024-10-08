package elevator

import (
	"project.com/pkg/timer"
)

func FSM(
	elevatorInit_ch 		     chan Elevator,
	requestUpdate_ch 		     chan [N_FLOORS][N_BUTTONS]bool,
	clearRequestToInfobank_ch    chan []ButtonEvent,
	stateToInfobank_ch 		     chan State,
	lightUpdate_ch               chan [N_FLOORS][N_BUTTONS]bool,
	obstructedStateToInfobank_ch chan bool,
	updateDiagnostics_ch 		 chan Elevator,
	obstructionDiagnose_ch       chan bool) {

	floorSensor_ch := make(chan int)
	obstruction_ch := make(chan bool)
	doorTimer_ch := make(chan bool)

	go PollFloorSensor(floorSensor_ch)
	go PollObstructionSwitch(obstruction_ch)

	elevator := new(Elevator)
	*elevator = <-elevatorInit_ch

	for {
		select {
		case requests := <- requestUpdate_ch:
			
			elevator.Requests = requests
			fsmNewRequests(elevator, doorTimer_ch)

			if elevator.Requests != requests {
				clearRequestToInfobank_ch <- getClearedRequests(requests, elevator.Requests)
			}

			stateToInfobank_ch <- elevator.State
			updateDiagnostics_ch <- *elevator

		case lights := <- lightUpdate_ch:
			
			elevator.Lights = lights
			setAllLights(elevator)

		case newFloor := <- floorSensor_ch:
			
			requestsBeforeNewFloor := elevator.Requests
			fsmOnFloorArrival(elevator, newFloor, doorTimer_ch)
			stateToInfobank_ch <- elevator.State

			if requestsBeforeNewFloor != elevator.Requests {
				clearRequestToInfobank_ch <- getClearedRequests(requestsBeforeNewFloor, elevator.Requests)
			}

			updateDiagnostics_ch <- *elevator

		case <- doorTimer_ch:
			
			requestsBeforeNewFloor := elevator.Requests
			handleDeparture(elevator, doorTimer_ch)
			stateToInfobank_ch <- elevator.State
			updateDiagnostics_ch <- *elevator

			if elevator.Requests != requestsBeforeNewFloor {
				clearRequestToInfobank_ch <- getClearedRequests(requestsBeforeNewFloor, elevator.Requests)
				setAllLights(elevator)
			}

		case obstruction := <- obstruction_ch:
			
			if !obstruction && elevator.State.Behaviour == EB_DoorOpen {
				go timer.Run_timer(3, doorTimer_ch)

				if elevator.State.OutOfService {
					elevator.State.OutOfService = false
					updateDiagnostics_ch <- *elevator
					obstructedStateToInfobank_ch <- false
				}
			}

		case <- obstructionDiagnose_ch:
			
			elevator.State.OutOfService = true
			obstructedStateToInfobank_ch <- true
		}
	}
}

func fsmNewRequests(e *Elevator, doorTimer_ch chan bool) {

	if e.State.Behaviour == EB_DoorOpen {
		if requestShouldClearImmediately(*e) {
			requestsAndLightsClearAtCurrentFloor(e)
			go timer.Run_timer(3, doorTimer_ch)
			setAllLights(e)
		}
		return
	}

	e.State.Dirn, e.State.Behaviour = getDirectionAndBehaviour(e)
	switch e.State.Behaviour {

	case EB_DoorOpen:
		SetDoorOpenLamp(true)
		go timer.Run_timer(3, doorTimer_ch)
		requestsAndLightsClearAtCurrentFloor(e)

	case EB_Moving:
		SetMotorDirection(e.State.Dirn)
	}
	setAllLights(e)
}

func handleDeparture(e *Elevator, doorTimer_ch chan bool) {
	if GetObstruction() && e.State.Behaviour == EB_DoorOpen {
		go timer.Run_timer(3, doorTimer_ch)
		return
	}

	e.State.Dirn, e.State.Behaviour = getDirectionAndBehaviour(e)

	switch e.State.Behaviour {

	case EB_DoorOpen:
		SetDoorOpenLamp(true)
		requestsAndLightsClearAtCurrentFloor(e)
		setAllLights(e)
		go timer.Run_timer(3, doorTimer_ch)

	case EB_Moving:
		SetMotorDirection(e.State.Dirn)
		SetDoorOpenLamp(false)

	case EB_Idle:
		SetDoorOpenLamp(false)
	}
}

func fsmOnFloorArrival(e *Elevator, newFloor int, doorTimer_ch chan bool) {
	e.State.Floor = newFloor
	SetFloorIndicator(newFloor)
	setAllLights(e)

	if requestShouldStop(*e) {
		SetMotorDirection(MD_Stop)
		SetDoorOpenLamp(true)
		requestsAndLightsClearAtCurrentFloor(e)
		go timer.Run_timer(3, doorTimer_ch)
		e.State.Behaviour = EB_DoorOpen
		e.State.Dirn = MD_Stop
		setAllLights(e)
	}
}

func setAllLights(e *Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			SetButtonLamp(ButtonType(btn), floor, e.Lights[floor][btn])
		}
	}
}

func getClearedRequests(oldRequests [N_FLOORS][N_BUTTONS]bool, newRequests [N_FLOORS][N_BUTTONS]bool) []ButtonEvent {
	var clearedRequests []ButtonEvent
	for i := 0; i < N_FLOORS; i++ {
		for j := 0; j < N_BUTTONS; j++ {
			if oldRequests[i][j] != newRequests[i][j] {
				clearedRequests = append(clearedRequests, ButtonEvent{i, ButtonType(j)})
			}
		}
	}
	return clearedRequests
}
