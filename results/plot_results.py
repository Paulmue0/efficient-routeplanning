import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt
import numpy as np
import os

# Set style for better looking plots
plt.style.use('seaborn-v0_8')
sns.set_palette("husl")

# Define the directory where the CSVs are located and where plots will be saved
results_dir = os.path.join(os.getcwd(), 'results')
output_dir = results_dir  # Save plots in results directory

# Load all datasets
df_cch_customization = pd.read_csv(os.path.join(results_dir, 'cch_customization_experiment_results.csv'))
df_cch_preprocess = pd.read_csv(os.path.join(results_dir, 'cch_preprocess_experiment_results.csv'))
df_cch_query = pd.read_csv(os.path.join(results_dir, 'cch_query_experiment_results.csv'))
df_ch_query = pd.read_csv(os.path.join(results_dir, 'query_experiment_results.csv'))

# Extract graph numbers for better plotting
def extract_graph_num(graph_name):
    if 'example' in graph_name:
        return 0
    return int(graph_name.replace('osm', '').replace('.txt', ''))

for df in [df_cch_customization, df_cch_preprocess, df_cch_query, df_ch_query]:
    df['GraphNum'] = df['Graph'].apply(extract_graph_num)
    df = df.sort_values('GraphNum')

# --- CCH Customization Experiment Results ---

fig, ax = plt.subplots(figsize=(14, 8))
x_pos = range(len(df_cch_customization))
ax.plot(x_pos, df_cch_customization['OriginalCustomizationTime(ms)'], 'o-', linewidth=2.5, markersize=8, label='Original Customization')
ax.plot(x_pos, df_cch_customization['AvgRandomCustomizationTime(ms)'], 's-', linewidth=2.5, markersize=8, label='Random Customization')
ax.set_yscale('log')
ax.set_xlabel('Graph Instance', fontsize=12, fontweight='bold')
ax.set_ylabel('Customization Time (ms)', fontsize=12, fontweight='bold')
ax.set_title('CCH Customization Performance', fontsize=16, fontweight='bold', pad=20)
ax.set_xticks(x_pos)
ax.set_xticklabels(df_cch_customization['Graph'], rotation=45, ha='right')
ax.legend(fontsize=11, frameon=True, fancybox=True, shadow=True)
ax.grid(True, alpha=0.3)
plt.tight_layout()
plt.savefig(os.path.join(output_dir, 'cch_customization_time.pdf'), bbox_inches='tight')
plt.close()

# --- CCH Preprocessing Experiment Results ---
df_cch_preprocess = pd.read_csv(os.path.join(output_dir, 'cch_preprocess_experiment_results.csv'))

# Preprocessing Time
plt.figure(figsize=(12, 7))
sns.lineplot(data=df_cch_preprocess, x='Graph', y='PreprocessingTime(ms)', marker='o')
plt.title('CCH Preprocessing Time')
plt.xlabel('Graph')
plt.ylabel('Time (ms)')
plt.yscale('log')
plt.xticks(rotation=45, ha='right')
plt.grid(True, which="both", ls="--", c='0.7')
plt.tight_layout()
plt.savefig(os.path.join(output_dir, 'cch_preprocess_time.pdf'))
plt.close()

# Shortcuts Added
plt.figure(figsize=(12, 7))
sns.lineplot(data=df_cch_preprocess, x='Graph', y='ShortcutsAdded', marker='o')
plt.title('CCH Shortcuts Added')
plt.xlabel('Graph')
plt.ylabel('Shortcuts Added')
plt.yscale('log')
plt.xticks(rotation=45, ha='right')
plt.grid(True, which="both", ls="--", c='0.7')
plt.tight_layout()
plt.savefig(os.path.join(output_dir, 'cch_shortcuts_added.pdf'))
plt.close()

# Avg and Max Triangles
plt.figure(figsize=(12, 7))
sns.lineplot(data=df_cch_preprocess, x='Graph', y='AvgTriangles', marker='o', label='Avg Triangles')
sns.lineplot(data=df_cch_preprocess, x='Graph', y='MaxTriangles', marker='o', label='Max Triangles')
plt.title('CCH Triangle Statistics')
plt.xlabel('Graph')
plt.ylabel('Count')
plt.xticks(rotation=45, ha='right')
plt.legend()
plt.grid(True, which="both", ls="--", c='0.7')
plt.tight_layout()
plt.savefig(os.path.join(output_dir, 'cch_triangle_statistics.pdf'))
plt.close()

# --- CCH Query Experiment Results ---
df_cch_query = pd.read_csv(os.path.join(output_dir, 'cch_query_experiment_results.csv'))

plt.figure(figsize=(12, 7))
sns.lineplot(data=df_cch_query, x='Graph', y='AvgDijkstraTime(ms)', marker='o', label='Avg Dijkstra Time')
sns.lineplot(data=df_cch_query, x='Graph', y='AvgCCHQueryTime(ms)', marker='o', label='Avg CCH Query Time')
plt.title('CCH Query Time vs Dijkstra')
plt.xlabel('Graph')
plt.ylabel('Time (ms)')
plt.yscale('log')
plt.xticks(rotation=45, ha='right')
plt.legend()
plt.grid(True, which="both", ls="--", c='0.7')
plt.tight_layout()
plt.savefig(os.path.join(output_dir, 'cch_query_time.pdf'))
plt.close()

# --- CH Query Experiment Results ---
df_ch_query = pd.read_csv(os.path.join(output_dir, 'query_experiment_results.csv'))

# CH Query Time
plt.figure(figsize=(12, 7))
sns.lineplot(data=df_ch_query, x='Graph', y='AvgDijkstraTime(ms)', marker='o', label='Avg Dijkstra Time')
sns.lineplot(data=df_ch_query, x='Graph', y='AvgCHDijkstraTime(ms)', marker='o', label='Avg CH Dijkstra Time')
plt.title('CH Query Time vs Dijkstra')
plt.xlabel('Graph')
plt.ylabel('Time (ms)')
plt.yscale('log')
plt.xticks(rotation=45, ha='right')
plt.legend()
plt.grid(True, which="both", ls="--", c='0.7')
plt.tight_layout()
plt.savefig(os.path.join(output_dir, 'ch_query_time.pdf'))
plt.close()

# CH Nodes Popped
plt.figure(figsize=(12, 7))
sns.lineplot(data=df_ch_query, x='Graph', y='AvgDijkstraNodesPopped', marker='o', label='Avg Dijkstra Nodes Popped')
sns.lineplot(data=df_ch_query, x='Graph', y='AvgCHNodesPopped', marker='o', label='Avg CH Nodes Popped')
plt.title('CH Nodes Popped vs Dijkstra')
plt.xlabel('Graph')
plt.ylabel('Nodes Popped')
plt.yscale('log')
plt.xticks(rotation=45, ha='right')
plt.legend()
plt.grid(True, which="both", ls="--", c='0.7')
plt.tight_layout()
plt.savefig(os.path.join(output_dir, 'ch_nodes_popped.pdf'))
plt.close()

# --- Combined Query Time Plot (Dijkstra vs CCH vs CH) ---
# Merge the two query dataframes
# Rename columns to avoid conflicts and for clarity in the combined plot
df_cch_query_renamed = df_cch_query[['Graph', 'AvgDijkstraTime(ms)', 'AvgCCHQueryTime(ms)']].copy()
df_cch_query_renamed.rename(columns={'AvgDijkstraTime(ms)': 'Dijkstra Time (ms)', 'AvgCCHQueryTime(ms)': 'CCH Query Time (ms)'}, inplace=True)

df_ch_query_renamed = df_ch_query[['Graph', 'AvgCHDijkstraTime(ms)']].copy()
df_ch_query_renamed.rename(columns={'AvgCHDijkstraTime(ms)': 'CH Query Time (ms)'}, inplace=True)

# Merge on Graph column
# Use outer merge to keep all graphs from both datasets
combined_query_df = pd.merge(df_cch_query_renamed, df_ch_query_renamed, on='Graph', how='outer')

# Melt the DataFrame for easier plotting with seaborn
combined_query_melted = combined_query_df.melt(id_vars=['Graph'], var_name='Algorithm', value_name='Time (ms)')

# Ensure correct order of Graph instances for plotting
# Extract graph numbers and sort to get the desired order
combined_query_melted['GraphNum'] = combined_query_melted['Graph'].apply(extract_graph_num)
# Get unique graph names in the desired numerical order
ordered_graphs = combined_query_melted.sort_values('GraphNum')['Graph'].unique()
# Convert 'Graph' column to a categorical type with the specified order
combined_query_melted['Graph'] = pd.Categorical(combined_query_melted['Graph'], categories=ordered_graphs, ordered=True)

plt.figure(figsize=(14, 8))
sns.lineplot(data=combined_query_melted, x='Graph', y='Time (ms)', hue='Algorithm', marker='o')
plt.title('Query Time Comparison: Dijkstra vs CCH vs CH')
plt.xlabel('Graph')
plt.ylabel('Time (ms)')
plt.yscale('log')
plt.xticks(rotation=45, ha='right')
plt.legend(title='Algorithm')
plt.grid(True, which="both", ls="--", c='0.7')
plt.tight_layout()
plt.savefig(os.path.join(output_dir, 'combined_query_time.pdf'))
plt.close()

print(f"Plots saved to {output_dir}")