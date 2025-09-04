#import "@preview/charged-ieee:0.1.4": ieee

#show: ieee.with(
  title: [Route Planning Algorithm Visualization System],
  abstract: [
    The process of scientific writing is often tangled up with the intricacies of typesetting, leading to frustration and wasted time for researchers. In this paper, we introduce Typst, a new typesetting system designed specifically for scientific writing. Typst untangles the typesetting process, allowing researchers to compose papers faster. In a series of experiments we demonstrate that Typst offers several advantages, including faster document creation, simplified syntax, and increased ease-of-use.
  ],
  authors: (
    (
      name: "Martin Haug",
      department: [Co-Founder],
      organization: [Typst GmbH],
      location: [Berlin, Germany],
      email: "haug@typst.app"
    ),
    (
      name: "Laurenz Mädje",
      department: [Co-Founder],
      organization: [Typst GmbH],
      location: [Berlin, Germany],
      email: "maedje@typst.app"
    ),
  ),
  index-terms: ("Scientific writing", "Typesetting", "Document creation", "Syntax"),
  bibliography: bibliography("refs.bib"),
  figure-supplement: [Fig.],
)

= Introduction
This application provides an interactive visualization platform for comparing pathfinding algorithms on real road networks. The system implements Dijkstra's algorithm, Contraction Hierarchies (CH), and Customizable Contraction Hierarchies (CCH) with a web-based interface for performance analysis.

= Implementation

== Backend Architecture
The backend implements several algorithmic optimizations for efficient preprocessing and querying. The Contraction Hierarchies implementation employs parallel batch processing during vertex contraction, where independent vertices are identified and contracted simultaneously to reduce preprocessing time. The priority function extends beyond simple edge difference calculations:

$
text("priority")(v) = text("shortcuts")(v) - text("degree")(v) + (text("shortcuts")(v))/(text("degree")(v) + 1),
+ 0.5 dot text("contractedneighbors")(v)
$

This formulation effectively prioritizes low-degree vertices, particularly degree-1 nodes which contribute minimal shortcuts while reducing graph complexity.

Customizable Contraction Hierarchies utilize nested dissection ordering through recursive separator decomposition implemented via KaHIP integration. The preprocessing pipeline recursively partitions graphs into balanced components with minimal separators, generating vertex orderings optimized for contraction efficiency.

Witness search optimization employs bidirectional Dijkstra terminating early when potential witness paths exceed shortcut costs, significantly reducing preprocessing overhead compared to full shortest path computations.

== Frontend Visualization
The frontend leverages WebGL-based rendering through deck.gl with Vue.js integration for interactive map visualization. Three-dimensional arc rendering distinguishes shortcut edges from original road segments, providing visual differentiation between preprocessed hierarchical structures and base network topology. The MapLibre GL foundation enables efficient geographic data handling with real-time layer control and performance metrics display.

== System Integration
A RESTful Go API provides communication between frontend and backend components, handling graph data serialization and query execution with measured performance metrics. The modular architecture maintains separation between algorithmic implementations and visualization layers.

Code quality is maintained through comprehensive unit testing and clean architectural patterns. Documentation is generated using pkgsite and accessible via pkgsite open . for complete API reference.

= Experimental Results
Performance evaluation used OpenStreetMap datasets ranging from small urban networks to larger regional graphs. Contraction Hierarchies achieved substantial search space reduction, exploring 91-99\% fewer nodes compared to Dijkstra's algorithm across all test cases. However, query time improvements remained modest, ranging from 0.62x to 1.09x speedup.

For smaller graphs (osm1-3), CH showed slower execution times due to bidirectional search overhead. Medium to larger graphs (osm8-10) demonstrated marginal speedup improvements correlating with graph size. The results indicate that while CH correctly reduces algorithmic complexity, the overhead costs dominate performance gains at these graph scales.

= Conclusion
The visualization system successfully demonstrates the theoretical advantages of hierarchical pathfinding methods while revealing their practical limitations on moderate-scale road networks. The implementation provides an effective platform for algorithm education and research, confirming that Contraction Hierarchies require larger graph sizes to achieve their documented performance advantages.

// Keeping the rest of the original content from main.typ
= Methods <sec:methods>
#lorem(45)

$ a + b = gamma $ <eq:gamma>

#lorem(80)

#figure(
  placement: none,
  circle(radius: 15pt),
  caption: [A circle representing the Sun.]
) <fig:sun>

In @fig:sun you can see a common representation of the Sun, which is a star that is located at the center of the solar system.

#lorem(120)

#figure(
  caption: [The Planets of the Solar System and Their Average Distance from the Sun],
  placement: top,
  table(
    // Table styling is not mandated by the IEEE. Feel free to adjust these
    // settings and potentially move them into a set rule.
    columns: (6em, auto),
    align: (left, right),
    inset: (x: 8pt, y: 4pt),
    stroke: (x, y) => if y <= 1 { (top: 0.5pt) },
    fill: (x, y) => if y > 0 and calc.rem(y, 2) == 0  { rgb("#efefef") },

    table.header[Planet][Distance (million km)],
    [Mercury], [57.9],
    [Venus], [108.2],
    [Earth], [149.6],
    [Mars], [227.9],
    [Jupiter], [778.6],
    [Saturn], [1,433.5],
    [Uranus], [2,872.5],
    [Neptune], [4,495.1],
  )
) <tab:planets>

In @tab:planets, you see the planets of the solar system and their average distance from the Sun.
The distances were calculated with @eq:gamma that we presented in @sec:methods.

#lorem(240)

#lorem(240)
