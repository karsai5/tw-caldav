
watch:
	nodemon --exec go run tw-caldav.go sync --signal SIGTERM

task-dev-mode:
	echo "data.location=~/.task-dev" > ~/.taskrc

task-normal-mode:
	yadm alt
