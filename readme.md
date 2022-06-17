# ReadRequest

Experimental HTTP1 header parser with aim of reducing number of allocations, whilst still getting a `http.Request`.

* Total number of allocations is independant of the number of header keys.

Currently for the QuickTest case benchmarked against `http.ReadRequest()` it has favourable results, but implementation is still incomplete.

```
BenchmarkReadRequest-4         	  387646	      2759 ns/op	    1649 B/op	       6 allocs/op
BenchmarkStdlibReadRequest-4   	  255250	      4225 ns/op	    1673 B/op	      19 allocs/op
```

## TODO

### Initialise `http.Request` properties
 - #### Body
	`http.body` & `io.LimitedReader` (~2 additional allocations)
 - #### GetBody

