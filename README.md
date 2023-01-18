# fastgifresize

Universal parallel gif resize example in golang

## algo description
Resizes gif frame by frame in parallel goroutines. 

Synchroneous part is drawing frame onto accumulation image in its default size and position and after that creates async task for resize of the accumualtion image.

The parallelization is done this way beacuse every frame in gif can be different size and resizing it one by one in parallel without reference to its original size and position can produce messy result on final image.

# usage
## run with go
You can configure goroutines number by arg `-poolsize`.
```bash
go run main.go -src=./in.gif -dims=400x400 -dst=./out.gif -poolsize=100
```

## check args description
```bash
go run main.go -h
```