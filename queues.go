package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	// CLI execution
	"github.com/codeskyblue/go-sh"
	// File path retirever
	"github.com/kardianos/osext"
	// Config file reader
	"github.com/spf13/viper"
	// Daemon
	"github.com/takama/daemon"
)

var wg sync.WaitGroup

func main() {
	queues := readConfigFile()

	service, serviceError := daemon.New("queue-daemon", "queue consumer")
	logError(serviceError, "Error initializing daemon: ")

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			installed, err := service.Install()
			logError(err, installed)
		case "remove":
			removed, err := service.Remove()
			logError(err, removed)
		case "start":
			started, err := service.Start()
			logError(err, started)
		case "stop":
			stopped, err := service.Stop()
			logError(err, stopped)
		case "status":
			status, err := service.Status()
			logError(err, status)
			fmt.Println(status)
		case "dry-run":
			wg.Add(len(queues))
			for i := 0; i < len(queues); i++ {
				go dryRun(queues[i])
			}
			wg.Wait()
		default:
			fmt.Println("Enter a valid input: install | start | stop | remove | status | dry-run")
		}
		return
	} // Only the code below will run as a daemon

	consumerController := consumeQueues(queues)
	waitForSignal()
	close(consumerController)
	wg.Wait()
	log.Println("All queues stopped")
}

func readConfigFile() []string {
	projectDirectory, pathError := osext.ExecutableFolder()
	logError(pathError, projectDirectory)
	viper.SetConfigName("config")
	viper.AddConfigPath(projectDirectory)
	readError := viper.ReadInConfig()
	logError(readError, "Error reading config.yaml file: ")
	queues := viper.AllKeys()
	return queues
}

func consumeQueues(queues []string) chan bool {
	consumerController := make(chan bool)
	wg.Add(len(queues))
	log.Println("Starting daemon")
	for i := 0; i < len(queues); i++ {
		go func(queueName string, consumerController chan bool) {
			processes, timeInterval, environment, project := getParameters(queueName)
			for j := 0; j < processes; j++ {
				go spawnProcesses(queueName, timeInterval, environment, project, consumerController)
			}
			wg.Done()
		}(queues[i], consumerController)
	}
	return consumerController
}

func spawnProcesses(queueName string, timeInterval int, environment string, project string, consumerController chan bool) {
	for {
		select {
		case <-consumerController:
			return
		default:
		}
		executionError := sh.Command("/uni/"+project+"/app/console", "uecode:qpush:receive", queueName, "--no-debug", "--env="+environment).Run()
		logError(executionError, "CLI execution error: ")
		select {
		case <-consumerController:
			return
		case <-time.After(time.Second * time.Duration(timeInterval)):
		}
	}
}

func dryRun(queueName string) {
	defer wg.Done()
	processes, timeInterval, environment, project := getParameters(queueName)
	for {
		for i := 0; i < processes; i++ {
			executionString := "/uni/" + project + "/app/console uecode:qpush:receive " + queueName + " --no-debug --env=" + environment
			fmt.Println(executionString)
		}
		time.Sleep(time.Second * time.Duration(timeInterval))
	}
}

func getParameters(queueName string) (int, int, string, string) {
	queue := viper.GetStringMapString(queueName)
	processes, processesError := strconv.Atoi(queue["processes"])
	logError(processesError, "Error converting the number of processes to a string")
	timeInterval, timeIntervalError := strconv.Atoi(queue["interval"])
	logError(timeIntervalError, "Error converting the time interval to a string")
	environment := queue["environment"]
	project := queue["project"]
	if processes < 1 || timeInterval < 1 || queueName == "" || environment == "" || project == "" {
		panic("Error: Enter valid parameters on the config.yaml file")
	}
	return processes, timeInterval, environment, project
}

func waitForSignal() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	log.Println("SIGNAL: ", <-sigs)
}

func logError(err error, comment string) {
	if err != nil {
		log.Println(comment, "\n", err)
	}
}
