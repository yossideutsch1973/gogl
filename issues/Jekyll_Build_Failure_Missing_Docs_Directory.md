# Jekyll Build Failure: Missing docs Directory

## Description:
The GitHub Pages automatic deployment is failing due to a missing `docs` directory. GitHub Pages is configured to build a Jekyll site from the `./docs` source directory, but this directory does not exist in the repository.

## Error Details:

**Workflow:** pages build and deployment  
**Job:** build ([Job #52629774236](https://github.com/yossideutsch1973/gogl/actions/runs/18472591456/job/52629774236))  
**Status:** Failed  
**Date:** 2025-10-13T16:44:06Z

### Key Error Messages:

```
Configuration file: none
Conversion error: Jekyll::Converters::Scss encountered an error while converting 'assets/css/style.scss':
                    No such file or directory @ dir_chdir0 - /github/workspace/docs
```

```
Source: /github/workspace/./docs
Destination: /github/workspace/./docs/_site
...
Error: No such file or directory @ dir_chdir0 - /github/workspace/docs
```

### Technical Details:

The Jekyll build action (`actions/jekyll-build-pages@v1`) is configured with:
- **source:** `./docs`
- **destination:** `./docs/_site`

However, the repository structure does not include a `docs` directory, causing the build process to fail when Jekyll attempts to access it.

## Root Cause:

GitHub Pages is likely configured in the repository settings to:
- **Source:** Deploy from a branch
- **Branch:** main
- **Folder:** /docs

But the repository does not contain this directory.

## Possible Solutions:

1. **Create the docs directory structure:**
   - Create a `docs/` directory in the repository root
   - Add appropriate Jekyll configuration (`_config.yml`)
   - Add content files (index.md, etc.)

2. **Change GitHub Pages settings:**
   - Navigate to repository Settings → Pages
   - Change the source folder from `/docs` to `/ (root)` if documentation should be built from the root
   - Or disable GitHub Pages if it's not needed for this Go project

3. **Disable automatic GitHub Pages deployment:**
   - If documentation via GitHub Pages is not required for this OpenGL library
   - The project already has comprehensive documentation in README.md and CLAUDE.md

## Recommendation:

Given that this is a Go OpenGL shader library with existing comprehensive documentation in Markdown files (README.md, CLAUDE.md, EXPERT_REVIEW.md), GitHub Pages may not be necessary. Consider disabling GitHub Pages in the repository settings unless there's a specific need for a hosted documentation site.

Alternatively, if GitHub Pages is desired for API documentation, create a proper `docs/` directory structure with:
- Jekyll configuration
- Auto-generated Go documentation (using tools like `godoc` or `pkgsite`)
- Project website content

## Action Items:
- [ ] Review repository Settings → Pages configuration
- [ ] Decide whether GitHub Pages should be used for this project
- [ ] Either:
  - Create proper `docs/` directory structure with Jekyll site, OR
  - Disable GitHub Pages in repository settings, OR  
  - Change Pages source to root directory and add appropriate Jekyll files

## References:
- Full job logs: https://github.com/yossideutsch1973/gogl/actions/runs/18472591456/job/52629774236
- Workflow run: https://github.com/yossideutsch1973/gogl/actions/runs/18472591456
- Commit: bd645c1651e35082684da13a64d627db70f722a5
