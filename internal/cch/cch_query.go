package cch

// function Query(s, t):
//     dist, parent = DijkstraOrEliminationTree(s, t, ℓ⁺)
//     # may use outdated shortcut costs
//     path = ReconstructPath(s, t, parent)
//     return Unpack(path)
//
// function Unpack(path):  # path is sequence of edges in E⁺
//     fullPath = []
//     for edge (x,y) in path:
//         if (x,y) ∈ E:             # original edge
//             fullPath.append((x,y), ℓ(x,y))  # recompute with base cost
//         else:                      # shortcut
//             (x,u,y) = getTriangle(edge)     # stored during customization
//             subpath = [(x,u), (u,y)]
//             fullPath += Unpack(subpath)     # recurse
//     return fullPath
