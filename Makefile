
watch:
	nodemon --exec go run tw-caldav.go sync --signal SIGTERM

task-dev-mode:
	cp ./taskrc ~/.taskrc

task-normal-mode:
	yadm alt
