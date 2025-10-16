### Title: CI Failure - OpenGL Context Initialization Errors in Headless Environment

---
**Workflow:** ci-go  
**Reference Job:** https://github.com/yossideutsch1973/gogl/actions/runs/18529569178/job/52890818503?pr=15

**Error:**
```
Failed to compile vertex shader: failed to create shader: OpenGL context may not be initialized
```

**Problem Description:**

The CI workflow fails when running OpenGL/GLFW tests in a headless Ubuntu environment. The tests require an OpenGL context which depends on X11 display server access. Without a virtual display, GLFW cannot create windows or OpenGL contexts, causing shader compilation tests to fail.

This is a common issue when running graphics tests in CI environments that don't have a physical display or GPU access.

**Root Cause:**

1. **Missing X11 Display**: Tests that use GLFW to create OpenGL contexts require an X11 server
2. **Headless CI Environment**: GitHub Actions runners don't have a display by default
3. **OpenGL Context Dependency**: The shader compilation tests in `tests/unit/shader_test.go` and `tests/unit/pipeline/pipeline_test.go` initialize GLFW and create OpenGL contexts in `TestMain`

**Solution:**

Use Xvfb (X Virtual Framebuffer) to provide a virtual display server in the CI environment. Xvfb allows running X11 applications without a physical display, which is perfect for headless testing.

**Implementation Steps:**

1. Install Xvfb and required dependencies:
```yaml
- name: Install X11 and OpenGL dependencies
  run: sudo apt-get update && sudo apt-get install -y libx11-dev libgl1-mesa-dev xorg-dev xvfb
```

2. Start Xvfb before running tests:
```yaml
- name: Start Xvfb
  run: Xvfb :99 -screen 0 1280x1024x24 &
```

3. Set DISPLAY environment variable for test execution:
```yaml
- name: Run tests
  run: go test ./...
  env:
    DISPLAY: :99
```

**Alternative Approach:**

Another option is to set DISPLAY in the environment step:
```yaml
- name: Start Xvfb
  run: |
    sudo apt-get update
    sudo apt-get install -y xvfb
    Xvfb :99 -screen 0 1280x1024x24 &
    export DISPLAY=:99
    echo "DISPLAY=:99" >> $GITHUB_ENV

- name: Run tests
  run: go test ./...
  env:
    DISPLAY: ${{ env.DISPLAY }}
```

**Current Status:**

✅ **RESOLVED** - The CI workflow (`.github/workflows/ci-go.yml`) has been updated with the Xvfb solution and tests now pass successfully.

**Technical Details:**

- **Affected Tests**: All tests that initialize GLFW and create OpenGL contexts
  - `tests/unit/shader_test.go`
  - `tests/unit/pipeline/pipeline_test.go`
  - `tests/unit/resource/resource_test.go`
  
- **OpenGL Version**: Tests use OpenGL 4.1 core profile for maximum compatibility (including macOS)

- **GLFW Configuration**: Tests create hidden windows (`glfw.WindowHint(glfw.Visible, glfw.False)`) specifically for headless testing

**Testing:**

To verify the solution locally:
```bash
# Start Xvfb
Xvfb :99 -screen 0 1280x1024x24 &
export DISPLAY=:99

# Run tests
go test ./tests/unit/...
```

**References:**

- Original CI failure: https://github.com/yossideutsch1973/gogl/actions/runs/18529569178/job/52890818503?pr=15
- Xvfb documentation: https://www.x.org/releases/X11R7.6/doc/man/man1/Xvfb.1.xhtml
- GLFW headless testing: https://www.glfw.org/docs/latest/window_guide.html

---
