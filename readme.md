# ReadRequest

Experimental HTTP1 header parser with aim of reducing number of allocations, whilst still getting a `http.Request`.

* Total number of allocations is independant of the number of header keys.

Currently for the QuickTest case benchmarked against `http.ReadRequest()` it has favourable results, but implementation is still incomplete.

```
BenchmarkReadRequest-4         	  364293	      2906 ns/op	    1697 B/op	       7 allocs/op
BenchmarkStdlibReadRequest-4   	  255628	      4244 ns/op	    1673 B/op	      19 allocs/op
```

## TODO

### Initialise `http.Request` properties
 - #### Body
	`http.body` & `io.LimitedReader` (~2 additional allocations)
 - #### GetBody

