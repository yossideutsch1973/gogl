# Changelog

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
