Solution to the elevator project in TTK4145. The solution is a peer-to-peer system with a mesh topology, where assignments are distrubited through an executable file, [hall_request_assigner v1.1.1](https://github.com/TTK4145/Project-resources/releases/tag/v1.1.1). This file is to be downloaded and placed in the "cmd" folder. To run the program, first start the elevator server through the command "elevatorserver" in the terminal on n different computers. Then, on the same computers, run the command "go run main.go port" in another terminal, where "port" is to be replaced with the port number that each elevator will use for communication between the primary and the backup processes, e.g. 20030. Give the port number without quotes in the command. It is important that you run the command with different port numbers on different computers, to ensure the backup processes only listen for heartbeats from the primary processes on their own computers. 

Authors: 
Trygve Tafjord,
Ole Mandius Harm Thorrud,
Per Martin Herdlevær
