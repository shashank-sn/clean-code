# queue shutdown

Once `Enqueue` accepts a job, `Stop` must wait until that job has been handled.
The command uses this guarantee before reporting a successful shutdown.
