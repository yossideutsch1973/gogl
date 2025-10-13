# CI Job Fails: Unused Import and Missing DISPLAY for GLFW Tests

## Description:
The CI workflow is currently failing due to two primary issues:

1. **Unused Import**  
   - **File:** pkg/shader/shader.go (ref d46d58f6d09e6000c29561c0bad7177eff373b47)  
   - **Error:** "strings" imported and not used  
   - **Solution:** Remove the unused "strings" import statement.

2. **GLFW/X11 Failure in CI**  
   - **Error:** PlatformError: X11: The DISPLAY environment variable is missing and panic: NotInitialized: The GLFW library is not initialized  
   - **Cause:** Tests requiring an X server (OpenGL/GLFW context) are running in a headless environment.  
   - **Solution:** Either configure CI to use a virtual framebuffer (e.g., xvfb), or skip/mask tests that require a graphical environment when running in CI.

## Action Items:
- Remove the unused import from pkg/shader/shader.go.
- Update the CI configuration to support headless OpenGL tests or skip them in CI.