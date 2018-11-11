NAME:=watcher
default:
	go build -o $(NAME)

run:
	./$(NAME)
