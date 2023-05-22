lork
=====
The flexible, extensible and structured logging for Go. Lork provides bridge and binder
for logger which can send log from logger to another logger you preferred. Lork also
provides unified writers, encoders and filters, it brings different logger with same
apis and flexible configurations.

## Features

* Logging with zero allocation
* Multiple writer
* Multiple structured format
* Flexible configurations

Usage
=====

Look at this [tutorial][1]

Benchmarks
==========
Benchmarks with complex log field, different encoder and writer.

```text
cpu: Intel(R) Core(TM) i7-8550U CPU @ 1.80GHz
BenchmarkJsonFileWriter-8      	  290490        3615 ns/op      0 B/op      0 allocs/op
BenchmarkPatternFileWriter-8   	  243361        4670 ns/op      0 B/op      0 allocs/op
BenchmarkAsyncFileWriter-8     	 1220036        998.8 ns/op     0 B/op      0 allocs/op
BenchmarkNoWriter-8            	 1591646        746.5 ns/op     0 B/op      0 allocs/op
```

Credits
======

* [slf4j][2]: Simple Logging Facade for Java
* [logback][3]: The reliable, generic, fast and flexible logging framework for Java.

Donate
=======
If you enjoy this project and want to support it, you can buy a coffee.

<a href="https://www.buymeacoffee.com/coolerfall" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>


License
=======

    Copyright (c) 2019-2023 Vincent Cheung (coolingfall@gmail.com).
    
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
    
         http://www.apache.org/licenses/LICENSE-2.0
    
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.

[1]: https://lork.coolerfall.com
[2]: https://github.com/qos-ch/slf4j
[3]: https://github.com/qos-ch/logback
