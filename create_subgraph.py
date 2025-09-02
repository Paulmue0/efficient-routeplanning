import sys

def create_subgraph(original_metis, nodes_file, output_metis):
    with open(nodes_file, 'r') as f:
        subgraph_nodes = {int(line.strip()) for line in f}

    node_map = {old_id: new_id for new_id, old_id in enumerate(sorted(list(subgraph_nodes)), 1)}

    new_edges = 0
    new_lines = []

    with open(original_metis, 'r') as f:
        header = f.readline()
        num_nodes, _ = map(int, header.strip().split())

        for i, line in enumerate(f, 1):
            if i in subgraph_nodes:
                new_adj = []
                adj = line.strip().split()
                for neighbor in adj:
                    neighbor_id = int(neighbor)
                    if neighbor_id in subgraph_nodes:
                        new_adj.append(str(node_map[neighbor_id]))
                        new_edges += 1
                new_lines.append(" ".join(new_adj))

    with open(output_metis, 'w') as f:
        f.write(f"{len(subgraph_nodes)} {new_edges // 2}\n")
        for line in new_lines:
            f.write(f"{line}\n")

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: python create_subgraph.py <original_metis> <nodes_file> <output_metis>")
        sys.exit(1)

    create_subgraph(sys.argv[1], sys.argv[2], sys.argv[3])
