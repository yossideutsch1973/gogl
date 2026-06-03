# Contributing to GoGL

Thanks for your interest in improving GoGL! Contributions of all kinds are
welcome — bug reports, feature requests, documentation, and code.

## Development setup

GoGL uses CGO bindings to OpenGL and GLFW, so you need the native OpenGL and
X11 development headers installed.

**Linux (Debian/Ubuntu):**

```bash
sudo apt-get install -y libx11-dev libgl1-mesa-dev libxcursor-dev \
  libxinerama-dev libxi-dev libxrandr-dev libxxf86vm-dev pkg-config
```

**macOS:** OpenGL and the required frameworks ship with the system; no extra
packages are needed.

**Windows:** Use the MinGW toolchain (provides the CGO compiler go-gl needs).

## Before you open a pull request

Run the same checks CI runs:

```bash
gofmt -l .          # must print nothing
go vet ./...
go build ./...
go build ./cmd/examples/...
go test ./tests/unit/...   # requires an OpenGL context (Xvfb in CI)
```

Code must be `gofmt`-clean — CI fails otherwise.

## Tests and OpenGL

The unit tests require a live OpenGL context (a GLFW window). On headless
machines and in CI they run under Xvfb with Mesa's software renderer:

```bash
Xvfb :99 -screen 0 1280x720x24 &
DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 \
  MESA_GL_VERSION_OVERRIDE=4.1 MESA_GLSL_VERSION_OVERRIDE=410 \
  go test ./tests/unit/...
```

## Guidelines

- Keep changes focused; one logical change per pull request.
- Match the surrounding code style and naming.
- Update `README.md` and `CHANGELOG.md` when behavior or the public API changes.
- New shaders go under `shaders/<type>/` and should be documented in
  `shaders/README.md`.

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE).
