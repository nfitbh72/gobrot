# GoBrot

GoBrot is an interactive command-line fractal generator written in Go. It renders the
Mandelbrot set (and a Burning Ship–style Julia variant) to PNG images, splits the result
into depth layers, and can export each layer as an STL 3D model for printing.

Coordinates are held as `math/big` floats, so you can zoom far deeper than `float64`
precision allows.

## Features

- **Fractals**: Mandelbrot (`mandel`) and Burning Ship Julia (`bsjulia`)
- **Arbitrary precision**: big-float coordinates with configurable precision (`big`)
- **PNG output** with several colour profiles (`gray`, `gray2`, `cycle`, `gpt1`)
- **Layers**: split the escape-time depth into bands, one PNG/STL per band
- **3D export**: STL meshes with base and sides, depth scaling and cut thresholds
- **Canvas effects**: project the fractal onto a shape (`dome`, `bowl`, `semiCylinder`, `test`)
- **Circle trim**: crop the output to a circle
- **ASCII preview** in the terminal
- **XML import** from the [HWS Mandelbrot viewer](https://math.hws.edu/eck/js/mandelbrot/MB.html)

## Requirements

- Go 1.20+

Dependencies are managed with Go modules and download automatically on first build.
STL output uses [`github.com/hschendel/stl`](https://github.com/hschendel/stl), pinned to
v1.0.4 (v1.0.5 needs a newer Go toolchain).

## Building

```powershell
go build -o GoBrot.exe .
```

Or install it straight from GitHub:

```powershell
go install github.com/nfitbh72/gobrot@latest
```

Or use `run.bat`, which builds each package, installs the binary to `%GOPATH%\bin` and starts it.

## Usage

Run the binary and type commands at the `>` prompt:

```text
Welcome to GOBROT

Add Commands / XML here:
> width 2048
> height 2048
> iterations 1000
> coords -0.75 -0.74 0.1 0.11
> cp cycle
> go
> 3d
> exit
```

`go` writes `out0.png`, `out1.png`, … (one per layer). `3d` writes `out20.stl`, `out21.stl`, …

Type `help` for the list of commands, or `help <command>` for details on one.

### Commands

| Command | Arguments | Description |
| --- | --- | --- |
| `fractal` | `mandel` \| `bsjulia` | Fractal type |
| `width` / `height` | `<int>` | Output size in pixels (default 1024×1024) |
| `iterations` | `<int>` | Maximum iterations (default 512) |
| `coords` | `<rmin> <rmax> <imin> <imax>` | View bounds (default -2 2 -2 2) |
| `big` | `<int>` | Big-float precision in bits (default 53) |
| `fname` | `<string>` | Base filename for PNG output (default `./out`) |
| `cp` | `gray` \| `gray2` \| `cycle` \| `gpt1` | Colour profile (default `gray2`) |
| `smooth` | `on` \| `off` | Fractional escape times instead of whole iterations |
| `supersample` | `<int>` | Render at n times the size and average down (1–16, default off) |
| `heightcurve` | `<float>` | Log curve `k` for 3D heights, 0 is linear (default 0) |
| `stretch` | `on` \| `off` | Stretch values across the full colour range |
| `shadow` | `on` \| `off` | Also generate a shadow fractal |
| `layers` | `<float> ...` | Layer break points between 0 and 1, e.g. `layers 0.25 0.5 0.75` |
| `step` | `<float>` | Quantise the height into steps |
| `go` | | Compute the fractal and save the PNG(s) |
| `img` | | Re-save the PNG(s) from existing data |
| `ascii` | `<width> <height>` | Print an ASCII preview |
| `canvas` | `[method]` | Apply a canvas shape and save the PNG(s) |
| `cstr` | `<float>` | Canvas strength relative to the fractal |
| `circle` | `<float>` | Trim the layers to a circle |
| `3d` | | Save each layer as an STL file |
| `depthscale` | `<float>` | Depth scale of the 3D output (default 0.1) |
| `cutbelow` / `cutabove` | `<float>` | Cut 3D geometry outside this 0–1 range |
| `test` | | Write a test STL file |
| `config` | | Show the current configuration |
| `usage` | | Report memory usage |
| `help` | `[command]` | Show help |
| `exit` | | Quit |

### Importing XML

Paste settings exported from the [HWS Mandelbrot viewer](https://math.hws.edu/eck/js/mandelbrot/MB.html)
straight into the prompt. Input starting with `<?xml version='1.0'?>` and ending with
`</mandelbrot_settings_2>` is read as one block; GoBrot takes the bounds and max iterations,
then runs `go` and `3d` automatically. Examples are in [`samples/`](samples/).

## Project layout

```text
main.go, commands.go, appdata.go   entry point, CLI commands, default config
pkg/cli             interactive command parser
pkg/fractals        fractal algorithms
pkg/bigfloat        big.Float helpers
pkg/layers          depth layer splitting
pkg/images          PNG and ASCII output
pkg/colourprofiles  colour profiles
pkg/canvas          canvas projections
pkg/shapes          circle trimming
pkg/threed          STL generation
pkg/usage           memory reporting
pkg/threads         goroutine pool (work in progress)
samples/            example XML settings
```
