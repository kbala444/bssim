import networkx as nx
from matplotlib import pyplot as plt

class BWGrapher():
    def __init__(self, filepath='../bwinfo'):
        # tuple of (source, dest, bw)
        self.conns = []
        with open(filepath, 'r') as f:
            for line in f.readlines():
                line = line.split('->')
                line[1:] = line[1].split()
                self.conns.append((int(line[0]), int(line[1]), float(line[2])))

    # abstract this for use with latgen?
    def graph_all_nodes(self, outfile=None):
        g = nx.DiGraph()

        for con in self.conns:
            g.add_edge(con[0], con[1], weight=con[2])

        print g.edges(data=True)
        edge_labels = dict([((u, v,), d['weight']) for u,v,d in g.edges(data=True)])
        print edge_labels

        pos = nx.circular_layout(g)
        nx.draw(g, pos, node_size=1000, alpha=.5)
        nx.draw_networkx_edge_labels(g, pos, edge_labels=edge_labels, font_size=14)
        nx.draw_networkx_labels(g, pos, font_size=12, label_pos=0)
        plt.draw()

        if outfile:
            plt.savefig(outfile)

        plt.show()

    def graph_node(self, node=0, outfile=None):
        g = nx.Graph()

        for con in self.conns:
            if con[0] == node:
                g.add_edge(con[0], con[1], weight=con[2])

        edge_labels = {}
        for u, v, data in g.edges(data=True):
            edge_labels[(u, v)] = data['weight']

        pos = nx.circular_layout(g)
        nx.draw(g, pos, node_size=1000, alpha=.5)
        nx.draw_networkx_edge_labels(g, pos, edge_labels=edge_labels, font_size=12)
        nx.draw_networkx_labels(g, pos, font_size=12, label_pos=0)

        plt.draw()

        if outfile:
            plt.savefig(outfile)

        plt.show()

if __name__ == "__main__":
    bwg = BWGrapher()
    while True:
        n = input('which node do you want to see upload data for?\n')
        bwg.graph_node(n)
