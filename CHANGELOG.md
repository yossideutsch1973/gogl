# Changelog

## v0.3.0 - Showcase demo, cross-platform CI, test reliability
- Add `cmd/examples/showcase`: fullscreen raymarched scene driven by a single fragment shader, with `-screenshot` flag for headless PNG capture
- Add `shaders/fragment/raymarch_showcase.frag` and `shaders/vertex/fullscreen_triangle.vert` (no-VBO fullscreen triangle)
- README now embeds a screenshot rendered headlessly from the showcase example
- CI: split into `lint`, `test-linux`, `test-macos`, `test-windows` jobs; add `go vet` and `staticcheck`; render the showcase PNG on every Linux build and upload it as an artifact
- Replace deprecated `gl.PtrOffset` calls with the `WithOffset` overloads in `pkg/resource/vertex_array.go` and the basic/compute examples
- Drop unused log-buffer `sync.Pool` in `pkg/shader` (triggered SA6002) in favor of a small per-call helper
- Add package doc comments to `pkg/pipeline` and `pkg/resource`
- Fix flaky GL tests: Go's test runner schedules individual tests on a different OS thread than `TestMain`; the GL context was stuck on the `TestMain` thread. Each test now acquires/releases the context via a `glSetup(t)` helper that calls `runtime.LockOSThread` + `MakeContextCurrent` and cleans up with `DetachCurrentContext` + `UnlockOSThread`

## v0.2.0 - Project cleanup and CI fixes
- Rewrite README to accurately reflect implemented features
- Fix CI workflow: add Mesa software renderer, Xvfb wait, example build step
- Remove development artifacts (CLAUDE.md, EXPERT_REVIEW.md, copilot-instructions.md)
- Remove local issue tracking (issues/ directory)
- Update Go version to 1.22
- Add MIT license

## v0.1.0 - Initial release
- Core shader compilation, linking, and program management
- Pipeline state management with builder pattern and state caching
- Resource management: VBO, IBO, UBO, VAO, textures, meshes, buffer pools
- Platform detection: GPU vendor, OpenGL version, hardware capabilities
- 28 production GLSL shaders (vertex, fragment, geometry, compute)
- 6 example applications (basic, compute, geometry, pipeline, platform, shader test)
- Complete unit test suite
