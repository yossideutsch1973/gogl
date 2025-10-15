# Shader Test Example

This example validates that all shaders in the GoGL shader library compile and link correctly.

## What it does

- Tests compilation of all vertex/fragment shader pairs
- Tests geometry shaders with appropriate vertex/fragment shaders
- Tests post-processing fragment shaders
- Reports pass/fail status for each shader

## Running

```bash
go run cmd/examples/shader_test/main.go
```

## Expected Output

```
Testing GoGL Shader Library
OpenGL Version: 4.1 ...

Testing: Basic
  ✅ Success
Testing: Flat Color
  ✅ Success
...

Test Summary: 17 passed, 0 failed
```

## Note

This example requires a display and OpenGL context. It won't run in headless CI environments without X11/Xvfb.
