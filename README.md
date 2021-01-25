# Introduce
Gin driver for [zeta](https://github.com/zeta-io/zeta).
# Feature
Features:
 - Parameters of the assembly
 - Parameter validation
# Usage
Sample:
```go

```
Request parameter tag:
 - query:"${name},${default}": Bind query parameters.
 - path:"${name},${default}": Bind path parameters.
 - body:"${name},${default}": Bind form parameters.
 - header:"${name},${default}": Bind body parameters.
 - cookie:"${name},${default}": Bind body parameters.
 - file:"${name},${default}": Bind body parameters.
 

# Validator
It's used [https://github.com/go-playground/validator](https://github.com/go-playground/validator) by default.

