## Issue: Failing CI Job in PR #7

The CI job in PR #7 is failing due to an issue in the unit test `TestPipelineStateApplication`. The error message is:

```
Failed to compile vertex shader: failed to create shader: OpenGL context may not be initialized.
```

This issue arises because there is a lack of a display/OpenGL context in the CI environment.

### Recommendation
To resolve this issue, it's recommended to start Xvfb before running the tests in the workflow file. This will provide a virtual display for OpenGL context creation.

### Suggested YAML Snippet for the Workflow File:
```yaml
- name: Start Xvfb
  run: Xvfb :99 -screen 0 1024x768x24 &
- name: Set DISPLAY
  run: export DISPLAY=:99
- name: Run tests
  run: go test ./...
```

### Reference
For more details, see the CI job: [Job Link](https://github.com/yossideutsch1973/gogl/actions/runs/18497830881/job/52745494965?pr=7)