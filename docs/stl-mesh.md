# STL mesh generation

How `pkg/threed` turns a layer into an STL solid, and why it is built this way. Rebuilt for #43.

## What the output is

Each layer is a grid of heights. The solid has three parts:

- a **roof** that follows the heights,
- a flat **base** at z = 0,
- vertical **walls** wherever material meets empty space or the edge of the layer.

A grid point with a height of 0 is empty. **Empty areas are left out of the solid**: there is no
floor under them, and they get walls around them like any other edge. (The question was raised
on #38, #41 first built it this way, and the developer kept it for this rebuild when asking for
#43 to be executed. With `InverseDepth`, which is the default, a layer value of 1.0 is height 0;
that is also what `cutbelow` and `cutabove` write.)

## What "valid" means

The output is meant to load into a slicer with nothing to repair:

- no triangle with zero area,
- every edge is used by exactly two triangles, running in opposite directions,
- the triangles around every point form a single fan (no two bodies touching at a point),
- all triangles face outwards (positive volume), with no NaN or infinite coordinates.

`meshProblems` in `pkg/threed/threed_test.go` checks all of these on the triangles as they are
written to the file. Every test mesh goes through it.

## How it is built

Three choices carry the design. The code is in `pkg/threed/mesh.go`.

**1. One position per grid point.** A vertex is a grid point plus "roof" or "base". Its
coordinates are computed from that alone, so two triangles that share a point share it bit for
bit. Nothing has to be matched up or welded afterwards.

**2. Triangles that halve.** The grid is covered by right-angled triangles. A triangle is split by
a line from the middle of its long side to the opposite corner, giving two triangles of the same
shape, and so on down to half a grid cell. A big triangle is used wherever a split would not
change anything that matters. (This is the scheme known as a right-triangulated irregular
network, RTIN.) The grid is padded up to a power of two cells per side; the padding is empty.

Each possible split has a cost: how far the surface could move if the split were left out. A split
is kept if its cost is over the tolerance. Costs are added up from the smallest triangles to the
largest, which does two things:

- the cost is a true upper bound, so **no point of the roof moves by more than the tolerance**;
- a split always costs at least as much as the finer splits that depend on it, so the kept
  triangles always meet edge to edge. No triangle ends up with a neighbour's corner part-way
  along its side, which is what causes open edges.

A triangle that is partly material and partly empty has an infinite cost, so the outline of the
material is always kept exactly as the layer has it. Simplifying never adds or removes material
in plan view.

**3. The base follows the walls, not the roof.** The base is flat, so it needs no splits for
shape. It only has to match the bottom of the walls. It is built the same way as the roof, with an
infinite cost on the points where the roof put a wall corner and zero everywhere else. That makes
the base far coarser than the roof while still meeting every wall exactly.

One more rule: where two areas of material touch at a single point, that point is emptied first.
Otherwise four walls would share one vertical edge. This removes at most the handful of
half-cells around such a point.

## Tolerance

`tolerance <float>` is in output units (the model is `Scale` = 100 wide, so think millimetres).
The default is 0.02. `tolerance 0` keeps every height exactly and still merges areas that are
exactly flat or exactly planar.

The tolerance only affects the roof's heights. The outline of the material, the base and the
walls' positions are exact at any setting.

The command refuses negative values, `inf` and `nan`. The mesher itself is safe with any value: it
treats a negative or not-a-number tolerance as 0, and always keeps the outline.

## Measured

Same renders on `main` before the rebuild and after, all passing the checks above. Sizes in MB.

| Render | Before | `tolerance 0` | default 0.02 |
| --- | --- | --- | --- |
| 1024×1024, `coords -0.75 -0.74 0.1 0.11`, 1000 iterations | 137.6 | 53.2 | 41.4 |
| 2048×2048, default view, 512 iterations | 579.2 | 26.7 | 17.1 |
| 1500×1000, same zoom as the first, `step 0.05` | 196.2 | 54.0 | 53.9 |

Peak memory for the 2048×2048 render went from 2.05 GB to 0.21 GB.

What limits the saving is the outline. The first render's material has about 65,000 grid-length
edges of outline, around the fractal's interior and its filaments. Each needs two wall triangles
and forces small triangles next to it in the roof and base. That floor is about 17 MB for that
render at any tolerance. Renders whose empty areas have a simple outline shrink far more.

## Not done

- **Other file formats.** STL repeats every vertex in every triangle (50 bytes per triangle). A
  format with shared vertices (3MF, OBJ, binary PLY) would cut the same mesh to roughly a third.
  The mesher already holds triangles as vertex references, so this is a writer, not a redesign.
- **Simplifying the outline.** Straightening the outline within a tolerance would lift the floor
  described above, at the cost of changing where material is.
- **Long, thin layers.** The grid is a square as wide as the layer's longer side, rounded up to a
  power of two, so memory and time follow that square and not the layer. Square and near-square
  layers use far less memory than the old generator; long thin ones use more (a 8200×125 layer
  took 3.2 GB against 1.3 GB). A layer over 16,385 points on its longer side is refused with a
  message. The fix is to cover the layer with several square blocks, side by side. Worth doing
  when a real render is past about 4:1, or over 4,097 points a side.
- **Non-square layers are stretched** to `Scale` × `Scale`, as before. Not changed here.
