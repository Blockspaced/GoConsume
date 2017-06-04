# Queue Consumer   
A small daemon with multithreading to consume several queues from sqs.
This tool reads input from `config.yaml` and based on those parameters it will execute X processes every Y seconds on a specific environment for N queues (in parallel).
This service will continue to run on the background until you stop it. It will also properly stop any running goroutines in case of a kill signal.

**Follow the steps:**
* Open `config.yaml`, enter valid parameters for at least one queue and save it
* Open terminal and enter `go build queues.go`
* Enter `sudo ./queues` followed by one of commands below
* (*You have to use `sudo` and enter your password because you need root privileges in order to install the service for the daemon*)

**Commands:**
* install - Installs a service named `queueDaemon`
* start - Starts the service
* stop - Stops the service
* remove - Removes the service
* status - Display the current status of the service *(installed or not, running or stopped)*
* dry-run - Prints the processes that would be executed if the users selects "run" *(used for testing)*

**Logs:**
* Logs will be saved on `/usr/local/var/log`
* To see the logs open a second terminal window and enter `sudo tail -f /usr/local/var/log/queue-daemon.err`

**Warning:**
* Don't use the `remove` command while the service is still running, otherwise you won't be able to stop it until you use `install` again
