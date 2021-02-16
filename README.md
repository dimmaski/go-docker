# Docker restrainer

Mini thang to experiment with go rotines and concurrent code.

This thing tries to perform a certain amount of work using the docker API while restricting the overal CPU usage to not go over a given limit.

Using the `progrium/stress` image.

```sh
go run main.go cpu.go data.go
```
