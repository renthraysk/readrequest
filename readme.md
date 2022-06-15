# ReadRequest

Experimental HTTP1 header parser with aim of reducing number of allocations, whilst still getting a `http.Request`.

* Total number of allocations is independant of the number of header keys present.

## TODO

### Initialise `http.Request` properties
 - #### Body
	`http.body` & `io.LimitedReader` (~2 additional allocations)
 - #### GetBody

