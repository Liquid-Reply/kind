# krustlet

This directory contains sources for building the `kind` krustlet "node" image.

>At this stage the Dockerfile is just a striped version of `images/base` and the Krustlet is downloaded directly in the dockerfile. `kind build krustlet-image` is not yet possible.

## Building

The image can be built with `make quick`.
